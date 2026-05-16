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
	// Resolve symlinks so the canonical path is used and os.ReadDir works on
	// all platforms (Windows junctions, Linux/macOS symlinks).
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	m := Mod{Path: path}

	entries, err := os.ReadDir(path)
	if err != nil {
		return m, err
	}

	for _, e := range entries {
		entryPath := filepath.Join(path, e.Name())
		info, statErr := os.Stat(entryPath) // Stat follows symlinks; e.IsDir() does not
		if statErr != nil || !info.IsDir() {
			continue
		}
		switch strings.ToLower(e.Name()) {
		case "addons":
			m.Addons, err = scanFiles(entryPath, ".pbo", ".ebo")
			if err != nil {
				return m, err
			}
		case "keys":
			m.Keys, err = scanFiles(entryPath, ".bikey")
			if err != nil {
				return m, err
			}
		case "optionals":
			m.Submods, err = loadSubmods(entryPath)
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
		if !strings.HasPrefix(e.Name(), "@") {
			continue
		}
		entryPath := filepath.Join(optionalsDir, e.Name())
		info, err := os.Stat(entryPath) // Stat follows symlinks
		if err != nil || !info.IsDir() {
			continue
		}
		sub, err := LoadMod(entryPath)
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
