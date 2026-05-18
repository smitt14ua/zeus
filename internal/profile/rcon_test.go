package profile

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/smitt14ua/zeus/internal/arma"
)

func TestDefaultRCon(t *testing.T) {
	t.Run("password_is_md5_hex", func(t *testing.T) {
		r := DefaultRCon("my-server", 2302)
		assert.Len(t, r.Password, 32, "MD5 hex is 32 chars")
		assert.True(t, isHex(r.Password), "password must be hex string")
	})

	t.Run("port_is_game_port_minus_one", func(t *testing.T) {
		r := DefaultRCon("x", 2302)
		assert.Equal(t, uint16(2301), r.Port)

		r2 := DefaultRCon("x", 2402)
		assert.Equal(t, uint16(2401), r2.Port)
	})

	t.Run("ip_empty_by_default", func(t *testing.T) {
		r := DefaultRCon("x", 2302)
		assert.Empty(t, r.IP)
	})

	t.Run("deterministic", func(t *testing.T) {
		r1 := DefaultRCon("server", 2302)
		r2 := DefaultRCon("server", 2302)
		assert.Equal(t, r1.Password, r2.Password)
	})

	t.Run("different_names_different_passwords", func(t *testing.T) {
		r1 := DefaultRCon("alpha", 2302)
		r2 := DefaultRCon("beta", 2302)
		assert.NotEqual(t, r1.Password, r2.Password)
	})

	t.Run("zero_game_port_uses_default", func(t *testing.T) {
		r := DefaultRCon("srv", 0)
		assert.Equal(t, arma.DefaultPort-1, r.Port, "port should be DefaultPort-1, not uint16 underflow")
	})
}

func TestDumpBEServerConfig(t *testing.T) {
	t.Run("contains_all_fields", func(t *testing.T) {
		r := RCon{Password: "secret", Port: 2301, IP: "127.0.0.1"}
		out := string(DumpBEServerConfig(r))
		assert.Contains(t, out, "RConPassword secret")
		assert.Contains(t, out, "RConPort 2301")
		assert.Contains(t, out, "RConIP 127.0.0.1")
	})

	t.Run("empty_ip_defaults_to_0000", func(t *testing.T) {
		r := RCon{Password: "pass", Port: 2301}
		out := string(DumpBEServerConfig(r))
		assert.Contains(t, out, "RConIP 0.0.0.0")
	})
}

func isHex(s string) bool {
	return strings.Trim(s, "0123456789abcdef") == ""
}
