package missions

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const etagSidecar = ".zeus-s3-etags"

// loadETags reads the ETag sidecar file from dir.
// Returns an empty map (not an error) if the file does not exist.
func loadETags(dir string) (map[string]string, error) {
	path := filepath.Join(dir, etagSidecar)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	var etags map[string]string
	if err := json.Unmarshal(data, &etags); err != nil {
		return nil, err
	}
	return etags, nil
}

// saveETags writes the ETag sidecar file atomically (temp file + rename).
func saveETags(dir string, etags map[string]string) error {
	path := filepath.Join(dir, etagSidecar)
	data, err := json.Marshal(etags)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}
