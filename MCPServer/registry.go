package MCPServer

import "sync"

// ToolDefinition is the provider-agnostic shape describing a tool to an
// LLM's function/tool-calling API. OpenAI, Gemini, and Ollama's
// OpenAI-compatible endpoint all expect roughly this {name, description,
// parameters} shape, so this is what internal/llm provider adapters will
// consume to advertise "available tools" in a chat request (Stage 3).
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// Registry holds every tool the assistant knows how to call, keyed by
// name. Safe for concurrent use: Register typically only happens once at
// startup, but Get/List/Definitions will be read concurrently from every
// in-flight chat request once the assistant loop (Stage 4) is wired up.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry. Registering a second tool under
// the same name silently overwrites the first — check List() first if
// you need to detect that.
func (r *Registry) Register(tool Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name()] = tool
}

func (r *Registry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, exists := r.tools[name]
	return tool, exists
}

func (r *Registry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	toolList := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		toolList = append(toolList, tool)
	}
	return toolList
}

// Definitions returns every registered tool in the shape an LLM provider
// adapter needs in order to advertise "available functions" to the
// model. Pass the result straight into llm.ChatRequest.Tools (Stage 3).
func (r *Registry) Definitions() []ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	defs := make([]ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		defs = append(defs, ToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters:  tool.Schema(),
		})
	}
	return defs
}
