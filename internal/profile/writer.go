package profile

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/smitt14ua/zeus/internal/arma"
)

// BattleyeDirName returns the BattlEye subdirectory name for the current OS.
// Linux uses lowercase "battleye" because its filesystem is case-sensitive and
// the Arma 3 Linux server expects the directory name in lowercase.
func BattleyeDirName() string {
	if runtime.GOOS == "linux" {
		return "battleye"
	}
	return "BattlEye"
}

type ProfileWriter struct{}

func (w ProfileWriter) Write(p Profile) error {
	dir := filepath.Join(p.InstallDir, ".zeus", p.Name)

	for _, sub := range []string{"configs", "mpmissions", "keys", "optionalkeys", BattleyeDirName()} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0700); err != nil {
			return err
		}
	}

	cfgDir := filepath.Join(dir, "configs")
	if err := os.WriteFile(filepath.Join(cfgDir, "server.cfg"), arma.DumpServerConfig(p.Config), 0600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "basic.cfg"), arma.DumpBasicServerConfig(p.Basic), 0600); err != nil {
		return err
	}

	rcon := p.RCon
	if rcon.Password == "" {
		gamePort := arma.DefaultPort
		if p.Params.Port != nil {
			gamePort = *p.Params.Port
		}
		rcon = DefaultRCon(p.Name, gamePort)
	}
	beData := DumpBEServerConfig(rcon)
	beDir := filepath.Join(dir, BattleyeDirName())
	if err := os.WriteFile(filepath.Join(beDir, "BEServer.cfg"), beData, 0600); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(beDir, "BEServer_x64.cfg"), beData, 0600)
}
