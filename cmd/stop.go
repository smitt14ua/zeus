package cmd

import "github.com/spf13/cobra"

var stopCmd = &cobra.Command{
	Use:     "stop <name>",
	Short:   "Stop a running profile (shortcut for 'profile stop')",
	Args:    cobra.ExactArgs(1),
	Example: "  zeus stop my-server",
	Run:     runProfileStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
