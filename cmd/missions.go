package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/smitt14ua/zeus/internal/missions"
	"github.com/smitt14ua/zeus/internal/process"
	"github.com/smitt14ua/zeus/internal/profile"
	"github.com/smitt14ua/zeus/internal/storage"
)

var missionsCmd = &cobra.Command{
	Use:   "missions",
	Short: "Manage mission files for a profile",
}

var missionsPullCmd = &cobra.Command{
	Use:   "pull <profile>",
	Short: "Sync .pbo mission files from the configured source into the profile mpmissions directory",
	Args:  cobra.ExactArgs(1),
	Example: `  zeus missions pull my-server
  zeus missions pull my-server --dry-run`,
	Run: runMissionsPull,
}

func runMissionsPull(cmd *cobra.Command, args []string) {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if err := execMissionsPull(cmd.Context(), args[0], dryRun, os.Stdout); err != nil {
		fatal(err)
	}
}

func execMissionsPull(ctx context.Context, name string, dryRun bool, w io.Writer) error {
	repo := storage.ProfileRepository{}
	p, err := repo.Get(name)
	if err != nil {
		return err
	}
	if p.MissionSource == nil {
		return fmt.Errorf("profile %q has no mission_source configured", name)
	}

	source := *p.MissionSource
	mode := source.Mode
	if mode == "" {
		mode = "copy"
	}

	serverRunning := false
	pm := process.Manager{}
	if entries, err := pm.List(); err == nil {
		for _, e := range entries {
			if e.Name == name {
				serverRunning = true
				break
			}
		}
	}

	if p.Hooks != nil {
		if len(p.Hooks.PrePullMissions) > 0 {
			fmt.Fprintf(w, "Running pre-pull hooks...\n")
		}
		if err := profile.RunHooks(p.Hooks.PrePullMissions, p); err != nil {
			return err
		}
	}

	targetDir := filepath.Join(p.InstallDir, ".zeus", p.Name, "mpmissions")

	header := fmt.Sprintf("Pulling missions for %q (%s, %s)", name, source.Driver, mode)
	if dryRun {
		header = "[dry-run] " + header
	}
	if serverRunning {
		header += " [server running — existing files locked]"
	}
	fmt.Fprintln(w, header)

	result, err := missions.Puller{}.Pull(source, targetDir, dryRun, serverRunning)
	if err != nil {
		return err
	}

	if mode == "symlink" {
		if result.Symlinked {
			fmt.Fprintf(w, "  → %s\n", source.Path)
		} else {
			fmt.Fprintf(w, "  = %s (symlink unchanged)\n", source.Path)
		}
	} else {
		sort.Strings(result.Added)
		sort.Strings(result.Updated)
		sort.Strings(result.Removed)
		sort.Strings(result.Skipped)
		sort.Strings(result.Locked)
		for _, f := range result.Added {
			fmt.Fprintf(w, "  + %s\n", f)
		}
		for _, f := range result.Updated {
			fmt.Fprintf(w, "  ~ %s\n", f)
		}
		for _, f := range result.Removed {
			fmt.Fprintf(w, "  - %s\n", f)
		}
		for _, f := range result.Skipped {
			fmt.Fprintf(w, "  = %s\n", f)
		}
		for _, f := range result.Locked {
			fmt.Fprintf(w, "  ! %s\n", f)
		}
		if len(result.Locked) > 0 {
			fmt.Fprintf(w, "%d file(s) not updated — stop the server then re-run to apply changes.\n", len(result.Locked))
		}
	}

	if dryRun {
		fmt.Fprintln(w, "Dry run complete. No changes made.")
		return nil
	}

	if mode == "symlink" {
		fmt.Fprintln(w, "Done.")
	} else {
		fmt.Fprintf(w, "Done. %d added, %d updated, %d removed, %d unchanged.\n",
			len(result.Added), len(result.Updated), len(result.Removed), len(result.Skipped))
	}

	if p.Hooks != nil {
		if len(p.Hooks.PostPullMissions) > 0 {
			fmt.Fprintf(w, "Running post-pull hooks...\n")
		}
		if err := profile.RunHooks(p.Hooks.PostPullMissions, p); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	missionsCmd.AddCommand(missionsPullCmd)
	missionsPullCmd.Flags().BoolP("dry-run", "n", false, "show what would change without making changes")
}
