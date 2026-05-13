package storage

import (
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
