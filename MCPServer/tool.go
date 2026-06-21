package MCPServer

import "context"

// Tool is a single capability the assistant can invoke — checking
// hydroponic sensor readings, sending an email, querying a calendar, etc.
// Tool logic and credentials live entirely on the backend; the LLM only
// ever sees Name/Description/Schema and a JSON args map — never API
// keys, bearer tokens, or database access.
type Tool interface {
	// Name is the stable identifier the LLM uses to call this tool (e.g.
	// "ehydrotel_get_status"). Must be unique across the registry.
	Name() string

	// Description tells the LLM what this tool does and when to use it.
	// This is the single biggest factor in whether the LLM picks the
	// right tool for a given request — be specific about what it does
	// and doesn't cover, not just a one-line label.
	Description() string

	// Schema describes the INPUT this tool expects, as a JSON Schema
	// object — i.e. the arguments the LLM must fill in before calling
	// Execute. It is NOT a description of what Execute returns.
	//
	// For a tool that takes no arguments, return an empty object schema:
	//   map[string]any{"type": "object", "properties": map[string]any{}}
	Schema() map[string]any

	// Execute runs the tool. args is whatever JSON object the LLM
	// produced for this call — it has the shape Schema() described, but
	// since it came from an LLM's output rather than a compiler, Execute
	// should still defensively type-check/validate before using it.
	Execute(c context.Context, args map[string]any) (any, error)
}
