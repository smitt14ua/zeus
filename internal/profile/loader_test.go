package profile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smitt14ua/zeus/internal/arma"
)

const testProfileJSON = `{
  "name": "my-server",
  "install_dir": "/opt/arma3",
  "params": {
    "port": 2302,
    "server": true
  },
  "config": {
    "hostname": "Test Server",
    "maxPlayers": 32,
    "passwordAdmin": "secret"
  },
  "basic": {
    "language": "English",
    "maxMsgSend": 128
  },
  "rcon": {
    "password": "abc123def456abc1",
    "port": 2301
  }
}`

func TestProfileLoader_FromBytes(t *testing.T) {
	loader := ProfileLoader{}

	t.Run("valid_json", func(t *testing.T) {
		p, err := loader.FromBytes([]byte(testProfileJSON))
		require.NoError(t, err)
		assert.Equal(t, "my-server", p.Name)
		assert.Equal(t, "/opt/arma3", p.InstallDir)
		assert.Equal(t, uint16(2302), *p.Params.Port)
		assert.Equal(t, true, *p.Params.Server)
		require.NotNil(t, p.Config.Hostname)
		assert.Equal(t, "Test Server", *p.Config.Hostname)
		require.NotNil(t, p.Config.MaxPlayers)
		assert.Equal(t, uint16(32), *p.Config.MaxPlayers)
		require.NotNil(t, p.Config.PasswordAdmin)
		assert.Equal(t, "secret", *p.Config.PasswordAdmin)
		require.NotNil(t, p.Basic.Language)
		assert.Equal(t, "English", *p.Basic.Language)
		require.NotNil(t, p.Basic.MaxMsgSend)
		assert.Equal(t, uint16(128), *p.Basic.MaxMsgSend)
		assert.Equal(t, "abc123def456abc1", p.RCon.Password)
		assert.Equal(t, uint16(2301), p.RCon.Port)
	})

	t.Run("invalid_json", func(t *testing.T) {
		_, err := loader.FromBytes([]byte("{invalid json"))
		assert.Error(t, err)
	})

	t.Run("empty_bytes_returns_zero_profile", func(t *testing.T) {
		p, err := loader.FromBytes([]byte{})
		require.NoError(t, err)
		assert.Nil(t, p.Config.MaxPlayers)
		assert.Nil(t, p.Config.BattlEye)
		assert.Nil(t, p.Basic.MaxMsgSend)
	})
}

func TestProfileLoader_FromFile(t *testing.T) {
	loader := ProfileLoader{}

	t.Run("valid_file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "profile.json")
		require.NoError(t, os.WriteFile(path, []byte(testProfileJSON), 0600))

		p, err := loader.FromFile(path)
		require.NoError(t, err)
		assert.Equal(t, "my-server", p.Name)
		assert.Equal(t, "/opt/arma3", p.InstallDir)
		assert.Equal(t, uint16(2302), *p.Params.Port)
		require.NotNil(t, p.Config.Hostname)
		assert.Equal(t, "Test Server", *p.Config.Hostname)
		require.NotNil(t, p.Basic.Language)
		assert.Equal(t, "English", *p.Basic.Language)
		assert.Equal(t, "abc123def456abc1", p.RCon.Password)
		assert.Equal(t, uint16(2301), p.RCon.Port)
	})

	t.Run("missing_file", func(t *testing.T) {
		_, err := loader.FromFile(filepath.Join(t.TempDir(), "nonexistent.yaml"))
		assert.Error(t, err)
	})
}

const testProfileTOML = `
name = "my-server"
install_dir = "/opt/arma3"

[params]
port = 2302
server = true

[config]
hostname = "Test Server"
max_players = 32
password_admin = "secret"

[basic]
language = "English"
max_msg_send = 128

[rcon]
password = "abc123def456abc1"
port = 2301
`

func TestProfileLoader_FromFile_TOML(t *testing.T) {
	loader := ProfileLoader{}

	t.Run("valid_toml_file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "profile.toml")
		require.NoError(t, os.WriteFile(path, []byte(testProfileTOML), 0600))

		p, err := loader.FromFile(path)
		require.NoError(t, err)
		assert.Equal(t, "my-server", p.Name)
		assert.Equal(t, "/opt/arma3", p.InstallDir)
		assert.Equal(t, uint16(2302), *p.Params.Port)
		assert.Equal(t, true, *p.Params.Server)
		require.NotNil(t, p.Config.Hostname)
		assert.Equal(t, "Test Server", *p.Config.Hostname)
		require.NotNil(t, p.Config.MaxPlayers)
		assert.Equal(t, uint16(32), *p.Config.MaxPlayers)
		require.NotNil(t, p.Config.PasswordAdmin)
		assert.Equal(t, "secret", *p.Config.PasswordAdmin)
		require.NotNil(t, p.Basic.Language)
		assert.Equal(t, "English", *p.Basic.Language)
		require.NotNil(t, p.Basic.MaxMsgSend)
		assert.Equal(t, uint16(128), *p.Basic.MaxMsgSend)
		assert.Equal(t, "abc123def456abc1", p.RCon.Password)
		assert.Equal(t, uint16(2301), p.RCon.Port)
	})

	t.Run("toml_empty_returns_zero_profile", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "empty.toml")
		require.NoError(t, os.WriteFile(path, []byte{}, 0600))

		p, err := loader.FromFile(path)
		require.NoError(t, err)
		assert.Nil(t, p.Config.MaxPlayers)
		assert.Nil(t, p.Config.BattlEye)
	})

	t.Run("toml_bandwidth_string", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "bw.toml")
		require.NoError(t, os.WriteFile(path, []byte("[basic]\nmax_bandwidth = \"750Mbps\"\n"), 0600))

		p, err := loader.FromFile(path)
		require.NoError(t, err)
		assert.Equal(t, uint64(750_000_000), p.Basic.MaxBandwidth.Bps())
	})

	t.Run("toml_bandwidth_integer", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "bwint.toml")
		require.NoError(t, os.WriteFile(path, []byte("[basic]\nmax_bandwidth = 750000000\n"), 0600))

		p, err := loader.FromFile(path)
		require.NoError(t, err)
		assert.Equal(t, uint64(750_000_000), p.Basic.MaxBandwidth.Bps())
	})
}

// TestProfileLoader_OnlySetValuesAppearInDump verifies that only explicitly set
// fields appear in the generated cfg files; nil fields are omitted.
func TestProfileLoader_OnlySetValuesAppearInDump(t *testing.T) {
	loader := ProfileLoader{}

	const minimalJSON = `{
  "config": {
    "hostname": "My Server",
    "maxPlayers": 40
  },
  "basic": {
    "maxMsgSend": 64,
    "maxSizeGuaranteed": 1324,
    "maxSizeNonguaranteed": 1324,
    "maxBandwidth": "750Mbps"
  }
}`
	p, err := loader.FromBytes([]byte(minimalJSON))
	require.NoError(t, err)

	serverOut := string(arma.DumpServerConfig(p.Config))

	assert.Contains(t, serverOut, `hostname = "My Server";`)
	assert.Contains(t, serverOut, `maxPlayers = 40;`)

	// unset fields must not appear
	assert.NotContains(t, serverOut, `maxPing`)
	assert.NotContains(t, serverOut, `maxPacketLoss`)
	assert.NotContains(t, serverOut, `maxDesync`)
	assert.NotContains(t, serverOut, `disconnectTimeout`)
	assert.NotContains(t, serverOut, `verifySignatures`)
	assert.NotContains(t, serverOut, `BattlEye`)
	assert.NotContains(t, serverOut, `voteThreshold`)
	assert.NotContains(t, serverOut, `statisticsEnabled`)
	assert.NotContains(t, serverOut, `AdvancedOptions`)
	assert.NotContains(t, serverOut, `AntiFlood`)

	basicOut := string(arma.DumpBasicServerConfig(p.Basic))

	assert.Contains(t, basicOut, `MaxMsgSend = 64;`)
	assert.Contains(t, basicOut, `MaxSizeGuaranteed = 1324;`)
	assert.Contains(t, basicOut, `MaxSizeNonguaranteed = 1324;`)
	assert.Contains(t, basicOut, `MaxBandwidth`)

	assert.NotContains(t, basicOut, `MinBandwidth`)
	assert.NotContains(t, basicOut, `MinErrorToSend`)
	assert.NotContains(t, basicOut, `sockets`)
}

// TestProfileLoader_ExplicitZeroOverridesDefault verifies that a field
// explicitly set to 0 in the YAML correctly overrides a non-zero default.
func TestProfileLoader_ExplicitZeroOverridesDefault(t *testing.T) {
	loader := ProfileLoader{}

	const jsonInput = `{
  "config": {
    "maxPing": 0,
    "maxPacketLoss": 0,
    "maxDesync": 0
  }
}`
	p, err := loader.FromBytes([]byte(jsonInput))
	require.NoError(t, err)

	out := string(arma.DumpServerConfig(p.Config))

	assert.Contains(t, out, `maxPing = 0;`)
	assert.Contains(t, out, `maxPacketLoss = 0;`)
	assert.Contains(t, out, `maxDesync = 0;`)
}

func TestProfileLoader_MissionSource_JSON(t *testing.T) {
	loader := ProfileLoader{}
	const jsonInput = `{
  "name": "srv",
  "install_dir": "/opt/arma3",
  "missionSource": {
    "driver": "path",
    "path": "/opt/missions",
    "mode": "copy"
  }
}`
	p, err := loader.FromBytes([]byte(jsonInput))
	require.NoError(t, err)
	require.NotNil(t, p.MissionSource)
	assert.Equal(t, "path", p.MissionSource.Driver)
	assert.Equal(t, "/opt/missions", p.MissionSource.Path)
	assert.Equal(t, "copy", p.MissionSource.Mode)
}

func TestProfileLoader_MissionSource_Absent(t *testing.T) {
	loader := ProfileLoader{}
	p, err := loader.FromBytes([]byte(`{"name":"srv","install_dir":"/opt/arma3"}`))
	require.NoError(t, err)
	assert.Nil(t, p.MissionSource)
}

func TestProfileLoader_MissionSource_S3_JSON(t *testing.T) {
	loader := ProfileLoader{}
	const jsonInput = `{
  "name": "test-s3",
  "install_dir": "/tmp/arma",
  "missionSource": {
    "driver": "s3",
    "bucket": "mpmissions",
    "prefix": "servers/prod/",
    "endpoint": "http://localhost:9000",
    "region": "us-east-1",
    "accessKeyId": "minioadmin",
    "secretAccessKey": "minioadmin"
  }
}`
	p, err := loader.FromBytes([]byte(jsonInput))
	require.NoError(t, err)
	require.NotNil(t, p.MissionSource)
	assert.Equal(t, "s3", p.MissionSource.Driver)
	assert.Equal(t, "mpmissions", p.MissionSource.Bucket)
	assert.Equal(t, "servers/prod/", p.MissionSource.Prefix)
	assert.Equal(t, "http://localhost:9000", p.MissionSource.Endpoint)
	assert.Equal(t, "us-east-1", p.MissionSource.Region)
	assert.Equal(t, "minioadmin", p.MissionSource.AccessKeyID)
	assert.Equal(t, "minioadmin", p.MissionSource.SecretAccessKey)
}

func TestProfileLoader_MissionSource_S3_TOML(t *testing.T) {
	loader := ProfileLoader{}
	const toml = `
name = "test-s3"
install_dir = "/tmp/arma"

[mission_source]
driver = "s3"
bucket = "mpmissions"
prefix = "servers/prod/"
endpoint = "http://localhost:9000"
region = "us-east-1"
access_key_id = "minioadmin"
secret_access_key = "minioadmin"
`
	path := filepath.Join(t.TempDir(), "profile.toml")
	require.NoError(t, os.WriteFile(path, []byte(toml), 0600))

	p, err := loader.FromFile(path)
	require.NoError(t, err)
	require.NotNil(t, p.MissionSource)
	assert.Equal(t, "s3", p.MissionSource.Driver)
	assert.Equal(t, "mpmissions", p.MissionSource.Bucket)
	assert.Equal(t, "servers/prod/", p.MissionSource.Prefix)
	assert.Equal(t, "http://localhost:9000", p.MissionSource.Endpoint)
	assert.Equal(t, "us-east-1", p.MissionSource.Region)
	assert.Equal(t, "minioadmin", p.MissionSource.AccessKeyID)
	assert.Equal(t, "minioadmin", p.MissionSource.SecretAccessKey)
}
