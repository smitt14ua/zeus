package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"go.yaml.in/yaml/v4"
)

type ProfileLoader struct{}

// FromBytes parses JSON bytes into a Profile.
func (l ProfileLoader) FromBytes(data []byte) (Profile, error) {
	return l.fromJSON(data)
}

// FromBytesFormat parses bytes using the given format ("yaml", "toml", or "json").
func (l ProfileLoader) FromBytesFormat(data []byte, format string) (Profile, error) {
	switch strings.ToLower(format) {
	case "toml":
		return l.fromTOML(data)
	case "yaml":
		return l.fromYAML(data)
	default:
		return l.fromJSON(data)
	}
}

// FormatFromPath returns the format string ("yaml", "toml", "json") for a file path.
func FormatFromPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".toml":
		return "toml"
	case ".yaml", ".yml":
		return "yaml"
	default:
		return "json"
	}
}

// FromFile parses a profile file, detecting format from the .yaml, .toml, or .json extension.
func (l ProfileLoader) FromFile(path string) (Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Profile{}, err
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".toml":
		return l.fromTOML(data)
	case ".json":
		return l.fromJSON(data)
	default:
		return l.fromYAML(data)
	}
}

func (l ProfileLoader) fromYAML(data []byte) (Profile, error) {
	var p Profile
	if err := yaml.Unmarshal(data, &p); err != nil {
		return Profile{}, err
	}
	return p, nil
}

func (l ProfileLoader) fromTOML(data []byte) (Profile, error) {
	var p Profile
	if _, err := toml.Decode(string(data), &p); err != nil {
		return Profile{}, err
	}
	return p, nil
}

func (l ProfileLoader) fromJSON(data []byte) (Profile, error) {
	var p Profile
	if len(data) == 0 {
		return p, nil
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return Profile{}, err
	}
	return p, nil
}
