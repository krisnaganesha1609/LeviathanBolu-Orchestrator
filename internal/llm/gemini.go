package llm

import (
	"context"
	"encoding/json"
	"fmt"

	genai "google.golang.org/genai"
)

// GeminiProvider implements Provider using Google's official Gen AI Go SDK
// (google.golang.org/genai v1.52.1+, GA since May 2025).
// NOT to be confused with the legacy github.com/google/generative-ai-go.
type GeminiProvider struct {
	client *genai.Client
	model  string
}

// NewGeminiProvider creates a Gemini client backed by the Gemini Developer
// API (api.google.ai). For Vertex AI, swap BackendGeminiAPI → BackendVertexAI
// and provide Project/Location in ClientConfig instead of APIKey.
func NewGeminiProvider(ctx context.Context, apiKey, model string) (*GeminiProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("gemini: GEMINI_API_KEY is required")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini: failed to create client: %w", err)
	}

	if model == "" {
		model = "gemini-2.5-flash"
	}

	return &GeminiProvider{client: client, model: model}, nil
}

// Chat implements [Provider].
func (g *GeminiProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = g.model
	}

	contents, err := messagesToGeminiContents(req.Messages)
	if err != nil {
		return nil, fmt.Errorf("gemini: %w", err)
	}

	cfg := &genai.GenerateContentConfig{}

	if req.SystemPrompt != "" {
		cfg.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{{Text: req.SystemPrompt}},
		}
	}

	if len(req.Tools) > 0 {
		cfg.Tools = []*genai.Tool{
			{FunctionDeclarations: toGeminiFunctionDeclarations(req.Tools)},
		}
	}

	resp, err := g.client.Models.GenerateContent(ctx, model, contents, cfg)
	if err != nil {
		return nil, fmt.Errorf("gemini: GenerateContent failed: %w", err)
	}

	return parseGeminiResponse(resp)
}

// messagesToGeminiContents converts our internal Message slice into the
// []*genai.Content shape Gemini's SDK expects.
func messagesToGeminiContents(messages []Message) ([]*genai.Content, error) {
	contents := make([]*genai.Content, 0, len(messages))

	for _, msg := range messages {
		switch msg.Role {
		case RoleUser:
			contents = append(contents, &genai.Content{
				Role:  "user",
				Parts: []*genai.Part{{Text: msg.Content}},
			})

		case RoleModel:
			if msg.ToolCall != nil {
				// Model's turn was a function call — replay it exactly.
				contents = append(contents, &genai.Content{
					Role: "model",
					Parts: []*genai.Part{{
						FunctionCall: &genai.FunctionCall{
							Name: msg.ToolCall.Name,
							Args: msg.ToolCall.Arguments,
						},
					}},
				})
			} else {
				contents = append(contents, &genai.Content{
					Role:  "model",
					Parts: []*genai.Part{{Text: msg.Content}},
				})
			}

		case RoleTool:
			if msg.ToolResult == nil {
				continue
			}
			// Gemini expects function responses as a "user" role turn.
			contents = append(contents, &genai.Content{
				Role: "user",
				Parts: []*genai.Part{{
					FunctionResponse: &genai.FunctionResponse{
						Name: msg.ToolResult.Name,
						Response: map[string]any{
							"result": msg.ToolResult.Content,
						},
					},
				}},
			})
		}
	}

	return contents, nil
}

// parseGeminiResponse extracts text or a tool call from the raw SDK response.
func parseGeminiResponse(resp *genai.GenerateContentResponse) (*ChatResponse, error) {
	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, fmt.Errorf("gemini: empty response from model")
	}

	result := &ChatResponse{}

	if resp.UsageMetadata != nil {
		result.InputTokens = int(resp.UsageMetadata.PromptTokenCount)
		result.OutputTokens = int(resp.UsageMetadata.CandidatesTokenCount)
	}

	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			result.Content += part.Text
		}
		if part.FunctionCall != nil {
			result.ToolCall = &ToolCall{
				Name:      part.FunctionCall.Name,
				Arguments: part.FunctionCall.Args,
			}
			// Only one tool call per turn for now (Gemini may return
			// multiple in parallel — we'll handle that in Stage 5).
			break
		}
	}

	return result, nil
}

// toGeminiFunctionDeclarations maps our generic ToolDef slice to the
// Gemini SDK's []*genai.FunctionDeclaration.
func toGeminiFunctionDeclarations(tools []ToolDef) []*genai.FunctionDeclaration {
	decls := make([]*genai.FunctionDeclaration, 0, len(tools))
	for _, t := range tools {
		decl := &genai.FunctionDeclaration{
			Name:        t.Name,
			Description: t.Description,
		}
		if t.Parameters != nil {
			decl.Parameters = jsonMapToGeminiSchema(t.Parameters)
		}
		decls = append(decls, decl)
	}
	return decls
}

// jsonMapToGeminiSchema converts a JSON Schema map[string]any (the shape
// MCPServer.Tool.Schema() returns) into a *genai.Schema.
// Marshal → Unmarshal is the cleanest approach because genai.Schema has
// proper json tags that mirror JSON Schema field names.
func jsonMapToGeminiSchema(m map[string]any) *genai.Schema {
	b, err := json.Marshal(m)
	if err != nil {
		return &genai.Schema{Type: "object"}
	}
	var s genai.Schema
	if err := json.Unmarshal(b, &s); err != nil {
		return &genai.Schema{Type: "object"}
	}
	return &s
}
