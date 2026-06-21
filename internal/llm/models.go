package llm

type ChatRequest struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Content string `json:"content"`

	ToolCall *ToolCall `json:"tool_call,omitempty"`
}

type ToolCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}
