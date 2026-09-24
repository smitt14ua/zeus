package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func TestRebootCommand_Windows(t *testing.T) {
	name, args, err := rebootCommand("windows")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "shutdown" {
		t.Errorf("name = %q, want shutdown", name)
	}
	want := []string{"/r", "/t", "5", "/c", rebootMessage}
	if !slices.Equal(args, want) {
		t.Errorf("args = %v, want %v", args, want)
	}
}

func TestRebootCommand_Unix(t *testing.T) {
	want := []string{"-r", "+1", rebootMessage}
	for _, goos := range []string{"linux", "darwin", "freebsd", "openbsd", "netbsd"} {
		name, args, err := rebootCommand(goos)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", goos, err)
		}
		if name != "shutdown" {
			t.Errorf("%s: name = %q, want shutdown", goos, name)
		}
		if !slices.Equal(args, want) {
			t.Errorf("%s: args = %v, want %v", goos, args, want)
		}
	}
}

func TestRebootCommand_Unsupported(t *testing.T) {
	if _, _, err := rebootCommand("plan9"); err == nil {
		t.Error("expected error for unsupported OS")
	}
}

// stubRunCmd replaces runCmd for the duration of the test.
func stubRunCmd(t *testing.T, fn func(*exec.Cmd) error) {
	t.Helper()
	orig := runCmd
	runCmd = fn
	t.Cleanup(func() { runCmd = orig })
}

func TestRebootMachine_Success(t *testing.T) {
	var got *exec.Cmd
	stubRunCmd(t, func(cmd *exec.Cmd) error {
		got = cmd
		return nil
	})

	var w bytes.Buffer
	if err := rebootMachine(context.Background(), "linux", &w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || !slices.Equal(got.Args[1:], []string{"-r", "+1", rebootMessage}) {
		t.Errorf("unexpected command: %v", got)
	}
	if got.WaitDelay != rebootWaitDelay {
		t.Errorf("WaitDelay = %v, want %v", got.WaitDelay, rebootWaitDelay)
	}
	out := w.String()
	if !strings.Contains(out, "Scheduling reboot: shutdown -r +1 \""+rebootMessage+"\"") {
		t.Errorf("missing scheduling line in %q", out)
	}
	if !strings.HasSuffix(out, "Reboot scheduled.\n") {
		t.Errorf("missing success line in %q", out)
	}
}

func TestRebootMachine_WaitDelayIsSuccess(t *testing.T) {
	stubRunCmd(t, func(*exec.Cmd) error { return exec.ErrWaitDelay })

	var w bytes.Buffer
	if err := rebootMachine(context.Background(), "darwin", &w); err != nil {
		t.Fatalf("ErrWaitDelay should be treated as success, got %v", err)
	}
	if !strings.Contains(w.String(), "Reboot scheduled.") {
		t.Errorf("missing success line in %q", w.String())
	}
}

func TestRebootMachine_FailureIncludesOutput(t *testing.T) {
	stubRunCmd(t, func(cmd *exec.Cmd) error {
		fmt.Fprintln(cmd.Stderr, "Access is denied.")
		return errors.New("exit status 5")
	})

	var w bytes.Buffer
	err := rebootMachine(context.Background(), "windows", &w)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "exit status 5") || !strings.Contains(err.Error(), "Access is denied.") {
		t.Errorf("error = %q, want exit status and command output", err)
	}
	if !strings.Contains(w.String(), "Access is denied.") {
		t.Errorf("command output not streamed: %q", w.String())
	}
	if strings.Contains(w.String(), "Reboot scheduled.") {
		t.Error("success line must not be written on failure")
	}
}

func TestRebootMachine_Unsupported(t *testing.T) {
	stubRunCmd(t, func(*exec.Cmd) error {
		t.Fatal("runCmd must not be called on unsupported OS")
		return nil
	})
	if err := rebootMachine(context.Background(), "plan9", &bytes.Buffer{}); err == nil {
		t.Error("expected error for unsupported OS")
	}
}
