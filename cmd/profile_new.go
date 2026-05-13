package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"

	"github.com/smitt14ua/zeus/internal/arma"
	"github.com/smitt14ua/zeus/internal/profile"
)

var profileNewCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Generate a new profile template with Arma 3 defaults",
	Args:  cobra.ExactArgs(1),
	Example: "  zeus profile new my-server\n" +
		"  zeus profile new my-server --json\n" +
		"  zeus profile new my-server --toml\n" +
		"  zeus profile new my-server --yaml > server.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		hostname := name + " server"
		p := profile.Profile{
			Name:   name,
			Config: arma.NewDefaultServerConfig(),
			Basic:  arma.NewDefaultBasicServerConfig(),
			RCon:   profile.DefaultRCon(name, 2302),
		}
		p.Config.Hostname = &hostname
		p.Config.Motd = []string{"Welcome to the " + name + " server!"}
		p.Config.Admins = []string{"00000000000000000"}

		asYAML, _ := cmd.Flags().GetBool("yaml")
		asTOML, _ := cmd.Flags().GetBool("toml")

		switch {
		case asYAML:
			data, err := yaml.Marshal(p)
			if err != nil {
				fatal(err)
			}
			out := string(data)
			out = strings.Replace(out, "install_dir:",
				"# Set absolute path to Arma 3 installation\ninstall_dir:", 1)
			out = strings.Replace(out, "params: {}\n",
				"params:\n"+
					"    # See available parameters on https://community.bistudio.com/wiki/Arma_3:_Startup_Parameters\n"+
					"    # Parameters must be in snake_case\n"+
					"    # port: 2302\n"+
					"    # load_mission_to_memory: true\n", 1)
			fmt.Print(out)
		case asTOML:
			var buf bytes.Buffer
			if err := toml.NewEncoder(&buf).Encode(p); err != nil {
				fatal(err)
			}
			out := buf.String()
			out = strings.Replace(out, "install_dir =",
				"# Set absolute path to Arma 3 installation\ninstall_dir =", 1)
			out = strings.Replace(out, "[params]\n",
				"[params]\n"+
					"  # Available parameters on https://community.bistudio.com/wiki/Arma_3:_Startup_Parameters\n"+
					"  # Parameters must be in snake_case\n"+
					"  # port = 2302\n"+
					"  # load_mission_to_memory = true\n", 1)
			fmt.Print(out)
		default:
			data, err := json.MarshalIndent(p, "", "  ")
			if err != nil {
				fatal(err)
			}
			fmt.Println(string(data))
		}
	},
}

func init() {
	profileCmd.AddCommand(profileNewCmd)
	profileNewCmd.Flags().Bool("yaml", false, "output as YAML")
	profileNewCmd.Flags().Bool("toml", false, "output as TOML")
	profileNewCmd.Flags().Bool("json", false, "output as JSON (default)")
	profileNewCmd.MarkFlagsMutuallyExclusive("yaml", "toml", "json")
}
