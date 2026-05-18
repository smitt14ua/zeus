package profile

import (
	"crypto/md5"
	"fmt"

	"github.com/smitt14ua/zeus/internal/arma"
)

type RCon struct {
	Password string `json:"password" yaml:"password" toml:"password"`
	Port     uint16 `json:"port"     yaml:"port"     toml:"port"`
	IP       string `json:"ip,omitempty" yaml:"ip,omitempty" toml:"ip,omitempty"`
}

// DefaultRCon generates RCon defaults: MD5 password from profile name, port = gamePort-1.
// Falls back to arma3 default game port when gamePort is 0.
func DefaultRCon(name string, gamePort uint16) RCon {
	if gamePort == 0 {
		gamePort = arma.DefaultPort
	}
	sum := md5.Sum([]byte(name))
	return RCon{
		Password: fmt.Sprintf("%x", sum),
		Port:     gamePort - 1,
	}
}

// DumpBEServerConfig formats RCon settings for a BattlEye config file.
func DumpBEServerConfig(r RCon) []byte {
	var ip string
	if r.IP != "" {
		ip = r.IP
	} else {
		ip = "0.0.0.0"
	}
	return []byte(fmt.Sprintf("RConPassword %s\nRConPort %d\nRConIP %s\n", r.Password, r.Port, ip))
}
