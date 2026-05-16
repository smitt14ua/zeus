package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

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
	name := args[0]

	mgr := process.Manager{}
	running, err := mgr.List()
	if err != nil {
		fatal(err)
	}
	for _, e := range running {
		if e.Name == name {
			fatalf("profile %q is already running (PID %d)", name, e.PID)
		}
	}

	repo := storage.ProfileRepository{}
	p, err := repo.Get(name)
	if err != nil {
		fatal(err)
	}

	runner := process.Runner{}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		command, err := runner.Command(p)
		if err != nil {
			fatal(err)
		}
		fmt.Println(command)
		return
	}

	if p.Hooks != nil {
		if err := profile.RunHooks(p.Hooks.PreProfileStart, p); err != nil {
			fatal(err)
		}
	}
	if p.Hooks != nil {
		if err := profile.RunHooks(p.Hooks.PreProfileRun, p); err != nil {
			fatal(err)
		}
	}
	if err := runner.Run(p); err != nil {
		fatal(err)
	}
	if p.Hooks != nil {
		if err := profile.RunHooks(p.Hooks.PostProfileRun, p); err != nil {
			fatal(err)
		}
		if err := profile.RunHooks(p.Hooks.PostProfileStart, p); err != nil {
			fatal(err)
		}
	}
}

var profileStopCmd = &cobra.Command{
	Use:     "stop <name>",
	Short:   "Stop a running profile",
	Args:    cobra.ExactArgs(1),
	Example: "  zeus profile stop my-server",
	Run:     runProfileStop,
}

func runProfileStop(cmd *cobra.Command, args []string) {
	name := args[0]

	repo := storage.ProfileRepository{}
	p, err := repo.Get(name)
	if err != nil {
		fatal(err)
	}

	mgr := process.Manager{}
	running, err := mgr.List()
	if err != nil {
		fatal(err)
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
		fatalf("profile %q is not running", name)
	}

	if p.Hooks != nil {
		if err := profile.RunHooks(p.Hooks.PreProfileStop, p); err != nil {
			fatal(err)
		}
	}
	if err := mgr.Kill(name); err != nil {
		fatal(err)
	}

	info("Waiting for process %d to terminate...", pid)
	if !mgr.WaitGone(pid, name, 30*time.Second) {
		fmt.Fprintf(os.Stderr, "process %d did not terminate within 30s\n", pid)
		os.Exit(1)
	}
	fmt.Printf("Profile %q stopped.\n", name)
	if p.Hooks != nil {
		if err := profile.RunHooks(p.Hooks.PostProfileStop, p); err != nil {
			fatal(err)
		}
	}
}

func init() {
	profileCmd.AddCommand(profileStartCmd)
	profileStartCmd.Flags().BoolP("dry-run", "n", false, "print the generated command without executing")

	profileCmd.AddCommand(profileStopCmd)
}
