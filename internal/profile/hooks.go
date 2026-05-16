package profile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// RunHooks executes each command sequentially.
// Each command runs with ZEUS_* env vars appended to the current environment
// and with the profile directory as the working directory.
// Execution stops on the first non-zero exit.
func RunHooks(commands []string, p Profile) error {
	if len(commands) == 0 {
		return nil
	}
	profileDir := filepath.Join(p.InstallDir, ".zeus", p.Name)
	extra := []string{
		"ZEUS_PROFILE=" + p.Name,
		"ZEUS_PROFILE_DIR=" + profileDir,
		"ZEUS_PROFILE_INSTALL_DIR=" + p.InstallDir,
	}
	for _, command := range commands {
		if err := runHookCommand(command, extra, profileDir); err != nil {
			return fmt.Errorf("hook %q: %w", command, err)
		}
	}
	return nil
}

func runHookCommand(command string, extraEnv []string, dir string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// cmd /C "..." lets Go escape the string, which mangles internal quotes.
		// Writing to a temp .bat file and running that bypasses all quoting issues.
		f, err := os.CreateTemp("", "zeus-hook-*.bat")
		if err != nil {
			return fmt.Errorf("create temp batch: %w", err)
		}
		defer os.Remove(f.Name())
		if _, err := fmt.Fprintf(f, "@echo off\r\n%s\r\n", command); err != nil {
			f.Close()
			return fmt.Errorf("write temp batch: %w", err)
		}
		f.Close()
		cmd = exec.Command("cmd", "/C", f.Name())
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Env = append(os.Environ(), extraEnv...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
