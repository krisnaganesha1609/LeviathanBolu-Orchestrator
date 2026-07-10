package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/google/uuid"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/internal/assistant"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/internal/llm"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/services"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/session"
)

// wsIncoming is the shape the Flutter client sends over the WebSocket
// (JSON/text frames). Binary frames are handled separately — see below.
type wsIncoming struct {
	Action      string `json:"action,omitempty"` // "clear_history" | "ping" | "audio_start" | "audio_end"
	Message     string `json:"message,omitempty"`
	Personality string `json:"personality,omitempty"`
	Purpose     string `json:"purpose,omitempty"` // utk audio_start: "wake_check" | "conversation"
}

// WSHandler wraps AssistantService for WebSocket connections.
type WSHandler struct {
	Service              *assistant.AssistantService
	SessionStore         *session.Store
	SystemPrompt         string
	DeviceSessionService services.DeviceSessionService

	// WakeChecker & STT bersifat OPSIONAL — kalau nil, fitur suara otomatis
	// nonaktif (server hanya melayani chat teks, sama seperti sebelumnya)
	// dan client yang coba kirim audio_start akan dapat error yang jelas,
	// bukan silently gagal.
	WakeChecker *assistant.WakeWordChecker
	STT         assistant.SpeechToText
}

func (h *WSHandler) HandleWS(conn *websocket.Conn) {
	deviceIDStr := conn.Params("device_id")
	if deviceIDStr == "" {
		_ = conn.WriteJSON(assistant.Event{Event: assistant.EventError, Message: "device_id is required in the URL path"})
		return
	}

	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		_ = conn.WriteJSON(assistant.Event{Event: assistant.EventError, Message: "device_id must be a valid UUID"})
		return
	}

	log.Printf("[ws] device %q connected", deviceIDStr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ── Buat sesi baru saat koneksi WS dibuka ──────────────────────────────
	sessionIDStr, err := h.DeviceSessionService.CreateSession(ctx, deviceID)
	if err != nil {
		log.Printf("[ws] device %q: gagal membuat session: %v", deviceIDStr, err)
		_ = conn.WriteJSON(assistant.Event{Event: assistant.EventError, Message: "failed to create device session"})
		return
	}

	conversationID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		log.Printf("[ws] device %q: session ID bukan UUID valid: %v", deviceIDStr, err)
		_ = conn.WriteJSON(assistant.Event{Event: assistant.EventError, Message: "internal session ID error"})
		return
	}

	log.Printf("[ws] device %q session %s dibuat", deviceIDStr, conversationID)

	// ── Tutup sesi saat koneksi WS ditutup ────────────────────────────────
	// disconnectReason diisi tepat sebelum return agar defer bisa membacanya.
	disconnectReason := "unknown"
	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer closeCancel()

		session, sErr := h.DeviceSessionService.FindByID(closeCtx, conversationID)
		if sErr == nil {
			session.DisconnectReason = truncate(disconnectReason, 200)
			_ = h.DeviceSessionService.CloseSession(closeCtx, conversationID)
		}

		log.Printf("[ws] device %q session %s ditutup (reason: %s)", deviceIDStr, conversationID, disconnectReason)
	}()

	history, err := h.SessionStore.GetHistory(ctx, deviceIDStr)
	if err != nil {
		log.Printf("[ws] device %q: failed to load history: %v", deviceIDStr, err)
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

	activePersonality := assistant.PersonalityBolu

	// State sesi audio saat ini. Aman tanpa mutex — semua diakses dari SATU
	// goroutine (loop ini), persis pola single-writer yang sudah dipakai
	// file ini untuk WS write.
	var (
		audioPurpose string // "" | "wake_check" | "conversation"
		audioBuffer  [][]byte
	)

	for {
		msgType, raw, err := conn.ReadMessage()
		if err != nil {
			disconnectReason = err.Error()
			cancel()
			return
		}
		conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

		// ── Binary frame = Opus packet mic dari Flutter ─────────────────
		if msgType == websocket.BinaryMessage {
			if audioPurpose == "" {
				// Frame nyasar di luar sesi audio_start/audio_end — dulu ini
				// bikin server nembak balik "invalid JSON" ke Flutter untuk
				// SETIAP frame. Sekarang cukup diabaikan & dicatat.
				log.Printf("[ws] device %q: binary frame diterima tanpa audio_start aktif, diabaikan", deviceIDStr)
				continue
			}
			audioBuffer = append(audioBuffer, append([]byte(nil), raw...))
			continue
		}

		// ── Text frame = JSON control ─────────────────────────────────────
		var incoming wsIncoming
		if err := json.Unmarshal(raw, &incoming); err != nil {
			_ = conn.WriteJSON(assistant.Event{
				Event:   assistant.EventError,
				Message: `invalid JSON — expected {"message": "..."} or an audio control action`,
			})
			continue
		}

		switch incoming.Action {
		case "clear_history":
			if err := h.SessionStore.ClearHistory(ctx, deviceIDStr); err != nil {
				log.Printf("[ws] device %q: clear history failed: %v", deviceIDStr, err)
			}
			history = nil
			_ = conn.WriteJSON(assistant.Event{Event: assistant.EventHistoryCleared})
			continue

		case "ping":
			// UpdatePing tidak blocking; error hanya dicatat.
			go func() {
				pingCtx, pingCancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer pingCancel()
				if pErr := h.DeviceSessionService.UpdatePing(pingCtx, conversationID); pErr != nil {
					log.Printf("[ws] device %q: UpdatePing failed: %v", deviceIDStr, pErr)
				}
			}()
			continue

		case "audio_start":
			audioPurpose = incoming.Purpose
			audioBuffer = nil
			continue

		case "audio_end":
			purpose := audioPurpose
			buffered := audioBuffer
			audioPurpose = ""
			audioBuffer = nil

			var writeErr error
			switch purpose {
			case "wake_check":
				h.handleWakeCheck(ctx, conn, deviceIDStr, conversationID, buffered)
			case "conversation":
				writeErr = h.handleVoiceTurn(ctx, conn, deviceIDStr, conversationID, buffered, activePersonality, &history)
			default:
				log.Printf("[ws] device %q: audio_end tanpa audio_start yang valid (purpose=%q)", deviceIDStr, purpose)
			}
			if writeErr != nil {
				disconnectReason = fmt.Sprintf("write error after voice turn: %v", writeErr)
				cancel()
				return
			}
			continue
		}

		// ── Chat teks manual (keyboard) ────────────────────────────────────
		if incoming.Message == "" {
			_ = conn.WriteJSON(assistant.Event{Event: assistant.EventError, Message: "message field is required"})
			continue
		}

		personality := assistant.Personality(incoming.Personality)
		if personality == "" {
			personality = activePersonality
		}
		activePersonality = personality

		if err := h.runTurn(ctx, conn, deviceIDStr, conversationID, incoming.Message, personality, &history, false); err != nil {
			disconnectReason = fmt.Sprintf("write error after text turn: %v", err)
			cancel()
			return
		}
	}
}

// ── Wake word check (round-trip singkat, lihat assistant/voice.go) ────────

func (h *WSHandler) handleWakeCheck(ctx context.Context, conn *websocket.Conn, deviceID string, conversationID uuid.UUID, opusFrames [][]byte) {
	if h.WakeChecker == nil {
		_ = conn.WriteJSON(assistant.Event{Event: assistant.EventWakeResult, Matched: false})
		return
	}

	matched, personality, err := h.WakeChecker.Check(ctx, deviceID, opusFrames, conversationID)
	if err != nil {
		log.Printf("[ws] device %q: wake check error: %v", deviceID, err)
		_ = conn.WriteJSON(assistant.Event{Event: assistant.EventWakeResult, Matched: false})
		return
	}

	_ = conn.WriteJSON(assistant.Event{
		Event:       assistant.EventWakeResult,
		Matched:     matched,
		Personality: string(personality),
	})
}

// ── Turn yang berasal dari suara: transcribe dulu, baru jalanin LLM+TTS ──

func (h *WSHandler) handleVoiceTurn(
	ctx context.Context,
	conn *websocket.Conn,
	deviceID string,
	conversationID uuid.UUID,
	opusFrames [][]byte,
	personality assistant.Personality,
	history *[]llm.Message,
) error {
	if h.STT == nil {
		_ = conn.WriteJSON(assistant.Event{Event: assistant.EventError, Message: "voice belum dikonfigurasi di server"})
		return nil
	}
	if len(opusFrames) == 0 {
		return nil
	}

	text, err := h.STT.Transcribe(ctx, opusFrames, conversationID)
	if err != nil {
		log.Printf("[ws] device %q: transcribe error: %v", deviceID, err)
		_ = conn.WriteJSON(assistant.Event{Event: assistant.EventError, Message: "gagal transkrip audio"})
		return nil
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil // wake word terucap tanpa perintah lanjutan — diam saja
	}

	if err := conn.WriteJSON(assistant.Event{Event: assistant.EventFinalTranscript, Text: text}); err != nil {
		return err
	}

	return h.runTurn(ctx, conn, deviceID, conversationID, text, personality, history, true)
}

// ── Satu turn LLM lengkap (dipakai baik oleh chat teks maupun voice) ──────

func (h *WSHandler) runTurn(
	ctx context.Context,
	conn *websocket.Conn,
	deviceID string,
	conversationID uuid.UUID,
	message string,
	personality assistant.Personality,
	history *[]llm.Message,
	voiceMode bool,
) error {
	systemPrompt := assistant.BuildSystemPrompt("LeviathanBolu", "id", personality)

	events := make(chan assistant.Event, 10)
	go h.Service.Stream(ctx, assistant.ChatRequest{
		Message:        message,
		History:        *history,
		SystemPrompt:   systemPrompt,
		Personality:    string(personality),
		VoiceMode:      voiceMode,
		ConversationID: conversationID,
	}, events)

	var updatedHistory []llm.Message
	for e := range events {
		if e.Event == assistant.EventDone && e.UpdatedHistory != nil {
			updatedHistory = e.UpdatedHistory
		}
		if err := conn.WriteJSON(e); err != nil {
			return err // client disconnected mid-stream
		}
	}

	if updatedHistory != nil {
		*history = updatedHistory
		if err := h.SessionStore.SaveHistory(ctx, deviceID, *history); err != nil {
			log.Printf("[ws] device %q: save history failed: %v", deviceID, err)
		}
	}
	return nil
}

// truncate memotong string ke maxLen rune pertama untuk menghindari nilai
// DisconnectReason yang terlalu panjang tersimpan ke database.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}
