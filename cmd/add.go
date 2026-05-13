package cmd

import "github.com/spf13/cobra"

var addCmd = &cobra.Command{
	Use:     "add [profile.yaml]",
	Short:   "Add a profile from a YAML file or stdin (shortcut for 'profile add')",
	Args:    cobra.MaximumNArgs(1),
	Example: "  zeus add server.yaml\n  zeus add server.yaml --name staging\n  cat server.yaml | zeus add",
	Run:     runProfileAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().String("name", "", "override profile name")
	addCmd.Flags().BoolP("copy-keys", "k", false, "copy all mod keys without prompting")
	addCmd.Flags().BoolP("force", "f", false, "skip all prompts using default answers")
}
