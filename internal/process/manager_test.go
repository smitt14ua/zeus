package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func newManager(t *testing.T) Manager {
	t.Helper()
	return Manager{HomeDir: t.TempDir()}
}

func writeRunningDir(t *testing.T, m Manager) string {
	t.Helper()
	dir := filepath.Join(m.HomeDir, ".zeus", "running")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func writePIDFile(t *testing.T, dir, name string, pid int) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name+".pid"), fmt.Appendf(nil, "%d", pid), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestManager_ListEmpty(t *testing.T) {
	m := newManager(t)

	entries, err := m.List()
	if err != nil {
		t.Fatalf("List on missing dir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty, got %v", entries)
	}
}

func TestManager_List(t *testing.T) {
	m := newManager(t)
	dir := writeRunningDir(t, m)
	writePIDFile(t, dir, "server1", 1234)
	writePIDFile(t, dir, "server2", 5678)

	entries, err := m.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	byName := make(map[string]int, len(entries))
	for _, e := range entries {
		byName[e.Name] = e.PID
	}
	if byName["server1"] != 1234 {
		t.Errorf("server1 PID mismatch: got %d", byName["server1"])
	}
	if byName["server2"] != 5678 {
		t.Errorf("server2 PID mismatch: got %d", byName["server2"])
	}
}

func TestManager_Kill(t *testing.T) {
	m := newManager(t)
	dir := writeRunningDir(t, m)

	cmd := exec.Command("ping", "-n", "30", "127.0.0.1")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start dummy process: %v", err)
	}
	pid := cmd.Process.Pid
	writePIDFile(t, dir, "dummy", pid)

	if err := m.Kill("dummy"); err != nil {
		t.Fatalf("Kill: %v", err)
	}

	pidFile := filepath.Join(dir, "dummy.pid")
	if _, err := os.Stat(pidFile); err == nil {
		t.Error("PID file should be removed after Kill")
	}

	if err := cmd.Wait(); err == nil {
		t.Error("expected process to be dead after Kill")
	}
}

func TestManager_KillMissing(t *testing.T) {
	m := newManager(t)
	writeRunningDir(t, m)

	if err := m.Kill("ghost"); err == nil {
		t.Fatal("expected error killing non-existent PID file")
	}
}

func TestManager_List_RejectsNonPositivePID(t *testing.T) {
	for _, pid := range []int{0, -1} {
		m := newManager(t)
		dir := writeRunningDir(t, m)
		writePIDFile(t, dir, "bad", pid)

		_, err := m.List()
		if err == nil {
			t.Fatalf("List should reject PID %d", pid)
		}
	}
}

func TestManager_Kill_RejectsNonPositivePID(t *testing.T) {
	for _, pid := range []int{0, -1} {
		m := newManager(t)
		dir := writeRunningDir(t, m)
		writePIDFile(t, dir, "bad", pid)

		if err := m.Kill("bad"); err == nil {
			t.Fatalf("Kill should reject PID %d", pid)
		}
	}
}
