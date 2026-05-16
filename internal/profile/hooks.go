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
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Env = append(os.Environ(), extraEnv...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
