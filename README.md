# ZEUS

[![codecov](https://codecov.io/gh/smitt14ua/zeus/graph/badge.svg)](https://codecov.io/gh/smitt14ua/zeus)
[![Downloads](https://img.shields.io/github/downloads/smitt14ua/zeus/total?label=downloads)](https://github.com/smitt14ua/zeus/releases)
[![Latest release](https://img.shields.io/github/v/release/smitt14ua/zeus)](https://github.com/smitt14ua/zeus/releases/latest)

CLI tool for managing Arma 3 dedicated server profiles on Windows and Linux.

A **profile** bundles your server config, startup parameters, mod list, and RCon settings into a single file (YAML, TOML, or JSON). ZEUS generates the required `.cfg` files, launches the server process, and tracks running instances.

## Installation

### Linux / macOS

```sh
curl -fsSL https://raw.githubusercontent.com/smitt14ua/zeus/main/install.sh | sh
```

Installs to `/usr/local/bin/zeus`. Uses `sudo` automatically if needed.

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/smitt14ua/zeus/main/install.ps1 | iex
```

Installs to `%LOCALAPPDATA%\Programs\Zeus\zeus.exe` and adds the directory to your user PATH.

### Manual installation

Download the binary for your platform from the [latest release](https://github.com/smitt14ua/zeus/releases/latest):

| Platform        | Asset                   |
|-----------------|-------------------------|
| Linux x86_64    | `zeus-linux-amd64`      |
| Linux arm64     | `zeus-linux-arm64`      |
| macOS x86_64    | `zeus-darwin-amd64`     |
| macOS arm64     | `zeus-darwin-arm64`     |
| Windows x86_64  | `zeus-windows-amd64.exe`|

```sh
# Linux example
curl -fsSL -o zeus https://github.com/smitt14ua/zeus/releases/latest/download/zeus-linux-amd64
chmod +x zeus
sudo mv zeus /usr/local/bin/
```

### Install with Go

```sh
go install github.com/smitt14ua/zeus@latest
```

### Build from source

```sh
git clone https://github.com/smitt14ua/zeus
cd zeus
go build -o zeus .
```

## Quick Start

**1. Generate a profile template:**

```bash
zeus new my-server --yaml > server.yaml
```

**2. Edit `server.yaml`** — set `install_dir` and tweak config as needed.

**3. Register the profile:**

```bash
zeus add server.yaml
```

**4. Start the server:**

```bash
zeus start my-server
```

## Commands

### Top-level shortcuts

These are aliases for their `zeus profile ...` equivalents.

| Command | Description |
|---------|-------------|
| `zeus new <name>` | Generate a profile template with Arma 3 defaults |
| `zeus add [file]` | Register or update a profile |
| `zeus start <name>` | Start a profile |
| `zeus stop <name>` | Stop a running profile |
| `zeus agent` | Connect to a remote web panel as a managed agent |

### Profile management (`zeus profile ...`)

| Command | Description |
|---------|-------------|
| `zeus profile new <name>` | Generate a profile template |
| `zeus profile add [file]` | Register or update a profile |
| `zeus profile ls` | List all profiles with status |
| `zeus profile info <name>` | Show profile details |
| `zeus profile start <name>` | Start a profile |
| `zeus profile stop <name>` | Stop a running profile |
| `zeus profile rm <name>` | Delete a profile and all its data |

### Mission management (`zeus missions ...`)

| Command | Description |
|---------|-------------|
| `zeus missions pull <name>` | Sync `.pbo` mission files from the configured source into the profile's mpmissions directory |

### `zeus agent`

Connect ZEUS to a remote web panel server over an outbound WebSocket (WSS) connection. The panel can then list profiles, start/stop servers, stream output, and pull missions without opening inbound ports on the Arma 3 machine.

```bash
zeus agent --url wss://panel.example.com/ws/agent --token my-secret
zeus agent --url wss://panel.example.com/ws/agent --token my-secret --name server-A
zeus agent --url wss://panel.example.com/ws/agent --token my-secret --allow view,control
```

| Flag | Description |
|------|-------------|
| `--url` | WebSocket URL of the web panel (required) |
| `--token` | Shared secret token (required) |
| `--name` | Agent name shown in the panel (default: hostname) |
| `--heartbeat` | Heartbeat interval (default `1s`) |
| `--reconnect` | Reconnect delay after disconnect (default `5s`) |
| `--allow` | Comma-separated scope allowlist: `view`, `control`, `manage`, `update`, `all` (default: all) |

The agent reconnects automatically on disconnect. See [`docs/agent-protocol.md`](docs/agent-protocol.md) for the full protocol specification and server implementation guide.

---

## Flags

### `new` / `profile new`

| Flag | Description |
|------|-------------|
| `--yaml` | Output as YAML |
| `--toml` | Output as TOML |
| `--json` | Output as JSON (default) |

### `profile ls`

| Flag | Description |
|------|-------------|
| `--json` | Output as compact JSON (one line per call) |

### `add` / `profile add`

Accepts YAML, TOML, or JSON files. Also reads from stdin.

```bash
zeus add server.yaml
zeus add server.toml
zeus add server.json
zeus add server.yaml --name staging
cat server.json | zeus add
```

> **Note:** Stdin input must be JSON. File input auto-detects format from the `.yaml`, `.toml`, or `.json` extension.

| Flag | Short | Description |
|------|-------|-------------|
| `--name` | `-n` | Override the profile name from the file |
| `--copy-keys` | `-k` | Copy all mod `.bikey` files without prompting |
| `--force` | `-f` | Skip all prompts using default answers |

### `start` / `profile start`

`zeus start` blocks until the server is confirmed running: it waits for Arma 3 to write its PID file and then holds a short stability window to ensure the process hasn't crashed immediately. Use `--start-timeout` to control how long it waits.

Before launching, ZEUS checks every running profile's effective game port. If another profile is already bound to the same port (default 2302), the command exits immediately with an error before hooks are run or the process is spawned.

| Flag | Short | Description |
|------|-------|-------------|
| `--dry-run` | `-n` | Print the generated launch command without executing |
| `--start-timeout` | | How long to wait for the server PID file to appear (default `1m0s`) |

### `profile info`

| Flag | Description |
|------|-------------|
| `--yaml` | Output as YAML |
| `--json` | Output as JSON |
| `--toml` | Output as TOML |
| _(none)_ | Human-readable console output (default) |

### `profile rm`

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Delete without confirmation prompt |

### `missions pull`

| Flag | Short | Description |
|------|-------|-------------|
| `--dry-run` | `-n` | Show what would change without making changes |

> **Note:** ZEUS sets the mpmissions directory via the `-mpmissions` startup parameter. This parameter had a bug in Arma 3 that was fixed in revision 153745 — on older builds it may not work correctly. See [feedback.bistudio.com/T199168](https://feedback.bistudio.com/T199168) for details.

## Profile format

Profiles are YAML, TOML, or JSON. Generate a template with `zeus new <name> --yaml`.

### Profile name rules

Profile names must match `[A-Za-z0-9][A-Za-z0-9_.-]{0,63}`:

- First character must be a letter or digit (no leading `.` or `-`)
- Remaining characters: letters, digits, `_`, `.`, `-`
- Maximum 64 characters
- Windows reserved device names are rejected cross-platform (`CON`, `NUL`, `COM1`–`COM9`, `LPT1`–`LPT9`)

These restrictions prevent path traversal and OS-level filename normalization issues on both Linux and Windows.

```yaml
name: my-server
install_dir: /opt/arma3

params:
  port: 2302
  limit_fps: 100
  mod:
    - /opt/arma3/@CBA_A3        # absolute path
    - @ACE                      # relative — resolved to <install_dir>/@ACE
  # See https://community.bistudio.com/wiki/Arma_3:_Startup_Parameters
  # Parameters use snake_case

config:
  hostname: My Arma 3 Server
  max_players: 32
  password_admin: secret

basic:
  max_msg_send: 128
  max_bandwidth: 100Mbps

rcon:
  password: changeme
  port: 2301
```

### Mission source (optional)

Configure `mission_source` to enable `zeus missions pull`.

**Local path — copy mode** (copies `.pbo` files into the profile):

```yaml
mission_source:
  driver: path
  path: /home/user/missions
  mode: copy        # default; omit to use copy
```

**Local path — symlink mode** (replaces mpmissions dir with a symlink):

```yaml
mission_source:
  driver: path
  path: /home/user/missions
  mode: symlink
```

**S3 bucket**:

```yaml
mission_source:
  driver: s3
  bucket: my-missions-bucket
  prefix: arma3/          # optional — scopes sync to this key prefix
  region: us-east-1
  endpoint: https://...   # optional, for S3-compatible stores (MinIO, etc.)
  access_key_id: AKID
  secret_access_key: secret
```

**Prefix scoping:** only `.pbo` objects stored *directly* at the prefix level are downloaded — objects in subdirectories are ignored.

```
bucket: missions
  mission1.VR.pbo          ← downloaded (prefix: "" or omitted)
  mission2.Altis.pbo       ← downloaded (prefix: "" or omitted)
  submissions/
    mission3.Stratis.pbo   ← ignored (prefix: ""); downloaded only if prefix: submissions
```

Set `prefix: submissions` to scope the pull to the `submissions/` folder exclusively.

## Hooks

Profiles can declare shell commands under a `hooks:` key. Commands run sequentially in the order listed; the first non-zero exit aborts the remaining commands for that event.

### Lifecycle points

| Hook | Fires when |
|------|-----------|
| `post_profile_add` | After `zeus add` / `zeus profile add` completes |
| `pre_profile_start` | Before the server process is prepared (`zeus start`) |
| `pre_profile_run` | Immediately before `arma3server_x64` is spawned |
| `post_profile_run` | After the server process is confirmed running (PID file written and stable) |
| `post_profile_start` | After `zeus start` completes (server confirmed running) |
| `pre_profile_stop` | Before the server process is killed (`zeus stop`) |
| `post_profile_stop` | After the server process has terminated |
| `pre_pull_missions` | Before `zeus missions pull` fetches files |
| `post_pull_missions` | After `zeus missions pull` completes |

### Environment variables

Every hook command receives these variables in addition to the current environment:

| Variable | Value |
|----------|-------|
| `ZEUS_PROFILE` | Profile name |
| `ZEUS_PROFILE_DIR` | Absolute path to `<install_dir>/.zeus/<name>` |
| `ZEUS_PROFILE_INSTALL_DIR` | Absolute path to `<install_dir>` |

The working directory is set to `<install_dir>/.zeus/<name>` (the profile directory).

### Shell used

- **Linux / macOS:** `sh -c "<command>"`
- **Windows:** written to a temporary `.bat` file in the profile directory and executed with `cmd /C`. Placing it in the profile directory (rather than the shared system temp) avoids quoting issues and multi-user temp-dir races.

### Examples

```yaml
hooks:
  # Download a CBA settings PBO after adding the profile
  post_profile_add:
    - 'mkdir -p servermods/@cba_settings_userconfig/addons'
    - >-
        curl -X POST "https://example.com/api/cba-settings"
        -H "Content-Type: text/plain; charset=utf-8"
        --data-binary "@cba_settings.sqf"
        -o "servermods/@cba_settings_userconfig/addons/cba_settings.pbo"

  # Notify a webhook when the server starts
  post_profile_start:
    - 'curl -s -X POST "https://hooks.example.com/notify" -d "{\"event\":\"start\",\"profile\":\"$ZEUS_PROFILE\"}"'

  # Pull fresh missions before each start
  pre_profile_start:
    - 'zeus missions pull %ZEUS_PROFILE%'   # Windows
    # - 'zeus missions pull $ZEUS_PROFILE'  # Linux

  # Sync modpack from a remote server before starting
  pre_profile_run:
    - 'rsync -av user@mod-server:/mods/ "$ZEUS_PROFILE_INSTALL_DIR/@mymod/"'

  # Archive logs after stopping
  post_profile_stop:
    - 'tar -czf "$ZEUS_PROFILE_DIR/logs/archive-$(date +%Y%m%d-%H%M%S).tar.gz" "$ZEUS_PROFILE_DIR/logs/"'
```

**Windows example** (cmd.exe syntax inside the batch file):

```yaml
hooks:
  post_profile_add:
    - 'mkdir servermods\@cba_settings_userconfig\addons || cd .'
    - >-
        curl -X POST "https://example.com/api/cba-settings"
        -H "Content-Type: text/plain; charset=utf-8"
        --data-binary "@cba_settings.sqf"
        -o "servermods\@cba_settings_userconfig\addons\cba_settings.pbo"

  post_profile_start:
    - 'echo Profile %ZEUS_PROFILE% started, dir is %ZEUS_PROFILE_DIR%'
```

> **Tip:** To debug the injected variables, add a hook like:
> ```yaml
> post_profile_add:
>   - 'echo ZEUS_PROFILE=%ZEUS_PROFILE%'          # Windows
>   - 'echo ZEUS_PROFILE_DIR=%ZEUS_PROFILE_DIR%'
>   - 'echo ZEUS_PROFILE_INSTALL_DIR=%ZEUS_PROFILE_INSTALL_DIR%'
>   - 'cd'
> ```

## File locations

| Path | Contents |
|------|----------|
| `~/.zeus/profiles/<name>.json` | Saved profile |
| `~/.zeus/running/<name>.pid` | PID file while server is running |
| `<install_dir>/.zeus/<name>/configs/` | Generated `server.cfg` and `basic.cfg` |
| `<install_dir>/.zeus/<name>/mpmissions/` | Mission `.pbo` files |
| `<install_dir>/.zeus/<name>/keys/` | BI signing keys (`.bikey` files) copied from mods |
| `<install_dir>/.zeus/<name>/optionalkeys/` | User-managed keys for optional mods (never modified by ZEUS) |

## License

MIT — see [LICENSE](LICENSE).
