package profile

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// validNameRe matches safe profile name characters.
// Names must start with a letter or digit, contain only letters, digits,
// underscores, hyphens, or dots, and be 1–64 characters long.
var validNameRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$`)

// windowsReserved lists Windows reserved device name stems (upper-cased).
// These are rejected on all platforms so profiles remain portable.
var windowsReserved = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true,
	"COM5": true, "COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true,
	"LPT5": true, "LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

func validProfileName(name string) error {
	if !validNameRe.MatchString(name) {
		return fmt.Errorf("invalid profile name %q: must start with a letter or digit and contain only [A-Za-z0-9_.-] (1–64 chars)", name)
	}
	// Block Windows reserved device names regardless of platform for portability.
	stem := strings.ToUpper(strings.SplitN(name, ".", 2)[0])
	if windowsReserved[stem] {
		return fmt.Errorf("invalid profile name %q: reserved OS filename", name)
	}
	return nil
}

// ValidateName returns an error if name is not a safe profile identifier.
// Use this when only the name is available (e.g. before the full profile is built).
func ValidateName(name string) error {
	return validProfileName(name)
}

// Validate returns an error if the profile cannot be safely persisted or launched.
// It checks that the name is valid and that install_dir is a non-empty absolute path.
func (p Profile) Validate() error {
	if err := validProfileName(p.Name); err != nil {
		return err
	}
	if p.InstallDir == "" {
		return fmt.Errorf("profile %q: install_dir must not be empty", p.Name)
	}
	if !filepath.IsAbs(p.InstallDir) {
		return fmt.Errorf("profile %q: install_dir must be an absolute path, got %q", p.Name, p.InstallDir)
	}
	return nil
}
