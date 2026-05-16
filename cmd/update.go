package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"
)

const repoSlug = "smitt14ua/zeus"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update zeus to the latest release",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := execUpdate(cmd.Context(), version, os.Stdout); err != nil {
			fatal(err)
		}
	},
}

func execUpdate(ctx context.Context, currentVersion string, w io.Writer) error {
	if currentVersion == "dev" {
		return fmt.Errorf("cannot update a dev build — install a released version first:\n  go install github.com/smitt14ua/zeus@latest")
	}

	fmt.Fprintf(w, "Current version: %s\nChecking for updates...\n", currentVersion)

	latest, found, err := selfupdate.DetectLatest(ctx, selfupdate.ParseSlug(repoSlug))
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}
	if !found {
		return fmt.Errorf("no releases found for %s", repoSlug)
	}

	if latest.LessOrEqual(currentVersion) {
		fmt.Fprintf(w, "Already up to date (%s).\n", currentVersion)
		return nil
	}

	fmt.Fprintf(w, "Update available: %s → %s\n", currentVersion, latest.Version())

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolving executable path: %w", err)
	}

	if err := selfupdate.UpdateTo(ctx, latest.AssetURL, latest.AssetName, exe); err != nil {
		return fmt.Errorf("applying update: %w", err)
	}

	fmt.Fprintf(w, "Updated to %s. Restart zeus to use the new version.\n", latest.Version())
	return nil
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
