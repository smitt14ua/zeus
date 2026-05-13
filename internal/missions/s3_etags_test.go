package missions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadETags_NonExistent(t *testing.T) {
	dir := t.TempDir()
	etags, err := loadETags(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(etags) != 0 {
		t.Fatalf("expected empty map, got %v", etags)
	}
}

func TestLoadETags_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := map[string]string{
		"op_cobra.pbo":     `"d41d8cd98f00b204e9800998ecf8427e"`,
		"op_patrol_v2.pbo": `"abc123"`,
	}
	if err := saveETags(dir, want); err != nil {
		t.Fatalf("saveETags: %v", err)
	}
	got, err := loadETags(dir)
	if err != nil {
		t.Fatalf("loadETags: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("len mismatch: got %d want %d", len(got), len(want))
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("key %q: got %q want %q", k, got[k], v)
		}
	}
}

func TestLoadETags_Overwrites(t *testing.T) {
	dir := t.TempDir()
	first := map[string]string{"a.pbo": `"etag1"`}
	second := map[string]string{"b.pbo": `"etag2"`}
	if err := saveETags(dir, first); err != nil {
		t.Fatal(err)
	}
	if err := saveETags(dir, second); err != nil {
		t.Fatal(err)
	}
	got, err := loadETags(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got["a.pbo"]; ok {
		t.Error("old key a.pbo should not exist after overwrite")
	}
	if got["b.pbo"] != `"etag2"` {
		t.Errorf("expected b.pbo etag2, got %q", got["b.pbo"])
	}
}

func TestLoadETags_SidecarNotPBO(t *testing.T) {
	// scanPBOs ignores non-.pbo files; verify sidecar filename constant is not .pbo
	if filepath.Ext(etagSidecar) == ".pbo" {
		t.Errorf("etagSidecar %q must not have .pbo extension", etagSidecar)
	}
}

func TestLoadETags_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, etagSidecar)
	if err := os.WriteFile(path, []byte("not json {{{"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := loadETags(dir)
	if err == nil {
		t.Fatal("expected error on corrupt JSON, got nil")
	}
}
