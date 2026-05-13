package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/smitt14ua/zeus/internal/missions"
	"github.com/smitt14ua/zeus/internal/process"
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
	name := args[0]
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	repo := storage.ProfileRepository{}
	p, err := repo.Get(name)
	if err != nil {
		fatal(err)
	}
	if p.MissionSource == nil {
		fatalf("profile %q has no mission_source configured", name)
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

	targetDir := filepath.Join(p.InstallDir, ".zeus", p.Name, "mpmissions")

	header := fmt.Sprintf("Pulling missions for %q (%s, %s)", name, source.Driver, mode)
	if dryRun {
		header = "[dry-run] " + header
	}
	if serverRunning {
		header += " [server running — existing files locked]"
	}
	fmt.Println(header)

	result, err := missions.Puller{}.Pull(source, targetDir, dryRun, serverRunning)
	if err != nil {
		fatal(err)
	}

	if mode == "symlink" {
		if result.Symlinked {
			fmt.Printf("  → %s\n", source.Path)
		} else {
			fmt.Printf("  = %s (symlink unchanged)\n", source.Path)
		}
	} else {
		sort.Strings(result.Added)
		sort.Strings(result.Updated)
		sort.Strings(result.Removed)
		sort.Strings(result.Skipped)
		sort.Strings(result.Locked)
		for _, f := range result.Added {
			fmt.Printf("  + %s\n", f)
		}
		for _, f := range result.Updated {
			fmt.Printf("  ~ %s\n", f)
		}
		for _, f := range result.Removed {
			fmt.Printf("  - %s\n", f)
		}
		for _, f := range result.Skipped {
			fmt.Printf("  = %s\n", f)
		}
		for _, f := range result.Locked {
			fmt.Printf("  ! %s\n", f)
		}
		if len(result.Locked) > 0 {
			fmt.Fprintf(os.Stderr, "%d file(s) not updated — stop the server then re-run to apply changes.\n", len(result.Locked))
		}
	}

	if dryRun {
		fmt.Println("Dry run complete. No changes made.")
		return
	}

	if mode == "symlink" {
		fmt.Println("Done.")
	} else {
		fmt.Printf("Done. %d added, %d updated, %d removed, %d unchanged.\n",
			len(result.Added), len(result.Updated), len(result.Removed), len(result.Skipped))
	}
}

func init() {
	missionsCmd.AddCommand(missionsPullCmd)
	missionsPullCmd.Flags().BoolP("dry-run", "n", false, "show what would change without making changes")
}
