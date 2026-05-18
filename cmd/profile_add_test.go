package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smitt14ua/zeus/internal/profile"
	"github.com/smitt14ua/zeus/internal/storage"
)

func marshalProfile(t *testing.T, p profile.Profile) []byte {
	t.Helper()
	data, err := json.Marshal(p)
	require.NoError(t, err)
	return data
}

func TestExecProfileAdd_JSON_SavesProfile(t *testing.T) {
	home := setTestHome(t)
	p := profile.Profile{Name: "new-srv", InstallDir: t.TempDir()}

	var buf bytes.Buffer
	require.NoError(t, execProfileAdd(context.Background(), bytes.NewReader(marshalProfile(t, p)), "json", "", false, true, &buf))

	assert.Contains(t, buf.String(), "new-srv")

	exists, err := storage.ProfileRepository{HomeDir: home}.Exists("new-srv")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestExecProfileAdd_NameOverride(t *testing.T) {
	home := setTestHome(t)
	p := profile.Profile{Name: "original", InstallDir: t.TempDir()}

	var buf bytes.Buffer
	require.NoError(t, execProfileAdd(context.Background(), bytes.NewReader(marshalProfile(t, p)), "json", "renamed", false, true, &buf))

	assert.Contains(t, buf.String(), "renamed")

	exists, err := storage.ProfileRepository{HomeDir: home}.Exists("renamed")
	require.NoError(t, err)
	assert.True(t, exists)

	// Original name must not be saved
	exists, err = storage.ProfileRepository{HomeDir: home}.Exists("original")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestExecProfileAdd_AlreadyExists_ShowsUpdateMessage(t *testing.T) {
	home := setTestHome(t)
	p := profile.Profile{Name: "existing", InstallDir: t.TempDir()}
	saveProfile(t, home, p)

	var buf bytes.Buffer
	require.NoError(t, execProfileAdd(context.Background(), bytes.NewReader(marshalProfile(t, p)), "json", "", false, true, &buf))

	assert.Contains(t, buf.String(), "updating")
}

func TestExecProfileAdd_PostHookMessage(t *testing.T) {
	setTestHome(t)
	p := profile.Profile{
		Name:       "hooked",
		InstallDir: t.TempDir(),
		Hooks: &profile.ProfileHooks{
			PostProfileAdd: []string{"echo hook_ran"},
		},
	}

	var buf bytes.Buffer
	require.NoError(t, execProfileAdd(context.Background(), bytes.NewReader(marshalProfile(t, p)), "json", "", false, true, &buf))

	assert.Contains(t, buf.String(), "Running post-add hooks...")
}

func TestExecProfileAdd_NoHookMessage_WhenNoHooks(t *testing.T) {
	setTestHome(t)
	p := profile.Profile{Name: "plain", InstallDir: t.TempDir()}

	var buf bytes.Buffer
	require.NoError(t, execProfileAdd(context.Background(), bytes.NewReader(marshalProfile(t, p)), "json", "", false, true, &buf))

	assert.NotContains(t, strings.ToLower(buf.String()), "hook")
}
