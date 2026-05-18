package process

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Entry struct {
	Name string
	PID  int
}

type Manager struct {
	HomeDir string
}

func (m Manager) runningDir() (string, error) {
	home := m.HomeDir
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(home, ".zeus", "running"), nil
}

func (m Manager) List() ([]Entry, error) {
	dir, err := m.runningDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var processes []Entry
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".pid") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".pid")
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil {
			return nil, fmt.Errorf("invalid PID in %s: %w", e.Name(), err)
		}
		if pid <= 0 {
			return nil, fmt.Errorf("invalid PID %d in %s", pid, e.Name())
		}
		processes = append(processes, Entry{Name: name, PID: pid})
	}
	return processes, nil
}

// WaitReady polls until the server writes its PID file, then holds for stabilityWindow
// to confirm the process hasn't crashed immediately. directPID is the OS PID returned
// by Runner.Run; if it exits before the PID file appears the function fails fast.
// Returns the PID read from the file on success.
func (m Manager) WaitReady(name string, directPID int, pidTimeout, stabilityWindow time.Duration) (int, error) {
	dir, err := m.runningDir()
	if err != nil {
		return 0, err
	}
	pidFile := filepath.Join(dir, name+".pid")
	deadline := time.Now().Add(pidTimeout)

	for time.Now().Before(deadline) {
		// Fast-fail: if the launch process already exited and the PID file still isn't there, it crashed.
		if !processExists(directPID) {
			if _, statErr := os.Stat(pidFile); os.IsNotExist(statErr) {
				return 0, fmt.Errorf("profile %q: launch process exited before writing PID file", name)
			}
			// PID file exists even though the launcher exited (server may have forked) — fall through.
		}

		data, readErr := os.ReadFile(pidFile)
		if readErr == nil {
			pid, parseErr := strconv.Atoi(strings.TrimSpace(string(data)))
			if parseErr != nil {
				return 0, fmt.Errorf("invalid PID in %s.pid: %w", name, parseErr)
			}
			// Stability window: confirm the process hasn't immediately exited.
			stableUntil := time.Now().Add(stabilityWindow)
			for time.Now().Before(stableUntil) {
				if !processExists(pid) {
					return 0, fmt.Errorf("profile %q: server process %d exited shortly after start", name, pid)
				}
				time.Sleep(200 * time.Millisecond)
			}
			return pid, nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return 0, fmt.Errorf("profile %q did not write a PID file within %s", name, pidTimeout)
}

// WaitGone polls until the process with pid is gone and the pid file is deleted.
// Returns true if both conditions are met within timeout.
func (m Manager) WaitGone(pid int, name string, timeout time.Duration) bool {
	dir, err := m.runningDir()
	if err != nil {
		return false
	}
	pidFile := filepath.Join(dir, name+".pid")
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, pidFileErr := os.Stat(pidFile)
		if os.IsNotExist(pidFileErr) && !processExists(pid) {
			return true
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}

func (m Manager) Kill(name string) error {
	dir, err := m.runningDir()
	if err != nil {
		return err
	}
	pidFile := filepath.Join(dir, name+".pid")
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return fmt.Errorf("invalid PID in %s.pid: %w", name, err)
	}
	if pid <= 0 {
		return fmt.Errorf("invalid PID %d in %s.pid", pid, name)
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if err := proc.Kill(); err != nil {
		return err
	}
	return os.Remove(pidFile)
}
