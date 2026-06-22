package dto

import "github.com/krisnaganesha1609/LeviathanBolu-BE/internal/llm"

// AssistantChatRequest is the HTTP request body for POST /api/assistant/chat.
type AssistantChatRequest struct {
	UserID  string `json:"user_id" validate:"required"`
	Message string `json:"message" validate:"required,min=1"`

	// History is optional. The client carries the conversation across
	// requests by echoing back what the server previously returned.
	// Stage 4 will add server-side session tracking in Redis so clients
	// won't need to do this manually.
	History []llm.Message `json:"history"`
}

// AssistantChatResponse is the HTTP response body.
type AssistantChatResponse struct {
	Reply     string   `json:"reply"`
	ToolsUsed []string `json:"tools_used,omitempty"`
}
