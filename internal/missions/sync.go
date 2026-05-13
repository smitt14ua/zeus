package missions

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// scanPBOs returns a map of filename → FileInfo for all .pbo files in dir.
// Non-.pbo files and subdirectories are ignored.
func scanPBOs(dir string) (map[string]os.FileInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]os.FileInfo{}, nil
		}
		return nil, err
	}
	result := make(map[string]os.FileInfo)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".pbo") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return nil, err
		}
		result[strings.ToLower(e.Name())] = info
	}
	return result, nil
}

// sameFile reports whether two FileInfo values have identical size and mtime.
func sameFile(a, b os.FileInfo) bool {
	return a.Size() == b.Size() && a.ModTime().Equal(b.ModTime())
}

// copyFile copies src to dst and preserves the source mtime so subsequent
// pulls can detect unchanged files via sameFile.
// Uses a temp-file-then-rename pattern to avoid corrupting dst if the copy fails.
func copyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp, err := os.CreateTemp(filepath.Dir(dst), ".pbo-tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := io.Copy(tmp, in); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("copying data: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Chtimes(tmpName, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("setting mtime: %w", err)
	}
	if err := os.Rename(tmpName, dst); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming to target: %w", err)
	}
	return nil
}

// pboPath joins dir and a .pbo filename.
func pboPath(dir, name string) string {
	return filepath.Join(dir, name)
}
