package arma

// --- Enum types ---

// FilePatchingMode controls which clients may join with -filePatching.
type FilePatchingMode uint8

const (
	FilePatchingNone    FilePatchingMode = 0 // no clients allowed
	FilePatchingHCOnly  FilePatchingMode = 1 // Headless Clients only
	FilePatchingAllowed FilePatchingMode = 2 // all clients
)

// KickSource identifies the origin of a kick for use in KickTimeout.
type KickSource uint8

const (
	KickSourceManual       KickSource = 0 // vote kick, admin kick, bruteforce detection
	KickSourceConnectivity KickSource = 1 // ping, timeout, packetloss, desync
	KickSourceBattlEye     KickSource = 2
	KickSourceHarmless     KickSource = 3 // wrong addons, steam timeout/checks, signatures
)

// Special KickTimeout.Timeout sentinel values.
const (
	KickTimeoutUntilMissionEnd    int32 = -1
	KickTimeoutUntilServerRestart int32 = -2
)

// ChannelID identifies an Arma 3 communication channel.
type ChannelID uint8

const (
	ChannelGlobal  ChannelID = 0
	ChannelSide    ChannelID = 1
	ChannelCommand ChannelID = 2
	ChannelGroup   ChannelID = 3
	ChannelVehicle ChannelID = 4
	ChannelDirect  ChannelID = 5
	ChannelSystem  ChannelID = 16
)

// SignatureVerification controls addon signature checking.
type SignatureVerification uint8

const (
	SignaturesDisabled SignatureVerification = 0 // disabled
	// Value 1 is accepted by the engine but defaults to SignaturesV2Only.
	SignaturesV2Only SignatureVerification = 2 // only v2 .bisign files accepted
)

// VonCodecType selects the Voice over Net audio codec.
type VonCodecType uint8

const (
	VonCodecSPEEX VonCodecType = 0
	VonCodecOPUS  VonCodecType = 1
)

// ZeusScriptLevel controls which scripts Zeus compositions may execute.
type ZeusScriptLevel uint8

const (
	ZeusScriptsForbidden      ZeusScriptLevel = 0 // no scripts
	ZeusScriptsAttributesOnly ZeusScriptLevel = 1 // only attributes (default)
	ZeusScriptsAll            ZeusScriptLevel = 2 // init scripts included
)

// RotorLibMode sets the flight model enforced server-side.
type RotorLibMode uint8

const (
	RotorLibPlayerChoice RotorLibMode = 0 // up to the player
	RotorLibAFM          RotorLibMode = 1 // Advanced Flight Model forced
	RotorLibSFM          RotorLibMode = 2 // Standard Flight Model forced
)

// TimestampFormat sets the format of timestamps in the server RPT log.
type TimestampFormat string

const (
	TimestampNone  TimestampFormat = "none"
	TimestampShort TimestampFormat = "short"
	TimestampFull  TimestampFormat = "full"
)

// Difficulty names the enforced difficulty level.
type Difficulty string

const (
	DifficultyRecruit Difficulty = "recruit"
	DifficultyRegular Difficulty = "regular"
	DifficultyVeteran Difficulty = "veteran"
	DifficultyCustom  Difficulty = "custom"
)

// HazeQualityLevel forces haze rendering quality on all MP clients.
type HazeQualityLevel int8

const (
	HazeQualityOff      HazeQualityLevel = -1 // do not force (default)
	HazeQualityVeryLow  HazeQualityLevel = 0
	HazeQualityLow      HazeQualityLevel = 1
	HazeQualityStandard HazeQualityLevel = 2
)

// --- Supporting structs ---

// KickTimeout defines how long a kicked player must wait before rejoining.
type KickTimeout struct {
	KickID KickSource `json:"kickID"  yaml:"kick_id"  toml:"kick_id"`
	// Seconds >0; KickTimeoutUntilMissionEnd (-1); KickTimeoutUntilServerRestart (-2)
	Timeout int32 `json:"timeout" yaml:"timeout" toml:"timeout"`
}

// DisabledChannel configures per-channel communication restrictions.
type DisabledChannel struct {
	ChannelID  ChannelID `json:"channelID"  yaml:"channel_id"  toml:"channel_id"`
	Text       bool      `json:"text"       yaml:"text"        toml:"text"`        // true = disable text chat
	Voice      bool      `json:"voice"      yaml:"voice"       toml:"voice"`       // true = disable VON
	MapMarkers bool      `json:"mapMarkers" yaml:"map_markers" toml:"map_markers"` // true = disable manual map markers
	DrawOnMap  bool      `json:"drawOnMap"  yaml:"draw_on_map" toml:"draw_on_map"` // true = disable Ctrl+LMB map drawing
}

// AdvancedOptions maps to class AdvancedOptions in server.cfg.
type AdvancedOptions struct {
	// false = skip "Server: Object not found" log messages
	LogObjectNotFound *bool `json:"logObjectNotFound,omitempty" yaml:"log_object_not_found,omitempty" toml:"log_object_not_found,omitempty"`
	// true = skip description.ext/mission.sqm parsing; faster mission list but breaks overviewText etc.
	SkipDescriptionParsing *bool `json:"skipDescriptionParsing,omitempty" yaml:"skip_description_parsing,omitempty" toml:"skip_description_parsing,omitempty"`
	// true = load mission regardless of loading errors; false = abort on errors
	IgnoreMissionLoadErrors *bool `json:"ignoreMissionLoadErrors,omitempty" yaml:"ignore_mission_load_errors,omitempty" toml:"ignore_mission_load_errors,omitempty"`
	// Dump player message types to log when Guaranteed Queue exceeds this byte threshold; 0 = disabled
	QueueSizeLogG *uint32 `json:"queueSizeLogG,omitempty" yaml:"queue_size_log_g,omitempty" toml:"queue_size_log_g,omitempty"`
}

// AntiFlood maps to class AntiFlood in server.cfg.
type AntiFlood struct {
	// Cycle length in seconds. Every cycle, if a player exceeds CycleLimit messages they are flagged.
	CycleTime *float32 `json:"cycleTime,omitempty" yaml:"cycle_time,omitempty" toml:"cycle_time,omitempty"`
	// Message count that triggers a flag for the cycle.
	CycleLimit *uint32 `json:"cycleLimit,omitempty" yaml:"cycle_limit,omitempty" toml:"cycle_limit,omitempty"`
	// Immediately triggers action (kick or log) when exceeded within one cycle.
	CycleHardLimit *uint32 `json:"cycleHardLimit,omitempty" yaml:"cycle_hard_limit,omitempty" toml:"cycle_hard_limit,omitempty"`
	// true = kick flagged players; false = log only
	EnableKick *bool `json:"enableKick,omitempty" yaml:"enable_kick,omitempty" toml:"enable_kick,omitempty"`
}

// VoteCommand defines one entry in allowedVoteCmds.
// All fields after Name are optional; omitted fields fall back to Arma 3 defaults.
// If VotingThreshold or PercentSideVotingThreshold are set, PreMissionStart and
// PostMissionStart are emitted with their default value (true) when they are nil.
type VoteCommand struct {
	Name                       string   `json:"name"                                 yaml:"name"                                    toml:"name"`
	PreMissionStart            *bool    `json:"preMissionStart,omitempty"            yaml:"pre_mission_start,omitempty"             toml:"pre_mission_start,omitempty"`
	PostMissionStart           *bool    `json:"postMissionStart,omitempty"           yaml:"post_mission_start,omitempty"            toml:"post_mission_start,omitempty"`
	VotingThreshold            *float32 `json:"votingThreshold,omitempty"            yaml:"voting_threshold,omitempty"              toml:"voting_threshold,omitempty"`
	PercentSideVotingThreshold *float32 `json:"percentSideVotingThreshold,omitempty" yaml:"percent_side_voting_threshold,omitempty" toml:"percent_side_voting_threshold,omitempty"`
}

// VotedAdminCommand defines one entry in allowedVotedAdminCmds.
// PreMissionStart and PostMissionStart are optional; omitted fields are emitted
// only when at least one of them is non-nil.
type VotedAdminCommand struct {
	Name             string `json:"name"                        yaml:"name"                         toml:"name"`
	PreMissionStart  *bool  `json:"preMissionStart,omitempty"   yaml:"pre_mission_start,omitempty"  toml:"pre_mission_start,omitempty"`
	PostMissionStart *bool  `json:"postMissionStart,omitempty"  yaml:"post_mission_start,omitempty" toml:"post_mission_start,omitempty"`
}

// Mission defines a single entry in the server mission rotation.
// Template and Difficulty are required; CustomName and Params are optional.
type Mission struct {
	// Class name used in the generated cfg; defaults to "MissionN" if empty.
	CustomName string `json:"customName,omitempty" yaml:"custom_name,omitempty" toml:"custom_name,omitempty"`
	// "missionName.terrainName", e.g. "MP_Marksmen_01.Altis".
	Template   string     `json:"template"         yaml:"template"   toml:"template"`
	Difficulty Difficulty `json:"difficulty"       yaml:"difficulty" toml:"difficulty"`
	// Overrides mission parameter defaults; keys match description.ext parameter names.
	Params map[string]int `json:"params,omitempty" yaml:"params,omitempty" toml:"params,omitempty"`
}

// --- Main config ---

// ServerConfig holds all Arma 3 server.cfg parameters.
// All fields are optional (pointer/slice); nil means the parameter is omitted from the cfg file.
type ServerConfig struct {
	// --- Authentication ---
	PasswordAdmin         *string `json:"passwordAdmin,omitempty"         yaml:"password_admin,omitempty"          toml:"password_admin,omitempty"`
	Password              *string `json:"password,omitempty"              yaml:"password,omitempty"                toml:"password,omitempty"`
	ServerCommandPassword *string `json:"serverCommandPassword,omitempty" yaml:"server_command_password,omitempty" toml:"server_command_password,omitempty"`
	// Visible in game browser. Defaults to local machine name when empty.
	Hostname *string `json:"hostname,omitempty" yaml:"hostname,omitempty" toml:"hostname,omitempty"`
	LogFile  *string `json:"logFile,omitempty"  yaml:"log_file,omitempty" toml:"log_file,omitempty"`

	// --- Message of the Day ---
	Motd         []string `json:"motd,omitempty"         yaml:"motd,omitempty"          toml:"motd,omitempty"`
	MotdInterval *uint16  `json:"motdInterval,omitempty" yaml:"motd_interval,omitempty" toml:"motd_interval,omitempty"`

	// --- Player limits ---
	MaxPlayers *uint16 `json:"maxPlayers,omitempty" yaml:"max_players,omitempty" toml:"max_players,omitempty"`

	// --- Access lists ---
	Admins                 []string `json:"admins,omitempty"                  yaml:"admins,omitempty"                    toml:"admins,omitempty"`
	HeadlessClients        []string `json:"headlessClients,omitempty"         yaml:"headless_clients,omitempty"          toml:"headless_clients,omitempty"`
	LocalClient            []string `json:"localClient,omitempty"             yaml:"local_client,omitempty"              toml:"local_client,omitempty"`
	FilePatchingExceptions []string `json:"filePatchingExceptions,omitempty"  yaml:"file_patching_exceptions,omitempty"  toml:"file_patching_exceptions,omitempty"`
	MissionWhitelist       []string `json:"missionWhitelist,omitempty"        yaml:"mission_whitelist,omitempty"         toml:"mission_whitelist,omitempty"`

	// --- Voting ---
	VoteThreshold         *float32             `json:"voteThreshold,omitempty"         yaml:"vote_threshold,omitempty"          toml:"vote_threshold,omitempty"`
	VoteMissionPlayers    *uint16              `json:"voteMissionPlayers,omitempty"    yaml:"vote_mission_players,omitempty"    toml:"vote_mission_players,omitempty"`
	// nil = omitted (Arma uses engine defaults); &[]VoteCommand{} = empty array (all commands disabled)
	AllowedVoteCmds *[]VoteCommand `json:"allowedVoteCmds,omitempty" yaml:"allowed_vote_cmds,omitempty" toml:"allowed_vote_cmds,omitempty"`
	// nil = omitted (Arma uses engine defaults); &[]VotedAdminCommand{} = empty array (all commands disabled)
	AllowedVotedAdminCmds *[]VotedAdminCommand `json:"allowedVotedAdminCmds,omitempty" yaml:"allowed_voted_admin_cmds,omitempty" toml:"allowed_voted_admin_cmds,omitempty"`

	// --- Server behaviour ---
	KickDuplicate      *bool `json:"kickduplicate,omitempty"       yaml:"kickduplicate,omitempty"         toml:"kickduplicate,omitempty"`
	Loopback           *bool `json:"loopback,omitempty"            yaml:"loopback,omitempty"              toml:"loopback,omitempty"`
	Upnp               *bool `json:"upnp,omitempty"                yaml:"upnp,omitempty"                  toml:"upnp,omitempty"`
	Persistent         *bool `json:"persistent,omitempty"          yaml:"persistent,omitempty"            toml:"persistent,omitempty"`
	AutoSelectMission  *bool `json:"autoSelectMission,omitempty"   yaml:"auto_select_mission,omitempty"   toml:"auto_select_mission,omitempty"`
	RandomMissionOrder *bool `json:"randomMissionOrder,omitempty"  yaml:"random_mission_order,omitempty"  toml:"random_mission_order,omitempty"`
	EnablePlayerDiag   *bool `json:"enablePlayerDiag,omitempty"    yaml:"enable_player_diag,omitempty"    toml:"enable_player_diag,omitempty"`

	// --- File access control ---
	AllowedFilePatching             *FilePatchingMode `json:"allowedFilePatching,omitempty"             yaml:"allowed_file_patching,omitempty"              toml:"allowed_file_patching,omitempty"`
	AllowedLoadFileExtensions       []string          `json:"allowedLoadFileExtensions,omitempty"       yaml:"allowed_load_file_extensions,omitempty"       toml:"allowed_load_file_extensions,omitempty"`
	AllowedPreprocessFileExtensions []string          `json:"allowedPreprocessFileExtensions,omitempty" yaml:"allowed_preprocess_file_extensions,omitempty" toml:"allowed_preprocess_file_extensions,omitempty"`
	AllowedHTMLLoadExtensions       []string          `json:"allowedHTMLLoadExtensions,omitempty"       yaml:"allowed_html_load_extensions,omitempty"       toml:"allowed_html_load_extensions,omitempty"`
	AllowedHTMLLoadURIs             []string          `json:"allowedHTMLLoadURIs,omitempty"             yaml:"allowed_html_load_uris,omitempty"             toml:"allowed_html_load_uris,omitempty"`

	// --- Network quality thresholds (negative = disabled) ---
	MaxPing       *int32 `json:"maxPing,omitempty"       yaml:"max_ping,omitempty"        toml:"max_ping,omitempty"`
	MaxPacketLoss *int32 `json:"maxPacketLoss,omitempty" yaml:"max_packet_loss,omitempty" toml:"max_packet_loss,omitempty"`
	MaxDesync     *int32 `json:"maxDesync,omitempty"     yaml:"max_desync,omitempty"      toml:"max_desync,omitempty"`
	// Seconds to wait before disconnecting on lost connection. Range 1–90.
	DisconnectTimeout *uint8 `json:"disconnectTimeout,omitempty" yaml:"disconnect_timeout,omitempty" toml:"disconnect_timeout,omitempty"`
	// Per-threshold action: 0=log, 1=kick. Order: {MaxPing, MaxPacketLoss, MaxDesync, DisconnectTimeout}.
	KickClientsOnSlowNetwork *[4]uint8     `json:"kickClientsOnSlowNetwork,omitempty" yaml:"kick_clients_on_slow_network,omitempty" toml:"kick_clients_on_slow_network,omitempty"`
	KickTimeout              []KickTimeout `json:"kickTimeout,omitempty"              yaml:"kick_timeout,omitempty"                 toml:"kick_timeout,omitempty"`
	// Warn in RPT when callExtension takes longer than this many milliseconds.
	CallExtReportLimit *uint32 `json:"callExtReportLimit,omitempty" yaml:"call_ext_report_limit,omitempty" toml:"call_ext_report_limit,omitempty"`

	// --- Screen timeouts (seconds) ---
	// Each field can be a single value or {ready, notReady} pair.
	VotingTimeOut     []uint32 `json:"votingTimeOut,omitempty"     yaml:"voting_time_out,omitempty"     toml:"voting_time_out,omitempty"`
	RoleTimeOut       []uint32 `json:"roleTimeOut,omitempty"       yaml:"role_time_out,omitempty"       toml:"role_time_out,omitempty"`
	BriefingTimeOut   []uint32 `json:"briefingTimeOut,omitempty"   yaml:"briefing_time_out,omitempty"   toml:"briefing_time_out,omitempty"`
	DebriefingTimeOut []uint32 `json:"debriefingTimeOut,omitempty" yaml:"debriefing_time_out,omitempty" toml:"debriefing_time_out,omitempty"`
	// At least MAX(votingTimeout, roleTimeout, briefingTimeout, debriefingTimeout) + 5s regardless of setting.
	LobbyIdleTimeout *uint32 `json:"lobbyIdleTimeout,omitempty" yaml:"lobby_idle_timeout,omitempty" toml:"lobby_idle_timeout,omitempty"`

	// --- Mission cycling ---
	// Restart/shutdown process after N mission ends. 0 = disabled.
	MissionsToServerRestart *uint32   `json:"missionsToServerRestart,omitempty" yaml:"missions_to_server_restart,omitempty" toml:"missions_to_server_restart,omitempty"`
	MissionsToShutdown      *uint32   `json:"missionsToShutdown,omitempty"      yaml:"missions_to_shutdown,omitempty"       toml:"missions_to_shutdown,omitempty"`
	Missions                []Mission `json:"missions,omitempty"                yaml:"missions,omitempty"                   toml:"missions,omitempty"`

	// --- Channel restrictions ---
	DisableChannels []DisabledChannel `json:"disableChannels,omitempty" yaml:"disable_channels,omitempty" toml:"disable_channels,omitempty"`

	// --- Security ---
	VerifySignatures *SignatureVerification `json:"verifySignatures,omitempty" yaml:"verify_signatures,omitempty"  toml:"verify_signatures,omitempty"`
	EqualModRequired *bool                  `json:"equalModRequired,omitempty" yaml:"equal_mod_required,omitempty" toml:"equal_mod_required,omitempty"`
	BattlEye         *bool                  `json:"battleye,omitempty"         yaml:"battleye,omitempty"           toml:"battleye,omitempty"`

	// --- Display / Voice ---
	DrawingInMap               *bool            `json:"drawingInMap,omitempty"               yaml:"drawing_in_map,omitempty"                toml:"drawing_in_map,omitempty"`
	DisableVoN                 *bool            `json:"disableVoN,omitempty"                 yaml:"disable_von,omitempty"                   toml:"disable_von,omitempty"`
	VonCodecQuality            *uint8           `json:"vonCodecQuality,omitempty"            yaml:"von_codec_quality,omitempty"             toml:"von_codec_quality,omitempty"`
	VonCodec                   *VonCodecType    `json:"vonCodec,omitempty"                   yaml:"von_codec,omitempty"                     toml:"von_codec,omitempty"`
	SkipLobby                  *bool            `json:"skipLobby,omitempty"                  yaml:"skip_lobby,omitempty"                    toml:"skip_lobby,omitempty"`
	AllowProfileGlasses        *bool            `json:"allowProfileGlasses,omitempty"        yaml:"allow_profile_glasses,omitempty"         toml:"allow_profile_glasses,omitempty"`
	ZeusCompositionScriptLevel *ZeusScriptLevel `json:"zeusCompositionScriptLevel,omitempty" yaml:"zeus_composition_script_level,omitempty" toml:"zeus_composition_script_level,omitempty"`

	// --- Server-side scripting hooks ---
	DoubleIdDetected   *string `json:"doubleIdDetected,omitempty"   yaml:"double_id_detected,omitempty"   toml:"double_id_detected,omitempty"`
	OnUserConnected    *string `json:"onUserConnected,omitempty"    yaml:"on_user_connected,omitempty"    toml:"on_user_connected,omitempty"`
	OnUserDisconnected *string `json:"onUserDisconnected,omitempty" yaml:"on_user_disconnected,omitempty" toml:"on_user_disconnected,omitempty"`
	OnHackedData       *string `json:"onHackedData,omitempty"       yaml:"on_hacked_data,omitempty"       toml:"on_hacked_data,omitempty"`
	OnDifferentData    *string `json:"onDifferentData,omitempty"    yaml:"on_different_data,omitempty"    toml:"on_different_data,omitempty"`
	OnUnsignedData     *string `json:"onUnsignedData,omitempty"     yaml:"on_unsigned_data,omitempty"     toml:"on_unsigned_data,omitempty"`
	OnUserKicked       *string `json:"onUserKicked,omitempty"       yaml:"on_user_kicked,omitempty"       toml:"on_user_kicked,omitempty"`
	RegularCheck       *string `json:"regularCheck,omitempty"       yaml:"regular_check,omitempty"        toml:"regular_check,omitempty"`
	// Called repeatedly while a player attempts to join. Must return "ACCEPT", "DELAY", or "REFUSE[_<message>]".
	OnPlayerJoinAttempt *string `json:"onPlayerJoinAttempt,omitempty" yaml:"on_player_join_attempt,omitempty" toml:"on_player_join_attempt,omitempty"`
	// Sends a "System" chat message to the specified user.
	SendChatMessage *string `json:"sendChatMessage,omitempty" yaml:"send_chat_message,omitempty" toml:"send_chat_message,omitempty"`

	// --- Misc ---
	TimeStampFormat            *TimestampFormat  `json:"timeStampFormat,omitempty"            yaml:"time_stamp_format,omitempty"              toml:"time_stamp_format,omitempty"`
	ForceRotorLibSimulation    *RotorLibMode     `json:"forceRotorLibSimulation,omitempty"    yaml:"force_rotor_lib_simulation,omitempty"     toml:"force_rotor_lib_simulation,omitempty"`
	RequiredBuild              *uint32           `json:"requiredBuild,omitempty"              yaml:"required_build,omitempty"                 toml:"required_build,omitempty"`
	StatisticsEnabled          *bool             `json:"statisticsEnabled,omitempty"          yaml:"statistics_enabled,omitempty"             toml:"statistics_enabled,omitempty"`
	ForcedDifficulty           *Difficulty       `json:"forcedDifficulty,omitempty"           yaml:"forced_difficulty,omitempty"              toml:"forced_difficulty,omitempty"`
	SteamProtocolMaxDataSize   *DataSize         `json:"steamProtocolMaxDataSize,omitempty"   yaml:"steam_protocol_max_data_size,omitempty"   toml:"steam_protocol_max_data_size,omitempty"`
	ArmaUnitsTimeout           *uint32           `json:"armaUnitsTimeout,omitempty"           yaml:"arma_units_timeout,omitempty"             toml:"arma_units_timeout,omitempty"`
	OverrideHazeQuality        *HazeQualityLevel `json:"overrideHazeQuality,omitempty"        yaml:"override_haze_quality,omitempty"          toml:"override_haze_quality,omitempty"`
	MissionHTTPDownloadBaseURL *string           `json:"missionHTTPDownloadBaseURL,omitempty" yaml:"mission_http_download_base_url,omitempty" toml:"mission_http_download_base_url,omitempty"`

	// --- Sub-configs ---
	AdvancedOptions *AdvancedOptions `json:"advancedOptions,omitempty" yaml:"advanced_options,omitempty" toml:"advanced_options,omitempty"`
	AntiFlood       *AntiFlood       `json:"antiFlood,omitempty"       yaml:"anti_flood,omitempty"       toml:"anti_flood,omitempty"`
}

// NewDefaultServerConfig returns ServerConfig populated with documented Arma 3 defaults.
func NewDefaultServerConfig() ServerConfig {
	arr := [4]uint8{1, 1, 1, 1}
	return ServerConfig{
		MaxPlayers:               ptr(uint16(64)),
		MotdInterval:             ptr(uint16(5)),
		VoteThreshold:            ptr(float32(0.5)),
		VoteMissionPlayers:       ptr(uint16(1)),
		MaxPing:                  ptr(int32(-1)),
		MaxPacketLoss:            ptr(int32(-1)),
		MaxDesync:                ptr(int32(-1)),
		DisconnectTimeout:        ptr(uint8(15)),
		KickClientsOnSlowNetwork: &arr,
		KickTimeout: []KickTimeout{
			{KickID: KickSourceManual, Timeout: 60},
			{KickID: KickSourceConnectivity, Timeout: 60},
			{KickID: KickSourceBattlEye, Timeout: 60},
			{KickID: KickSourceHarmless, Timeout: 60},
		},
		CallExtReportLimit:         ptr(uint32(1000)),
		VotingTimeOut:              []uint32{60, 90},
		RoleTimeOut:                []uint32{90, 120},
		BriefingTimeOut:            []uint32{60, 90},
		DebriefingTimeOut:          []uint32{45, 60},
		VerifySignatures:           ptr(SignaturesV2Only),
		BattlEye:                   ptr(true),
		DrawingInMap:               ptr(true),
		VonCodecQuality:            ptr(uint8(3)),
		VonCodec:                   ptr(VonCodecOPUS),
		AllowProfileGlasses:        ptr(true),
		ZeusCompositionScriptLevel: ptr(ZeusScriptsAttributesOnly),
		StatisticsEnabled:          ptr(true),
		SteamProtocolMaxDataSize:   ptr(NewDataSize(1024)),
		ArmaUnitsTimeout:           ptr(uint32(30)),
		OverrideHazeQuality:        ptr(HazeQualityOff),
		AdvancedOptions: &AdvancedOptions{
			LogObjectNotFound: ptr(true),
		},
		AntiFlood: &AntiFlood{
			CycleTime:      ptr(float32(0.5)),
			CycleLimit:     ptr(uint32(400)),
			CycleHardLimit: ptr(uint32(4000)),
		},
	}
}
