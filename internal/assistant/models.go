package assistant

import "github.com/krisnaganesha1609/LeviathanBolu-BE/internal/llm"

type ChatRequest struct {
	UserID  string
	Message string
}

type ChatResponse struct {
	Message string

	ToolCall *llm.ToolCall `json:"tool_call,omitempty"`
}
