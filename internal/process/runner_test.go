package process

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/smitt14ua/zeus/internal/arma"
	"github.com/smitt14ua/zeus/internal/profile"
)

func ptr[T any](v T) *T { return &v }

// ── quoteIfNeeded ─────────────────────────────────────────────────────────────

func TestQuoteIfNeeded(t *testing.T) {
	cases := []struct{ in, want string }{
		{"simple", "simple"},
		{"/path/to/file", "/path/to/file"},
		{"/path with spaces/file", `"/path with spaces/file"`},
		{"tab\there", `"tab` + "\t" + `here"`},
	}
	for _, c := range cases {
		if got := quoteIfNeeded(c.in); got != c.want {
			t.Errorf("quoteIfNeeded(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ── absModPaths ───────────────────────────────────────────────────────────────

func TestAbsModPaths(t *testing.T) {
	installDir := t.TempDir()
	absMod := t.TempDir()

	t.Run("absolute_unchanged", func(t *testing.T) {
		got := absModPaths(installDir, []string{absMod})
		if got[0] != absMod {
			t.Fatalf("got %q", got[0])
		}
	})
	t.Run("relative_resolved", func(t *testing.T) {
		got := absModPaths(installDir, []string{"@mod"})
		want := filepath.Join(installDir, "@mod")
		if got[0] != want {
			t.Fatalf("got %q want %q", got[0], want)
		}
	})
	t.Run("empty_install_dir_leaves_relative", func(t *testing.T) {
		got := absModPaths("", []string{"@mod"})
		if got[0] != "@mod" {
			t.Fatalf("got %q", got[0])
		}
	})
	t.Run("nil_slice_returns_empty", func(t *testing.T) {
		got := absModPaths(installDir, nil)
		if len(got) != 0 {
			t.Fatalf("got %v", got)
		}
	})
}

// ── buildArgs ─────────────────────────────────────────────────────────────────

func TestBuildArgs_NilPointers(t *testing.T) {
	got := buildArgs(arma.StartupParams{})
	if len(got) != 0 {
		t.Fatalf("expected no args, got %v", got)
	}
}

func TestBuildArgs_BoolTrue(t *testing.T) {
	params := arma.StartupParams{Server: ptr(true)}
	args := buildArgs(params)
	if len(args) != 1 || args[0] != "-server" {
		t.Fatalf("got %v", args)
	}
}

func TestBuildArgs_BoolFalse(t *testing.T) {
	params := arma.StartupParams{Server: ptr(false)}
	args := buildArgs(params)
	if len(args) != 0 {
		t.Fatalf("false bool should produce no arg, got %v", args)
	}
}

func TestBuildArgs_StringPointer(t *testing.T) {
	params := arma.StartupParams{Config: ptr("/etc/server.cfg")}
	args := buildArgs(params)
	if len(args) != 1 || args[0] != "-config=/etc/server.cfg" {
		t.Fatalf("got %v", args)
	}
}

func TestBuildArgs_StringPointerWithSpaces(t *testing.T) {
	params := arma.StartupParams{Config: ptr("/path with spaces/server.cfg")}
	args := buildArgs(params)
	want := `-config="/path with spaces/server.cfg"`
	if len(args) != 1 || args[0] != want {
		t.Fatalf("got %v want %v", args, want)
	}
}

func TestBuildArgs_Uint16Pointer(t *testing.T) {
	params := arma.StartupParams{Port: ptr(uint16(2302))}
	args := buildArgs(params)
	if len(args) != 1 || args[0] != "-port=2302" {
		t.Fatalf("got %v", args)
	}
}

func TestBuildArgs_Slice(t *testing.T) {
	params := arma.StartupParams{Mod: []string{"/mod1", "/mod2", "/mod3"}}
	args := buildArgs(params)
	if len(args) != 1 || args[0] != "-mod=/mod1;/mod2;/mod3" {
		t.Fatalf("got %v", args)
	}
}

func TestBuildArgs_EmptySlice(t *testing.T) {
	params := arma.StartupParams{Mod: []string{}}
	args := buildArgs(params)
	if len(args) != 0 {
		t.Fatalf("empty slice should produce no arg, got %v", args)
	}
}

// ── resolveExecutable ─────────────────────────────────────────────────────────

func TestResolveExecutable_Default(t *testing.T) {
	installDir := t.TempDir()
	p := profile.Profile{InstallDir: installDir}
	exe := resolveExecutable(p)
	wantBin := "arma3server_x64"
	if runtime.GOOS == "windows" {
		wantBin += ".exe"
	}
	if exe != filepath.Join(installDir, wantBin) {
		t.Fatalf("got %q", exe)
	}
}

func TestResolveExecutable_Custom(t *testing.T) {
	installDir := t.TempDir()
	p := profile.Profile{InstallDir: installDir, Executable: "arma3server"}
	exe := resolveExecutable(p)
	want := filepath.Join(installDir, "arma3server")
	if exe != want {
		t.Fatalf("got %q want %q", exe, want)
	}
}

// ── prepareParams ─────────────────────────────────────────────────────────────

func TestPrepareParams(t *testing.T) {
	r := Runner{HomeDir: t.TempDir()}
	p := profile.Profile{
		Name:       "myserver",
		InstallDir: t.TempDir(),
	}

	prepared, err := r.prepareParams(p)
	if err != nil {
		t.Fatalf("prepareParams: %v", err)
	}

	zeusDir := filepath.Join(".zeus", "myserver")

	if prepared.Params.MpMissions == nil || *prepared.Params.MpMissions != filepath.Join(zeusDir, "mpmissions") {
		t.Errorf("MpMissions = %v", prepared.Params.MpMissions)
	}
	if prepared.Params.Cfg == nil || *prepared.Params.Cfg != filepath.Join(zeusDir, "configs", "basic.cfg") {
		t.Errorf("Cfg = %v", prepared.Params.Cfg)
	}
	if prepared.Params.Config == nil || *prepared.Params.Config != filepath.Join(zeusDir, "configs", "server.cfg") {
		t.Errorf("Config = %v", prepared.Params.Config)
	}
	if prepared.Params.Pid == nil || !strings.HasSuffix(*prepared.Params.Pid, filepath.Join(".zeus", "running", "myserver.pid")) {
		t.Errorf("Pid = %v", prepared.Params.Pid)
	}
	if prepared.Params.Profiles == nil || *prepared.Params.Profiles != zeusDir {
		t.Errorf("Profiles = %v", prepared.Params.Profiles)
	}
	if prepared.Params.Name == nil || *prepared.Params.Name != "myserver" {
		t.Errorf("Name = %v", prepared.Params.Name)
	}

	wantKeys := []string{
		filepath.Join(zeusDir, "keys"),
		filepath.Join(zeusDir, "optionalkeys"),
	}
	if len(prepared.Params.KeysFolder) != len(wantKeys) {
		t.Fatalf("KeysFolder = %v, want %v", prepared.Params.KeysFolder, wantKeys)
	}
	for i, want := range wantKeys {
		if prepared.Params.KeysFolder[i] != want {
			t.Errorf("KeysFolder[%d] = %q, want %q", i, prepared.Params.KeysFolder[i], want)
		}
	}
}

func TestPrepareParams_PreservesMpMissions(t *testing.T) {
	r := Runner{HomeDir: t.TempDir()}
	custom := "/custom/mpmissions"
	p := profile.Profile{
		Name: "srv",
		Params: arma.StartupParams{MpMissions: ptr(custom)},
	}

	prepared, err := r.prepareParams(p)
	if err != nil {
		t.Fatalf("prepareParams: %v", err)
	}
	if *prepared.Params.MpMissions != custom {
		t.Errorf("MpMissions overwritten: got %q", *prepared.Params.MpMissions)
	}
}

func TestPrepareParams_ResolvesMods(t *testing.T) {
	r := Runner{HomeDir: t.TempDir()}
	installDir := t.TempDir()
	absMod := t.TempDir()
	p := profile.Profile{
		Name:       "srv",
		InstallDir: installDir,
		Params:     arma.StartupParams{Mod: []string{"@CBA", absMod}},
	}

	prepared, err := r.prepareParams(p)
	if err != nil {
		t.Fatalf("prepareParams: %v", err)
	}
	if prepared.Params.Mod[0] != filepath.Join(installDir, "@CBA") {
		t.Errorf("relative mod not resolved: %q", prepared.Params.Mod[0])
	}
	if prepared.Params.Mod[1] != absMod {
		t.Errorf("absolute mod changed: %q", prepared.Params.Mod[1])
	}
}
