# Arma 3 Startup Parameters (Server)

## Purpose

Startup parameters are CLI arguments passed to the Arma 3 server executable. ZEUS builds
the full command from a profile's `params:` YAML section, then adds managed paths
(configs, profiles, PID file) automatically.

Parameter names are **case-insensitive**. Values containing spaces must be quoted.

## ZEUS-Managed Parameters

These are set automatically by ZEUS and **must not be set manually** in the profile YAML.
Overriding them will produce unexpected behavior.

| Parameter | Managed value |
|-----------|--------------|
| `-config` | `<install_dir>/.zeus/<name>/configs/server.cfg` |
| `-cfg` | `<install_dir>/.zeus/<name>/configs/basic.cfg` |
| `-profiles` | `<install_dir>/.zeus/<name>` |
| `-pid` | `~/.zeus/running/<name>.pid` |
| `-name` | profile name |
| `-keysFolder` | `<install_dir>/.zeus/<name>/keys` (appended) |
| `-mpmissions` | `<install_dir>/.zeus/<name>/mpmissions` (if not set) |

## User-Configurable Parameters

These map directly to fields in the `params:` YAML section.

### Core Server

| CLI Parameter | YAML key | Description |
|---------------|----------|-------------|
| `-server` | `server: true` | Start as dedicated server (not needed for `arma3server_x64.exe`). |
| `-port=2302` | `port: 2302` | UDP port to listen on. Default: 2302. |
| `-mod=path1;path2` | `mod: [path1, path2]` | Load mod folders (semicolon-separated). Relative to install dir or absolute. |
| `-serverMod=path1` | `server_mod: [path1]` | Server-side mods not broadcast to clients. |

### Paths

| CLI Parameter | YAML key | Description |
|---------------|----------|-------------|
| `-mpmissions=dir` | `mp_missions: "dir"` | Alternative MPMissions directory (relative to Arma 3 root). Overrides ZEUS default if set. |

### Performance

| CLI Parameter | YAML key | Description |
|---------------|----------|-------------|
| `-limitFPS=100` | `limit_fps: 100` | Server simulation cycle rate limit (5–1000). Default: 50. Servers with no players auto-cap at 30 regardless. *(since 2.00)* |
| `-maxMem=4096` | `max_mem: 4096` | Memory allocation limit in MiB. Engine auto-selects if omitted. Minimum: 1024. |

> **Warning (Linux, before 2.14):** `-maxMem` parsed as signed int; values like 4096
> were negative. Use `4095`, `8191`, `16383` etc. to avoid.

## Other Commonly Used Parameters

These are not in the profile YAML struct but are worth knowing.

### Startup Behavior

| Parameter | Description |
|-----------|-------------|
| `-autoInit` | Initialize mission immediately on server start (requires `persistent = 1` in server.cfg). **Breaks mission parameters** — only defaults are returned. |
| `-loadMissionToMemory` | Cache mission in RAM on first client download; reused for subsequent clients. |
| `-netlog` | Enable network traffic logging to `net.log`. |
| `-ranking=path` | Write player stats (kills, score) to file. |

### Process Control

| Parameter | Description |
|-----------|-------------|
| `-pid=path` | File to write server PID. Removed automatically on clean exit. Dedicated server only. |

### BattlEye

| Parameter | Description |
|-----------|-------------|
| `-bePath=dir` | Custom BattlEye folder. Defaults to `BattlEye/` inside profile folder. |

### Network

| Parameter | Description |
|-----------|-------------|
| `-ip=x.x.x.x` | Bind to specific IP (multihome servers). |
| `-bandwidthAlg=2` | Use experimental networking algorithm. |
| `-disableServerThread` | Disable send messaging thread (may resolve crashes at performance cost). |

### CPU and Threading

| Parameter | Description |
|-----------|-------------|
| `-cpuCount=8` | Override CPU core count detection. Minimum 2 since 2.20. |
| `-enableHT` | Use all logical cores including HT/SMT. Overridden by `-cpuCount` or `-cpuAffinity`. |
| `-exThreads=7` | Extra thread flags: `0`=none, `1`=file ops, `3`=file+texture, `5`=file+geometry, `7`=all. |
| `-malloc=name` | Use specific memory allocator. |
| `-hugePages` | Enable hugepages with default allocator. |

> **Warning:** `-setThreadCharacteristics` can freeze Windows Server. Do not use on
> dedicated server OS.

## Path Rules

- **Relative paths** are relative to the Arma 3 installation directory (where the
  executable resides), not the current working directory.
- **Spaces** in paths require quoting: `-profiles="E:\Arma 3\Profiles"` or
  `"-profiles=E:\Arma 3\Profiles"`.
- `-mod`, `-serverMod`, `-keysFolder` values are semicolon-separated. On Linux,
  escape semicolons: `-mod=mod1\;mod2`.

## Key File: PID

The `-pid` file is created on server start and deleted on clean exit. ZEUS uses it to
track running instances. If the server crashes, the PID file may remain; ZEUS's
`WaitGone` check uses both process existence and PID file presence.

## Executable Names

| OS | Executable |
|----|-----------|
| Windows | `arma3server_x64.exe` |
| Linux | `arma3server_x64` |

ZEUS resolves this automatically from `runtime.GOOS` unless `executable` is set in the
profile YAML.

## Related

- [server.cfg Reference](../formats/server_cfg.md)
- [basic.cfg Reference](../formats/basic_cfg.md)
- [Mission Rotation](../formats/server_cfg_missions.md)
