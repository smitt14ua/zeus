package profile

import "github.com/smitt14ua/zeus/internal/arma"

// ProfileHooks holds ordered shell commands for each lifecycle event.
// Commands run sequentially; the first failure aborts the remaining commands.
type ProfileHooks struct {
	PrePullMissions  []string `json:"prePullMissions,omitempty"  yaml:"pre_pull_missions,omitempty"  toml:"pre_pull_missions,omitempty"`
	PostPullMissions []string `json:"postPullMissions,omitempty" yaml:"post_pull_missions,omitempty" toml:"post_pull_missions,omitempty"`
	PostProfileAdd   []string `json:"postProfileAdd,omitempty"   yaml:"post_profile_add,omitempty"   toml:"post_profile_add,omitempty"`
	PreProfileStart  []string `json:"preProfileStart,omitempty"  yaml:"pre_profile_start,omitempty"  toml:"pre_profile_start,omitempty"`
	PostProfileStart []string `json:"postProfileStart,omitempty" yaml:"post_profile_start,omitempty" toml:"post_profile_start,omitempty"`
	PreProfileStop   []string `json:"preProfileStop,omitempty"   yaml:"pre_profile_stop,omitempty"   toml:"pre_profile_stop,omitempty"`
	PostProfileStop  []string `json:"postProfileStop,omitempty"  yaml:"post_profile_stop,omitempty"  toml:"post_profile_stop,omitempty"`
	PreProfileRun    []string `json:"preProfileRun,omitempty"    yaml:"pre_profile_run,omitempty"    toml:"pre_profile_run,omitempty"`
	PostProfileRun   []string `json:"postProfileRun,omitempty"   yaml:"post_profile_run,omitempty"   toml:"post_profile_run,omitempty"`
}

type Profile struct {
	Name          string                 `json:"name"                     yaml:"name"                     toml:"name"`
	Executable    string                 `json:"executable,omitempty"     yaml:"executable,omitempty"     toml:"executable,omitempty"`
	InstallDir    string                 `json:"install_dir"              yaml:"install_dir"              toml:"install_dir"`
	Params        arma.StartupParams     `json:"params"                   yaml:"params"                   toml:"params"`
	Config        arma.ServerConfig      `json:"config"                   yaml:"config"                   toml:"config"`
	Basic         arma.BasicServerConfig `json:"basic"                    yaml:"basic"                    toml:"basic"`
	MissionSource *arma.MissionSource    `json:"missionSource,omitempty"  yaml:"mission_source,omitempty" toml:"mission_source,omitempty"`
	RCon          RCon                   `json:"rcon"                     yaml:"rcon"                     toml:"rcon"`
	Hooks         *ProfileHooks          `json:"hooks,omitempty"          yaml:"hooks,omitempty"          toml:"hooks,omitempty"`
}
