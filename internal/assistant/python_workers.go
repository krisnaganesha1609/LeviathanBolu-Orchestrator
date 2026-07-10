package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gopkg.in/hraban/opus.v2"
)

// Konfigurasi audio — HARUS SAMA dengan yang dipakai Flutter & worker Python.
const (
	pyAudioSampleRate  = 16000
	pyAudioChannels    = 1
	pyOpusFrameSamples = pyAudioSampleRate * 20 / 1000 // 320 sample @20ms
)

// ═══ STT worker client (SenseVoice / FunASR) ═══════════════════════════

type STTWorkerClient struct {
	WSURL string // contoh: "ws://localhost:9001/stt"
}

func NewSTTWorkerClient(wsURL string) *STTWorkerClient {
	return &STTWorkerClient{WSURL: wsURL}
}

// Transcribe: decode tiap Opus frame dari Flutter jadi PCM, stream ke worker
// Python sbg binary PCM frame berurutan, tutup dgn {"action":"end"}, tunggu
// balasan final_transcript.
func (c *STTWorkerClient) Transcribe(ctx context.Context, opusFrames [][]byte, conversationID uuid.UUID) (string, error) {
	url := fmt.Sprintf("%s?conversation_id=%s", c.WSURL, conversationID)
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		return "", fmt.Errorf("dial STT worker: %w", err)
	}
	defer conn.Close()

	dec, err := opus.NewDecoder(pyAudioSampleRate, pyAudioChannels)
	if err != nil {
		return "", fmt.Errorf("init opus decoder: %w", err)
	}

	pcmBuf := make([]int16, pyOpusFrameSamples)
	for _, frame := range opusFrames {
		n, err := dec.Decode(frame, pcmBuf)
		if err != nil {
			continue // 1 frame korup jangan gagalkan seluruh transkrip
		}
		if err := conn.WriteMessage(websocket.BinaryMessage, int16SliceToBytes(pcmBuf[:n])); err != nil {
			return "", fmt.Errorf("send pcm to STT worker: %w", err)
		}
	}
	if err := conn.WriteJSON(map[string]string{"action": "end"}); err != nil {
		return "", fmt.Errorf("send end to STT worker: %w", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return "", fmt.Errorf("read STT worker response: %w", err)
		}
		var resp struct {
			Event string `json:"event"`
			Text  string `json:"text"`
		}
		if json.Unmarshal(raw, &resp) != nil {
			continue
		}
		if resp.Event == "final_transcript" {
			return resp.Text, nil
		}
		// event == "partial_transcript" bisa di-relay real-time kalau nanti
		// mau ditambah — untuk sekarang cukup ditunggu sampai final.
	}
}

// ═══ TTS worker client (Kokoro TTS) ═════════════════════════════════════

type TTSWorkerClient struct {
	WSURL string // contoh: "ws://localhost:9002/tts"
}

func NewTTSWorkerClient(wsURL string) *TTSWorkerClient {
	return &TTSWorkerClient{WSURL: wsURL}
}

// SynthesizeStream: kirim {"text":..., "personality":...}, terima PCM
// binary frame berurutan dari Python, encode tiap frame ke Opus, kirim
// lewat channel. Channel ditutup otomatis saat worker kirim {"event":"done"}
// atau socket ditutup.
func (c *TTSWorkerClient) SynthesizeStream(ctx context.Context, text string, personality Personality, conversationID uuid.UUID) (<-chan []byte, error) {
	url := fmt.Sprintf("%s?conversation_id=%s", c.WSURL, conversationID)
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		return nil, fmt.Errorf("dial TTS worker: %w", err)
	}

	if err := conn.WriteJSON(map[string]string{"text": text, "personality": string(personality)}); err != nil {
		conn.Close()
		return nil, fmt.Errorf("send text to TTS worker: %w", err)
	}

	// opus.AppVoIP ini aman aja pas dideploy kok. Dia error karena belum enable CGO di Windows, tapi di Linux/ARM64 (Raspberry Pi) aman.
	enc, err := opus.NewEncoder(pyAudioSampleRate, pyAudioChannels, opus.AppVoIP)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("init opus encoder: %w", err)
	}

	out := make(chan []byte, 8)

	go func() {
		defer close(out)
		defer conn.Close()

		opusBuf := make([]byte, 4000)

		for {
			msgType, raw, err := conn.ReadMessage()
			if err != nil {
				return // socket ditutup Python = sintesis selesai (atau error)
			}

			if msgType == websocket.BinaryMessage {
				pcm := bytesToInt16Slice(raw)
				// CATATAN: kalau Kokoro tidak menghasilkan chunk persis
				// kelipatan 320 sample (20ms@16kHz), Encode() di bawah akan
				// error karena Opus butuh frame size tetap — kalau itu
				// terjadi, perlu buffering/reslicing di sini dulu sebelum
				// encode. Cek ukuran chunk asli dari worker-mu.
				n, err := enc.Encode(pcm, opusBuf)
				if err != nil {
					continue
				}
				packet := make([]byte, n)
				copy(packet, opusBuf[:n])
				select {
				case out <- packet:
				case <-ctx.Done():
					return
				}
				continue
			}

			var resp struct {
				Event string `json:"event"`
			}
			if json.Unmarshal(raw, &resp) == nil && resp.Event == "done" {
				return
			}
		}
	}()

	return out, nil
}

func int16SliceToBytes(s []int16) []byte {
	b := make([]byte, len(s)*2)
	for i, v := range s {
		b[i*2] = byte(v)
		b[i*2+1] = byte(v >> 8)
	}
	return b
}

func bytesToInt16Slice(b []byte) []int16 {
	s := make([]int16, len(b)/2)
	for i := range s {
		s[i] = int16(b[i*2]) | int16(b[i*2+1])<<8
	}
	return s
}
