package assistant

import "github.com/krisnaganesha1609/LeviathanBolu-BE/internal/llm"

// EventType is the "event" field the client receives over WebSocket.
// Flutter's Orb UI uses these to decide which animation state to render.
type EventType string

const (
	// EventThinking: LLM is being called (spinner / orb pulsing)
	EventThinking EventType = "assistant_thinking"

	// EventToolCalled: a tool is about to be executed
	// (orb → "scanning" state, show tool name in UI)
	EventToolCalled EventType = "tool_called"

	// EventToolResult: tool finished, result is available
	// (can be shown as a sub-card under the reply)
	EventToolResult EventType = "tool_result"

	// EventReply: final text from the model
	// (orb → "speaking" state, begin TTS if enabled)
	EventReply EventType = "assistant_reply"

	// EventDone: turn is fully complete
	// (orb → idle, persist history)
	EventDone EventType = "done"

	// EventError: something went wrong
	// (orb → error state, show message)
	EventError EventType = "error"

	// EventHistoryCleared: Redis history was reset for this device
	EventHistoryCleared EventType = "history_cleared"
)

// Event is one unit of real-time feedback streamed to the Flutter client.
// All exported fields except UpdatedHistory are serialized over WebSocket.
type Event struct {
	Event   EventType `json:"event"`
	Tool    string    `json:"tool,omitempty"`
	Message string    `json:"message,omitempty"`
	Data    any       `json:"data,omitempty"`

	// UpdatedHistory carries the full conversation after a turn completes.
	// It is intentionally NOT serialized (json:"-") — the WebSocket handler
	// reads it to persist to Redis, but the client never sees the raw history.
	UpdatedHistory []llm.Message `json:"-"`
}
