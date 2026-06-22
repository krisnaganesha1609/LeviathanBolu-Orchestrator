package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const openRouterBaseURL = "https://openrouter.ai/api/v1/chat/completions"

// OpenRouterProvider implements Provider using OpenRouter's
// OpenAI-compatible REST API — no extra SDK needed, plain net/http.
// Swap Gemini → OpenRouter in config by changing LLM_PROVIDER=openrouter.
type OpenRouterProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewOpenRouterProvider(apiKey, model string) *OpenRouterProvider {
	if model == "" {
		model = "google/gemini-2.5-flash" // change to any model OpenRouter hosts
	}
	return &OpenRouterProvider{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ── internal OpenAI-compatible wire types ────────────────────────────────

type orRequest struct {
	Model    string      `json:"model"`
	Messages []orMessage `json:"messages"`
	Tools    []orTool    `json:"tools,omitempty"`
}

type orMessage struct {
	Role       string       `json:"role"`
	Content    any          `json:"content"` // string | null
	ToolCalls  []orToolCall `json:"tool_calls,omitempty"`
	ToolCallID string       `json:"tool_call_id,omitempty"`
	Name       string       `json:"name,omitempty"`
}

type orToolCall struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"` // always "function"
	Function orToolCallFunc `json:"function"`
}

type orToolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string, not object
}

type orTool struct {
	Type     string     `json:"type"` // always "function"
	Function orToolFunc `json:"function"`
}

type orToolFunc struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type orResponse struct {
	Choices []struct {
		Message struct {
			Content   *string      `json:"content"`
			ToolCalls []orToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

// ── Provider implementation ───────────────────────────────────────────────

// Chat implements [Provider].
func (o *OpenRouterProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = o.model
	}

	msgs := make([]orMessage, 0, len(req.Messages)+1)

	if req.SystemPrompt != "" {
		msgs = append(msgs, orMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}

	for _, msg := range req.Messages {
		switch msg.Role {
		case RoleUser:
			msgs = append(msgs, orMessage{
				Role:    "user",
				Content: msg.Content,
			})

		case RoleModel:
			if msg.ToolCall != nil {
				argsJSON, _ := json.Marshal(msg.ToolCall.Arguments)
				msgs = append(msgs, orMessage{
					Role:    "assistant",
					Content: nil, // must be null when tool_calls is present
					ToolCalls: []orToolCall{{
						ID:   msg.ToolCall.ID,
						Type: "function",
						Function: orToolCallFunc{
							Name:      msg.ToolCall.Name,
							Arguments: string(argsJSON),
						},
					}},
				})
			} else {
				msgs = append(msgs, orMessage{
					Role:    "assistant",
					Content: msg.Content,
				})
			}

		case RoleTool:
			if msg.ToolResult == nil {
				continue
			}
			contentJSON, _ := json.Marshal(msg.ToolResult.Content)
			msgs = append(msgs, orMessage{
				Role:       "tool",
				Content:    string(contentJSON),
				ToolCallID: msg.ToolResult.ToolCallID,
				Name:       msg.ToolResult.Name,
			})
		}
	}

	payload := orRequest{Model: model, Messages: msgs}

	if len(req.Tools) > 0 {
		payload.Tools = make([]orTool, 0, len(req.Tools))
		for _, t := range req.Tools {
			payload.Tools = append(payload.Tools, orTool{
				Type: "function",
				Function: orToolFunc{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  t.Parameters,
				},
			})
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("openrouter: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, openRouterBaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("openrouter: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)
	// OpenRouter recommends these headers for attribution
	httpReq.Header.Set("HTTP-Referer", "https://github.com/krisnaganesha1609/LeviathanBolu-BE")
	httpReq.Header.Set("X-Title", "LeviathanBolu")

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openrouter: HTTP request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openrouter: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openrouter: API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var orResp orResponse
	if err := json.Unmarshal(respBody, &orResp); err != nil {
		return nil, fmt.Errorf("openrouter: unmarshal response: %w", err)
	}

	if orResp.Error != nil {
		return nil, fmt.Errorf("openrouter: API error %d: %s", orResp.Error.Code, orResp.Error.Message)
	}

	if len(orResp.Choices) == 0 {
		return nil, fmt.Errorf("openrouter: empty choices in response")
	}

	result := &ChatResponse{
		InputTokens:  orResp.Usage.PromptTokens,
		OutputTokens: orResp.Usage.CompletionTokens,
	}

	choice := orResp.Choices[0].Message
	if choice.Content != nil {
		result.Content = *choice.Content
	}

	if len(choice.ToolCalls) > 0 {
		tc := choice.ToolCalls[0]
		var args map[string]any
		json.Unmarshal([]byte(tc.Function.Arguments), &args)
		result.ToolCall = &ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: args,
		}
	}

	return result, nil
}
