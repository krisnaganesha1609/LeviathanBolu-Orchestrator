package llm

import "context"

type Provider interface {
	Chat(c context.Context, req ChatRequest) (*ChatResponse, error)
}
