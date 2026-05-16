package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
		if err := execProfileList(cmd.Context(), os.Stdout); err != nil {
			fatal(err)
		}
	},
}

func execProfileList(ctx context.Context, w io.Writer) error {
	repo := storage.ProfileRepository{}
	profiles, err := repo.List()
	if err != nil {
		return err
	}

	mgr := process.Manager{}
	running, _ := mgr.List()
	runningSet := make(map[string]bool, len(running))
	for _, e := range running {
		runningSet[e.Name] = true
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS")
	for _, p := range profiles {
		status := "-"
		if runningSet[p.Name] {
			status = "running"
		}
		fmt.Fprintf(tw, "%s\t%s\n", p.Name, status)
	}
	return tw.Flush()
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
		asYAML, _ := cmd.Flags().GetBool("yaml")
		asJSON, _ := cmd.Flags().GetBool("json")
		asTOML, _ := cmd.Flags().GetBool("toml")
		format := "console"
		if asYAML {
			format = "yaml"
		} else if asJSON {
			format = "json"
		} else if asTOML {
			format = "toml"
		}
		if err := execProfileInfo(cmd.Context(), args[0], format, os.Stdout); err != nil {
			fatal(err)
		}
	},
}

func execProfileInfo(ctx context.Context, name, format string, w io.Writer) error {
	repo := storage.ProfileRepository{}
	p, err := repo.Get(name)
	if err != nil {
		return err
	}
	switch format {
	case "yaml":
		data, err := yaml.Marshal(p)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, string(data))
		return err
	case "json":
		data, err := json.MarshalIndent(p, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w, string(data))
		return err
	case "toml":
		data, err := toml.Marshal(p)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, string(data))
		return err
	default:
		printProfileConsole(p, w)
		return nil
	}
}

func printProfileConsole(p profile.Profile, w io.Writer) {
	fmt.Fprintf(w, "Profile:     %s\n", p.Name)
	fmt.Fprintf(w, "Install Dir: %s\n", p.InstallDir)

	fmt.Fprintln(w, "\nServer Config")
	if p.Config.Hostname != nil {
		fmt.Fprintf(w, "  Hostname:    %s\n", *p.Config.Hostname)
	}
	if p.Config.MaxPlayers != nil {
		fmt.Fprintf(w, "  Max Players: %d\n", *p.Config.MaxPlayers)
	}
	if p.Config.BattlEye != nil {
		fmt.Fprintf(w, "  BattlEye:    %v\n", *p.Config.BattlEye)
	}
	if p.Config.ForcedDifficulty != nil {
		fmt.Fprintf(w, "  Difficulty:  %s\n", *p.Config.ForcedDifficulty)
	}
	if len(p.Config.Missions) > 0 {
		fmt.Fprintf(w, "  Missions:    %d\n", len(p.Config.Missions))
		for _, m := range p.Config.Missions {
			fmt.Fprintf(w, "    - %s (%s)\n", m.Template, m.Difficulty)
		}
	}

	fmt.Fprintln(w, "\nStartup Params")
	if p.Params.Port != nil {
		fmt.Fprintf(w, "  Port:       %d\n", *p.Params.Port)
	}
	if p.Params.LimitFPS != nil {
		fmt.Fprintf(w, "  FPS Limit:  %d\n", *p.Params.LimitFPS)
	}
	if p.Params.MaxMem != nil {
		fmt.Fprintf(w, "  Max Mem:    %d MiB\n", *p.Params.MaxMem)
	}
	if len(p.Params.Mod) > 0 {
		fmt.Fprintf(w, "  Mods:       %s\n", strings.Join(p.Params.Mod, ", "))
	}
	if len(p.Params.ServerMod) > 0 {
		fmt.Fprintf(w, "  Server Mods: %s\n", strings.Join(p.Params.ServerMod, ", "))
	}
}

var profileRmCmd = &cobra.Command{
	Use:     "rm <name>",
	Short:   "Delete a saved profile and its data",
	Args:    cobra.ExactArgs(1),
	Example: "  zeus profile rm my-server\n  zeus profile rm my-server --force",
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")
		if !force && stdinIsPipe() {
			fatalf("stdin is not a terminal; pass --force to delete without confirmation")
		}
		if err := execProfileRm(cmd.Context(), args[0], force, os.Stdin, os.Stderr); err != nil {
			fatal(err)
		}
	},
}

func execProfileRm(ctx context.Context, name string, force bool, in io.Reader, w io.Writer) error {
	repo := storage.ProfileRepository{}
	p, err := repo.Get(name)
	if err != nil {
		return err
	}

	mgr := process.Manager{}
	running, _ := mgr.List()
	for _, e := range running {
		if e.Name == name {
			return fmt.Errorf("profile %q is currently running (PID %d); stop it before deleting", name, e.PID)
		}
	}

	if !force {
		profileDir := filepath.Join(p.InstallDir, ".zeus", name)
		fmt.Fprintf(w, "The following will be permanently removed:\n")
		fmt.Fprintf(w, "  profile data : %s\n", profileDir)
		fmt.Fprintf(w, "  contents     : configs, mpmissions, keys, logs, and all other profile data\n")
		fmt.Fprintf(w, "\nType \"yes\" to confirm, anything else aborts: ")

		scanner := bufio.NewScanner(in)
		scanner.Scan()
		if strings.TrimSpace(scanner.Text()) != "yes" {
			fmt.Fprintln(w, "Aborted.")
			return nil
		}
	}

	profileDir := filepath.Join(p.InstallDir, ".zeus", name)
	if err := os.RemoveAll(profileDir); err != nil {
		return err
	}
	return repo.Delete(name)
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
