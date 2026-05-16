package agent

import (
	"context"
	"fmt"
	"io"
	"sync"
)

// HandlerFunc is the signature every command handler must implement.
// args mirrors the Command.Args map from the protocol envelope.
// w receives all output lines that will be streamed back to the server.
type HandlerFunc func(ctx context.Context, args map[string]any, w io.Writer) error

// Executor is a registry of command handlers. It is safe for concurrent use.
type Executor struct {
	mu       sync.RWMutex
	handlers map[string]HandlerFunc
	allowed  map[string]bool // nil = allow all registered commands
}

func NewExecutor() *Executor {
	return &Executor{handlers: make(map[string]HandlerFunc)}
}

// SetAllowed restricts dispatch to the given command set.
// Pass nil to remove restrictions (allow all).
func (e *Executor) SetAllowed(allowed map[string]bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.allowed = allowed
}

// Register associates a command name with a handler. Calling Register twice
// for the same name replaces the previous handler.
func (e *Executor) Register(cmd string, h HandlerFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers[cmd] = h
}

// Dispatch looks up the handler for cmd and runs it.
// Returns an error if the command is unknown, not in the allowed set, or the handler fails.
func (e *Executor) Dispatch(ctx context.Context, cmd string, args map[string]any, w io.Writer) error {
	e.mu.RLock()
	h, ok := e.handlers[cmd]
	allowed := e.allowed
	e.mu.RUnlock()

	if !ok {
		return fmt.Errorf("unknown command: %s", cmd)
	}
	if allowed != nil && !allowed[cmd] {
		return fmt.Errorf("command %q is not permitted by this agent's scope configuration", cmd)
	}
	return h(ctx, args, w)
}

// StringArg extracts a string from args, returning "" when absent.
func StringArg(args map[string]any, key string) string {
	v, _ := args[key].(string)
	return v
}

// BoolArg extracts a bool from args, returning false when absent.
func BoolArg(args map[string]any, key string) bool {
	v, _ := args[key].(bool)
	return v
}
