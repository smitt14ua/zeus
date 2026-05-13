package missions

import (
	"fmt"

	"github.com/smitt14ua/zeus/internal/arma"
)

// Result holds the outcome of a Pull operation.
type Result struct {
	Added     []string
	Updated   []string
	Removed   []string
	Skipped   []string
	Locked    []string // existing files skipped because the server is running
	Symlinked bool     // true when a new symlink was created or replaced
}

// Puller dispatches mission sync operations by driver.
type Puller struct{}

// Pull syncs missions from source into targetDir.
// When dryRun is true, the diff is computed but no filesystem changes are made.
// When serverRunning is true, existing files are not modified or removed — only new files are added.
func (p Puller) Pull(source arma.MissionSource, targetDir string, dryRun, serverRunning bool) (Result, error) {
	switch source.Driver {
	case "path":
		return PathDriver{}.pull(source, targetDir, dryRun, serverRunning)
	case "s3":
		return S3Driver{}.pull(source, targetDir, dryRun, serverRunning)
	default:
		return Result{}, fmt.Errorf("unsupported driver %q", source.Driver)
	}
}
