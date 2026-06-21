package assistant

import (
	"context"

	"github.com/krisnaganesha1609/LeviathanBolu-BE/MCPServer"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/internal/llm"
)

type AssistantService struct {
	llm llm.Provider

	executor *MCPServer.Executor
}

func NewAssistantService(llm llm.Provider, executor *MCPServer.Executor) *AssistantService {
	return &AssistantService{
		llm:      llm,
		executor: executor,
	}
}

func (s *AssistantService) Chat(c context.Context, req ChatRequest) (ChatResponse, error) {
	// Create a chat request for the LLM
	llmReq := llm.ChatRequest{
		Messages: []llm.Message{
			{
				Role:    "user",
				Content: req.Message,
			},
		},
	}

	llmResp, err := s.llm.Chat(c, llmReq)
	if err != nil {
		return ChatResponse{}, err
	}
	return ChatResponse{Message: llmResp.Content}, nil
}
