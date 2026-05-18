package arma

import (
	"fmt"
	"strings"
)

// cleanRateStr returns the first string representation of d (largest binary
// unit first, then SI) whose formatted value contains no '.' or ','.
// Returns empty string when all representations are imprecise (caller falls
// back to raw number).
func cleanRateStr(d DataTransferRate) string {
	bps := d.Bps()
	if bps == 0 {
		return ""
	}
	type unit struct {
		mult uint64
		fn   func() string
	}
	candidates := []unit{
		{uint64(gibi), d.AsGibps},
		{uint64(mebi), d.AsMibps},
		{uint64(kibi), d.AsKibps},
		{uint64(giga), d.AsGbps},
		{uint64(mega), d.AsMbps},
		{uint64(kilo), d.AsKbps},
		{1, d.AsBps},
	}
	for _, c := range candidates {
		if bps%c.mult != 0 {
			continue
		}
		s := c.fn()
		if !strings.ContainsAny(s, ".,") {
			return s
		}
	}
	return ""
}

// cfgWriter builds Arma 3 .cfg (C++-like class syntax) output.
type cfgWriter struct {
	buf    strings.Builder
	indent int
}

func (w *cfgWriter) line(s string) {
	w.buf.WriteString(strings.Repeat("\t", w.indent))
	w.buf.WriteString(s)
	w.buf.WriteByte('\n')
}

func (w *cfgWriter) str(key, val string) {
	w.line(fmt.Sprintf(`%s = "%s";`, key, strings.ReplaceAll(val, `"`, `\"`)))
}

func (w *cfgWriter) numU8(key string, val uint8) {
	w.line(fmt.Sprintf(`%s = %d;`, key, val))
}

func (w *cfgWriter) numI8(key string, val int8) {
	w.line(fmt.Sprintf(`%s = %d;`, key, val))
}

func (w *cfgWriter) numU16(key string, val uint16) {
	w.line(fmt.Sprintf(`%s = %d;`, key, val))
}

func (w *cfgWriter) numU32(key string, val uint32) {
	w.line(fmt.Sprintf(`%s = %d;`, key, val))
}

func (w *cfgWriter) numI32(key string, val int32) {
	w.line(fmt.Sprintf(`%s = %d;`, key, val))
}

func (w *cfgWriter) numU64(key string, val uint64) {
	w.line(fmt.Sprintf(`%s = %d;`, key, val))
}

func (w *cfgWriter) flt(key string, val float32) {
	w.line(fmt.Sprintf(`%s = %g;`, key, val))
}

func (w *cfgWriter) boolean(key string, val bool) {
	if val {
		w.line(key + " = true;")
	} else {
		w.line(key + " = false;")
	}
}

func (w *cfgWriter) strSlice(key string, vals []string) {
	quoted := make([]string, len(vals))
	for i, v := range vals {
		quoted[i] = `"` + strings.ReplaceAll(v, `"`, `\"`) + `"`
	}
	w.line(fmt.Sprintf(`%s[] = {%s};`, key, strings.Join(quoted, ", ")))
}

func (w *cfgWriter) uint32Slice(key string, vals []uint32) {
	parts := make([]string, len(vals))
	for i, v := range vals {
		parts[i] = fmt.Sprintf(`%d`, v)
	}
	w.line(fmt.Sprintf(`%s[] = {%s};`, key, strings.Join(parts, ", ")))
}

func (w *cfgWriter) uint8Array4(key string, vals [4]uint8) {
	w.line(fmt.Sprintf(`%s[] = {%d, %d, %d, %d};`, key, vals[0], vals[1], vals[2], vals[3]))
}

func (w *cfgWriter) kickTimeouts(key string, vals []KickTimeout) {
	parts := make([]string, len(vals))
	for i, v := range vals {
		parts[i] = fmt.Sprintf(`{%d, %d}`, v.KickID, v.Timeout)
	}
	w.line(fmt.Sprintf(`%s[] = {%s};`, key, strings.Join(parts, ", ")))
}

func (w *cfgWriter) disabledChannels(key string, vals []DisabledChannel) {
	parts := make([]string, len(vals))
	for i, v := range vals {
		parts[i] = fmt.Sprintf(`{%d, %s, %s, %s, %s}`,
			v.ChannelID,
			boolLit(v.Text),
			boolLit(v.Voice),
			boolLit(v.MapMarkers),
			boolLit(v.DrawOnMap),
		)
	}
	w.line(fmt.Sprintf(`%s[] = {%s};`, key, strings.Join(parts, ", ")))
}

func boolLit(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func (w *cfgWriter) beginClass(name string) {
	w.line(fmt.Sprintf("class %s", name))
	w.line("{")
	w.indent++
}

func (w *cfgWriter) endClass() {
	w.indent--
	w.line("};")
}

func (w *cfgWriter) missions(vals []Mission) {
	if len(vals) == 0 {
		w.line("class Missions {};")
		return
	}
	w.beginClass("Missions")
	for i, m := range vals {
		name := m.CustomName
		if name == "" {
			name = fmt.Sprintf("Mission%d", i+1)
		}
		w.beginClass(name)
		w.str("template", m.Template)
		w.str("difficulty", string(m.Difficulty))
		if len(m.Params) > 0 {
			w.beginClass("Params")
			for k, v := range m.Params {
				w.line(fmt.Sprintf(`%s = %d;`, k, v))
			}
			w.endClass()
		}
		w.endClass()
	}
	w.endClass()
}

func (w *cfgWriter) bytes() []byte {
	return []byte(w.buf.String())
}

// DumpServerConfig serialises cfg into Arma 3 server.cfg format.
// All non-nil fields are written; nil fields are omitted (Arma uses engine defaults).
func DumpServerConfig(cfg ServerConfig) []byte {
	var w cfgWriter

	if cfg.PasswordAdmin != nil {
		w.str("passwordAdmin", *cfg.PasswordAdmin)
	}
	if cfg.Password != nil {
		w.str("password", *cfg.Password)
	}
	if cfg.ServerCommandPassword != nil {
		w.str("serverCommandPassword", *cfg.ServerCommandPassword)
	}
	if cfg.Hostname != nil {
		w.str("hostname", *cfg.Hostname)
	}
	if cfg.LogFile != nil {
		w.str("logFile", *cfg.LogFile)
	}

	if cfg.Motd != nil {
		w.strSlice("motd", cfg.Motd)
	}
	if cfg.MotdInterval != nil {
		w.numU16("motdInterval", *cfg.MotdInterval)
	}

	if cfg.MaxPlayers != nil {
		w.numU16("maxPlayers", *cfg.MaxPlayers)
	}

	if cfg.Admins != nil {
		w.strSlice("admins", cfg.Admins)
	}
	if cfg.HeadlessClients != nil {
		w.strSlice("headlessClients", cfg.HeadlessClients)
	}
	if cfg.LocalClient != nil {
		w.strSlice("localClient", cfg.LocalClient)
	}
	if cfg.FilePatchingExceptions != nil {
		w.strSlice("filePatchingExceptions", cfg.FilePatchingExceptions)
	}
	if cfg.MissionWhitelist != nil {
		w.strSlice("missionWhitelist", cfg.MissionWhitelist)
	}

	if cfg.VoteThreshold != nil {
		w.flt("voteThreshold", *cfg.VoteThreshold)
	}
	if cfg.VoteMissionPlayers != nil {
		w.numU16("voteMissionPlayers", *cfg.VoteMissionPlayers)
	}

	if cfg.KickDuplicate != nil {
		w.boolean("kickduplicate", *cfg.KickDuplicate)
	}
	if cfg.Loopback != nil {
		w.boolean("loopback", *cfg.Loopback)
	}
	if cfg.Upnp != nil {
		w.boolean("upnp", *cfg.Upnp)
	}
	if cfg.Persistent != nil {
		w.boolean("persistent", *cfg.Persistent)
	}
	if cfg.AutoSelectMission != nil {
		w.boolean("autoSelectMission", *cfg.AutoSelectMission)
	}
	if cfg.RandomMissionOrder != nil {
		w.boolean("randomMissionOrder", *cfg.RandomMissionOrder)
	}
	if cfg.EnablePlayerDiag != nil {
		w.boolean("enablePlayerDiag", *cfg.EnablePlayerDiag)
	}

	if cfg.AllowedFilePatching != nil {
		w.numU8("allowedFilePatching", uint8(*cfg.AllowedFilePatching))
	}
	if cfg.AllowedLoadFileExtensions != nil {
		w.strSlice("allowedLoadFileExtensions", cfg.AllowedLoadFileExtensions)
	}
	if cfg.AllowedPreprocessFileExtensions != nil {
		w.strSlice("allowedPreprocessFileExtensions", cfg.AllowedPreprocessFileExtensions)
	}
	if cfg.AllowedHTMLLoadExtensions != nil {
		w.strSlice("allowedHTMLLoadExtensions", cfg.AllowedHTMLLoadExtensions)
	}
	if cfg.AllowedHTMLLoadURIs != nil {
		w.strSlice("allowedHTMLLoadURIs", cfg.AllowedHTMLLoadURIs)
	}

	if cfg.MaxPing != nil {
		w.numI32("maxPing", *cfg.MaxPing)
	}
	if cfg.MaxPacketLoss != nil {
		w.numI32("maxPacketLoss", *cfg.MaxPacketLoss)
	}
	if cfg.MaxDesync != nil {
		w.numI32("maxDesync", *cfg.MaxDesync)
	}
	if cfg.DisconnectTimeout != nil {
		w.numU8("disconnectTimeout", *cfg.DisconnectTimeout)
	}
	if cfg.KickClientsOnSlowNetwork != nil {
		w.uint8Array4("kickClientsOnSlowNetwork", *cfg.KickClientsOnSlowNetwork)
	}
	if cfg.KickTimeout != nil {
		w.kickTimeouts("kickTimeout", cfg.KickTimeout)
	}
	if cfg.CallExtReportLimit != nil {
		w.numU32("callExtReportLimit", *cfg.CallExtReportLimit)
	}

	if cfg.VotingTimeOut != nil {
		w.uint32Slice("votingTimeOut", cfg.VotingTimeOut)
	}
	if cfg.RoleTimeOut != nil {
		w.uint32Slice("roleTimeOut", cfg.RoleTimeOut)
	}
	if cfg.BriefingTimeOut != nil {
		w.uint32Slice("briefingTimeOut", cfg.BriefingTimeOut)
	}
	if cfg.DebriefingTimeOut != nil {
		w.uint32Slice("debriefingTimeOut", cfg.DebriefingTimeOut)
	}
	if cfg.LobbyIdleTimeout != nil {
		w.numU32("lobbyIdleTimeout", *cfg.LobbyIdleTimeout)
	}

	if cfg.MissionsToServerRestart != nil {
		w.numU32("missionsToServerRestart", *cfg.MissionsToServerRestart)
	}
	if cfg.MissionsToShutdown != nil {
		w.numU32("missionsToShutdown", *cfg.MissionsToShutdown)
	}
	if cfg.Missions != nil {
		w.missions(cfg.Missions)
	}

	if cfg.DisableChannels != nil {
		w.disabledChannels("disableChannels", cfg.DisableChannels)
	}

	if cfg.VerifySignatures != nil {
		w.numU8("verifySignatures", uint8(*cfg.VerifySignatures))
	}
	if cfg.EqualModRequired != nil {
		w.boolean("equalModRequired", *cfg.EqualModRequired)
	}
	if cfg.BattlEye != nil {
		w.boolean("BattlEye", *cfg.BattlEye)
	}

	if cfg.DrawingInMap != nil {
		w.boolean("drawingInMap", *cfg.DrawingInMap)
	}
	if cfg.DisableVoN != nil {
		w.boolean("disableVoN", *cfg.DisableVoN)
	}
	if cfg.VonCodecQuality != nil {
		w.numU8("vonCodecQuality", *cfg.VonCodecQuality)
	}
	if cfg.VonCodec != nil {
		w.numU8("vonCodec", uint8(*cfg.VonCodec))
	}
	if cfg.SkipLobby != nil {
		w.boolean("skipLobby", *cfg.SkipLobby)
	}
	if cfg.AllowProfileGlasses != nil {
		w.boolean("allowProfileGlasses", *cfg.AllowProfileGlasses)
	}
	if cfg.ZeusCompositionScriptLevel != nil {
		w.numU8("zeusCompositionScriptLevel", uint8(*cfg.ZeusCompositionScriptLevel))
	}

	if cfg.DoubleIdDetected != nil {
		w.str("doubleIdDetected", *cfg.DoubleIdDetected)
	}
	if cfg.OnUserConnected != nil {
		w.str("onUserConnected", *cfg.OnUserConnected)
	}
	if cfg.OnUserDisconnected != nil {
		w.str("onUserDisconnected", *cfg.OnUserDisconnected)
	}
	if cfg.OnHackedData != nil {
		w.str("onHackedData", *cfg.OnHackedData)
	}
	if cfg.OnDifferentData != nil {
		w.str("onDifferentData", *cfg.OnDifferentData)
	}
	if cfg.OnUnsignedData != nil {
		w.str("onUnsignedData", *cfg.OnUnsignedData)
	}
	if cfg.OnUserKicked != nil {
		w.str("onUserKicked", *cfg.OnUserKicked)
	}
	if cfg.RegularCheck != nil {
		w.str("regularCheck", *cfg.RegularCheck)
	}

	if cfg.TimeStampFormat != nil {
		w.str("timeStampFormat", string(*cfg.TimeStampFormat))
	}
	if cfg.ForceRotorLibSimulation != nil {
		w.numU8("forceRotorLibSimulation", uint8(*cfg.ForceRotorLibSimulation))
	}
	if cfg.RequiredBuild != nil {
		w.numU32("requiredBuild", *cfg.RequiredBuild)
	}
	if cfg.StatisticsEnabled != nil {
		w.boolean("statisticsEnabled", *cfg.StatisticsEnabled)
	}
	if cfg.ForcedDifficulty != nil {
		w.str("forcedDifficulty", string(*cfg.ForcedDifficulty))
	}
	if cfg.SteamProtocolMaxDataSize != nil {
		w.numU64("steamProtocolMaxDataSize", cfg.SteamProtocolMaxDataSize.B())
	}
	if cfg.ArmaUnitsTimeout != nil {
		w.numU32("armaUnitsTimeout", *cfg.ArmaUnitsTimeout)
	}
	if cfg.OverrideHazeQuality != nil {
		w.numI8("overrideHazeQuality", int8(*cfg.OverrideHazeQuality))
	}
	if cfg.MissionHTTPDownloadBaseURL != nil {
		w.str("missionHTTPDownloadBaseURL", *cfg.MissionHTTPDownloadBaseURL)
	}

	if cfg.AdvancedOptions != nil {
		ao := cfg.AdvancedOptions
		w.beginClass("AdvancedOptions")
		if ao.LogObjectNotFound != nil {
			w.boolean("logObjectNotFound", *ao.LogObjectNotFound)
		}
		if ao.SkipDescriptionParsing != nil {
			w.boolean("skipDescriptionParsing", *ao.SkipDescriptionParsing)
		}
		if ao.IgnoreMissionLoadErrors != nil {
			w.boolean("ignoreMissionLoadErrors", *ao.IgnoreMissionLoadErrors)
		}
		if ao.QueueSizeLogG != nil {
			w.numU32("queueSizeLogG", *ao.QueueSizeLogG)
		}
		w.endClass()
	}

	if cfg.AntiFlood != nil {
		af := cfg.AntiFlood
		w.beginClass("AntiFlood")
		if af.CycleTime != nil {
			w.flt("cycleTime", *af.CycleTime)
		}
		if af.CycleLimit != nil {
			w.numU32("cycleLimit", *af.CycleLimit)
		}
		if af.CycleHardLimit != nil {
			w.numU32("cycleHardLimit", *af.CycleHardLimit)
		}
		if af.EnableKick != nil {
			w.boolean("enableKick", *af.EnableKick)
		}
		w.endClass()
	}

	return w.bytes()
}

// DumpBasicServerConfig serialises cfg into Arma 3 basic.cfg format.
// All non-nil fields are written; nil fields are omitted (Arma uses engine defaults).
func DumpBasicServerConfig(cfg BasicServerConfig) []byte {
	var w cfgWriter

	if cfg.Language != nil {
		w.str("language", *cfg.Language)
	}
	if cfg.MaxMsgSend != nil {
		w.numU16("MaxMsgSend", *cfg.MaxMsgSend)
	}
	if cfg.MaxSizeGuaranteed != nil {
		w.numU16("MaxSizeGuaranteed", *cfg.MaxSizeGuaranteed)
	}
	if cfg.MaxSizeNonguaranteed != nil {
		w.numU16("MaxSizeNonguaranteed", *cfg.MaxSizeNonguaranteed)
	}
	if cfg.MinBandwidth != nil {
		if s := cleanRateStr(*cfg.MinBandwidth); s != "" {
			w.str("MinBandwidth", s)
		} else {
			w.numU64("MinBandwidth", cfg.MinBandwidth.Bps())
		}
	}
	if cfg.MaxBandwidth != nil {
		if s := cleanRateStr(*cfg.MaxBandwidth); s != "" {
			w.str("MaxBandwidth", s)
		} else {
			w.numU64("MaxBandwidth", cfg.MaxBandwidth.Bps())
		}
	}
	if cfg.MinErrorToSend != nil {
		w.flt("MinErrorToSend", *cfg.MinErrorToSend)
	}
	if cfg.MinErrorToSendNear != nil {
		w.flt("MinErrorToSendNear", *cfg.MinErrorToSendNear)
	}
	if cfg.MaxCustomFileSize != nil {
		w.numU16("MaxCustomFileSize", *cfg.MaxCustomFileSize)
	}

	if cfg.Sockets != nil {
		w.beginClass("sockets")
		if cfg.Sockets.MaxPacketSize != nil {
			w.numU16("maxPacketSize", *cfg.Sockets.MaxPacketSize)
		}
		w.endClass()
	}

	return w.bytes()
}
