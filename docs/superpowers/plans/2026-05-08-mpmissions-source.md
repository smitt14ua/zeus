# mpmissions Source Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `mission_source` profile field and `zeus missions pull <profile>` command that syncs `.pbo` files from a local path into the profile's mpmissions directory (copy or symlink mode).

**Architecture:** New `internal/missions` package owns all sync logic; `internal/arma` gets the `MissionSource` domain struct; `cmd/missions.go` wires the CLI. No imports from `cmd/` into `internal/`. Follows existing `HomeDir`-pattern and cobra command structure.

**Tech Stack:** Go stdlib only (`os`, `io`, `path/filepath`). No new dependencies.

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `internal/arma/mission_source.go` | Create | `MissionSource` struct with YAML/TOML/JSON tags |
| `internal/profile/profile.go` | Modify | Add `MissionSource *arma.MissionSource` field |
| `internal/missions/puller.go` | Create | `Puller`, `Result`, driver dispatch |
| `internal/missions/sync.go` | Create | `scanPBOs`, `sameFile`, `copyFile` helpers |
| `internal/missions/path_driver.go` | Create | `PathDriver` — copy and symlink implementations |
| `internal/missions/puller_test.go` | Create | Full test suite for all modes |
| `cmd/missions.go` | Create | `missionsCmd`, `missionsPullCmd`, output formatting |
| `cmd/root.go` | Modify | Register `missionsCmd` |
| `conf/example.yaml` | Modify | Add `mission_source` example block |
| `conf/example.toml` | Modify | Add `[mission_source]` example block |

---

## Task 1: MissionSource struct + Profile field

**Files:**
- Create: `internal/arma/mission_source.go`
- Modify: `internal/profile/profile.go`

- [ ] **Step 1: Write a failing test for Profile YAML round-trip with MissionSource**

Add to `internal/profile/loader_test.go`:

```go
func TestProfileLoader_MissionSource_YAML(t *testing.T) {
    loader := ProfileLoader{}
    const yaml = `
name: srv
install_dir: /opt/arma3
mission_source:
  driver: path
  path: /opt/missions
  mode: copy
`
    p, err := loader.FromBytes([]byte(yaml))
    require.NoError(t, err)
    require.NotNil(t, p.MissionSource)
    assert.Equal(t, "path", p.MissionSource.Driver)
    assert.Equal(t, "/opt/missions", p.MissionSource.Path)
    assert.Equal(t, "copy", p.MissionSource.Mode)
}

func TestProfileLoader_MissionSource_Absent(t *testing.T) {
    loader := ProfileLoader{}
    p, err := loader.FromBytes([]byte("name: srv\ninstall_dir: /opt/arma3\n"))
    require.NoError(t, err)
    assert.Nil(t, p.MissionSource)
}
```

- [ ] **Step 2: Run to confirm failure**

```
go test ./internal/profile/... -run TestProfileLoader_MissionSource -v
```

Expected: `FAIL` — field does not exist yet.

- [ ] **Step 3: Create `internal/arma/mission_source.go`**

```go
package arma

// MissionSource describes where mission .pbo files are fetched from.
// Driver "path" is the only supported driver in v1.
type MissionSource struct {
	Driver string `json:"driver"         yaml:"driver"         toml:"driver"`
	Path   string `json:"path,omitempty" yaml:"path,omitempty" toml:"path,omitempty"`
	// Mode controls how files are placed: "copy" (default) or "symlink".
	// symlink replaces the entire mpmissions directory with a symlink to Path.
	Mode string `json:"mode,omitempty" yaml:"mode,omitempty" toml:"mode,omitempty"`
}
```

- [ ] **Step 4: Add `MissionSource` field to Profile**

In `internal/profile/profile.go`, change:

```go
type Profile struct {
	Name          string                  `json:"name"                 yaml:"name"                 toml:"name"`
	Executable    string                  `json:"executable,omitempty" yaml:"executable,omitempty" toml:"executable,omitempty"`
	InstallDir    string                  `json:"install_dir"          yaml:"install_dir"          toml:"install_dir"`
	Params        arma.StartupParams      `json:"params"               yaml:"params"               toml:"params"`
	Config        arma.ServerConfig       `json:"config"               yaml:"config"               toml:"config"`
	Basic         arma.BasicServerConfig  `json:"basic"                yaml:"basic"                toml:"basic"`
	MissionSource *arma.MissionSource     `json:"missionSource,omitempty" yaml:"mission_source,omitempty" toml:"mission_source,omitempty"`
}
```

- [ ] **Step 5: Run tests to confirm pass**

```
go test ./internal/profile/... -run TestProfileLoader_MissionSource -v
```

Expected: `PASS`.

- [ ] **Step 6: Run full test suite to confirm no regressions**

```
go test ./... 2>&1
```

Expected: all packages `ok`.

- [ ] **Step 7: Commit**

```
git add internal/arma/mission_source.go internal/profile/profile.go internal/profile/loader_test.go
git commit -m "feat: add MissionSource struct and Profile.MissionSource field"
```

---

## Task 2: `internal/missions` package — helpers and copy mode

**Files:**
- Create: `internal/missions/sync.go`
- Create: `internal/missions/puller.go`
- Create: `internal/missions/path_driver.go`
- Create: `internal/missions/puller_test.go`

- [ ] **Step 1: Write failing tests for copy mode**

Create `internal/missions/puller_test.go`:

```go
package missions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smitt14ua/zeus/internal/arma"
)

// mkpbo writes a .pbo file with given content to dir.
func mkpbo(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0644))
}

func TestPuller_UnknownDriver(t *testing.T) {
	_, err := Puller{}.Pull(arma.MissionSource{Driver: "ftp"}, t.TempDir(), false)
	assert.ErrorContains(t, err, `unsupported driver "ftp"`)
}

func TestPathDriver_Copy_SourceNotExist(t *testing.T) {
	src := arma.MissionSource{Driver: "path", Path: "/no/such/path", Mode: "copy"}
	_, err := Puller{}.Pull(src, t.TempDir(), false)
	assert.ErrorContains(t, err, "does not exist")
}

func TestPathDriver_Copy_AddNew(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	mkpbo(t, src, "op_cobra.pbo", "data")

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "copy"}, dst, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"op_cobra.pbo"}, result.Added)
	assert.Empty(t, result.Updated)
	assert.Empty(t, result.Removed)
	assert.FileExists(t, filepath.Join(dst, "op_cobra.pbo"))
}

func TestPathDriver_Copy_RemoveStale(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	mkpbo(t, dst, "old.pbo", "stale")

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "copy"}, dst, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"old.pbo"}, result.Removed)
	assert.NoFileExists(t, filepath.Join(dst, "old.pbo"))
}

func TestPathDriver_Copy_UpdateChanged(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	mkpbo(t, src, "op.pbo", "new content longer")
	mkpbo(t, dst, "op.pbo", "old")

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "copy"}, dst, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"op.pbo"}, result.Updated)
	assert.Empty(t, result.Skipped)
}

func TestPathDriver_Copy_SkipUnchanged(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	content := []byte("pbo data")
	srcPath := filepath.Join(src, "mission.pbo")
	dstPath := filepath.Join(dst, "mission.pbo")
	require.NoError(t, os.WriteFile(srcPath, content, 0644))
	require.NoError(t, os.WriteFile(dstPath, content, 0644))
	// Sync mtime so sameFile returns true.
	srcInfo, err := os.Stat(srcPath)
	require.NoError(t, err)
	require.NoError(t, os.Chtimes(dstPath, srcInfo.ModTime(), srcInfo.ModTime()))

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "copy"}, dst, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"mission.pbo"}, result.Skipped)
	assert.Empty(t, result.Updated)
}

func TestPathDriver_Copy_DryRun(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	mkpbo(t, src, "op.pbo", "data")

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "copy"}, dst, true)
	require.NoError(t, err)
	assert.Equal(t, []string{"op.pbo"}, result.Added)
	assert.NoFileExists(t, filepath.Join(dst, "op.pbo")) // no actual copy
}

func TestPathDriver_Copy_IgnoresNonPBO(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(src, "readme.txt"), []byte("text"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dst, "keep.txt"), []byte("keep"), 0644))

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "copy"}, dst, false)
	require.NoError(t, err)
	assert.Empty(t, result.Added)
	assert.Empty(t, result.Removed)
	assert.FileExists(t, filepath.Join(dst, "keep.txt")) // non-.pbo left alone
}

func TestPathDriver_Copy_DefaultModeCopy(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	mkpbo(t, src, "op.pbo", "data")

	// Mode empty → defaults to copy
	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: ""}, dst, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"op.pbo"}, result.Added)
	assert.FileExists(t, filepath.Join(dst, "op.pbo"))
}
```

- [ ] **Step 2: Run to confirm failure**

```
go test ./internal/missions/... -v
```

Expected: `FAIL` — package does not exist.

- [ ] **Step 3: Create `internal/missions/sync.go`**

```go
package missions

import (
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
		result[e.Name()] = info
	}
	return result, nil
}

// sameFile reports whether two FileInfo values have identical size and mtime.
func sameFile(a, b os.FileInfo) bool {
	return a.Size() == b.Size() && a.ModTime().Equal(b.ModTime())
}

// copyFile copies src to dst and preserves the source mtime so subsequent
// pulls can detect unchanged files via sameFile.
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

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}

	return os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
}

// pboPath joins dir and a .pbo filename.
func pboPath(dir, name string) string {
	return filepath.Join(dir, name)
}
```

- [ ] **Step 4: Create `internal/missions/puller.go`**

```go
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
	Symlinked bool // true when a new symlink was created or replaced
}

// Puller dispatches mission sync operations by driver.
type Puller struct{}

// Pull syncs missions from source into targetDir.
// When dryRun is true, the diff is computed but no filesystem changes are made.
func (p Puller) Pull(source arma.MissionSource, targetDir string, dryRun bool) (Result, error) {
	switch source.Driver {
	case "path":
		return PathDriver{}.pull(source, targetDir, dryRun)
	default:
		return Result{}, fmt.Errorf("unsupported driver %q", source.Driver)
	}
}
```

- [ ] **Step 5: Create `internal/missions/path_driver.go` (copy mode only for now)**

```go
package missions

import (
	"fmt"
	"os"

	"github.com/smitt14ua/zeus/internal/arma"
)

// PathDriver syncs missions from a local filesystem path.
type PathDriver struct{}

func (d PathDriver) pull(source arma.MissionSource, targetDir string, dryRun bool) (Result, error) {
	if _, err := os.Stat(source.Path); os.IsNotExist(err) {
		return Result{}, fmt.Errorf("source path %q does not exist", source.Path)
	}

	mode := source.Mode
	if mode == "" {
		mode = "copy"
	}

	switch mode {
	case "symlink":
		return Result{}, fmt.Errorf("symlink mode not yet implemented")
	default:
		return d.pullCopy(source.Path, targetDir, dryRun)
	}
}

func (d PathDriver) pullCopy(sourcePath, targetDir string, dryRun bool) (Result, error) {
	var result Result

	srcFiles, err := scanPBOs(sourcePath)
	if err != nil {
		return Result{}, fmt.Errorf("scanning source: %w", err)
	}

	if !dryRun {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return Result{}, err
		}
	}

	dstFiles, err := scanPBOs(targetDir)
	if err != nil {
		return Result{}, fmt.Errorf("scanning target: %w", err)
	}

	// Remove .pbo files in target not present in source.
	for name := range dstFiles {
		if _, ok := srcFiles[name]; !ok {
			result.Removed = append(result.Removed, name)
			if !dryRun {
				if err := os.Remove(pboPath(targetDir, name)); err != nil {
					return Result{}, err
				}
			}
		}
	}

	// Add or update .pbo files from source.
	for name, srcInfo := range srcFiles {
		dstInfo, exists := dstFiles[name]
		switch {
		case !exists:
			result.Added = append(result.Added, name)
			if !dryRun {
				if err := copyFile(pboPath(sourcePath, name), pboPath(targetDir, name)); err != nil {
					return Result{}, err
				}
			}
		case sameFile(srcInfo, dstInfo):
			result.Skipped = append(result.Skipped, name)
		default:
			result.Updated = append(result.Updated, name)
			if !dryRun {
				if err := copyFile(pboPath(sourcePath, name), pboPath(targetDir, name)); err != nil {
					return Result{}, err
				}
			}
		}
	}

	return result, nil
}
```

- [ ] **Step 6: Run copy-mode tests**

```
go test ./internal/missions/... -run "TestPuller_UnknownDriver|TestPathDriver_Copy" -v
```

Expected: all `PASS`.

- [ ] **Step 7: Commit**

```
git add internal/missions/ internal/arma/mission_source.go internal/profile/profile.go internal/profile/loader_test.go
git commit -m "feat: add internal/missions package with path driver copy mode"
```

---

## Task 3: Symlink mode

**Files:**
- Modify: `internal/missions/path_driver.go`
- Modify: `internal/missions/puller_test.go`

- [ ] **Step 1: Write failing symlink tests**

Append to `internal/missions/puller_test.go`:

```go
// checkSymlinkSupport skips the test if the OS does not support symlinks
// in the temp directory (e.g. Windows without developer mode).
func checkSymlinkSupport(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	target := filepath.Join(dir, "link")
	if err := os.Symlink(dir, target); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}
	os.Remove(target)
}

func TestPathDriver_Symlink_Create(t *testing.T) {
	checkSymlinkSupport(t)
	src := t.TempDir()
	parent := t.TempDir()
	targetDir := filepath.Join(parent, "mpmissions")

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "symlink"}, targetDir, false)
	require.NoError(t, err)
	assert.True(t, result.Symlinked)

	link, err := os.Readlink(targetDir)
	require.NoError(t, err)
	assert.Equal(t, src, link)
}

func TestPathDriver_Symlink_AlreadyCorrect(t *testing.T) {
	checkSymlinkSupport(t)
	src := t.TempDir()
	parent := t.TempDir()
	targetDir := filepath.Join(parent, "mpmissions")
	require.NoError(t, os.Symlink(src, targetDir))

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "symlink"}, targetDir, false)
	require.NoError(t, err)
	assert.False(t, result.Symlinked)
}

func TestPathDriver_Symlink_ReplaceWrongLink(t *testing.T) {
	checkSymlinkSupport(t)
	src := t.TempDir()
	other := t.TempDir()
	parent := t.TempDir()
	targetDir := filepath.Join(parent, "mpmissions")
	require.NoError(t, os.Symlink(other, targetDir))

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "symlink"}, targetDir, false)
	require.NoError(t, err)
	assert.True(t, result.Symlinked)

	link, err := os.Readlink(targetDir)
	require.NoError(t, err)
	assert.Equal(t, src, link)
}

func TestPathDriver_Symlink_RealDirError(t *testing.T) {
	checkSymlinkSupport(t)
	src := t.TempDir()
	dst := t.TempDir() // real directory

	_, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "symlink"}, dst, false)
	assert.ErrorContains(t, err, "real directory")
}

func TestPathDriver_Symlink_DryRun(t *testing.T) {
	checkSymlinkSupport(t)
	src := t.TempDir()
	parent := t.TempDir()
	targetDir := filepath.Join(parent, "mpmissions")

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "symlink"}, targetDir, true)
	require.NoError(t, err)
	assert.True(t, result.Symlinked)

	_, statErr := os.Lstat(targetDir)
	assert.True(t, os.IsNotExist(statErr)) // no actual symlink created
}
```

- [ ] **Step 2: Run to confirm failure**

```
go test ./internal/missions/... -run "TestPathDriver_Symlink" -v
```

Expected: `FAIL` — symlink mode returns "not yet implemented".

- [ ] **Step 3: Replace symlink stub in `path_driver.go`**

Replace the `switch mode` block inside `PathDriver.pull`:

```go
switch mode {
case "symlink":
	return d.pullSymlink(source.Path, targetDir, dryRun)
default:
	return d.pullCopy(source.Path, targetDir, dryRun)
}
```

Then add `pullSymlink` method at the end of `path_driver.go`:

```go
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
			// Already points to the right place.
			return Result{Symlinked: false}, nil
		}
		// Points elsewhere — remove before re-creating.
		if !dryRun {
			if err := os.Remove(targetDir); err != nil {
				return Result{}, err
			}
		}
	}

	if !dryRun {
		if err := os.Symlink(sourcePath, targetDir); err != nil {
			return Result{}, err
		}
	}
	return Result{Symlinked: true}, nil
}
```

- [ ] **Step 4: Run symlink tests**

```
go test ./internal/missions/... -run "TestPathDriver_Symlink" -v
```

Expected: all `PASS` (or `SKIP` on Windows without developer mode).

- [ ] **Step 5: Run full missions test suite**

```
go test ./internal/missions/... -v
```

Expected: all `PASS`.

- [ ] **Step 6: Commit**

```
git add internal/missions/path_driver.go internal/missions/puller_test.go
git commit -m "feat: add symlink mode to PathDriver"
```

---

## Task 4: `zeus missions pull` command + example configs

**Files:**
- Create: `cmd/missions.go`
- Modify: `cmd/root.go`
- Modify: `conf/example.yaml`
- Modify: `conf/example.toml`

- [ ] **Step 1: Create `cmd/missions.go`**

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/smitt14ua/zeus/internal/missions"
	"github.com/smitt14ua/zeus/internal/storage"
)

var missionsCmd = &cobra.Command{
	Use:   "missions",
	Short: "Manage mission files for a profile",
}

var missionsPullCmd = &cobra.Command{
	Use:   "pull <profile>",
	Short: "Sync .pbo mission files from the configured source into the profile mpmissions directory",
	Args:  cobra.ExactArgs(1),
	Example: `  zeus missions pull my-server
  zeus missions pull my-server --dry-run`,
	Run: runMissionsPull,
}

func runMissionsPull(cmd *cobra.Command, args []string) {
	name := args[0]
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	repo := storage.ProfileRepository{}
	p, err := repo.Get(name)
	if err != nil {
		fatal(err)
	}
	if p.MissionSource == nil {
		fatalf("profile %q has no mission_source configured", name)
	}

	source := *p.MissionSource
	mode := source.Mode
	if mode == "" {
		mode = "copy"
	}

	targetDir := filepath.Join(p.InstallDir, ".zeus", p.Name, "mpmissions")

	header := fmt.Sprintf("Pulling missions for %q (%s, %s)", name, source.Driver, mode)
	if dryRun {
		header = "[dry-run] " + header
	}
	fmt.Println(header)

	result, err := missions.Puller{}.Pull(source, targetDir, dryRun)
	if err != nil {
		fatal(err)
	}

	if mode == "symlink" {
		if result.Symlinked {
			fmt.Printf("  → %s\n", source.Path)
		} else {
			fmt.Printf("  = %s (symlink unchanged)\n", source.Path)
		}
	} else {
		// Sort each slice for deterministic output.
		sort.Strings(result.Added)
		sort.Strings(result.Updated)
		sort.Strings(result.Removed)
		sort.Strings(result.Skipped)
		for _, f := range result.Added {
			fmt.Printf("  + %s\n", f)
		}
		for _, f := range result.Updated {
			fmt.Printf("  ~ %s\n", f)
		}
		for _, f := range result.Removed {
			fmt.Printf("  - %s\n", f)
		}
		for _, f := range result.Skipped {
			fmt.Printf("  = %s\n", f)
		}
	}

	if dryRun {
		fmt.Println("Dry run complete. No changes made.")
		return
	}

	if mode == "symlink" {
		fmt.Println("Done.")
	} else {
		fmt.Printf("Done. %d added, %d updated, %d removed, %d unchanged.\n",
			len(result.Added), len(result.Updated), len(result.Removed), len(result.Skipped))
	}
}

func init() {
	missionsCmd.AddCommand(missionsPullCmd)
	missionsPullCmd.Flags().Bool("dry-run", false, "show what would change without making changes")
}
```

Note: `os` import is unused above — remove it. The `os.MkdirAll` for copy mode is handled inside `PathDriver.pullCopy`. Remove `"os"` from the import block.

Corrected import block:

```go
import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/smitt14ua/zeus/internal/missions"
	"github.com/smitt14ua/zeus/internal/storage"
)
```

- [ ] **Step 2: Register `missionsCmd` in `cmd/root.go`**

Add at the end of the `func init()` block (or create one if absent). Current `root.go` has no `init()`, so add:

```go
func init() {
	rootCmd.AddCommand(missionsCmd)
}
```

- [ ] **Step 3: Build to verify no compile errors**

```
go build ./...
```

Expected: no output (success).

- [ ] **Step 4: Add `mission_source` block to `conf/example.yaml`**

Append to `conf/example.yaml`:

```yaml
# Optional: sync .pbo mission files from a local directory.
# mode: copy   — copies .pbo files into profile mpmissions dir (default)
# mode: symlink — replaces mpmissions dir with a symlink to path
mission_source:
  driver: path
  path: "D:\\Missions"
  mode: copy
```

- [ ] **Step 5: Add `mission_source` block to `conf/example.toml`**

Append to `conf/example.toml`:

```toml
# Optional: sync .pbo mission files from a local directory.
# mode = "copy"    — copies .pbo files into profile mpmissions dir (default)
# mode = "symlink" — replaces mpmissions dir with a symlink to path
[mission_source]
driver = "path"
path = 'D:\Missions'
mode = "copy"
```

- [ ] **Step 6: Run full test suite**

```
go test ./... 2>&1
```

Expected: all packages `ok`.

- [ ] **Step 7: Commit**

```
git add cmd/missions.go cmd/root.go conf/example.yaml conf/example.toml
git commit -m "feat: add zeus missions pull command with path driver"
```

---

## Self-Review Checklist

- [x] **Spec coverage:**
  - `MissionSource` struct — Task 1
  - `Profile.MissionSource` field — Task 1
  - `internal/missions` package with `Puller`, `Result`, `PathDriver` — Tasks 2–3
  - Copy mode sync (add/update/remove/skip, .pbo only, non-.pbo untouched) — Task 2
  - Symlink mode (create, no-op, replace, real-dir error) — Task 3
  - dry-run for both modes — Tasks 2 & 3
  - `zeus missions pull <profile> [--dry-run]` command — Task 4
  - Output format (copy and symlink variants) — Task 4
  - Example configs updated — Task 4
  - All error cases from spec: source not exist, unknown driver, real-dir-in-symlink-mode, mode empty → copy — covered in tests and implementation

- [x] **No placeholders:** All steps contain full code.

- [x] **Type consistency:**
  - `Result` defined in `puller.go`, used in `path_driver.go` and `cmd/missions.go` — consistent
  - `PathDriver.pull` called from `Puller.Pull` — consistent
  - `arma.MissionSource` used throughout — consistent
  - `pullCopy` / `pullSymlink` are unexported methods on `PathDriver` — consistent
