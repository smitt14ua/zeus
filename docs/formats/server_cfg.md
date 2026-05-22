# server.cfg Reference

## Purpose

`server.cfg` is the primary Arma 3 dedicated server configuration file. It controls
authentication, player limits, voting behavior, network thresholds, and server
lifecycle settings.

ZEUS generates this file at `<install_dir>/.zeus/<profile_name>/configs/server.cfg`
from the `config:` section of a profile YAML. Only values that differ from Arma 3
defaults are written.

## File Loading

The filename is arbitrary. The real path is passed via the `-config` startup parameter.
When no `-config` is specified, no server config is loaded (there is no implicit default).

## Parameter Reference

### Authentication

| Parameter | Default | Description |
|-----------|---------|-------------|
| `passwordAdmin = "xyz";` | `""` | Password for `#login` admin access. |
| `password = "xyz";` | `""` | Password required to connect. |
| `serverCommandPassword = "xyz";` | `""` | Password for `serverCommand` scripting (case-sensitive). |
| `admins[] = { "<UID>" };` | `{}` | Steam UIDs that can `#login` without password. *(since 1.70)* |
| `filePatchingExceptions[] = { "<UID>" };` | `{}` | Steam UIDs exempt from `allowedFilePatching` and `verifySignatures`. Signature errors are still logged. *(operational since 2.10)* |

### Server Identity

| Parameter | Default | Description |
|-----------|---------|-------------|
| `hostname = "My Server";` | local machine name | Name shown in server browser. |
| `maxPlayers = 64;` | `64` (DS) | Maximum players. Capped by mission slot count. |
| `motd[] = { "line1", "line2" };` | `{}` | Message of the Day lines shown on join. |
| `motdInterval = 5;` | `5` | Seconds between MOTD lines. |
| `headlessClients[] = { "<IP>" };` | `{}` | IPs of connected Headless Clients. |
| `localClient[] = { "<IP>" };` | `{}` | IPs treated as unlimited bandwidth (HC use). |

### Server Behaviour

| Parameter | Default | Description |
|-----------|---------|-------------|
| `voteThreshold = 0.33;` | `0.5` | Fraction of players needed to confirm a vote. |
| `voteMissionPlayers = 1;` | `1` | Players required before mission voting starts. |
| `allowedVoteCmds[] = {...};` | *(see below)* | Permitted vote commands and their thresholds. |
| `allowedVotedAdminCmds[] = {...};` | *(see below)* | Commands available to voted-in admins. |
| `kickduplicate = 1;` | `0` | Kick players with duplicate game IDs (`1` = active). |
| `loopback = 1;` | `false` | Force LAN mode (allows multiple local instances; blocks external). |
| `upnp = 1;` | `false` | Auto-create UPnP/IGD port mappings. **Warning:** can delay startup by 600s if blocked. |
| `allowedFilePatching = 0;` | `0` | `0` = no clients, `1` = HC only, `2` = all. *(since 1.50)* |
| `persistent = 1;` | `0` | Mission keeps running after all players disconnect. Only effective with `base`/`instant` respawn types. |
| `requiredBuild = 12345;` | `0` | Minimum client build number. Values above current server build are clamped down. |
| `missionsToServerRestart = 8;` | `0` | Mission-end events before server process restarts. Cannot combine with `missionsToShutdown`. |
| `missionsToShutdown = 8;` | `0` | Mission-end events before server process shuts down. |
| `autoSelectMission = true;` | `false` | Auto-start next mission in cycle without admin. |
| `randomMissionOrder = true;` | `false` | Randomize mission selection order. |

#### allowedVoteCmds

Controls which vote commands players may initiate and under what conditions.
Each entry is a positional tuple; trailing optional fields may be omitted.

```cpp
allowedVoteCmds[] = {
    // { commandName, preMissionStart, postMissionStart, votingThreshold, percentSideVotingThreshold }
    { "admin",   true, true },
    { "kick",    true, true, 0.33 },
    { "mission", true, true, 0.5,  0.5 }
};
```

| Position | Type | Default | Description |
|----------|------|---------|-------------|
| 0 — `commandName` | String | *(required)* | Vote command name, e.g. `"admin"`, `"kick"`. |
| 1 — `preMissionStart` | Boolean | `true` | Allow this vote before the mission starts. |
| 2 — `postMissionStart` | Boolean | `true` | Allow this vote after the mission starts. |
| 3 — `votingThreshold` | Number 0–1 | `voteThreshold` value | Per-command override for the fraction of players required. |
| 4 — `percentSideVotingThreshold` | Number 0–1 | `0.5` | Side-specific vote threshold. *(since Arma 3 1.90+ PerformanceBranch)* |

#### allowedVotedAdminCmds

Controls which admin commands a *voted-in* admin may use.

```cpp
allowedVotedAdminCmds[] = {
    { "mission",  true, true },
    { "missions", true, true },
    { "restart",  true, true },
    { "reassign", true, true },
    { "kick",     true, true }
};
```

| Position | Type | Default | Description |
|----------|------|---------|-------------|
| 0 — `commandName` | String | *(required)* | Admin command name, e.g. `"kick"`, `"restart"`. |
| 1 — `preMissionStart` | Boolean | `true` | Allow command before the mission starts. |
| 2 — `postMissionStart` | Boolean | `true` | Allow command after the mission starts. |

> **Note:** `allowedVoteCmds[] = {};` (empty array) disables *all* player vote commands.
> Omitting `allowedVoteCmds` entirely allows every voting command.

> **Note:** `allowedVotedAdminCmds[] = {};` (empty array) disables *all* voted-admin commands.
> Omitting `allowedVotedAdminCmds` entirely grants voted-in admins unrestricted access.

In ZEUS profiles the two fields use snake_case keys:

```yaml
config:
  allowed_vote_cmds:
    - name: "admin"
      pre_mission_start: true
      post_mission_start: true
    - name: "kick"
      voting_threshold: 0.33
  allowed_voted_admin_cmds:
    - name: "mission"
      pre_mission_start: true
      post_mission_start: true
    - name: "restart"
      pre_mission_start: true
      post_mission_start: true
```

### File Access Restrictions

| Parameter | Default | Description |
|-----------|---------|-------------|
| `allowedLoadFileExtensions[] = { "sqf", "txt" };` | (undefined = all) | Extensions permitted for `loadFile`. Empty array = nothing allowed. |
| `allowedPreprocessFileExtensions[] = { "sqf" };` | (undefined = all) | Extensions permitted for `preprocessFile`. |
| `allowedHTMLLoadExtensions[] = { "htm" };` | (undefined = all) | Extensions permitted for `htmlLoad`. |
| `allowedHTMLLoadURIs[] = { "http://..." };` | (undefined = all) | Allowed URIs for `htmlLoad`. Comment out to let missions decide. |

> **Note:** `loadFile`, `preprocessFile`, `preprocessFileLineNumbers` only access files
> within the Arma 3 server directory tree. These restrictions are server-side only.

### Network Thresholds

Values of `-1` mean unlimited (no kicking/logging).

| Parameter | Default | Description |
|-----------|---------|-------------|
| `MaxPing = 200;` | `-1` | Max ping (ms) before kick action. *(since 1.56)* |
| `MaxPacketLoss = 50;` | `-1` | Max packet loss (%) before kick action. *(since 1.56)* |
| `MaxDesync = 150;` | `-1` | Max desync value before kick action. *(since 1.56)* |
| `DisconnectTimeout = 5;` | `15` | Seconds to wait for reconnect before disconnecting (range 1–90). *(since 1.56)* |
| `kickClientsOnSlowNetwork[] = { 0, 0, 0, 0 };` | `{ 1, 1, 1, 1 }` | Per-threshold action: `0` = log only, `1` = kick. Order: `{ MaxPing, MaxPacketLoss, MaxDesync, DisconnectTimeout }`. *(since 1.56)* |

### Kick Timeouts

```cpp
kickTimeout[] = {
    { 0, -1  },   // manual kick (vote/admin/bruteforce) — until mission end
    { 1, 180 },   // connectivity kick — 180s
    { 2, 180 },   // BattlEye kick — 180s
    { 3, 180 }    // harmless kick (signatures/addons/steam) — 180s
};
```

Defaults: all categories = 60s. Timeout values: `> 0` seconds, `-1` = until mission end, `-2` = until server restart. *(since 1.90)*

### Lobby & Phase Timeouts

| Parameter | Default | Description |
|-----------|---------|-------------|
| `votingTimeOut = 60;` or `votingTimeOut[] = { 60, 90 };` | `60` / `{ 60, 90 }` | Mission voting timeout. Array form: `{ ready, notReady }`. *(array since 1.90)* |
| `roleTimeOut = 90;` or `roleTimeOut[] = { 90, 120 };` | `90` / `{ 90, 120 }` | Role selection timeout. |
| `briefingTimeOut = 60;` or `briefingTimeOut[] = { 60, 90 };` | `60` / `{ 60, 90 }` | Briefing (map) screen timeout. |
| `debriefingTimeOut = 45;` or `debriefingTimeOut[] = { 45, 60 };` | `45` / `{ 45, 60 }` | Debriefing screen timeout. |
| `lobbyIdleTimeout = 300;` | `0` | Time before server force-starts without admin. Effective minimum is `MAX(voting, lobby, briefing, debriefing) + 5s`. *(since 1.90)* |

### Voice over Network (VoN)

| Parameter | Default | Description |
|-----------|---------|-------------|
| `disableVoN = 1;` | `0` | `1` = disable VoN globally. |
| `vonCodecQuality = 10;` | `3` | Codec quality 1–30. 0–10 = 8kHz, 11–20 = 16kHz, 21–30 = 32kHz/48kHz. |
| `vonCodec = 1;` | `1` | `0` = SPEEX, `1` = OPUS (IETF standard, recommended). *(since 1.58)* |

### Channel Control *(since 1.60, extended in 2.20)*

```cpp
disableChannels[] = {
    { 0, false, true, false, true },   // Global: voice+drawing disabled
    { 3, true,  true }                 // Group: text+voice disabled
};
// { channelID, disableText, disableVoice [, disableMapMarkers, disableDrawing] }
// Channel IDs: 0=Global 1=Side 2=Command 3=Group 4=Vehicle 5=Direct 16=System
```

> **Note:** Mission `Description.ext#disableChannels` overrides server.cfg settings.

### Other Options

| Parameter | Default | Description |
|-----------|---------|-------------|
| `verifySignatures = 2;` | `2` | `0` = disabled, `1` = deprecated (falls back to 2), `2` = v2 signatures only. |
| `equalModRequired = 1;` | `0` | Outdated. Require identical `-mod=` on clients. |
| `drawingInMap = false;` | `true` | Allow map markers and drawing. *(since 1.64)* |
| `skipLobby = false;` | `false` | Skip role selection for joining players (overridden by mission config). |
| `allowProfileGlasses = false;` | `true` | Allow player-profile glasses. *(since 2.06)* |
| `zeusCompositionScriptLevel = 0;` | `1` | `0` = no scripts, `1` = attributes only, `2` = all scripts. *(since 2.06)* |
| `BattlEye = 1;` | `1` | Enable BattlEye anti-cheat. |
| `timeStampFormat = "short";` | `""` | RPT log timestamp: `"none"`, `"short"`, `"full"`. |
| `forceRotorLibSimulation = 0;` | `0` | `0` = player choice, `1` = force AFM, `2` = force SFM. *(since 1.34)* |
| `forcedDifficulty = "veteran";` | `""` | Enforce difficulty. Mission cycle difficulty overrides this. *(since 1.56)* |
| `missionWhitelist[] = { "intro.altis" };` | `{}` | Restrict admin's mission-change options. *(since 1.56)* |
| `steamProtocolMaxDataSize = 1024;` | `1024` | Steam Query packet limit (bytes). Increasing risks fragmented UDP on some routers. *(since 2.00)* |
| `armaUnitsTimeout = 30;` | `30` | Seconds to wait for Arma Units data on connect. *(since 2.06)* |
| `overrideHazeQuality = 1;` | `-1` | Force haze quality: `0`=VeryLow, `1`=Low, `2`=Standard, `-1`=no forcing. Mission config has priority. *(since 2.16)* |
| `statisticsEnabled = 1;` | `1` | `0` to opt out of Arma 3 analytics. *(since 1.56)* |
| `logFile = "server_console.log";` | `""` | Path for dedicated server console log. Does not affect `net.log` (`-netlog`). |
| `missionHTTPDownloadBaseURL = "https://...";` | `""` | HTTP base URL for client mission download. URL must start with `http(s)://` and end with `/` or `=`. Max 512 MB. Falls back to server transfer on failure. *(since 2.20)* |

### Server-Side Scripting Callbacks

These are SQF commands executed on specific events. Assign a string containing SQF code.

| Parameter | Event |
|-----------|-------|
| `onUserConnected = "command";` | Player connected |
| `onUserDisconnected = "command";` | Player disconnected |
| `onUserKicked = "command";` | Player kicked |
| `doubleIdDetected = "command";` | Duplicate game ID |
| `onHackedData = "command";` | Signature tampering detected |
| `onDifferentData = "command";` | Valid signature but wrong version |
| `onUnsignedData = "command";` | Unsigned data detected |
| `regularCheck = "command";` | Periodic check |
| `onPlayerJoinAttempt = "command";` | Called repeatedly while a player is joining. Must return `"ACCEPT"`, `"DELAY"`, or `"REFUSE"` (optionally `"REFUSE_<message>"`). |
| `sendChatMessage = "command";` | Send a "System" chat message to a specified user. |

### AdvancedOptions Class *(since 2.02)*

```cpp
class AdvancedOptions
{
    logObjectNotFound = 1;        // log "Server: Object not found" (default: 1)
    skipDescriptionParsing = 0;   // skip description.ext/mission.sqm parsing (default: 0)
    ignoreMissionLoadErrors = 0;  // load mission despite errors (default: 0, since 2.04)
    queueSizeLogG = 1000000;      // dump player queue to log if > N bytes (0=disabled, since 2.08)
};
```

### AntiFlood Class *(since 2.18)*

```cpp
class AntiFlood
{
    cycleTime = 0.5;       // seconds per cycle
    cycleLimit = 400;      // messages per cycle before flagging
    cycleHardLimit = 4000; // messages per cycle for immediate action
    enableKick = 0;        // 1 = kick; 0 = log only
};
```

A player is kicked if: last 4 of 8 cycles were flagged, **or** the hard limit was exceeded in one cycle.

## Security Best Practices

Recommended settings for public servers:

```cpp
BattlEye = 1;
verifySignatures = 2;
allowedFilePatching = 0;
allowedLoadFileExtensions[] = {
    "hpp","sqs","sqf","fsm","cpp","paa","txt","xml","inc","ext",
    "sqm","ods","fxy","lip","csv","kb","bik","bikb","html","htm","biedi"
};
allowedPreprocessFileExtensions[] = {
    "hpp","sqs","sqf","fsm","cpp","paa","txt","xml","inc","ext",
    "sqm","ods","fxy","lip","csv","kb","bik","bikb","html","htm","biedi"
};
allowedHTMLLoadExtensions[] = { "htm", "html", "xml", "txt" };
passwordAdmin = "strongpassword";
serverCommandPassword = "anotherpassword";
```

> **Warning:** Allowed extension arrays cover both files inside and outside PBOs.
> Changing defaults without testing may break game features.

## Server Administration

Only one non-BattlEye-RCon admin can be active at a time.

**Voted admin:** elected by player vote; limited command set.

**Logged-in admin:** uses `#login <password>`. If UID is in `admins[]`, `#login` works without password.

Rules:
- If a logged-in admin is present, no second admin can log in until the first logs out.
- A logged-in admin overrides a voted-in admin.
- Both login methods grant identical permissions.

## Related

- [Mission Rotation](server_cfg_missions.md)
- [Startup Parameters](../protocols/startup_params.md)
- [basic.cfg Reference](basic_cfg.md)
