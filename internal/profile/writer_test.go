package profile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/smitt14ua/zeus/internal/arma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileWriter_Write(t *testing.T) {
	w := ProfileWriter{}

	t.Run("creates_dir_and_files", func(t *testing.T) {
		p := Profile{
			Name:       "my-server",
			InstallDir: t.TempDir(),
			Config: arma.ServerConfig{
				Hostname: ptrOf("Test Server"),
			},
			Basic: arma.BasicServerConfig{
				MaxMsgSend: ptrOf(uint16(256)),
			},
		}

		require.NoError(t, w.Write(p))

		cfgDir := filepath.Join(p.InstallDir, ".zeus", "my-server", "configs")

		serverCfg, err := os.ReadFile(filepath.Join(cfgDir, "server.cfg"))
		require.NoError(t, err)
		assert.Contains(t, string(serverCfg), `hostname = "Test Server"`)

		basicCfg, err := os.ReadFile(filepath.Join(cfgDir, "basic.cfg"))
		require.NoError(t, err)
		assert.Contains(t, string(basicCfg), `MaxMsgSend = 256`)
	})

	t.Run("idempotent", func(t *testing.T) {
		p := Profile{
			Name:       "my-server",
			InstallDir: t.TempDir(),
		}

		require.NoError(t, w.Write(p))
		require.NoError(t, w.Write(p))
	})

	t.Run("zeus_dir_nested_in_install_dir", func(t *testing.T) {
		base := t.TempDir()
		p := Profile{
			Name:       "alpha",
			InstallDir: base,
		}

		require.NoError(t, w.Write(p))

		assert.DirExists(t, filepath.Join(base, ".zeus", "alpha"))
		assert.FileExists(t, filepath.Join(base, ".zeus", "alpha", "configs", "server.cfg"))
		assert.FileExists(t, filepath.Join(base, ".zeus", "alpha", "configs", "basic.cfg"))
	})

	t.Run("optionalkeys_dir_created", func(t *testing.T) {
		base := t.TempDir()
		p := Profile{Name: "srv", InstallDir: base}

		require.NoError(t, w.Write(p))

		assert.DirExists(t, filepath.Join(base, ".zeus", "srv", "keys"))
		assert.DirExists(t, filepath.Join(base, ".zeus", "srv", "optionalkeys"))
	})

	t.Run("optionalkeys_files_preserved_on_update", func(t *testing.T) {
		base := t.TempDir()
		p := Profile{Name: "srv", InstallDir: base}

		require.NoError(t, w.Write(p))

		sentinel := filepath.Join(base, ".zeus", "srv", "optionalkeys", "mymod.bikey")
		require.NoError(t, os.WriteFile(sentinel, []byte("key"), 0644))

		require.NoError(t, w.Write(p))

		assert.FileExists(t, sentinel)
	})

	t.Run("battleye_dir_and_configs_created", func(t *testing.T) {
		p := Profile{
			Name:       "be-server",
			InstallDir: t.TempDir(),
			RCon:       RCon{Password: "testpass", Port: 2301},
		}

		require.NoError(t, w.Write(p))

		beDir := filepath.Join(p.InstallDir, ".zeus", "be-server", BattleyeDirName())
		assert.DirExists(t, beDir)

		for _, name := range []string{"BEServer.cfg", "BEServer_x64.cfg"} {
			data, err := os.ReadFile(filepath.Join(beDir, name))
			require.NoError(t, err)
			assert.Contains(t, string(data), "RConPassword testpass")
			assert.Contains(t, string(data), "RConPort 2301")
		}
	})

	t.Run("battleye_uses_default_rcon_when_password_empty", func(t *testing.T) {
		p := Profile{
			Name:       "auto-rcon",
			InstallDir: t.TempDir(),
		}

		require.NoError(t, w.Write(p))

		beDir := filepath.Join(p.InstallDir, ".zeus", "auto-rcon", BattleyeDirName())
		data, err := os.ReadFile(filepath.Join(beDir, "BEServer.cfg"))
		require.NoError(t, err)
		assert.Contains(t, string(data), "RConPassword ")
		assert.Contains(t, string(data), "RConPort 2301")
	})
}
