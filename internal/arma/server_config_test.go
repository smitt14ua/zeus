package arma

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

func TestNewDefaultServerConfig(t *testing.T) {
	cfg := NewDefaultServerConfig()

	require.NotNil(t, cfg.MaxPlayers)
	assert.Equal(t, uint16(64), *cfg.MaxPlayers)
	require.NotNil(t, cfg.MotdInterval)
	assert.Equal(t, uint16(5), *cfg.MotdInterval)
	require.NotNil(t, cfg.VoteThreshold)
	assert.Equal(t, float32(0.5), *cfg.VoteThreshold)
	require.NotNil(t, cfg.VoteMissionPlayers)
	assert.Equal(t, uint16(1), *cfg.VoteMissionPlayers)
	require.NotNil(t, cfg.MaxPing)
	assert.Equal(t, int32(-1), *cfg.MaxPing)
	require.NotNil(t, cfg.MaxPacketLoss)
	assert.Equal(t, int32(-1), *cfg.MaxPacketLoss)
	require.NotNil(t, cfg.MaxDesync)
	assert.Equal(t, int32(-1), *cfg.MaxDesync)
	require.NotNil(t, cfg.DisconnectTimeout)
	assert.Equal(t, uint8(15), *cfg.DisconnectTimeout)
	require.NotNil(t, cfg.KickClientsOnSlowNetwork)
	assert.Equal(t, [4]uint8{1, 1, 1, 1}, *cfg.KickClientsOnSlowNetwork)
	assert.Equal(t, []KickTimeout{{0, 60}, {1, 60}, {2, 60}, {3, 60}}, cfg.KickTimeout)
	require.NotNil(t, cfg.CallExtReportLimit)
	assert.Equal(t, uint32(1000), *cfg.CallExtReportLimit)
	assert.Equal(t, []uint32{60, 90}, cfg.VotingTimeOut)
	assert.Equal(t, []uint32{90, 120}, cfg.RoleTimeOut)
	assert.Equal(t, []uint32{60, 90}, cfg.BriefingTimeOut)
	assert.Equal(t, []uint32{45, 60}, cfg.DebriefingTimeOut)
	require.NotNil(t, cfg.VerifySignatures)
	assert.Equal(t, SignaturesV2Only, *cfg.VerifySignatures)
	require.NotNil(t, cfg.BattlEye)
	assert.True(t, *cfg.BattlEye)
	require.NotNil(t, cfg.DrawingInMap)
	assert.True(t, *cfg.DrawingInMap)
	require.NotNil(t, cfg.VonCodecQuality)
	assert.Equal(t, uint8(3), *cfg.VonCodecQuality)
	require.NotNil(t, cfg.VonCodec)
	assert.Equal(t, VonCodecOPUS, *cfg.VonCodec)
	require.NotNil(t, cfg.AllowProfileGlasses)
	assert.True(t, *cfg.AllowProfileGlasses)
	require.NotNil(t, cfg.ZeusCompositionScriptLevel)
	assert.Equal(t, ZeusScriptsAttributesOnly, *cfg.ZeusCompositionScriptLevel)
	require.NotNil(t, cfg.StatisticsEnabled)
	assert.True(t, *cfg.StatisticsEnabled)
	require.NotNil(t, cfg.SteamProtocolMaxDataSize)
	assert.Equal(t, uint64(1024), cfg.SteamProtocolMaxDataSize.B())
	require.NotNil(t, cfg.ArmaUnitsTimeout)
	assert.Equal(t, uint32(30), *cfg.ArmaUnitsTimeout)
	require.NotNil(t, cfg.OverrideHazeQuality)
	assert.Equal(t, HazeQualityOff, *cfg.OverrideHazeQuality)

	require.NotNil(t, cfg.AdvancedOptions)
	require.NotNil(t, cfg.AdvancedOptions.LogObjectNotFound)
	assert.True(t, *cfg.AdvancedOptions.LogObjectNotFound)
	assert.Nil(t, cfg.AdvancedOptions.SkipDescriptionParsing)
	assert.Nil(t, cfg.AdvancedOptions.IgnoreMissionLoadErrors)
	assert.Nil(t, cfg.AdvancedOptions.QueueSizeLogG)

	require.NotNil(t, cfg.AntiFlood)
	require.NotNil(t, cfg.AntiFlood.CycleTime)
	assert.Equal(t, float32(0.5), *cfg.AntiFlood.CycleTime)
	require.NotNil(t, cfg.AntiFlood.CycleLimit)
	assert.Equal(t, uint32(400), *cfg.AntiFlood.CycleLimit)
	require.NotNil(t, cfg.AntiFlood.CycleHardLimit)
	assert.Equal(t, uint32(4000), *cfg.AntiFlood.CycleHardLimit)
	assert.Nil(t, cfg.AntiFlood.EnableKick)

	// nil fields (not set by default)
	assert.Nil(t, cfg.PasswordAdmin)
	assert.Nil(t, cfg.Password)
	assert.Nil(t, cfg.Hostname)
	assert.Nil(t, cfg.Motd)
	assert.Nil(t, cfg.Admins)
	assert.Nil(t, cfg.HeadlessClients)
	assert.Nil(t, cfg.KickDuplicate)
	assert.Nil(t, cfg.Persistent)
	assert.Nil(t, cfg.DisableVoN)
	assert.Nil(t, cfg.MissionsToServerRestart)
	assert.Nil(t, cfg.MissionsToShutdown)
}

func TestServerConfigUnmarshal(t *testing.T) {
	t.Run("basic_fields", func(t *testing.T) {
		content := `
hostname: "My Arma Server"
password: "secret"
password_admin: "adminpass"
server_command_password: "cmdpass"
log_file: "server.log"
max_players: 32
motd_interval: 2
vote_threshold: 0.33
vote_mission_players: 3
`
		var cfg ServerConfig
		require.NoError(t, yaml.Load([]byte(content), &cfg))

		require.NotNil(t, cfg.Hostname)
		assert.Equal(t, "My Arma Server", *cfg.Hostname)
		require.NotNil(t, cfg.Password)
		assert.Equal(t, "secret", *cfg.Password)
		require.NotNil(t, cfg.PasswordAdmin)
		assert.Equal(t, "adminpass", *cfg.PasswordAdmin)
		require.NotNil(t, cfg.ServerCommandPassword)
		assert.Equal(t, "cmdpass", *cfg.ServerCommandPassword)
		require.NotNil(t, cfg.LogFile)
		assert.Equal(t, "server.log", *cfg.LogFile)
		require.NotNil(t, cfg.MaxPlayers)
		assert.Equal(t, uint16(32), *cfg.MaxPlayers)
		require.NotNil(t, cfg.MotdInterval)
		assert.Equal(t, uint16(2), *cfg.MotdInterval)
		require.NotNil(t, cfg.VoteThreshold)
		assert.Equal(t, float32(0.33), *cfg.VoteThreshold)
		require.NotNil(t, cfg.VoteMissionPlayers)
		assert.Equal(t, uint16(3), *cfg.VoteMissionPlayers)
	})

	t.Run("motd_and_lists", func(t *testing.T) {
		content := `
motd:
  - "Welcome to our server"
  - ""
  - "Have fun!"
admins:
  - "76561198000000001"
  - "76561198000000002"
headless_clients:
  - "127.0.0.1"
local_client:
  - "127.0.0.1"
file_patching_exceptions:
  - "76561198000000001"
mission_whitelist:
  - "intro.altis"
  - "coop.stratis"
`
		var cfg ServerConfig
		require.NoError(t, yaml.Load([]byte(content), &cfg))

		assert.Equal(t, []string{"Welcome to our server", "", "Have fun!"}, cfg.Motd)
		assert.Equal(t, []string{"76561198000000001", "76561198000000002"}, cfg.Admins)
		assert.Equal(t, []string{"127.0.0.1"}, cfg.HeadlessClients)
		assert.Equal(t, []string{"127.0.0.1"}, cfg.LocalClient)
		assert.Equal(t, []string{"76561198000000001"}, cfg.FilePatchingExceptions)
		assert.Equal(t, []string{"intro.altis", "coop.stratis"}, cfg.MissionWhitelist)
	})

	t.Run("behaviour_flags", func(t *testing.T) {
		content := `
kickduplicate: true
loopback: false
upnp: true
persistent: true
auto_select_mission: true
random_mission_order: true
enable_player_diag: true
equal_mod_required: false
battleye: true
drawing_in_map: false
disable_von: true
skip_lobby: true
allow_profile_glasses: false
`
		var cfg ServerConfig
		require.NoError(t, yaml.Load([]byte(content), &cfg))

		require.NotNil(t, cfg.KickDuplicate)
		assert.True(t, *cfg.KickDuplicate)
		require.NotNil(t, cfg.Loopback)
		assert.False(t, *cfg.Loopback)
		require.NotNil(t, cfg.Upnp)
		assert.True(t, *cfg.Upnp)
		require.NotNil(t, cfg.Persistent)
		assert.True(t, *cfg.Persistent)
		require.NotNil(t, cfg.AutoSelectMission)
		assert.True(t, *cfg.AutoSelectMission)
		require.NotNil(t, cfg.RandomMissionOrder)
		assert.True(t, *cfg.RandomMissionOrder)
		require.NotNil(t, cfg.EnablePlayerDiag)
		assert.True(t, *cfg.EnablePlayerDiag)
		require.NotNil(t, cfg.EqualModRequired)
		assert.False(t, *cfg.EqualModRequired)
		require.NotNil(t, cfg.BattlEye)
		assert.True(t, *cfg.BattlEye)
		require.NotNil(t, cfg.DrawingInMap)
		assert.False(t, *cfg.DrawingInMap)
		require.NotNil(t, cfg.DisableVoN)
		assert.True(t, *cfg.DisableVoN)
		require.NotNil(t, cfg.SkipLobby)
		assert.True(t, *cfg.SkipLobby)
		require.NotNil(t, cfg.AllowProfileGlasses)
		assert.False(t, *cfg.AllowProfileGlasses)
	})

	t.Run("network_thresholds", func(t *testing.T) {
		content := `
max_ping: 200
max_packet_loss: 50
max_desync: 150
disconnect_timeout: 5
kick_clients_on_slow_network: [0, 0, 0, 0]
call_ext_report_limit: 500
`
		var cfg ServerConfig
		require.NoError(t, yaml.Load([]byte(content), &cfg))

		require.NotNil(t, cfg.MaxPing)
		assert.Equal(t, int32(200), *cfg.MaxPing)
		require.NotNil(t, cfg.MaxPacketLoss)
		assert.Equal(t, int32(50), *cfg.MaxPacketLoss)
		require.NotNil(t, cfg.MaxDesync)
		assert.Equal(t, int32(150), *cfg.MaxDesync)
		require.NotNil(t, cfg.DisconnectTimeout)
		assert.Equal(t, uint8(5), *cfg.DisconnectTimeout)
		require.NotNil(t, cfg.KickClientsOnSlowNetwork)
		assert.Equal(t, [4]uint8{0, 0, 0, 0}, *cfg.KickClientsOnSlowNetwork)
		require.NotNil(t, cfg.CallExtReportLimit)
		assert.Equal(t, uint32(500), *cfg.CallExtReportLimit)
	})

	t.Run("kick_timeout", func(t *testing.T) {
		content := `
kick_timeout:
  - kick_id: 0
    timeout: -1
  - kick_id: 1
    timeout: 180
  - kick_id: 2
    timeout: 180
  - kick_id: 3
    timeout: 180
`
		var cfg ServerConfig
		require.NoError(t, yaml.Load([]byte(content), &cfg))

		require.Len(t, cfg.KickTimeout, 4)
		assert.Equal(t, KickSourceManual, cfg.KickTimeout[0].KickID)
		assert.Equal(t, int32(-1), cfg.KickTimeout[0].Timeout)
		assert.Equal(t, KickSourceConnectivity, cfg.KickTimeout[1].KickID)
		assert.Equal(t, int32(180), cfg.KickTimeout[1].Timeout)
	})

	t.Run("screen_timeouts", func(t *testing.T) {
		content := `
voting_time_out: [60, 90]
role_time_out: [90, 120]
briefing_time_out: [60, 90]
debriefing_time_out: [45, 60]
lobby_idle_timeout: 300
`
		var cfg ServerConfig
		require.NoError(t, yaml.Load([]byte(content), &cfg))

		assert.Equal(t, []uint32{60, 90}, cfg.VotingTimeOut)
		assert.Equal(t, []uint32{90, 120}, cfg.RoleTimeOut)
		assert.Equal(t, []uint32{60, 90}, cfg.BriefingTimeOut)
		assert.Equal(t, []uint32{45, 60}, cfg.DebriefingTimeOut)
		require.NotNil(t, cfg.LobbyIdleTimeout)
		assert.Equal(t, uint32(300), *cfg.LobbyIdleTimeout)
	})

	t.Run("file_access_control", func(t *testing.T) {
		content := `
allowed_file_patching: 2
allowed_load_file_extensions:
  - "sqf"
  - "txt"
  - "xml"
allowed_preprocess_file_extensions:
  - "sqf"
  - "sqs"
allowed_html_load_extensions:
  - "htm"
  - "html"
allowed_html_load_uris:
  - "https://arma3.com"
`
		var cfg ServerConfig
		require.NoError(t, yaml.Load([]byte(content), &cfg))

		require.NotNil(t, cfg.AllowedFilePatching)
		assert.Equal(t, FilePatchingAllowed, *cfg.AllowedFilePatching)
		assert.Equal(t, []string{"sqf", "txt", "xml"}, cfg.AllowedLoadFileExtensions)
		assert.Equal(t, []string{"sqf", "sqs"}, cfg.AllowedPreprocessFileExtensions)
		assert.Equal(t, []string{"htm", "html"}, cfg.AllowedHTMLLoadExtensions)
		assert.Equal(t, []string{"https://arma3.com"}, cfg.AllowedHTMLLoadURIs)
	})

	t.Run("disable_channels", func(t *testing.T) {
		content := `
disable_channels:
  - channel_id: 0
    text: false
    voice: true
    map_markers: false
    draw_on_map: true
  - channel_id: 3
    text: true
    voice: true
`
		var cfg ServerConfig
		require.NoError(t, yaml.Load([]byte(content), &cfg))

		require.Len(t, cfg.DisableChannels, 2)
		assert.Equal(t, ChannelGlobal, cfg.DisableChannels[0].ChannelID)
		assert.False(t, cfg.DisableChannels[0].Text)
		assert.True(t, cfg.DisableChannels[0].Voice)
		assert.False(t, cfg.DisableChannels[0].MapMarkers)
		assert.True(t, cfg.DisableChannels[0].DrawOnMap)
		assert.Equal(t, ChannelGroup, cfg.DisableChannels[1].ChannelID)
		assert.True(t, cfg.DisableChannels[1].Text)
		assert.True(t, cfg.DisableChannels[1].Voice)
	})

	t.Run("scripting_hooks", func(t *testing.T) {
		content := `
double_id_detected: "kick (_this select 0)"
on_hacked_data: "kick (_this select 0)"
on_unsigned_data: "kick (_this select 0)"
on_user_connected: ""
on_different_data: ""
`
		var cfg ServerConfig
		require.NoError(t, yaml.Load([]byte(content), &cfg))

		require.NotNil(t, cfg.DoubleIdDetected)
		assert.Equal(t, "kick (_this select 0)", *cfg.DoubleIdDetected)
		require.NotNil(t, cfg.OnHackedData)
		assert.Equal(t, "kick (_this select 0)", *cfg.OnHackedData)
		require.NotNil(t, cfg.OnUnsignedData)
		assert.Equal(t, "kick (_this select 0)", *cfg.OnUnsignedData)
		require.NotNil(t, cfg.OnUserConnected)
		assert.Empty(t, *cfg.OnUserConnected)
		require.NotNil(t, cfg.OnDifferentData)
		assert.Empty(t, *cfg.OnDifferentData)
	})

	t.Run("misc_fields", func(t *testing.T) {
		content := `
verify_signatures: 2
von_codec: 1
von_codec_quality: 30
time_stamp_format: "short"
force_rotor_lib_simulation: 1
required_build: 12345
statistics_enabled: false
forced_difficulty: "veteran"
steam_protocol_max_data_size: "2KB"
arma_units_timeout: 60
override_haze_quality: 2
mission_http_download_base_url: "https://example.com/missions/"
zeus_composition_script_level: 2
missions_to_server_restart: 8
missions_to_shutdown: 0
`
		var cfg ServerConfig
		require.NoError(t, yaml.Load([]byte(content), &cfg))

		require.NotNil(t, cfg.VerifySignatures)
		assert.Equal(t, SignaturesV2Only, *cfg.VerifySignatures)
		require.NotNil(t, cfg.VonCodec)
		assert.Equal(t, VonCodecOPUS, *cfg.VonCodec)
		require.NotNil(t, cfg.VonCodecQuality)
		assert.Equal(t, uint8(30), *cfg.VonCodecQuality)
		require.NotNil(t, cfg.TimeStampFormat)
		assert.Equal(t, TimestampShort, *cfg.TimeStampFormat)
		require.NotNil(t, cfg.ForceRotorLibSimulation)
		assert.Equal(t, RotorLibAFM, *cfg.ForceRotorLibSimulation)
		require.NotNil(t, cfg.RequiredBuild)
		assert.Equal(t, uint32(12345), *cfg.RequiredBuild)
		require.NotNil(t, cfg.StatisticsEnabled)
		assert.False(t, *cfg.StatisticsEnabled)
		require.NotNil(t, cfg.ForcedDifficulty)
		assert.Equal(t, DifficultyVeteran, *cfg.ForcedDifficulty)
		require.NotNil(t, cfg.SteamProtocolMaxDataSize)
		assert.Equal(t, uint64(2000), cfg.SteamProtocolMaxDataSize.B())
		require.NotNil(t, cfg.ArmaUnitsTimeout)
		assert.Equal(t, uint32(60), *cfg.ArmaUnitsTimeout)
		require.NotNil(t, cfg.OverrideHazeQuality)
		assert.Equal(t, HazeQualityStandard, *cfg.OverrideHazeQuality)
		require.NotNil(t, cfg.MissionHTTPDownloadBaseURL)
		assert.Equal(t, "https://example.com/missions/", *cfg.MissionHTTPDownloadBaseURL)
		require.NotNil(t, cfg.ZeusCompositionScriptLevel)
		assert.Equal(t, ZeusScriptsAll, *cfg.ZeusCompositionScriptLevel)
		require.NotNil(t, cfg.MissionsToServerRestart)
		assert.Equal(t, uint32(8), *cfg.MissionsToServerRestart)
		require.NotNil(t, cfg.MissionsToShutdown)
		assert.Equal(t, uint32(0), *cfg.MissionsToShutdown)
	})

	t.Run("steam_protocol_max_data_size_as_number", func(t *testing.T) {
		content := `steam_protocol_max_data_size: 2048`
		var cfg ServerConfig
		require.NoError(t, yaml.Load([]byte(content), &cfg))
		require.NotNil(t, cfg.SteamProtocolMaxDataSize)
		assert.Equal(t, uint64(2048), cfg.SteamProtocolMaxDataSize.B())
	})

	t.Run("advanced_options", func(t *testing.T) {
		content := `
advanced_options:
  log_object_not_found: false
  skip_description_parsing: true
  ignore_mission_load_errors: true
  queue_size_log_g: 1000000
`
		var cfg ServerConfig
		require.NoError(t, yaml.Load([]byte(content), &cfg))

		require.NotNil(t, cfg.AdvancedOptions)
		require.NotNil(t, cfg.AdvancedOptions.LogObjectNotFound)
		assert.False(t, *cfg.AdvancedOptions.LogObjectNotFound)
		require.NotNil(t, cfg.AdvancedOptions.SkipDescriptionParsing)
		assert.True(t, *cfg.AdvancedOptions.SkipDescriptionParsing)
		require.NotNil(t, cfg.AdvancedOptions.IgnoreMissionLoadErrors)
		assert.True(t, *cfg.AdvancedOptions.IgnoreMissionLoadErrors)
		require.NotNil(t, cfg.AdvancedOptions.QueueSizeLogG)
		assert.Equal(t, uint32(1000000), *cfg.AdvancedOptions.QueueSizeLogG)
	})

	t.Run("anti_flood", func(t *testing.T) {
		content := `
anti_flood:
  cycle_time: 1.0
  cycle_limit: 200
  cycle_hard_limit: 2000
  enable_kick: true
`
		var cfg ServerConfig
		require.NoError(t, yaml.Load([]byte(content), &cfg))

		require.NotNil(t, cfg.AntiFlood)
		require.NotNil(t, cfg.AntiFlood.CycleTime)
		assert.Equal(t, float32(1.0), *cfg.AntiFlood.CycleTime)
		require.NotNil(t, cfg.AntiFlood.CycleLimit)
		assert.Equal(t, uint32(200), *cfg.AntiFlood.CycleLimit)
		require.NotNil(t, cfg.AntiFlood.CycleHardLimit)
		assert.Equal(t, uint32(2000), *cfg.AntiFlood.CycleHardLimit)
		require.NotNil(t, cfg.AntiFlood.EnableKick)
		assert.True(t, *cfg.AntiFlood.EnableKick)
	})

	t.Run("missions_rotation", func(t *testing.T) {
		content := `
missions:
  - template: "MP_Marksmen_01.Altis"
    difficulty: "recruit"
    params:
      RespawnDelay: 15
      EndGameRespawnDelay: 30
  - template: "EscapeFromMalden.Malden"
    difficulty: "regular"
  - template: "MP_CombatPatrol_01.Altis"
    difficulty: "veteran"
    params:
      BIS_CP_tickets: 5
  - template: "EXP_m01.Tanoa"
    difficulty: "custom"
`
		var cfg ServerConfig
		require.NoError(t, yaml.Load([]byte(content), &cfg))

		require.Len(t, cfg.Missions, 4)

		assert.Equal(t, "MP_Marksmen_01.Altis", cfg.Missions[0].Template)
		assert.Equal(t, DifficultyRecruit, cfg.Missions[0].Difficulty)
		assert.Equal(t, map[string]int{"RespawnDelay": 15, "EndGameRespawnDelay": 30}, cfg.Missions[0].Params)

		assert.Equal(t, "EscapeFromMalden.Malden", cfg.Missions[1].Template)
		assert.Equal(t, DifficultyRegular, cfg.Missions[1].Difficulty)
		assert.Empty(t, cfg.Missions[1].Params)

		assert.Equal(t, "MP_CombatPatrol_01.Altis", cfg.Missions[2].Template)
		assert.Equal(t, DifficultyVeteran, cfg.Missions[2].Difficulty)
		assert.Equal(t, map[string]int{"BIS_CP_tickets": 5}, cfg.Missions[2].Params)

		assert.Equal(t, "EXP_m01.Tanoa", cfg.Missions[3].Template)
		assert.Equal(t, DifficultyCustom, cfg.Missions[3].Difficulty)
	})

	t.Run("missions_empty", func(t *testing.T) {
		content := `missions: []`
		var cfg ServerConfig
		require.NoError(t, yaml.Load([]byte(content), &cfg))
		assert.Empty(t, cfg.Missions)
	})
}
