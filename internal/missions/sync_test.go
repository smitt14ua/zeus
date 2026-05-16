package missions

import (
	"os"
	"path/filepath"
	"testing"
)

// ── scanPBOs ──────────────────────────────────────────────────────────────────

func TestScanPBOs_MissingDir(t *testing.T) {
	result, err := scanPBOs(filepath.Join(t.TempDir(), "nonexistent"))
	if err != nil {
		t.Fatalf("expected no error for missing dir, got %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty map, got %v", result)
	}
}

func TestScanPBOs_MixedFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "op1.pbo"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(dir, "op2.PBO"), []byte("data"), 0644) // case-insensitive
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("ignored"), 0644)
	os.MkdirAll(filepath.Join(dir, "subdir"), 0755)

	result, err := scanPBOs(dir)
	if err != nil {
		t.Fatalf("scanPBOs: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 PBOs, got %d: %v", len(result), result)
	}
	if _, ok := result["op1.pbo"]; !ok {
		t.Error("op1.pbo missing from result")
	}
	if _, ok := result["op2.pbo"]; !ok {
		t.Error("op2.pbo missing from result (case normalization)")
	}
}

func TestScanPBOs_Empty(t *testing.T) {
	result, err := scanPBOs(t.TempDir())
	if err != nil {
		t.Fatalf("scanPBOs: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty, got %v", result)
	}
}

// ── copyFile ──────────────────────────────────────────────────────────────────

func TestCopyFile_Success(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src.pbo")
	dst := filepath.Join(t.TempDir(), "dst.pbo")

	if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil || string(data) != "content" {
		t.Fatalf("dst content = %q err = %v", data, err)
	}

	// mtime is preserved
	srcInfo, _ := os.Stat(src)
	dstInfo, _ := os.Stat(dst)
	if !srcInfo.ModTime().Equal(dstInfo.ModTime()) {
		t.Errorf("mtime not preserved: src=%v dst=%v", srcInfo.ModTime(), dstInfo.ModTime())
	}
}

func TestCopyFile_MissingSource(t *testing.T) {
	src := filepath.Join(t.TempDir(), "nonexistent.pbo")
	dst := filepath.Join(t.TempDir(), "dst.pbo")

	if err := copyFile(src, dst); err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestCopyFile_DestDirMissing(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src.pbo")
	os.WriteFile(src, []byte("data"), 0644)

	dst := filepath.Join(t.TempDir(), "nonexistent", "dst.pbo")
	// Destination directory doesn't exist — CreateTemp will fail.
	if err := copyFile(src, dst); err == nil {
		t.Fatal("expected error when dest dir is missing")
	}
}

// ── sameFile ──────────────────────────────────────────────────────────────────

func TestSameFile_Identical(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.pbo")
	os.WriteFile(path, []byte("abc"), 0644)
	info, _ := os.Stat(path)
	if !sameFile(info, info) {
		t.Error("sameFile(info, info) should be true")
	}
}

func TestSameFile_DifferentSize(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.pbo")
	b := filepath.Join(dir, "b.pbo")
	os.WriteFile(a, []byte("short"), 0644)
	os.WriteFile(b, []byte("longer content"), 0644)
	infoA, _ := os.Stat(a)
	infoB, _ := os.Stat(b)
	if sameFile(infoA, infoB) {
		t.Error("sameFile should be false for different sizes")
	}
}
