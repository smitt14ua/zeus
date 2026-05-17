package agent

import (
	"errors"
	"os"
	"os/exec"
)

// ErrRestartRequested is returned by a command handler to signal that the agent
// should restart itself after sending a success result to the panel.
var ErrRestartRequested = errors.New("agent restart requested")

// restartSelf spawns a new process from the current executable with the same
// arguments, then exits. go-selfupdate has already replaced the binary on disk
// before this is called, so the new process runs the updated version.
func restartSelf() {
	exe, err := os.Executable()
	if err != nil {
		os.Exit(1)
	}
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
