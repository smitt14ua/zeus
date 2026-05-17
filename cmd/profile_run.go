package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/smitt14ua/zeus/internal/arma"
	"github.com/smitt14ua/zeus/internal/process"
	"github.com/smitt14ua/zeus/internal/profile"
	"github.com/smitt14ua/zeus/internal/storage"
)

var profileStartCmd = &cobra.Command{
	Use:     "start <name>",
	Short:   "Start a saved profile",
	Args:    cobra.ExactArgs(1),
	Example: "  zeus profile start my-server\n  zeus profile start my-server --dry-run",
	Run:     runProfileStart,
}

func runProfileStart(cmd *cobra.Command, args []string) {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	startTimeout, _ := cmd.Flags().GetDuration("start-timeout")
	if err := execProfileStart(cmd.Context(), args[0], dryRun, startTimeout, os.Stdout); err != nil {
		fatal(err)
	}
}

const startStabilityWindow = 3 * time.Second

func execProfileStart(ctx context.Context, name string, dryRun bool, startTimeout time.Duration, w io.Writer) error {
	mgr := process.Manager{}
	running, err := mgr.List()
	if err != nil {
		return err
	}
	for _, e := range running {
		if e.Name == name {
			return fmt.Errorf("profile %q is already running (PID %d)", name, e.PID)
		}
	}

	repo := storage.ProfileRepository{}
	p, err := repo.Get(name)
	if err != nil {
		return err
	}

	wantPort := effectivePort(p)
	for _, e := range running {
		other, err := repo.Get(e.Name)
		if err != nil {
			continue // stale PID file — skip silently
		}
		if effectivePort(other) == wantPort {
			return fmt.Errorf("profile %q is already running on port %d (PID %d)", e.Name, wantPort, e.PID)
		}
	}

	runner := process.Runner{}

	if dryRun {
		command, err := runner.Command(p)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w, command)
		return err
	}

	if p.Hooks != nil {
		if err := profile.RunHooks(p.Hooks.PreProfileStart, p); err != nil {
			return err
		}
		if err := profile.RunHooks(p.Hooks.PreProfileRun, p); err != nil {
			return err
		}
	}
	directPID, err := runner.Run(p)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "Waiting for profile %q to start (timeout %s)...\n", name, startTimeout)
	pid, err := mgr.WaitReady(name, directPID, startTimeout, startStabilityWindow)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "Profile %q started (PID %d).\n", name, pid)
	if p.Hooks != nil {
		if err := profile.RunHooks(p.Hooks.PostProfileRun, p); err != nil {
			return err
		}
		if err := profile.RunHooks(p.Hooks.PostProfileStart, p); err != nil {
			return err
		}
	}
	return nil
}

var profileStopCmd = &cobra.Command{
	Use:     "stop <name>",
	Short:   "Stop a running profile",
	Args:    cobra.ExactArgs(1),
	Example: "  zeus profile stop my-server",
	Run:     runProfileStop,
}

func runProfileStop(cmd *cobra.Command, args []string) {
	if err := execProfileStop(cmd.Context(), args[0], os.Stdout); err != nil {
		fatal(err)
	}
}

func execProfileStop(ctx context.Context, name string, w io.Writer) error {
	repo := storage.ProfileRepository{}
	p, err := repo.Get(name)
	if err != nil {
		return err
	}

	mgr := process.Manager{}
	running, err := mgr.List()
	if err != nil {
		return err
	}

	var pid int
	found := false
	for _, e := range running {
		if e.Name == name {
			pid = e.PID
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("profile %q is not running", name)
	}

	if p.Hooks != nil {
		if err := profile.RunHooks(p.Hooks.PreProfileStop, p); err != nil {
			return err
		}
	}
	if err := mgr.Kill(name); err != nil {
		return err
	}

	fmt.Fprintf(w, "Waiting for process %d to terminate...\n", pid)
	if !mgr.WaitGone(pid, name, 30*time.Second) {
		return fmt.Errorf("process %d did not terminate within 30s", pid)
	}
	fmt.Fprintf(w, "Profile %q stopped.\n", name)
	if p.Hooks != nil {
		if err := profile.RunHooks(p.Hooks.PostProfileStop, p); err != nil {
			return err
		}
	}
	return nil
}

// effectivePort returns the game port for a profile, falling back to the
// Arma 3 default when the profile does not specify one explicitly.
func effectivePort(p profile.Profile) uint16 {
	if p.Params.Port != nil {
		return *p.Params.Port
	}
	return arma.DefaultPort
}

func init() {
	profileCmd.AddCommand(profileStartCmd)
	profileStartCmd.Flags().BoolP("dry-run", "n", false, "print the generated command without executing")
	profileStartCmd.Flags().Duration("start-timeout", 60*time.Second, "how long to wait for the server PID file to appear")

	profileCmd.AddCommand(profileStopCmd)
}
