package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/smitt14ua/zeus/internal/profile"
)

type ProfileRepository struct {
	HomeDir string
}

func (r ProfileRepository) profilesDir() (string, error) {
	home := r.HomeDir
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(home, ".zeus", "profiles"), nil
}

func (r ProfileRepository) Save(p profile.Profile) error {
	if err := profile.ValidateName(p.Name); err != nil {
		return err
	}
	dir, err := r.profilesDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, p.Name+".json"), data, 0600)
}

func (r ProfileRepository) Delete(name string) error {
	if err := profile.ValidateName(name); err != nil {
		return err
	}
	dir, err := r.profilesDir()
	if err != nil {
		return err
	}
	deleted := 0
	for _, ext := range []string{".json", ".yaml", ".toml"} {
		err := os.Remove(filepath.Join(dir, name+ext))
		if err == nil {
			deleted++
			continue
		}
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if deleted == 0 {
		return fmt.Errorf("profile %q not found", name)
	}
	return nil
}

func (r ProfileRepository) List() ([]profile.Profile, error) {
	dir, err := r.profilesDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	loader := profile.ProfileLoader{}
	var profiles []profile.Profile
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".json") && !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".toml") {
			continue
		}
		p, err := loader.FromFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("loading profile %s: %w", name, err)
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

func (r ProfileRepository) Get(name string) (profile.Profile, error) {
	if err := profile.ValidateName(name); err != nil {
		return profile.Profile{}, err
	}
	dir, err := r.profilesDir()
	if err != nil {
		return profile.Profile{}, err
	}
	loader := profile.ProfileLoader{}
	for _, ext := range []string{".json", ".yaml", ".toml"} {
		p, err := loader.FromFile(filepath.Join(dir, name+ext))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		return p, err
	}
	return profile.Profile{}, fmt.Errorf("profile %q not found", name)
}

func (r ProfileRepository) Exists(name string) (bool, error) {
	if err := profile.ValidateName(name); err != nil {
		return false, err
	}
	dir, err := r.profilesDir()
	if err != nil {
		return false, err
	}
	for _, ext := range []string{".json", ".yaml", ".toml"} {
		_, err := os.Stat(filepath.Join(dir, name+ext))
		if err == nil {
			return true, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return false, err
		}
	}
	return false, nil
}
