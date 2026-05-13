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
		processes = append(processes, Entry{Name: name, PID: pid})
	}
	return processes, nil
}

// WaitGone polls until the process with pid is gone and the pid file is deleted.
// Returns true if both conditions are met within timeout.
func (m Manager) WaitGone(pid int, name string, timeout time.Duration) bool {
	dir, _ := m.runningDir()
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
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if err := proc.Kill(); err != nil {
		return err
	}
	return os.Remove(pidFile)
}
