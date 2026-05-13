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
	_, err := Puller{}.Pull(arma.MissionSource{Driver: "ftp"}, t.TempDir(), false, false)
	assert.ErrorContains(t, err, `unsupported driver "ftp"`)
}

func TestPathDriver_Copy_SourceNotExist(t *testing.T) {
	src := arma.MissionSource{Driver: "path", Path: filepath.Join(t.TempDir(), "nonexistent"), Mode: "copy"}
	_, err := Puller{}.Pull(src, t.TempDir(), false, false)
	assert.ErrorContains(t, err, "does not exist")
}

func TestPathDriver_Copy_AddNew(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	mkpbo(t, src, "op_cobra.pbo", "data")

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "copy"}, dst, false, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"op_cobra.pbo"}, result.Added)
	assert.Empty(t, result.Updated)
	assert.Empty(t, result.Removed)
	assert.FileExists(t, filepath.Join(dst, "op_cobra.pbo"))
}

func TestPathDriver_Copy_RemoveStale(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	mkpbo(t, dst, "old.pbo", "stale")

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "copy"}, dst, false, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"old.pbo"}, result.Removed)
	assert.NoFileExists(t, filepath.Join(dst, "old.pbo"))
}

func TestPathDriver_Copy_UpdateChanged(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	mkpbo(t, src, "op.pbo", "new content longer")
	mkpbo(t, dst, "op.pbo", "old")

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "copy"}, dst, false, false)
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

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "copy"}, dst, false, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"mission.pbo"}, result.Skipped)
	assert.Empty(t, result.Updated)
}

func TestPathDriver_Copy_DryRun(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	mkpbo(t, src, "op.pbo", "data")

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "copy"}, dst, true, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"op.pbo"}, result.Added)
	assert.NoFileExists(t, filepath.Join(dst, "op.pbo")) // no actual copy
}

func TestPathDriver_Copy_IgnoresNonPBO(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(src, "readme.txt"), []byte("text"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dst, "keep.txt"), []byte("text"), 0644))

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "copy"}, dst, false, false)
	require.NoError(t, err)
	assert.Empty(t, result.Added)
	assert.Empty(t, result.Removed)
	assert.FileExists(t, filepath.Join(dst, "keep.txt")) // non-.pbo left alone
}

func TestPathDriver_Copy_DefaultModeCopy(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	mkpbo(t, src, "op.pbo", "data")

	// Mode empty → defaults to copy
	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: ""}, dst, false, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"op.pbo"}, result.Added)
	assert.FileExists(t, filepath.Join(dst, "op.pbo"))
}

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

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "symlink"}, targetDir, false, false)
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

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "symlink"}, targetDir, false, false)
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

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "symlink"}, targetDir, false, false)
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

	_, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "symlink"}, dst, false, false)
	assert.ErrorContains(t, err, "real directory")
}

func TestPathDriver_Symlink_DryRun(t *testing.T) {
	checkSymlinkSupport(t)
	src := t.TempDir()
	parent := t.TempDir()
	targetDir := filepath.Join(parent, "mpmissions")

	result, err := Puller{}.Pull(arma.MissionSource{Driver: "path", Path: src, Mode: "symlink"}, targetDir, true, false)
	require.NoError(t, err)
	assert.True(t, result.Symlinked)

	_, statErr := os.Lstat(targetDir)
	assert.True(t, os.IsNotExist(statErr)) // no actual symlink created
}
