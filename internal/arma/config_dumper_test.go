package arma

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDumpServerConfig_EmptyProducesEmpty(t *testing.T) {
	out := string(DumpServerConfig(ServerConfig{}))
	assert.Empty(t, strings.TrimSpace(out))
}

func TestDumpServerConfig_SetScalars(t *testing.T) {
	var cfg ServerConfig
	cfg.MaxPlayers = ptr(uint16(32))
	cfg.DisconnectTimeout = ptr(uint8(5))
	cfg.CallExtReportLimit = ptr(uint32(2000))
	cfg.LobbyIdleTimeout = ptr(uint32(600))
	out := string(DumpServerConfig(cfg))

	assert.Contains(t, out, `maxPlayers = 32;`)
	assert.Contains(t, out, `disconnectTimeout = 5;`)
	assert.Contains(t, out, `callExtReportLimit = 2000;`)
	assert.Contains(t, out, `lobbyIdleTimeout = 600;`)
	assert.NotContains(t, out, `maxPing`)
	assert.NotContains(t, out, `verifySignatures`)
}

func TestDumpServerConfig_SetBooleans(t *testing.T) {
	var cfg ServerConfig
	cfg.Persistent = ptr(true)
	cfg.DisableVoN = ptr(true)
	cfg.BattlEye = ptr(false)
	cfg.DrawingInMap = ptr(false)
	out := string(DumpServerConfig(cfg))

	assert.Contains(t, out, `persistent = true;`)
	assert.Contains(t, out, `disableVoN = true;`)
	assert.Contains(t, out, `BattlEye = false;`)
	assert.Contains(t, out, `drawingInMap = false;`)
	assert.NotContains(t, out, `statisticsEnabled`)
	assert.NotContains(t, out, `allowProfileGlasses`)
}

func TestDumpServerConfig_SetStrings(t *testing.T) {
	var cfg ServerConfig
	cfg.Hostname = ptr("My Server")
	cfg.Password = ptr("secret")
	cfg.LogFile = ptr("server.log")
	cfg.TimeStampFormat = ptr(TimestampShort)
	cfg.ForcedDifficulty = ptr(DifficultyVeteran)
	cfg.OnUnsignedData = ptr("kick (_this select 0);")
	out := string(DumpServerConfig(cfg))

	assert.Contains(t, out, `hostname = "My Server";`)
	assert.Contains(t, out, `password = "secret";`)
	assert.Contains(t, out, `logFile = "server.log";`)
	assert.Contains(t, out, `timeStampFormat = "short";`)
	assert.Contains(t, out, `forcedDifficulty = "veteran";`)
	assert.Contains(t, out, `onUnsignedData = "kick (_this select 0);";`)
	assert.NotContains(t, out, `passwordAdmin`)
	assert.NotContains(t, out, `serverCommandPassword`)
}

func TestDumpServerConfig_SetSlices(t *testing.T) {
	var cfg ServerConfig
	cfg.Admins = []string{"11111111111111111"}
	cfg.Motd = []string{"Welcome", "Have fun"}
	cfg.AllowedLoadFileExtensions = []string{"sqf", "hpp"}
	out := string(DumpServerConfig(cfg))

	assert.Contains(t, out, `admins[] = {"11111111111111111"};`)
	assert.Contains(t, out, `motd[] = {"Welcome", "Have fun"};`)
	assert.Contains(t, out, `allowedLoadFileExtensions[] = {"sqf", "hpp"};`)
	assert.NotContains(t, out, `headlessClients`)
	assert.NotContains(t, out, `missionWhitelist`)
}

func TestDumpServerConfig_NetworkThresholds(t *testing.T) {
	var cfg ServerConfig
	cfg.MaxPing = ptr(int32(200))
	cfg.MaxPacketLoss = ptr(int32(50))
	cfg.MaxDesync = ptr(int32(150))
	out := string(DumpServerConfig(cfg))

	assert.Contains(t, out, `maxPing = 200;`)
	assert.Contains(t, out, `maxPacketLoss = 50;`)
	assert.Contains(t, out, `maxDesync = 150;`)
}

func TestDumpServerConfig_KickTimeout(t *testing.T) {
	var cfg ServerConfig
	cfg.KickTimeout = []KickTimeout{
		{KickID: KickSourceManual, Timeout: KickTimeoutUntilMissionEnd},
		{KickID: KickSourceConnectivity, Timeout: 180},
		{KickID: KickSourceBattlEye, Timeout: 180},
		{KickID: KickSourceHarmless, Timeout: 180},
	}
	out := string(DumpServerConfig(cfg))

	assert.Contains(t, out, `kickTimeout[] = {{0, -1}, {1, 180}, {2, 180}, {3, 180}};`)
}

func TestDumpServerConfig_KickClientsOnSlowNetwork(t *testing.T) {
	arr := [4]uint8{0, 0, 0, 0}
	var cfg ServerConfig
	cfg.KickClientsOnSlowNetwork = &arr
	out := string(DumpServerConfig(cfg))

	assert.Contains(t, out, `kickClientsOnSlowNetwork[] = {0, 0, 0, 0};`)
}

func TestDumpServerConfig_Timeouts(t *testing.T) {
	var cfg ServerConfig
	cfg.VotingTimeOut = []uint32{30}
	cfg.RoleTimeOut = []uint32{60, 120}
	out := string(DumpServerConfig(cfg))

	assert.Contains(t, out, `votingTimeOut[] = {30};`)
	assert.Contains(t, out, `roleTimeOut[] = {60, 120};`)
	assert.NotContains(t, out, `briefingTimeOut`)
	assert.NotContains(t, out, `debriefingTimeOut`)
}

func TestDumpServerConfig_MissionsNilOmitted(t *testing.T) {
	out := string(DumpServerConfig(ServerConfig{}))
	assert.NotContains(t, out, "Missions")
}

func TestDumpServerConfig_MissionsEmptySliceEmitsEmptyClass(t *testing.T) {
	var cfg ServerConfig
	cfg.Missions = []Mission{}
	out := string(DumpServerConfig(cfg))
	assert.Contains(t, out, "class Missions {};")
}

func TestDumpServerConfig_MissionRotation(t *testing.T) {
	var cfg ServerConfig
	cfg.Missions = []Mission{
		{
			Template:   "MP_Marksmen_01.Altis",
			Difficulty: DifficultyRecruit,
			Params:     map[string]int{"RespawnDelay": 15},
		},
		{
			Template:   "EscapeFromMalden.Malden",
			Difficulty: DifficultyRegular,
		},
	}
	out := string(DumpServerConfig(cfg))

	assert.Contains(t, out, "class Missions")
	assert.Contains(t, out, "class Mission1")
	assert.Contains(t, out, `template = "MP_Marksmen_01.Altis";`)
	assert.Contains(t, out, `difficulty = "recruit";`)
	assert.Contains(t, out, "class Params")
	assert.Contains(t, out, "RespawnDelay = 15;")
	assert.Contains(t, out, "class Mission2")
	assert.Contains(t, out, `template = "EscapeFromMalden.Malden";`)
	assert.Contains(t, out, `difficulty = "regular";`)
}

func TestDumpServerConfig_MissionCustomName(t *testing.T) {
	var cfg ServerConfig
	cfg.Missions = []Mission{
		{CustomName: "TestMission01", Template: "MP_Marksmen_01.Altis", Difficulty: DifficultyRecruit},
		{Template: "EscapeFromMalden.Malden", Difficulty: DifficultyRegular},
	}
	out := string(DumpServerConfig(cfg))

	assert.Contains(t, out, "class TestMission01")
	assert.Contains(t, out, "class Mission2")
	assert.NotContains(t, out, "class Mission1")
}

func TestDumpServerConfig_DisabledChannels(t *testing.T) {
	var cfg ServerConfig
	cfg.DisableChannels = []DisabledChannel{
		{ChannelID: ChannelGlobal, Text: false, Voice: true, MapMarkers: false, DrawOnMap: true},
		{ChannelID: ChannelGroup, Text: true, Voice: true},
	}
	out := string(DumpServerConfig(cfg))

	assert.Contains(t, out, `disableChannels[] = {{0, false, true, false, true}, {3, true, true, false, false}};`)
}

func TestDumpServerConfig_AdvancedOptionsPartialChange(t *testing.T) {
	var cfg ServerConfig
	cfg.AdvancedOptions = &AdvancedOptions{
		SkipDescriptionParsing: ptr(true),
	}
	out := string(DumpServerConfig(cfg))

	assert.Contains(t, out, "class AdvancedOptions")
	assert.Contains(t, out, "skipDescriptionParsing = true;")
	assert.NotContains(t, out, "logObjectNotFound")
	assert.NotContains(t, out, "queueSizeLogG")
}

func TestDumpServerConfig_AdvancedOptionsDefaultOmitted(t *testing.T) {
	out := string(DumpServerConfig(ServerConfig{}))
	assert.NotContains(t, out, "AdvancedOptions")
}

func TestDumpServerConfig_AntiFloodPartialChange(t *testing.T) {
	var cfg ServerConfig
	cfg.AntiFlood = &AntiFlood{
		EnableKick: ptr(true),
	}
	out := string(DumpServerConfig(cfg))

	assert.Contains(t, out, "class AntiFlood")
	assert.Contains(t, out, "enableKick = true;")
	assert.NotContains(t, out, "cycleTime")
	assert.NotContains(t, out, "cycleLimit")
}

func TestDumpServerConfig_AntiFloodDefaultOmitted(t *testing.T) {
	out := string(DumpServerConfig(ServerConfig{}))
	assert.NotContains(t, out, "AntiFlood")
}

func TestDumpServerConfig_Indentation(t *testing.T) {
	var cfg ServerConfig
	cfg.AdvancedOptions = &AdvancedOptions{SkipDescriptionParsing: ptr(true)}
	cfg.AntiFlood = &AntiFlood{EnableKick: ptr(true)}
	out := string(DumpServerConfig(cfg))

	assert.Contains(t, out, "\tskipDescriptionParsing = true;")
	assert.Contains(t, out, "\tenableKick = true;")
}

func TestDumpServerConfig_DefaultValuesAppear(t *testing.T) {
	out := string(DumpServerConfig(NewDefaultServerConfig()))

	assert.Contains(t, out, `maxPlayers = 64;`)
	assert.Contains(t, out, `voteThreshold = 0.5;`)
	assert.Contains(t, out, `BattlEye = true;`)
	assert.Contains(t, out, `drawingInMap = true;`)
	assert.Contains(t, out, `statisticsEnabled = true;`)
}

// --- BasicServerConfig ---

func TestDumpBasicServerConfig_EmptyProducesEmpty(t *testing.T) {
	out := string(DumpBasicServerConfig(BasicServerConfig{}))
	assert.Empty(t, strings.TrimSpace(out))
}

func TestDumpBasicServerConfig_SetFields(t *testing.T) {
	var cfg BasicServerConfig
	cfg.MaxMsgSend = ptr(uint16(256))
	cfg.MinBandwidth = ptr(NewDataTransferRate(768000))
	cfg.MaxBandwidth = ptr(NewDataTransferRate(10000000000))
	cfg.MaxCustomFileSize = ptr(uint16(0))
	out := string(DumpBasicServerConfig(cfg))

	assert.Contains(t, out, `MaxMsgSend = 256;`)
	assert.Contains(t, out, `MinBandwidth = "750Kibps";`)
	assert.Contains(t, out, `MaxBandwidth = "9765625Kibps";`)
	assert.Contains(t, out, `MaxCustomFileSize = 0;`)
	assert.NotContains(t, out, `MaxSizeGuaranteed`)
	assert.NotContains(t, out, `MinErrorToSend`)
}

func TestDumpBasicServerConfig_SocketsDefaultOmitted(t *testing.T) {
	out := string(DumpBasicServerConfig(BasicServerConfig{}))
	assert.NotContains(t, out, "sockets")
}

func TestDumpBasicServerConfig_SocketsSet(t *testing.T) {
	var cfg BasicServerConfig
	cfg.Sockets = &Sockets{MaxPacketSize: ptr(uint16(1200))}
	out := string(DumpBasicServerConfig(cfg))

	assert.Contains(t, out, "class sockets")
	assert.Contains(t, out, "\tmaxPacketSize = 1200;")
}

func TestDumpBasicServerConfig_LanguageSet(t *testing.T) {
	var cfg BasicServerConfig
	cfg.Language = ptr("Czech")
	out := string(DumpBasicServerConfig(cfg))
	assert.Contains(t, out, `language = "Czech";`)
}

func TestDumpBasicServerConfig_DefaultValuesAppear(t *testing.T) {
	out := string(DumpBasicServerConfig(NewDefaultBasicServerConfig()))

	assert.Contains(t, out, `MaxMsgSend = 128;`)
	assert.Contains(t, out, `MaxSizeGuaranteed = 512;`)
	assert.Contains(t, out, `class sockets`)
}
