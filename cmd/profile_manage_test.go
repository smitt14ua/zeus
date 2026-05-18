package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smitt14ua/zeus/internal/arma"
	"github.com/smitt14ua/zeus/internal/profile"
	"github.com/smitt14ua/zeus/internal/storage"
)

// setTestHome redirects os.UserHomeDir to a temp directory for the duration of the test.
// Both process.Manager and storage.ProfileRepository fall back to os.UserHomeDir when
// their HomeDir field is empty, so this lets cmd-level functions work against a temp dir.
func setTestHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	return dir
}

// saveProfile stores p in homeDir using ProfileRepository directly.
func saveProfile(t *testing.T, homeDir string, p profile.Profile) {
	t.Helper()
	require.NoError(t, storage.ProfileRepository{HomeDir: homeDir}.Save(p))
}

// ── execProfileList ───────────────────────────────────────────────────────────

func TestExecProfileList_JSON_Compact(t *testing.T) {
	home := setTestHome(t)
	saveProfile(t, home, profile.Profile{Name: "alpha", InstallDir: t.TempDir()})

	var buf bytes.Buffer
	require.NoError(t, execProfileList(context.Background(), "json", &buf))

	out := strings.TrimSuffix(buf.String(), "\n")
	assert.NotContains(t, out, "\n", "JSON output must be a single compact line (not indented)")

	var entries []profileListEntry
	require.NoError(t, json.Unmarshal([]byte(out), &entries))
	require.Len(t, entries, 1)
	assert.Equal(t, "alpha", entries[0].Name)
	assert.False(t, entries[0].Running)
	assert.Zero(t, entries[0].PID)
}

func TestExecProfileList_JSON_Empty(t *testing.T) {
	setTestHome(t)

	var buf bytes.Buffer
	require.NoError(t, execProfileList(context.Background(), "json", &buf))

	assert.Equal(t, "[]\n", buf.String())
}

func TestExecProfileList_JSON_MultipleProfiles(t *testing.T) {
	home := setTestHome(t)
	saveProfile(t, home, profile.Profile{Name: "alpha", InstallDir: t.TempDir()})
	saveProfile(t, home, profile.Profile{Name: "beta", InstallDir: t.TempDir()})

	var buf bytes.Buffer
	require.NoError(t, execProfileList(context.Background(), "json", &buf))

	var entries []profileListEntry
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &entries))
	assert.Len(t, entries, 2)
}

func TestExecProfileList_Console_ShowsHeaderAndName(t *testing.T) {
	home := setTestHome(t)
	saveProfile(t, home, profile.Profile{Name: "beta", InstallDir: t.TempDir()})

	var buf bytes.Buffer
	require.NoError(t, execProfileList(context.Background(), "console", &buf))

	out := buf.String()
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "STATUS")
	assert.Contains(t, out, "beta")
	assert.Contains(t, out, "-")
}

// ── execProfileInfo ───────────────────────────────────────────────────────────

func TestExecProfileInfo_JSON_RoundTrip(t *testing.T) {
	home := setTestHome(t)
	hostname := "Test Server"
	p := profile.Profile{
		Name:       "gamma",
		InstallDir: t.TempDir(),
		Config:     arma.ServerConfig{Hostname: &hostname},
	}
	saveProfile(t, home, p)

	var buf bytes.Buffer
	require.NoError(t, execProfileInfo(context.Background(), "gamma", "json", &buf))

	var got profile.Profile
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &got))
	assert.Equal(t, "gamma", got.Name)
	require.NotNil(t, got.Config.Hostname)
	assert.Equal(t, "Test Server", *got.Config.Hostname)
}

func TestExecProfileInfo_YAML(t *testing.T) {
	home := setTestHome(t)
	saveProfile(t, home, profile.Profile{Name: "delta", InstallDir: "/srv/arma3"})

	var buf bytes.Buffer
	require.NoError(t, execProfileInfo(context.Background(), "delta", "yaml", &buf))

	out := buf.String()
	assert.Contains(t, out, "name:")
	assert.Contains(t, out, "delta")
}

func TestExecProfileInfo_TOML(t *testing.T) {
	home := setTestHome(t)
	saveProfile(t, home, profile.Profile{Name: "epsilon", InstallDir: "/srv/arma3"})

	var buf bytes.Buffer
	require.NoError(t, execProfileInfo(context.Background(), "epsilon", "toml", &buf))

	out := buf.String()
	assert.Contains(t, out, "epsilon")
}

func TestExecProfileInfo_Console(t *testing.T) {
	home := setTestHome(t)
	hostname := "My Server"
	maxPlayers := uint16(64)
	p := profile.Profile{
		Name:       "zeta",
		InstallDir: "/opt/arma3",
		Config: arma.ServerConfig{
			Hostname:   &hostname,
			MaxPlayers: &maxPlayers,
		},
	}
	saveProfile(t, home, p)

	var buf bytes.Buffer
	require.NoError(t, execProfileInfo(context.Background(), "zeta", "console", &buf))

	out := buf.String()
	assert.Contains(t, out, "zeta")
	assert.Contains(t, out, "/opt/arma3")
	assert.Contains(t, out, "My Server")
	assert.Contains(t, out, "64")
}

func TestExecProfileInfo_NotFound(t *testing.T) {
	setTestHome(t)
	err := execProfileInfo(context.Background(), "ghost", "json", &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ghost")
}

// ── execProfileRm ─────────────────────────────────────────────────────────────

func TestExecProfileRm_NotFound(t *testing.T) {
	setTestHome(t)
	err := execProfileRm(context.Background(), "ghost", true, nil, &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ghost")
}

func TestExecProfileRm_Force_RemovesProfile(t *testing.T) {
	home := setTestHome(t)
	saveProfile(t, home, profile.Profile{Name: "srv", InstallDir: t.TempDir()})

	require.NoError(t, execProfileRm(context.Background(), "srv", true, nil, &bytes.Buffer{}))

	exists, err := storage.ProfileRepository{HomeDir: home}.Exists("srv")
	require.NoError(t, err)
	assert.False(t, exists)
}
