package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage profiles",
}

func stdinIsPipe() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func promptYN(scanner *bufio.Scanner, question string) bool {
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", question)
	scanner.Scan()
	return strings.ToLower(strings.TrimSpace(scanner.Text())) == "y"
}

func promptYn(scanner *bufio.Scanner, question string) bool {
	fmt.Fprintf(os.Stderr, "%s [Y/n]: ", question)
	scanner.Scan()
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "" || answer == "y" || answer == "yes"
}

func copyKeysTo(keys []string, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	for _, src := range keys {
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(destDir, filepath.Base(src)), data, 0644); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(profileCmd)
}
