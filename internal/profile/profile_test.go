package profile

import (
	"testing"

	"github.com/smitt14ua/zeus/internal/arma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

func ptrOf[T any](v T) *T { return &v }

func TestProfile_YAMLRoundTrip(t *testing.T) {
	port := uint16(2302)
	server := true

	original := Profile{
		Name:       "my-server",
		Executable: "arma3server_custom",
		InstallDir: "/opt/arma3",
		Params: arma.StartupParams{
			Port:   &port,
			Server: &server,
		},
		Config: arma.ServerConfig{
			Hostname:      ptrOf("Test Server"),
			MaxPlayers:    ptrOf(uint16(32)),
			PasswordAdmin: ptrOf("secret"),
		},
		Basic: arma.BasicServerConfig{
			Language:   ptrOf("English"),
			MaxMsgSend: ptrOf(uint16(128)),
		},
	}

	data, err := yaml.Marshal(original)
	require.NoError(t, err)

	var result Profile
	require.NoError(t, yaml.Unmarshal(data, &result))

	assert.Equal(t, original.Name, result.Name)
	assert.Equal(t, original.Executable, result.Executable)
	assert.Equal(t, original.InstallDir, result.InstallDir)
	assert.Equal(t, *original.Params.Port, *result.Params.Port)
	assert.Equal(t, *original.Params.Server, *result.Params.Server)
	assert.Equal(t, original.Config.Hostname, result.Config.Hostname)
	assert.Equal(t, original.Config.MaxPlayers, result.Config.MaxPlayers)
	assert.Equal(t, original.Config.PasswordAdmin, result.Config.PasswordAdmin)
	assert.Equal(t, original.Basic.Language, result.Basic.Language)
	assert.Equal(t, original.Basic.MaxMsgSend, result.Basic.MaxMsgSend)
}

func TestProfile_ZeroValue(t *testing.T) {
	data, err := yaml.Marshal(Profile{})
	require.NoError(t, err)

	var result Profile
	require.NoError(t, yaml.Unmarshal(data, &result))

	assert.Empty(t, result.Name)
	assert.Empty(t, result.InstallDir)
	assert.Nil(t, result.Params.Port)
	assert.Nil(t, result.Config.Hostname)
	assert.Nil(t, result.Basic.Language)
}
