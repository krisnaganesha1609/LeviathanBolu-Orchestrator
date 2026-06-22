package assistant

import "github.com/krisnaganesha1609/LeviathanBolu-BE/internal/llm"

// ChatRequest is what the handler layer passes to AssistantService.Chat.
type ChatRequest struct {
	// UserID identifies whose settings/preferences to load.
	UserID string

	// Message is the new text from the user for this turn.
	Message string

	// History is the conversation so far, NOT including the new Message.
	// The handler populates this from whatever the client sent.
	// Stage 4 will persist history in Redis so clients don't need to
	// carry the full transcript themselves.
	History []llm.Message

	// SystemPrompt overrides the default prompt built by BuildSystemPrompt.
	// Leave empty to use the default.
	SystemPrompt string
}

// ChatResponse is what AssistantService.Chat returns to the handler.
type ChatResponse struct {
	Message string `json:"message"`

	// ToolsUsed lists the names of every tool that was called during
	// this turn (useful for the Flutter Orb to animate per-tool states).
	ToolsUsed []string `json:"tools_used,omitempty"`
}
