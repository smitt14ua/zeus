package profile

import (
	"path/filepath"
	"testing"
)

func TestValidateName_Valid(t *testing.T) {
	valid := []string{
		"server", "my-server", "server1", "a", "A",
		"server.backup", "arma3-test_2", "x64",
	}
	for _, name := range valid {
		if err := ValidateName(name); err != nil {
			t.Errorf("ValidateName(%q) unexpected error: %v", name, err)
		}
	}
}

func TestValidateName_Invalid(t *testing.T) {
	cases := []struct {
		name   string
		reason string
	}{
		{"", "empty"},
		{".hidden", "leading dot"},
		{"../evil", "path traversal with slash"},
		{"foo/bar", "forward slash"},
		{"foo\\bar", "backslash"},
		{"foo\x00bar", "null byte"},
		{"CON", "Windows reserved"},
		{"con", "Windows reserved (lower)"},
		{"NUL.cfg", "Windows reserved stem"},
		{"COM1", "Windows reserved COM port"},
		{"-starts-with-dash", "leading dash not alphanumeric"},
		{"_starts-with-underscore", "leading underscore"},
		{string(make([]byte, 65)), "too long (65 chars)"},
	}
	for _, tc := range cases {
		if err := ValidateName(tc.name); err == nil {
			t.Errorf("ValidateName(%q) should have failed (%s)", tc.name, tc.reason)
		}
	}
}

func TestValidateName_MaxLength(t *testing.T) {
	// exactly 64 chars starting with a letter
	name64 := "a" + string(make([]byte, 63))
	for i := range name64[1:] {
		name64 = name64[:i+1] + "b" + name64[i+2:]
	}
	if len(name64) != 64 {
		t.Fatalf("test setup: expected 64, got %d", len(name64))
	}
	if err := ValidateName(name64); err != nil {
		t.Errorf("64-char name should be valid: %v", err)
	}
}

func TestProfileValidate_Valid(t *testing.T) {
	p := Profile{Name: "my-server", InstallDir: t.TempDir()}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

func TestProfileValidate_EmptyInstallDir(t *testing.T) {
	p := Profile{Name: "srv"}
	if err := p.Validate(); err == nil {
		t.Error("expected error for empty InstallDir")
	}
}

func TestProfileValidate_RelativeInstallDir(t *testing.T) {
	p := Profile{Name: "srv", InstallDir: "relative/path"}
	if err := p.Validate(); err == nil {
		t.Error("expected error for relative InstallDir")
	}
}

func TestProfileValidate_AbsoluteInstallDir(t *testing.T) {
	p := Profile{Name: "srv", InstallDir: filepath.FromSlash("/opt/arma3")}
	// On Windows this fails because /opt/arma3 is not absolute; use t.TempDir() instead.
	dir := t.TempDir()
	p.InstallDir = dir
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() unexpected error with absolute dir: %v", err)
	}
}
