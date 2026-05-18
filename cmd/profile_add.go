package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/smitt14ua/zeus/internal/arma"
	"github.com/smitt14ua/zeus/internal/profile"
	"github.com/smitt14ua/zeus/internal/storage"
)

var profileAddCmd = &cobra.Command{
	Use:   "add [profile.yaml|profile.toml|profile.json]",
	Short: "Add a profile from a YAML, TOML, or JSON file or stdin",
	Args:  cobra.MaximumNArgs(1),
	Example: `  zeus profile add server.yaml
  zeus profile add server.toml
  zeus profile add server.json
  zeus profile add server.yaml --name staging
  cat server.json | zeus profile add`,
	Run: runProfileAdd,
}

func runProfileAdd(cmd *cobra.Command, args []string) {
	nameOverride, _ := cmd.Flags().GetString("name")
	copyKeys, _ := cmd.Flags().GetBool("copy-keys")
	force, _ := cmd.Flags().GetBool("force")

	var r io.Reader
	var format string
	if len(args) == 1 {
		f, err := os.Open(args[0])
		if err != nil {
			fatal(err)
		}
		defer f.Close()
		r = f
		format = profile.FormatFromPath(args[0])
	} else if stdinIsPipe() {
		r = os.Stdin
		format = "json"
		force = true
	} else {
		fatalf("provide a YAML, TOML, or JSON file path or pipe content via stdin")
	}

	if err := execProfileAdd(cmd.Context(), r, format, nameOverride, copyKeys, force, os.Stderr); err != nil {
		fatal(err)
	}
}

func execProfileAdd(ctx context.Context, r io.Reader, format, nameOverride string, copyKeys, force bool, w io.Writer) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	loader := profile.ProfileLoader{}
	p, err := loader.FromBytesFormat(data, format)
	if err != nil {
		return err
	}

	if nameOverride != "" {
		p.Name = nameOverride
	}

	repo := storage.ProfileRepository{}
	if exists, _ := repo.Exists(p.Name); exists {
		fmt.Fprintf(w, "profile %q already exists, updating.\n", p.Name)
	}

	if err := processProfileMods(&p, copyKeys, force); err != nil {
		return err
	}
	if err := repo.Save(p); err != nil {
		return err
	}

	writer := profile.ProfileWriter{}
	if err := writer.Write(p); err != nil {
		return err
	}
	if p.Hooks != nil {
		if len(p.Hooks.PostProfileAdd) > 0 {
			fmt.Fprintf(w, "Running post-add hooks...\n")
		}
		if err := profile.RunHooks(p.Hooks.PostProfileAdd, p); err != nil {
			return err
		}
	}
	fmt.Fprintf(w, "profile %q saved. To start it: zeus start %s\n", p.Name, p.Name)
	return nil
}

func processProfileMods(p *profile.Profile, copyAll bool, force bool) error {
	if len(p.Params.Mod) == 0 {
		return nil
	}

	keysDir := filepath.Join(p.InstallDir, ".zeus", p.Name, "keys")
	var scanner *bufio.Scanner
	if !force {
		scanner = bufio.NewScanner(os.Stdin)
	}
	original := make([]string, len(p.Params.Mod))
	for i, m := range p.Params.Mod {
		if !filepath.IsAbs(m) && p.InstallDir != "" {
			original[i] = filepath.Join(p.InstallDir, m)
		} else {
			original[i] = m
		}
	}

	knownKeys := collectAllModKeys(original)

	for _, modPath := range original {
		mod, err := arma.LoadMod(modPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot load mod %q: %v (skipping)\n", modPath, err)
			continue
		}

		modName := filepath.Base(modPath)

		for _, sub := range mod.Submods {
			subName := filepath.Base(sub.Path)
			if !force && promptYN(scanner, fmt.Sprintf("Include submod %s (from %s)?", subName, modName)) {
				p.Params.Mod = append(p.Params.Mod, sub.Path)
				if len(sub.Keys) > 0 {
					if copyAll || promptYn(scanner, fmt.Sprintf("Copy %d key(s) from %s to profile keys dir?", len(sub.Keys), subName)) {
						if err := copyKeysTo(sub.Keys, keysDir); err != nil {
							fmt.Fprintf(os.Stderr, "copy keys failed: %v\n", err)
						}
					}
				}
			}
		}

		if len(mod.Keys) > 0 {
			if copyAll || force || promptYn(scanner, fmt.Sprintf("Copy %d key(s) from %s to profile keys dir?", len(mod.Keys), modName)) {
				if err := copyKeysTo(mod.Keys, keysDir); err != nil {
					fmt.Fprintf(os.Stderr, "copy keys failed: %v\n", err)
				}
			}
		}
	}

	unknowns, err := findUnknownKeys(keysDir, knownKeys)
	if err != nil {
		return err
	}
	if len(unknowns) > 0 {
		info("Found %d unknown key(s) in keys dir (not from any listed mod):", len(unknowns))
		for _, u := range unknowns {
			info("  %s", u)
		}
		if !force && promptYN(scanner, "Delete unknown keys?") {
			for _, u := range unknowns {
				if err := os.Remove(filepath.Join(keysDir, u)); err != nil {
					fmt.Fprintf(os.Stderr, "could not delete key %q: %v\n", u, err)
				}
			}
		}
	}

	return nil
}

func collectAllModKeys(modPaths []string) map[string]struct{} {
	known := make(map[string]struct{})
	for _, path := range modPaths {
		mod, err := arma.LoadMod(path)
		if err != nil {
			continue
		}
		for _, k := range mod.Keys {
			known[filepath.Base(k)] = struct{}{}
		}
		for _, sub := range mod.Submods {
			for _, k := range sub.Keys {
				known[filepath.Base(k)] = struct{}{}
			}
		}
	}
	return known
}

func findUnknownKeys(keysDir string, known map[string]struct{}) ([]string, error) {
	entries, err := os.ReadDir(keysDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var unknown []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".bikey") {
			continue
		}
		if _, ok := known[e.Name()]; !ok {
			unknown = append(unknown, e.Name())
		}
	}
	return unknown, nil
}

func init() {
	profileCmd.AddCommand(profileAddCmd)
	profileAddCmd.Flags().String("name", "", "override profile name")
	profileAddCmd.Flags().BoolP("copy-keys", "k", false, "copy all mod keys without prompting")
	profileAddCmd.Flags().BoolP("force", "f", false, "skip all prompts using default answers")
}
