package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"

	"github.com/smitt14ua/zeus/internal/arma"
	"github.com/smitt14ua/zeus/internal/profile"
)

// buildNewProfile mirrors the Profile construction in profileNewCmd.Run.
func buildNewProfile(name string) profile.Profile {
	hostname := name + " server"
	p := profile.Profile{
		Name:   name,
		Config: arma.NewDefaultServerConfig(),
		Basic:  arma.NewDefaultBasicServerConfig(),
		RCon:   profile.DefaultRCon(name, 2302),
	}
	p.Config.Hostname = &hostname
	p.Config.Motd = []string{"Welcome to the " + name + " server!"}
	p.Config.Admins = []string{"00000000000000000"}
	return p
}

func TestProfileNew_TOML_ContainsRCon(t *testing.T) {
	p := buildNewProfile("test-server")
	var buf bytes.Buffer
	require.NoError(t, toml.NewEncoder(&buf).Encode(p))
	out := buf.String()

	assert.Contains(t, out, "[rcon]")
	assert.Contains(t, out, "password =")
	assert.Contains(t, out, "port =")
}

func TestProfileNew_JSON_ContainsRCon(t *testing.T) {
	p := buildNewProfile("test-server")
	data, err := json.MarshalIndent(p, "", "  ")
	require.NoError(t, err)

	var top map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &top))
	rconRaw, ok := top["rcon"]
	require.True(t, ok, "rcon key missing from JSON output")

	var rcon struct {
		Password string `json:"password"`
		Port     uint16 `json:"port"`
	}
	require.NoError(t, json.Unmarshal(rconRaw, &rcon))
	assert.NotEmpty(t, rcon.Password)
	assert.NotZero(t, rcon.Port)
}

func TestProfileNew_YAML_ContainsRCon(t *testing.T) {
	p := buildNewProfile("test-server")
	data, err := yaml.Marshal(p)
	require.NoError(t, err)
	out := string(data)

	assert.Contains(t, out, "rcon:")
	assert.Contains(t, out, "password:")
	assert.Contains(t, out, "port:")
}

func TestProfileNew_RCon_DeterministicPassword(t *testing.T) {
	p1 := buildNewProfile("server-a")
	p2 := buildNewProfile("server-b")
	p3 := buildNewProfile("server-a")

	assert.NotEqual(t, p1.RCon.Password, p2.RCon.Password, "different names must produce different passwords")
	assert.Equal(t, p1.RCon.Password, p3.RCon.Password, "same name must produce same password")
}

func TestProfileNew_RCon_Port(t *testing.T) {
	p := buildNewProfile("my-server")
	assert.Equal(t, uint16(2301), p.RCon.Port)
}
