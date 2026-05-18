package missions

import (
	"fmt"
	"os"

	"github.com/smitt14ua/zeus/internal/arma"
)

// PathDriver syncs missions from a local filesystem path.
type PathDriver struct{}

func (d PathDriver) pull(source arma.MissionSource, targetDir string, dryRun, serverRunning bool) (Result, error) {
	if _, err := os.Stat(source.Path); os.IsNotExist(err) {
		return Result{}, fmt.Errorf("source path %q does not exist", source.Path)
	}

	mode := source.Mode
	if mode == "" {
		mode = "copy"
	}

	switch mode {
	case "copy":
		return d.pullCopy(source.Path, targetDir, dryRun, serverRunning)
	case "symlink":
		return d.pullSymlink(source.Path, targetDir, dryRun)
	default:
		return Result{}, fmt.Errorf("unknown mode %q", mode)
	}
}

func (d PathDriver) pullSymlink(sourcePath, targetDir string, dryRun bool) (Result, error) {
	info, err := os.Lstat(targetDir)
	if err != nil && !os.IsNotExist(err) {
		return Result{}, err
	}

	if err == nil {
		if info.Mode()&os.ModeSymlink == 0 {
			return Result{}, fmt.Errorf("target %q is a real directory; remove it manually before switching to symlink mode", targetDir)
		}
		current, err := os.Readlink(targetDir)
		if err != nil {
			return Result{}, err
		}
		if current == sourcePath {
			return Result{Symlinked: false}, nil
		}
	}

	if !dryRun {
		if err := replaceSymlink(sourcePath, targetDir); err != nil {
			return Result{}, err
		}
	}
	return Result{Symlinked: true}, nil
}

func (d PathDriver) pullCopy(sourcePath, targetDir string, dryRun, serverRunning bool) (Result, error) {
	var result Result

	srcFiles, err := scanPBOs(sourcePath)
	if err != nil {
		return Result{}, fmt.Errorf("scanning source: %w", err)
	}

	if !dryRun {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return Result{}, fmt.Errorf("creating target directory: %w", err)
		}
	}

	dstFiles, err := scanPBOs(targetDir)
	if err != nil {
		return Result{}, fmt.Errorf("scanning target: %w", err)
	}

	// Remove .pbo files in target not present in source (skip if server is running).
	for name := range dstFiles {
		if _, ok := srcFiles[name]; !ok {
			realName := dstFiles[name].Name()
			if serverRunning {
				result.Locked = append(result.Locked, realName)
			} else {
				result.Removed = append(result.Removed, realName)
				if !dryRun {
					if err := os.Remove(pboPath(targetDir, realName)); err != nil {
						return Result{}, err
					}
				}
			}
		}
	}

	// Add or update .pbo files from source.
	for name, srcInfo := range srcFiles {
		dstInfo, exists := dstFiles[name]
		realName := srcInfo.Name()
		switch {
		case !exists:
			result.Added = append(result.Added, realName)
			if !dryRun {
				if err := copyFile(pboPath(sourcePath, realName), pboPath(targetDir, realName)); err != nil {
					return Result{}, err
				}
			}
		case sameFile(srcInfo, dstInfo):
			result.Skipped = append(result.Skipped, realName)
		default:
			if serverRunning {
				result.Locked = append(result.Locked, realName)
			} else {
				result.Updated = append(result.Updated, realName)
				if !dryRun {
					if err := copyFile(pboPath(sourcePath, realName), pboPath(targetDir, realName)); err != nil {
						return Result{}, err
					}
				}
			}
		}
	}

	return result, nil
}
