package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/smitt14ua/zeus/internal/profile"
)

type Runner struct {
	HomeDir string
}

func (r Runner) homeDir() (string, error) {
	if r.HomeDir != "" {
		return r.HomeDir, nil
	}
	return os.UserHomeDir()
}

func (r Runner) prepareParams(p profile.Profile) (profile.Profile, error) {
	home, err := r.homeDir()
	if err != nil {
		return p, err
	}

	zeusDir := filepath.Join(".zeus", p.Name)

	if p.Params.MpMissions == nil || *p.Params.MpMissions == "" {
		mp := filepath.Join(zeusDir, "mpmissions")
		p.Params.MpMissions = &mp
	}

	p.Params.KeysFolder = append(p.Params.KeysFolder, filepath.Join(zeusDir, "keys"))
	p.Params.KeysFolder = append(p.Params.KeysFolder, filepath.Join(zeusDir, "optionalkeys"))

	cfg := filepath.Join(zeusDir, "configs", "basic.cfg")
	p.Params.Cfg = &cfg

	config := filepath.Join(zeusDir, "configs", "server.cfg")
	p.Params.Config = &config

	pid := filepath.Join(home, ".zeus", "running", p.Name+".pid")
	p.Params.Pid = &pid

	profiles := zeusDir
	p.Params.Profiles = &profiles

	name := p.Name
	p.Params.Name = &name

	p.Params.Mod = absModPaths(p.InstallDir, p.Params.Mod)
	p.Params.ServerMod = absModPaths(p.InstallDir, p.Params.ServerMod)

	return p, nil
}

func resolveExecutable(p profile.Profile) string {
	name := p.Executable
	if name == "" {
		name = "arma3server_x64"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
	}
	return filepath.Join(p.InstallDir, name)
}

func (r Runner) Command(p profile.Profile) (string, error) {
	prepared, err := r.prepareParams(p)
	if err != nil {
		return "", err
	}

	exe := resolveExecutable(prepared)
	args := buildArgs(prepared.Params)

	parts := make([]string, 0, len(args)+1)
	parts = append(parts, quoteIfNeeded(exe))
	for _, a := range args {
		parts = append(parts, a)
	}
	return strings.Join(parts, " "), nil
}

func (r Runner) Run(p profile.Profile) error {
	prepared, err := r.prepareParams(p)
	if err != nil {
		return err
	}

	home, err := r.homeDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(home, ".zeus", "running"), 0755); err != nil {
		return err
	}

	exe := resolveExecutable(prepared)
	args := buildArgs(prepared.Params)

	cmd := exec.Command(exe, args...)
	cmd.Dir = prepared.InstallDir
	return cmd.Start()
}

func buildArgs(params any) []string {
	var args []string
	v := reflect.ValueOf(params)
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		val := v.Field(i)

		argName := field.Tag.Get("arg")
		if argName == "" {
			continue
		}

		switch val.Kind() {
		case reflect.Pointer:
			if val.IsNil() {
				continue
			}
			elem := val.Elem()
			if elem.Kind() == reflect.Bool {
				if elem.Bool() {
					args = append(args, "-"+argName)
				}
			} else {
				args = append(args, fmt.Sprintf("-%s=%s", argName, quoteIfNeeded(fmt.Sprintf("%v", elem.Interface()))))
			}
		case reflect.Slice:
			if val.Len() == 0 {
				continue
			}
			parts := make([]string, val.Len())
			for j := 0; j < val.Len(); j++ {
				parts[j] = val.Index(j).String()
			}
			args = append(args, fmt.Sprintf("-%s=%s", argName, quoteIfNeeded(strings.Join(parts, ";"))))
		}
	}

	return args
}

func absModPaths(installDir string, paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		if !filepath.IsAbs(p) && installDir != "" {
			out[i] = filepath.Join(installDir, p)
		} else {
			out[i] = p
		}
	}
	return out
}

func quoteIfNeeded(s string) string {
	if strings.ContainsAny(s, " \t") {
		return `"` + s + `"`
	}
	return s
}
