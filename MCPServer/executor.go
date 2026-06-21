package MCPServer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
)

// DefaultToolTimeout caps how long a single tool execution may run
// before it's treated as failed. Without this, one slow/hanging tool
// (a flaky upstream API, a device that stopped responding, ...) could
// block the whole assistant indefinitely.
const DefaultToolTimeout = 15 * time.Second

// ToolTimeout is an optional extension a Tool can implement to override
// DefaultToolTimeout — e.g. a tool that calls a slow external API and
// genuinely needs more than 15s.
type ToolTimeout interface {
	Timeout() time.Duration
}

type Executor struct {
	registry *Registry
}

func NewExecutor(registry *Registry) *Executor {
	return &Executor{
		registry: registry,
	}
}

var ErrToolNotFound = errors.New("tool not found")

// Execute runs a registered tool by name.
//
// ctx is propagated from the caller (the originating HTTP/WebSocket
// request, or the assistant loop's own context) instead of being
// silently replaced with context.Background() — so cancelling the
// parent request also cancels any tool call it triggered. A per-call
// timeout is layered on top so a slow tool can't hang the assistant
// forever, and a panic inside a tool is recovered here and turned into
// a normal error instead of taking down the whole process.
func (e *Executor) Execute(ctx context.Context, toolName string, args map[string]any) (output any, err error) {
	tool, exists := e.registry.Get(toolName)
	if !exists {
		return nil, fmt.Errorf("%w: %q", ErrToolNotFound, toolName)
	}

	timeout := DefaultToolTimeout
	if t, ok := tool.(ToolTimeout); ok {
		timeout = t.Timeout()
	}
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[MCPServer] tool %q panicked: %v", toolName, r)
			err = fmt.Errorf("tool %q panicked: %v", toolName, r)
			output = nil
		}
	}()

	start := time.Now()
	output, err = tool.Execute(execCtx, args)
	elapsed := time.Since(start)

	if err != nil {
		log.Printf("[MCPServer] tool %q failed after %s: %v", toolName, elapsed, err)
		return nil, err
	}
	log.Printf("[MCPServer] tool %q completed in %s", toolName, elapsed)
	return output, nil
}
