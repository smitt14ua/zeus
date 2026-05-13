package profile

import "github.com/smitt14ua/zeus/internal/arma"

type Profile struct {
	Name          string                 `json:"name"                    yaml:"name"                    toml:"name"`
	Executable    string                 `json:"executable,omitempty"    yaml:"executable,omitempty"    toml:"executable,omitempty"`
	InstallDir    string                 `json:"install_dir"             yaml:"install_dir"             toml:"install_dir"`
	Params        arma.StartupParams     `json:"params"                  yaml:"params"                  toml:"params"`
	Config        arma.ServerConfig      `json:"config"                  yaml:"config"                  toml:"config"`
	Basic         arma.BasicServerConfig `json:"basic"                   yaml:"basic"                   toml:"basic"`
	MissionSource *arma.MissionSource    `json:"missionSource,omitempty" yaml:"mission_source,omitempty" toml:"mission_source,omitempty"`
	RCon          RCon                   `json:"rcon"                    yaml:"rcon"                    toml:"rcon"`
}
