//go:build !windows

package missions

import "os"

// replaceSymlink atomically replaces targetDir with a symlink pointing to
// sourcePath. On POSIX, rename(2) atomically replaces an existing symlink
// so there is no window where the target is absent.
func replaceSymlink(sourcePath, targetDir string) error {
	tmp := targetDir + ".new"
	os.Remove(tmp) // clean up any leftover from a prior interrupted run
	if err := os.Symlink(sourcePath, tmp); err != nil {
		return err
	}
	if err := os.Rename(tmp, targetDir); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}
