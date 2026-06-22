package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/internal/assistant"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/dto"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/utils"
)

type AssistantHandler interface {
	Chat(c fiber.Ctx) error
}

type AssistantHandlerImpl struct {
	Service      *assistant.AssistantService
	SystemPrompt string
}

func InitAssistantHandler(svc *assistant.AssistantService, systemPrompt string) AssistantHandler {
	return &AssistantHandlerImpl{
		Service:      svc,
		SystemPrompt: systemPrompt,
	}
}

// Chat handles POST /api/assistant/chat — the single most important
// endpoint in the whole project. When this returns a reply that mentions
// a tool result, JARVIS is officially alive.
func (h *AssistantHandlerImpl) Chat(c fiber.Ctx) error {
	var req dto.AssistantChatRequest
	if err := c.Bind().Body(&req); err != nil {
		return utils.ResponseBadRequest(c, []utils.ValidationError{
			{Field: "body", Message: "invalid request body"},
		})
	}
	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		return utils.ResponseBadRequest(c, errs)
	}

	resp, err := h.Service.Chat(c.Context(), assistant.ChatRequest{
		UserID:       req.UserID,
		Message:      req.Message,
		History:      req.History,
		SystemPrompt: h.SystemPrompt,
	})
	if err != nil {
		return err // caught by globalErrorHandler in main.go
	}

	return utils.ResponseOK(c, "success", dto.AssistantChatResponse{
		Reply:     resp.Message,
		ToolsUsed: resp.ToolsUsed,
	})
}
