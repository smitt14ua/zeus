package cmd

import "github.com/spf13/cobra"

var startCmd = &cobra.Command{
	Use:     "start <name>",
	Short:   "Start a saved profile (shortcut for 'profile start')",
	Args:    cobra.ExactArgs(1),
	Example: "  zeus start my-server\n  zeus start my-server --dry-run",
	Run:     runProfileStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolP("dry-run", "n", false, "print the generated command without executing")
}
