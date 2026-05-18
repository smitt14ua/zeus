package cmd

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smitt14ua/zeus/internal/arma"
	"github.com/smitt14ua/zeus/internal/profile"
)

// ── effectivePort ─────────────────────────────────────────────────────────────

func TestEffectivePort_ExplicitPort(t *testing.T) {
	port := uint16(2400)
	p := profile.Profile{Params: arma.StartupParams{Port: &port}}
	assert.Equal(t, uint16(2400), effectivePort(p))
}

func TestEffectivePort_DefaultWhenNil(t *testing.T) {
	assert.Equal(t, arma.DefaultPort, effectivePort(profile.Profile{}))
}

func TestEffectivePort_ZeroTreatedAsDefault(t *testing.T) {
	port := uint16(0)
	p := profile.Profile{Params: arma.StartupParams{Port: &port}}
	assert.Equal(t, arma.DefaultPort, effectivePort(p))
}

// ── execProfileStart error paths ──────────────────────────────────────────────

func TestExecProfileStart_ProfileNotFound(t *testing.T) {
	setTestHome(t)
	err := execProfileStart(context.Background(), "ghost", false, 30*time.Second, &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ghost")
}

func TestExecProfileStart_DryRun_ProfileNotFound(t *testing.T) {
	setTestHome(t)
	err := execProfileStart(context.Background(), "ghost", true, 30*time.Second, &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ghost")
}

// ── execProfileStop error paths ───────────────────────────────────────────────

func TestExecProfileStop_ProfileNotFound(t *testing.T) {
	setTestHome(t)
	err := execProfileStop(context.Background(), "ghost", &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ghost")
}

func TestExecProfileStop_NotRunning(t *testing.T) {
	home := setTestHome(t)
	saveProfile(t, home, profile.Profile{Name: "srv", InstallDir: t.TempDir()})

	err := execProfileStop(context.Background(), "srv", &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}
