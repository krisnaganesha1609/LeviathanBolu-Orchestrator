package assistant

import (
	"github.com/google/uuid"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/internal/llm"
)

// ChatRequest is what the handler layer passes to AssistantService.Chat.
type ChatRequest struct {
	UserID  string
	Message string
	History []llm.Message

	// Personality dipilih berdasarkan wake word di Flutter.
	// "bolu" → warm companion | "leviathan" → ancient dragon advisor
	// Default ke "bolu" jika kosong.
	Personality string

	// SystemPrompt override manual. Kalau kosong, diisi otomatis
	// berdasarkan Personality field di atas.
	SystemPrompt string

	VoiceMode bool

	// ConversationID is the DeviceSession.ID created when the WebSocket
	// connection was established. Passed through to STT / TTS workers so
	// they can correlate audio with a specific session.
	ConversationID uuid.UUID
}

// ChatResponse is what AssistantService.Chat returns to the handler.
type ChatResponse struct {
	Message string `json:"message"`

	// ToolsUsed lists the names of every tool that was called during
	// this turn (useful for the Flutter Orb to animate per-tool states).
	ToolsUsed []string `json:"tools_used,omitempty"`
}
