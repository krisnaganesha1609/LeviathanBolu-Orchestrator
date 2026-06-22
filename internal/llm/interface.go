package llm

import "context"

// Provider is the single interface every LLM adapter must implement.
// Adding a new provider = new file in this package + satisfy this interface.
// The assistant service never needs to change.
type Provider interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}
