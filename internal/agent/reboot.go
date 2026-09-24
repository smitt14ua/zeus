package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// rebootMessage is shown to logged-in users on platforms that support it.
const rebootMessage = "zeus agent: reboot requested by panel"

// rebootWaitDelay bounds how long we wait for shutdown's output pipes to close
// after the process exits. BSD/macOS shutdown forks a background child that
// inherits the pipes; without this, Run would block until the machine goes down
// and the result would never reach the panel.
const rebootWaitDelay = 3 * time.Second

// runCmd executes a prepared command. Replaced in tests to avoid rebooting.
var runCmd = func(cmd *exec.Cmd) error { return cmd.Run() }

// rebootCommand returns the OS command that schedules a machine reboot.
// The reboot is scheduled with a short delay rather than executed immediately
// so the agent has time to send the command result back to the panel, and so
// privilege errors surface synchronously as a command failure.
func rebootCommand(goos string) (name string, args []string, err error) {
	switch goos {
	case "windows":
		// Note: any /t > 0 implies /f — running applications are force-closed.
		return "shutdown", []string{"/r", "/t", "5", "/c", rebootMessage}, nil
	case "linux", "darwin", "freebsd", "openbsd", "netbsd":
		// shutdown(8) only accepts whole minutes; +1 is the shortest non-immediate delay.
		return "shutdown", []string{"-r", "+1", rebootMessage}, nil
	default:
		return "", nil, fmt.Errorf("reboot is not supported on %s", goos)
	}
}

// RebootMachine schedules a reboot of the machine the agent runs on.
// Output of the underlying shutdown command is written to w and, on failure,
// included in the returned error.
func RebootMachine(ctx context.Context, w io.Writer) error {
	return rebootMachine(ctx, runtime.GOOS, w)
}

func rebootMachine(ctx context.Context, goos string, w io.Writer) error {
	name, args, err := rebootCommand(goos)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "Scheduling reboot: %s %s %q\n", name, strings.Join(args[:len(args)-1], " "), args[len(args)-1])

	var out bytes.Buffer
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = io.MultiWriter(w, &out)
	cmd.Stderr = io.MultiWriter(w, &out)
	cmd.WaitDelay = rebootWaitDelay
	if err := runCmd(cmd); err != nil && !errors.Is(err, exec.ErrWaitDelay) {
		if msg := strings.TrimSpace(out.String()); msg != "" {
			return fmt.Errorf("scheduling reboot: %w: %s", err, msg)
		}
		return fmt.Errorf("scheduling reboot: %w", err)
	}
	fmt.Fprintln(w, "Reboot scheduled.")
	return nil
}
