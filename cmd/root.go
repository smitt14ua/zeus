package cmd

import (
	"os"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

// version is set at build time via -ldflags "-X github.com/smitt14ua/zeus/cmd.version=X.Y.Z"
// Falls back to the module version embedded by go install.
var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "zeus",
	Short: "Arma 3 dedicated server manager",
	Long: `zeus manages Arma 3 dedicated server profiles.

A profile bundles server configuration, startup parameters, and mod lists
into a single YAML file. zeus generates the required .cfg files, manages
the server process, and tracks running instances via PID files.

Examples:
  zeus add server.yaml          add or update a profile
  zeus start my-server          start a profile
  zeus stop my-server           stop a running server
  zeus profile ls               list all profiles

Report issues: https://github.com/smitt14ua/zeus/issues`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok &&
			info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = strings.TrimPrefix(info.Main.Version, "v")
		}
	}
	rootCmd.Version = version
	rootCmd.AddCommand(missionsCmd)
}
