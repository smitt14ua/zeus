package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/smitt14ua/zeus/internal/profile"
)

func newRepo(t *testing.T) ProfileRepository {
	t.Helper()
	return ProfileRepository{HomeDir: t.TempDir()}
}

func sampleProfile(name string) profile.Profile {
	return profile.Profile{Name: name, InstallDir: "/some/dir"}
}

func TestProfileRepository_SaveAndExists(t *testing.T) {
	r := newRepo(t)

	exists, err := r.Exists("alpha")
	if err != nil || exists {
		t.Fatalf("expected not exists before save, got exists=%v err=%v", exists, err)
	}

	if err := r.Save(sampleProfile("alpha")); err != nil {
		t.Fatalf("Save: %v", err)
	}

	exists, err = r.Exists("alpha")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !exists {
		t.Fatal("expected profile to exist after save")
	}
}

func TestProfileRepository_List(t *testing.T) {
	r := newRepo(t)

	profiles, err := r.List()
	if err != nil || len(profiles) != 0 {
		t.Fatalf("expected empty list, got %v err=%v", profiles, err)
	}

	if err := r.Save(sampleProfile("alpha")); err != nil {
		t.Fatal(err)
	}
	if err := r.Save(sampleProfile("beta")); err != nil {
		t.Fatal(err)
	}

	profiles, err = r.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}
}

func TestProfileRepository_SavePreservesFields(t *testing.T) {
	r := newRepo(t)
	p := profile.Profile{Name: "test", InstallDir: "/opt/arma3"}

	if err := r.Save(p); err != nil {
		t.Fatal(err)
	}

	profiles, err := r.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	got := profiles[0]
	if got.Name != p.Name || got.InstallDir != p.InstallDir {
		t.Fatalf("round-trip mismatch: got %+v, want %+v", got, p)
	}
}

func TestProfileRepository_Delete(t *testing.T) {
	r := newRepo(t)

	if err := r.Save(sampleProfile("alpha")); err != nil {
		t.Fatal(err)
	}

	if err := r.Delete("alpha"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	exists, err := r.Exists("alpha")
	if err != nil || exists {
		t.Fatalf("expected not exists after delete, got exists=%v err=%v", exists, err)
	}
}

func TestProfileRepository_DeleteMissing(t *testing.T) {
	r := newRepo(t)
	if err := r.Delete("ghost"); err == nil {
		t.Fatal("expected error deleting non-existent profile")
	}
}

func TestProfileRepository_Get(t *testing.T) {
	r := newRepo(t)

	want := profile.Profile{Name: "gamma", InstallDir: "/opt/arma3"}
	if err := r.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := r.Get("gamma")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != want.Name {
		t.Errorf("Name: got %q want %q", got.Name, want.Name)
	}
	if got.InstallDir != want.InstallDir {
		t.Errorf("InstallDir: got %q want %q", got.InstallDir, want.InstallDir)
	}
}

func TestProfileRepository_Get_NotFound(t *testing.T) {
	r := newRepo(t)
	_, err := r.Get("ghost")
	if err == nil {
		t.Fatal("expected error for missing profile")
	}
}

// ── List skips irrelevant entries ─────────────────────────────────────────────

func TestProfileRepository_List_SkipsDirectories(t *testing.T) {
	r := newRepo(t)
	if err := r.Save(sampleProfile("alpha")); err != nil {
		t.Fatal(err)
	}

	// Plant a subdirectory inside the profiles dir — must be skipped.
	dir, _ := r.profilesDir()
	if err := os.MkdirAll(filepath.Join(dir, "somedir"), 0755); err != nil {
		t.Fatal(err)
	}

	profiles, err := r.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile (directory should be skipped), got %d", len(profiles))
	}
}

func TestProfileRepository_List_SkipsUnknownExtensions(t *testing.T) {
	r := newRepo(t)
	if err := r.Save(sampleProfile("alpha")); err != nil {
		t.Fatal(err)
	}

	// Plant a .txt file — must be skipped.
	dir, _ := r.profilesDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignored"), 0644); err != nil {
		t.Fatal(err)
	}

	profiles, err := r.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile (.txt should be skipped), got %d", len(profiles))
	}
}

// ── Get and Exists with YAML/TOML files ──────────────────────────────────────

const yamlProfile = `
name: yaml-srv
install_dir: /opt/arma3
`

func TestProfileRepository_Get_YAML(t *testing.T) {
	r := newRepo(t)
	dir, _ := r.profilesDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "yaml-srv.yaml"), []byte(yamlProfile), 0644); err != nil {
		t.Fatal(err)
	}

	p, err := r.Get("yaml-srv")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if p.Name != "yaml-srv" {
		t.Errorf("Name = %q", p.Name)
	}
}

func TestProfileRepository_Exists_YAML(t *testing.T) {
	r := newRepo(t)
	dir, _ := r.profilesDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "yaml-srv.yaml"), []byte(yamlProfile), 0644); err != nil {
		t.Fatal(err)
	}

	exists, err := r.Exists("yaml-srv")
	if err != nil || !exists {
		t.Fatalf("Exists = %v err = %v", exists, err)
	}
}

func TestProfileRepository_Delete_YAML(t *testing.T) {
	r := newRepo(t)
	dir, _ := r.profilesDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "yaml-srv.yaml"), []byte(yamlProfile), 0644); err != nil {
		t.Fatal(err)
	}

	if err := r.Delete("yaml-srv"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	exists, err := r.Exists("yaml-srv")
	if err != nil || exists {
		t.Fatalf("expected not exists after delete, got exists=%v err=%v", exists, err)
	}
}
