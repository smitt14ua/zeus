//go:build windows

package missions

import "os"

// replaceSymlink replaces targetDir with a symlink pointing to sourcePath.
// On Windows, MoveFileEx cannot replace an existing directory-typed reparse
// point, so we fall back to remove-then-create (non-atomic but correct).
func replaceSymlink(sourcePath, targetDir string) error {
	if err := os.Remove(targetDir); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Symlink(sourcePath, targetDir)
}
