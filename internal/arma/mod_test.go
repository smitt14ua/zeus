package arma

import (
	"os"
	"path/filepath"
	"testing"
)

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
}

func touch(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, nil, 0644); err != nil {
		t.Fatal(err)
	}
}

func buildModDir(t *testing.T, root string) {
	t.Helper()
	// addons — mixed case dir, mixed extensions
	mkdirAll(t, filepath.Join(root, "Addons"))
	touch(t, filepath.Join(root, "Addons", "cba_main.pbo"))
	touch(t, filepath.Join(root, "Addons", "cba_a3.pbo"))
	touch(t, filepath.Join(root, "Addons", "cba_xeh.ebo"))
	touch(t, filepath.Join(root, "Addons", "readme.txt")) // should be ignored

	// keys — uppercase dir
	mkdirAll(t, filepath.Join(root, "KEYS"))
	touch(t, filepath.Join(root, "KEYS", "cba.bikey"))
	touch(t, filepath.Join(root, "KEYS", "not_a_key.txt")) // should be ignored

	// optionals with two submods
	mkdirAll(t, filepath.Join(root, "optionals", "@cba_xeh", "addons"))
	touch(t, filepath.Join(root, "optionals", "@cba_xeh", "addons", "cba_xeh.pbo"))
	mkdirAll(t, filepath.Join(root, "optionals", "@cba_xeh", "keys"))
	touch(t, filepath.Join(root, "optionals", "@cba_xeh", "keys", "cba_xeh.bikey"))

	mkdirAll(t, filepath.Join(root, "optionals", "@cba_legacy"))
	mkdirAll(t, filepath.Join(root, "optionals", "@cba_legacy", "addons"))
	touch(t, filepath.Join(root, "optionals", "@cba_legacy", "addons", "cba_legacy.pbo"))

	// non-@ dir inside optionals — should be ignored
	mkdirAll(t, filepath.Join(root, "optionals", "docs"))
}

func TestLoadMod_Path(t *testing.T) {
	root := t.TempDir()
	buildModDir(t, root)

	m, err := LoadMod(root)
	if err != nil {
		t.Fatalf("LoadMod: %v", err)
	}
	if m.Path != root {
		t.Errorf("Path = %q, want %q", m.Path, root)
	}
}

func TestLoadMod_Addons(t *testing.T) {
	root := t.TempDir()
	buildModDir(t, root)

	m, err := LoadMod(root)
	if err != nil {
		t.Fatalf("LoadMod: %v", err)
	}

	if len(m.Addons) != 3 {
		t.Errorf("Addons count = %d, want 3; got %v", len(m.Addons), m.Addons)
	}
	for _, a := range m.Addons {
		ext := filepath.Ext(a)
		if ext != ".pbo" && ext != ".ebo" {
			t.Errorf("unexpected addon extension %q in %q", ext, a)
		}
	}
}

func TestLoadMod_Keys(t *testing.T) {
	root := t.TempDir()
	buildModDir(t, root)

	m, err := LoadMod(root)
	if err != nil {
		t.Fatalf("LoadMod: %v", err)
	}

	if len(m.Keys) != 1 {
		t.Errorf("Keys count = %d, want 1; got %v", len(m.Keys), m.Keys)
	}
	if filepath.Ext(m.Keys[0]) != ".bikey" {
		t.Errorf("unexpected key extension in %q", m.Keys[0])
	}
}

func TestLoadMod_Submods(t *testing.T) {
	root := t.TempDir()
	buildModDir(t, root)

	m, err := LoadMod(root)
	if err != nil {
		t.Fatalf("LoadMod: %v", err)
	}

	if len(m.Submods) != 2 {
		t.Errorf("Submods count = %d, want 2; got %v", len(m.Submods), m.Submods)
	}

	byName := make(map[string]Mod, len(m.Submods))
	for _, s := range m.Submods {
		byName[filepath.Base(s.Path)] = s
	}

	xeh, ok := byName["@cba_xeh"]
	if !ok {
		t.Fatal("missing submod @cba_xeh")
	}
	if len(xeh.Addons) != 1 {
		t.Errorf("@cba_xeh addons = %d, want 1", len(xeh.Addons))
	}
	if len(xeh.Keys) != 1 {
		t.Errorf("@cba_xeh keys = %d, want 1", len(xeh.Keys))
	}

	legacy, ok := byName["@cba_legacy"]
	if !ok {
		t.Fatal("missing submod @cba_legacy")
	}
	if len(legacy.Addons) != 1 {
		t.Errorf("@cba_legacy addons = %d, want 1", len(legacy.Addons))
	}
}

func TestLoadMod_CaseInsensitiveDirs(t *testing.T) {
	root := t.TempDir()

	// All uppercase
	mkdirAll(t, filepath.Join(root, "ADDONS"))
	touch(t, filepath.Join(root, "ADDONS", "mod.pbo"))
	mkdirAll(t, filepath.Join(root, "Keys"))
	touch(t, filepath.Join(root, "Keys", "mod.bikey"))

	m, err := LoadMod(root)
	if err != nil {
		t.Fatalf("LoadMod: %v", err)
	}
	if len(m.Addons) != 1 {
		t.Errorf("Addons = %d, want 1", len(m.Addons))
	}
	if len(m.Keys) != 1 {
		t.Errorf("Keys = %d, want 1", len(m.Keys))
	}
}

func TestLoadMod_EmptyMod(t *testing.T) {
	root := t.TempDir()

	m, err := LoadMod(root)
	if err != nil {
		t.Fatalf("LoadMod: %v", err)
	}
	if len(m.Addons) != 0 || len(m.Keys) != 0 || len(m.Submods) != 0 {
		t.Errorf("expected empty mod, got %+v", m)
	}
}

func TestLoadMod_NonExistentPath(t *testing.T) {
	_, err := LoadMod(filepath.Join(t.TempDir(), "does_not_exist"))
	if err == nil {
		t.Fatal("expected error for non-existent path")
	}
}
