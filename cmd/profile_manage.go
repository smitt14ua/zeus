package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"

	"github.com/smitt14ua/zeus/internal/process"
	"github.com/smitt14ua/zeus/internal/profile"
	"github.com/smitt14ua/zeus/internal/storage"
)

var profileLsCmd = &cobra.Command{
	Use:     "ls",
	Short:   "List saved profiles",
	Args:    cobra.NoArgs,
	Example: "  zeus profile ls",
	Run: func(cmd *cobra.Command, args []string) {
		repo := storage.ProfileRepository{}
		profiles, err := repo.List()
		if err != nil {
			fatal(err)
		}

		mgr := process.Manager{}
		running, _ := mgr.List()
		runningSet := make(map[string]bool, len(running))
		for _, e := range running {
			runningSet[e.Name] = true
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSTATUS")
		for _, p := range profiles {
			status := "-"
			if runningSet[p.Name] {
				status = "running"
			}
			fmt.Fprintf(w, "%s\t%s\n", p.Name, status)
		}
		w.Flush()
	},
}

var profileInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Print detailed info about a profile",
	Args:  cobra.ExactArgs(1),
	Example: "  zeus profile info my-server\n" +
		"  zeus profile info my-server --yaml\n" +
		"  zeus profile info my-server --json\n" +
		"  zeus profile info my-server --toml",
	Run: func(cmd *cobra.Command, args []string) {
		repo := storage.ProfileRepository{}
		p, err := repo.Get(args[0])
		if err != nil {
			fatal(err)
		}

		asYAML, _ := cmd.Flags().GetBool("yaml")
		asJSON, _ := cmd.Flags().GetBool("json")
		asTOML, _ := cmd.Flags().GetBool("toml")

		switch {
		case asYAML:
			data, err := yaml.Marshal(p)
			if err != nil {
				fatal(err)
			}
			fmt.Print(string(data))
		case asJSON:
			data, err := json.MarshalIndent(p, "", "  ")
			if err != nil {
				fatal(err)
			}
			fmt.Println(string(data))
		case asTOML:
			data, err := toml.Marshal(p)
			if err != nil {
				fatal(err)
			}
			fmt.Print(string(data))
		default:
			printProfileConsole(p)
		}
	},
}

func printProfileConsole(p profile.Profile) {
	fmt.Printf("Profile:     %s\n", p.Name)
	fmt.Printf("Install Dir: %s\n", p.InstallDir)

	fmt.Println("\nServer Config")
	if p.Config.Hostname != nil {
		fmt.Printf("  Hostname:    %s\n", *p.Config.Hostname)
	}
	if p.Config.MaxPlayers != nil {
		fmt.Printf("  Max Players: %d\n", *p.Config.MaxPlayers)
	}
	if p.Config.BattlEye != nil {
		fmt.Printf("  BattlEye:    %v\n", *p.Config.BattlEye)
	}
	if p.Config.ForcedDifficulty != nil {
		fmt.Printf("  Difficulty:  %s\n", *p.Config.ForcedDifficulty)
	}
	if len(p.Config.Missions) > 0 {
		fmt.Printf("  Missions:    %d\n", len(p.Config.Missions))
		for _, m := range p.Config.Missions {
			fmt.Printf("    - %s (%s)\n", m.Template, m.Difficulty)
		}
	}

	fmt.Println("\nStartup Params")
	if p.Params.Port != nil {
		fmt.Printf("  Port:       %d\n", *p.Params.Port)
	}
	if p.Params.LimitFPS != nil {
		fmt.Printf("  FPS Limit:  %d\n", *p.Params.LimitFPS)
	}
	if p.Params.MaxMem != nil {
		fmt.Printf("  Max Mem:    %d MiB\n", *p.Params.MaxMem)
	}
	if len(p.Params.Mod) > 0 {
		fmt.Printf("  Mods:       %s\n", strings.Join(p.Params.Mod, ", "))
	}
	if len(p.Params.ServerMod) > 0 {
		fmt.Printf("  Server Mods: %s\n", strings.Join(p.Params.ServerMod, ", "))
	}
}

var profileRmCmd = &cobra.Command{
	Use:     "rm <name>",
	Short:   "Delete a saved profile and its data",
	Args:    cobra.ExactArgs(1),
	Example: "  zeus profile rm my-server\n  zeus profile rm my-server --force",
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		repo := storage.ProfileRepository{}
		p, err := repo.Get(name)
		if err != nil {
			fatal(err)
		}

		mgr := process.Manager{}
		running, _ := mgr.List()
		for _, e := range running {
			if e.Name == name {
				fatalf("profile %q is currently running (PID %d); stop it before deleting", name, e.PID)
			}
		}

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			if stdinIsPipe() {
				fatalf("stdin is not a terminal; pass --force to delete without confirmation")
			}
			profileDir := filepath.Join(p.InstallDir, ".zeus", name)
			fmt.Fprintf(os.Stderr, "The following will be permanently removed:\n")
			fmt.Fprintf(os.Stderr, "  profile data : %s\n", profileDir)
			fmt.Fprintf(os.Stderr, "  contents     : configs, mpmissions, keys, logs, and all other profile data\n")
			fmt.Fprintf(os.Stderr, "\nType \"yes\" to confirm, anything else aborts: ")

			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			if strings.TrimSpace(scanner.Text()) != "yes" {
				fmt.Fprintln(os.Stderr, "Aborted.")
				return
			}
		}

		profileDir := filepath.Join(p.InstallDir, ".zeus", name)
		if err := os.RemoveAll(profileDir); err != nil {
			fatal(err)
		}
		if err := repo.Delete(name); err != nil {
			fatal(err)
		}
	},
}

func init() {
	profileCmd.AddCommand(profileLsCmd)

	profileCmd.AddCommand(profileInfoCmd)
	profileInfoCmd.Flags().Bool("yaml", false, "output as YAML")
	profileInfoCmd.Flags().Bool("json", false, "output as JSON")
	profileInfoCmd.Flags().Bool("toml", false, "output as TOML")
	profileInfoCmd.MarkFlagsMutuallyExclusive("yaml", "json", "toml")

	profileCmd.AddCommand(profileRmCmd)
	profileRmCmd.Flags().BoolP("force", "f", false, "delete without confirmation prompt")
}
