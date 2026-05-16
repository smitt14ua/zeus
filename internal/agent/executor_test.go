package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
)

func TestExecutor_UnknownCommand(t *testing.T) {
	exec := NewExecutor()
	err := exec.Dispatch(context.Background(), "no.such", nil, io.Discard)
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}

func TestExecutor_HandlerCalled(t *testing.T) {
	exec := NewExecutor()
	called := false
	exec.Register("test.cmd", func(ctx context.Context, args map[string]any, w io.Writer) error {
		called = true
		fmt.Fprintf(w, "hello")
		return nil
	})

	var buf bytes.Buffer
	if err := exec.Dispatch(context.Background(), "test.cmd", nil, &buf); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
	if buf.String() != "hello" {
		t.Errorf("output = %q, want %q", buf.String(), "hello")
	}
}

func TestExecutor_HandlerError(t *testing.T) {
	exec := NewExecutor()
	exec.Register("fail.cmd", func(ctx context.Context, args map[string]any, w io.Writer) error {
		return errors.New("boom")
	})
	err := exec.Dispatch(context.Background(), "fail.cmd", nil, io.Discard)
	if err == nil || err.Error() != "boom" {
		t.Errorf("expected boom, got %v", err)
	}
}

func TestExecutor_RegisterOverwrite(t *testing.T) {
	exec := NewExecutor()
	exec.Register("cmd", func(_ context.Context, _ map[string]any, w io.Writer) error {
		fmt.Fprint(w, "first")
		return nil
	})
	exec.Register("cmd", func(_ context.Context, _ map[string]any, w io.Writer) error {
		fmt.Fprint(w, "second")
		return nil
	})
	var buf bytes.Buffer
	exec.Dispatch(context.Background(), "cmd", nil, &buf)
	if buf.String() != "second" {
		t.Errorf("expected second handler to win, got %q", buf.String())
	}
}

func TestStringArg(t *testing.T) {
	args := map[string]any{"name": "my-server", "other": 42}
	if got := StringArg(args, "name"); got != "my-server" {
		t.Errorf("got %q", got)
	}
	if got := StringArg(args, "missing"); got != "" {
		t.Errorf("missing key should return empty, got %q", got)
	}
	if got := StringArg(args, "other"); got != "" {
		t.Errorf("non-string value should return empty, got %q", got)
	}
}

func TestBoolArg(t *testing.T) {
	args := map[string]any{"flag": true, "off": false}
	if !BoolArg(args, "flag") {
		t.Error("expected true")
	}
	if BoolArg(args, "off") {
		t.Error("expected false")
	}
	if BoolArg(args, "missing") {
		t.Error("missing key should return false")
	}
}
