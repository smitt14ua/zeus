package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smitt14ua/zeus/internal/profile"
)

func TestExecMissionsPull_ProfileNotFound(t *testing.T) {
	setTestHome(t)
	err := execMissionsPull(context.Background(), "ghost", false, &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ghost")
}

func TestExecMissionsPull_NoMissionSource(t *testing.T) {
	home := setTestHome(t)
	saveProfile(t, home, profile.Profile{Name: "srv", InstallDir: t.TempDir()})

	err := execMissionsPull(context.Background(), "srv", false, &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mission_source")
}
