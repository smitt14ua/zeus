package arma

import (
	"os"
	"path/filepath"
	"strings"
)

type Mod struct {
	Path    string
	Addons  []string
	Keys    []string
	Submods []Mod
}

func LoadMod(path string) (Mod, error) {
	m := Mod{Path: path}

	entries, err := os.ReadDir(path)
	if err != nil {
		return m, err
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		switch strings.ToLower(e.Name()) {
		case "addons":
			m.Addons, err = scanFiles(filepath.Join(path, e.Name()), ".pbo", ".ebo")
			if err != nil {
				return m, err
			}
		case "keys":
			m.Keys, err = scanFiles(filepath.Join(path, e.Name()), ".bikey")
			if err != nil {
				return m, err
			}
		case "optionals":
			m.Submods, err = loadSubmods(filepath.Join(path, e.Name()))
			if err != nil {
				return m, err
			}
		}
	}

	return m, nil
}

func loadSubmods(optionalsDir string) ([]Mod, error) {
	entries, err := os.ReadDir(optionalsDir)
	if err != nil {
		return nil, err
	}

	var submods []Mod
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "@") {
			continue
		}
		sub, err := LoadMod(filepath.Join(optionalsDir, e.Name()))
		if err != nil {
			return nil, err
		}
		submods = append(submods, sub)
	}
	return submods, nil
}

func scanFiles(dir string, exts ...string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	extSet := make(map[string]struct{}, len(exts))
	for _, ext := range exts {
		extSet[strings.ToLower(ext)] = struct{}{}
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if _, ok := extSet[strings.ToLower(filepath.Ext(e.Name()))]; ok {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	return files, nil
}
