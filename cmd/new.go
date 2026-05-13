package cmd

import "github.com/spf13/cobra"

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Generate a new profile template with Arma 3 defaults (shortcut for 'profile new')",
	Args:  cobra.ExactArgs(1),
	Example: "  zeus new my-server\n" +
		"  zeus new my-server --json\n" +
		"  zeus new my-server --yaml > server.yaml",
	Run: profileNewCmd.Run,
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().Bool("yaml", false, "output as YAML")
	newCmd.Flags().Bool("toml", false, "output as TOML")
	newCmd.Flags().Bool("json", false, "output as JSON (default)")
	newCmd.MarkFlagsMutuallyExclusive("yaml", "toml", "json")
}
