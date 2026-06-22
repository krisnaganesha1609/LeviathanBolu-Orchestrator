package handlers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/internal/assistant"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/internal/llm"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/session"
)

// wsIncoming is the shape the Flutter client sends over the WebSocket.
type wsIncoming struct {
	// Action is an optional command ("clear_history").
	// If empty, Message is treated as a chat input.
	Action  string `json:"action,omitempty"`
	Message string `json:"message,omitempty"`
}

// WSHandler wraps AssistantService for WebSocket connections.
type WSHandler struct {
	Service      *assistant.AssistantService
	SessionStore *session.Store
	SystemPrompt string
}

// HandleWS is the WebSocket connection handler registered with Fiber.
//
// One goroutine per active connection. The connection stays open until
// the client disconnects. History is loaded from Redis at connect-time
// and saved back after every completed turn.
//
// Concurrency: all WebSocket writes happen on this goroutine (the channel
// loop below). Stream() runs in its own goroutine and writes to the
// events channel. Single writer → no mutex needed.
func (h *WSHandler) HandleWS(conn *websocket.Conn) {
	deviceID := conn.Params("device_id")
	if deviceID == "" {
		_ = conn.WriteJSON(assistant.Event{
			Event:   assistant.EventError,
			Message: "device_id is required in the URL path",
		})
		return
	}

	log.Printf("[ws] device %q connected", deviceID)
	defer log.Printf("[ws] device %q disconnected", deviceID)

	// Per-connection context. Cancelling it stops any in-progress
	// Stream() goroutine when the client disconnects.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load existing history for this device (nil if fresh session).
	history, err := h.SessionStore.GetHistory(ctx, deviceID)
	if err != nil {
		log.Printf("[ws] device %q: failed to load history: %v", deviceID, err)
	}

	// Set a generous read deadline so idle connections don't hold
	// Redis/goroutine resources forever. Clients should send a ping
	// (empty message) every ~30s if they want to keep the session alive.
	conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			// Client disconnected or read timed out — cancel any in-progress
			// Stream() and exit cleanly.
			cancel()
			return
		}
		// Reset deadline on every received message.
		conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

		var incoming wsIncoming
		if err := json.Unmarshal(raw, &incoming); err != nil {
			_ = conn.WriteJSON(assistant.Event{
				Event:   assistant.EventError,
				Message: "invalid JSON — expected {\"message\": \"...\"}",
			})
			continue
		}

		// ── Special actions ───────────────────────────────────────────────
		switch incoming.Action {
		case "clear_history":
			if err := h.SessionStore.ClearHistory(ctx, deviceID); err != nil {
				log.Printf("[ws] device %q: clear history failed: %v", deviceID, err)
			}
			history = nil
			_ = conn.WriteJSON(assistant.Event{Event: assistant.EventHistoryCleared})
			continue
		case "ping":
			// No-op keepalive — just resets the read deadline (done above).
			continue
		}

		if incoming.Message == "" {
			_ = conn.WriteJSON(assistant.Event{
				Event:   assistant.EventError,
				Message: "message field is required",
			})
			continue
		}

		// ── Streaming turn ────────────────────────────────────────────────
		// Stream() runs in a goroutine and sends events to this channel.
		// This goroutine reads from the channel and writes to the WebSocket.
		// Buffer of 10 lets Stream() stay a few events ahead without blocking.
		events := make(chan assistant.Event, 10)

		go h.Service.Stream(ctx, assistant.ChatRequest{
			Message:      incoming.Message,
			History:      history,
			SystemPrompt: h.SystemPrompt,
		}, events)

		var updatedHistory []llm.Message

		for e := range events {
			// Grab the updated history before serializing (json:"-" hides it).
			if e.Event == assistant.EventDone && e.UpdatedHistory != nil {
				updatedHistory = e.UpdatedHistory
			}
			if err := conn.WriteJSON(e); err != nil {
				// Client disconnected mid-stream.
				cancel()
				return
			}
		}

		// Persist history after the turn completes.
		if updatedHistory != nil {
			history = updatedHistory
			if err := h.SessionStore.SaveHistory(ctx, deviceID, history); err != nil {
				log.Printf("[ws] device %q: save history failed: %v", deviceID, err)
			}
		}
	}
}
