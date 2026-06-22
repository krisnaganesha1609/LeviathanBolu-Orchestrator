package assistant

import (
	"context"
	"errors"
	"fmt"
	"log"

	MCPServer "github.com/krisnaganesha1609/LeviathanBolu-BE/MCPServer"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/internal/llm"
)

// maxToolRounds is the maximum number of tool-call → tool-result cycles
// per user turn. Guards against an LLM stuck in an infinite calling loop.
const maxToolRounds = 5

// AssistantService orchestrates the full tool-calling loop:
//
//	user message → LLM → [tool call → execute → LLM]×N → text reply
type AssistantService struct {
	llm      llm.Provider
	registry *MCPServer.Registry
	executor *MCPServer.Executor
}

func NewAssistantService(
	provider llm.Provider,
	registry *MCPServer.Registry,
	executor *MCPServer.Executor,
) *AssistantService {
	return &AssistantService{
		llm:      provider,
		registry: registry,
		executor: executor,
	}
}

// Stream runs the full tool-calling loop and writes an Event to the
// provided channel at every meaningful step. The caller is responsible
// for creating the channel; Stream closes it when it returns.
//
// Design: Stream() is the canonical implementation. Chat() is just a
// thin wrapper that drains the channel and extracts the final reply —
// so both the HTTP and WebSocket endpoints share the exact same logic.
func (s *AssistantService) Stream(
	ctx context.Context,
	req ChatRequest,
	events chan<- Event,
) {
	defer close(events)

	send := func(e Event) {
		select {
		case events <- e:
		case <-ctx.Done():
		}
	}

	toolDefs := registryToLLMTools(s.registry)

	systemPrompt := req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = BuildSystemPrompt("LeviathanBolu", "id")
	}

	// Build working history: prior turns + new user message.
	history := make([]llm.Message, 0, len(req.History)+1)
	history = append(history, req.History...)
	history = append(history, llm.Message{
		Role:    llm.RoleUser,
		Content: req.Message,
	})

	var toolsUsed []string

	for round := 0; round < maxToolRounds; round++ {
		// Signal to the Orb: LLM is being called.
		send(Event{Event: EventThinking})

		llmResp, err := s.llm.Chat(ctx, llm.ChatRequest{
			SystemPrompt: systemPrompt,
			Messages:     history,
			Tools:        toolDefs,
		})
		if err != nil {
			send(Event{Event: EventError, Message: fmt.Sprintf("LLM call failed: %v", err)})
			return
		}

		// ── No tool call: model returned a text reply → we're done. ──────
		if llmResp.ToolCall == nil {
			history = append(history, llm.Message{
				Role:    llm.RoleModel,
				Content: llmResp.Content,
			})
			send(Event{Event: EventReply, Message: llmResp.Content})
			send(Event{
				Event:          EventDone,
				Data:           map[string]any{"tools_used": toolsUsed},
				UpdatedHistory: history, // picked up by WebSocket handler, not sent to client
			})
			return
		}

		// ── Tool-calling path ─────────────────────────────────────────────
		toolCall := llmResp.ToolCall
		toolsUsed = append(toolsUsed, toolCall.Name)

		log.Printf("[assistant] round %d/%d → calling tool %q args=%v",
			round+1, maxToolRounds, toolCall.Name, toolCall.Arguments)

		// Signal to the Orb: a specific tool is running.
		send(Event{Event: EventToolCalled, Tool: toolCall.Name})

		// Append the model's tool-call turn before executing so the next
		// LLM call gets the full context.
		history = append(history, llm.Message{
			Role:     llm.RoleModel,
			ToolCall: toolCall,
		})

		// Execute the tool.
		toolResult, toolErr := s.executor.Execute(ctx, toolCall.Name, toolCall.Arguments)
		if toolErr != nil {
			log.Printf("[assistant] tool %q failed: %v", toolCall.Name, toolErr)
			// Surface the error as the tool result so the LLM can respond
			// gracefully ("sorry, the sensor is offline") rather than
			// crashing the whole turn.
			if errors.Is(toolErr, MCPServer.ErrToolNotFound) {
				toolResult = map[string]any{
					"error": fmt.Sprintf("tool %q is not registered", toolCall.Name),
				}
			} else {
				toolResult = map[string]any{"error": toolErr.Error()}
			}
		}

		// Signal to the Orb: tool finished, here's what it returned.
		send(Event{Event: EventToolResult, Tool: toolCall.Name, Data: toolResult})

		// Append tool result for the next LLM call.
		history = append(history, llm.Message{
			Role: llm.RoleTool,
			ToolResult: &llm.ToolResult{
				ToolCallID: toolCall.ID,
				Name:       toolCall.Name,
				Content:    toolResult,
			},
		})
		// Loop → LLM gets the tool result and (hopefully) returns a reply.
	}

	send(Event{
		Event:   EventError,
		Message: fmt.Sprintf("exceeded max tool rounds (%d) without a text reply — possible LLM loop", maxToolRounds),
	})
}

// Chat is a simple request/response wrapper around Stream. It is used by
// the HTTP endpoint (POST /api/assistant/chat) which doesn't need events.
func (s *AssistantService) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	events := make(chan Event, 20)

	go s.Stream(ctx, req, events)

	var reply string
	var toolsUsed []string
	var lastErr string

	for e := range events {
		switch e.Event {
		case EventReply:
			reply = e.Message
		case EventDone:
			if data, ok := e.Data.(map[string]any); ok {
				if tu, ok := data["tools_used"].([]string); ok {
					toolsUsed = tu
				}
			}
		case EventError:
			lastErr = e.Message
		}
	}

	if lastErr != "" {
		return nil, fmt.Errorf("assistant: %s", lastErr)
	}
	return &ChatResponse{Message: reply, ToolsUsed: toolsUsed}, nil
}

// registryToLLMTools translates MCPServer definitions → llm.ToolDef so
// the llm package stays decoupled from MCPServer.
func registryToLLMTools(registry *MCPServer.Registry) []llm.ToolDef {
	defs := registry.Definitions()
	out := make([]llm.ToolDef, 0, len(defs))
	for _, d := range defs {
		out = append(out, llm.ToolDef{
			Name:        d.Name,
			Description: d.Description,
			Parameters:  d.Parameters,
		})
	}
	return out
}
