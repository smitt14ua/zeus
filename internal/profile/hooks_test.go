package profile

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeProfileWithDir(t *testing.T) Profile {
	t.Helper()
	base := t.TempDir()
	p := Profile{Name: "test-srv", InstallDir: base}
	dir := filepath.Join(base, ".zeus", p.Name)
	require.NoError(t, os.MkdirAll(dir, 0755))
	return p
}

// shellWrite returns a shell command that writes text to a file (cross-platform).
func shellWrite(text, path string) string {
	if runtime.GOOS == "windows" {
		return "echo " + text + " > " + path
	}
	return "echo " + text + " > " + path
}

// exitOneCmd returns a command that exits with code 1 on the current OS.
func exitOneCmd() string {
	if runtime.GOOS == "windows" {
		return "exit /b 1"
	}
	return "exit 1"
}

func TestRunHooks_Empty(t *testing.T) {
	p := makeProfileWithDir(t)
	require.NoError(t, RunHooks(nil, p))
	require.NoError(t, RunHooks([]string{}, p))
}

func TestRunHooks_Success(t *testing.T) {
	p := makeProfileWithDir(t)
	require.NoError(t, RunHooks([]string{"echo hello"}, p))
}

func TestRunHooks_StopsOnError(t *testing.T) {
	p := makeProfileWithDir(t)
	sentinel := filepath.Join(p.InstallDir, ".zeus", p.Name, "ran")
	err := RunHooks([]string{exitOneCmd(), shellWrite("ran", sentinel)}, p)
	require.Error(t, err)
	assert.NoFileExists(t, sentinel)
}

func TestRunHooks_EnvVars(t *testing.T) {
	p := makeProfileWithDir(t)
	profileDir := filepath.Join(p.InstallDir, ".zeus", p.Name)
	out := filepath.Join(p.InstallDir, "env_out.txt")

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "echo %ZEUS_PROFILE% %ZEUS_PROFILE_DIR% %ZEUS_PROFILE_INSTALL_DIR% > " + out
	} else {
		cmd = "echo $ZEUS_PROFILE $ZEUS_PROFILE_DIR $ZEUS_PROFILE_INSTALL_DIR > " + out
	}

	require.NoError(t, RunHooks([]string{cmd}, p))

	data, err := os.ReadFile(out)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, p.Name)
	assert.Contains(t, content, profileDir)
	assert.Contains(t, content, p.InstallDir)
}

func TestRunHooks_WorkingDir(t *testing.T) {
	p := makeProfileWithDir(t)
	profileDir := filepath.Join(p.InstallDir, ".zeus", p.Name)
	out := filepath.Join(profileDir, "pwd_out.txt")

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "cd > " + out
	} else {
		cmd = "pwd > " + out
	}

	require.NoError(t, RunHooks([]string{cmd}, p))

	data, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Contains(t, string(data), ".zeus")
}
