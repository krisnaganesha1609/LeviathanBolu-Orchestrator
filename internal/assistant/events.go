package assistant

import "github.com/krisnaganesha1609/LeviathanBolu-BE/internal/llm"

type EventType string

const (
	EventThinking       EventType = "assistant_thinking"
	EventToolCalled     EventType = "tool_called"
	EventToolResult     EventType = "tool_result"
	EventReply          EventType = "assistant_reply"
	EventDone           EventType = "done"
	EventError          EventType = "error"
	EventHistoryCleared EventType = "history_cleared"

	// ── Protokol audio ──────────────────────────────────────────────────
	EventWakeResult        EventType = "wake_result"
	EventPartialTranscript EventType = "partial_transcript"
	EventFinalTranscript   EventType = "final_transcript"

	// Penamaan pakai titik sesuai spek TTS streaming kamu.
	EventTTSStart EventType = "tts.start"
	EventTTSChunk EventType = "tts.chunk"
	EventTTSEnd   EventType = "tts.end"
)

type Event struct {
	Event   EventType `json:"event"`
	Tool    string    `json:"tool,omitempty"`
	Message string    `json:"message,omitempty"`
	Data    any       `json:"data,omitempty"`

	Text        string `json:"text,omitempty"`        // partial/final_transcript
	Matched     bool   `json:"matched,omitempty"`     // wake_result
	Personality string `json:"personality,omitempty"` // wake_result

	// tts.chunk
	Seq   int    `json:"seq,omitempty"`
	Audio string `json:"audio,omitempty"` // base64 Opus packet

	UpdatedHistory []llm.Message `json:"-"`
}
