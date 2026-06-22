package llm

// Role constants used in Message.Role.
//
// NOTE: Gemini calls the assistant's turn "model" while
// OpenAI/OpenRouter calls it "assistant". Internally we use "model"
// (Gemini's convention) and the OpenRouter adapter translates it.
const (
	RoleUser   = "user"
	RoleModel  = "model" // maps to "assistant" in OpenAI-compatible APIs
	RoleSystem = "system"
	RoleTool   = "tool"
)

// ToolDef is the provider-agnostic description of a callable tool.
// It mirrors MCPServer.ToolDefinition so that this package stays
// decoupled from MCPServer. The assistant service does the translation.
type ToolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// ToolCall carries a model's intent to invoke a tool.
// ID is a provider-generated correlation id (OpenAI always sets it,
// Gemini may leave it empty).
type ToolCall struct {
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// ToolResult carries the output of a tool back to the model.
type ToolResult struct {
	ToolCallID string `json:"tool_call_id,omitempty"`
	Name       string `json:"name"`
	Content    any    `json:"content"`
}

// Message is one turn in the conversation.
type Message struct {
	Role string `json:"role"` // RoleUser | RoleModel | RoleTool

	// Content is the plain text for user/model turns.
	Content string `json:"content,omitempty"`

	// ToolCall is set when Role == RoleModel and the model wants to call
	// a tool instead of (or in addition to) returning text.
	ToolCall *ToolCall `json:"tool_call,omitempty"`

	// ToolResult is set when Role == RoleTool, carrying the tool output
	// back to the model for the next turn.
	ToolResult *ToolResult `json:"tool_result,omitempty"`
}

// ChatRequest is everything a single LLM call needs.
type ChatRequest struct {
	// SystemPrompt is injected as a system instruction.
	SystemPrompt string `json:"system_prompt,omitempty"`

	// Messages is the full conversation history including the newest
	// user message. Must contain at least one message.
	Messages []Message `json:"messages"`

	// Tools is the list of tools the model may call.
	// Empty = no tool calling, model always responds with text.
	Tools []ToolDef `json:"tools,omitempty"`

	// Model overrides the default model for this specific call.
	Model string `json:"model,omitempty"`
}

// ChatResponse is what a single LLM call returns.
type ChatResponse struct {
	// Content is the model's text reply. Empty when ToolCall is set.
	Content string `json:"content,omitempty"`

	// ToolCall is set when the model wants to invoke a tool instead of
	// replying with text. The assistant service executes it and sends
	// the result back in a follow-up ChatRequest.
	ToolCall *ToolCall `json:"tool_call,omitempty"`

	// Token usage — useful for cost tracking and context-window management.
	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`
}
