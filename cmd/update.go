package cmd

import (
	"fmt"
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
		if version == "dev" {
			fatalf("cannot update a dev build — install a released version first:\n  go install github.com/smitt14ua/zeus@latest")
		}

		fmt.Printf("Current version: %s\nChecking for updates...\n", version)

		latest, found, err := selfupdate.DetectLatest(cmd.Context(), selfupdate.ParseSlug(repoSlug))
		if err != nil {
			fatalf("checking for updates: %v", err)
		}
		if !found {
			fatalf("no releases found for %s", repoSlug)
		}

		if latest.LessOrEqual(version) {
			fmt.Printf("Already up to date (%s).\n", version)
			return
		}

		fmt.Printf("Update available: %s → %s\n", version, latest.Version())

		exe, err := os.Executable()
		if err != nil {
			fatalf("resolving executable path: %v", err)
		}

		if err := selfupdate.UpdateTo(cmd.Context(), latest.AssetURL, latest.AssetName, exe); err != nil {
			fatalf("applying update: %v", err)
		}

		fmt.Printf("Updated to %s. Restart zeus to use the new version.\n", latest.Version())
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
