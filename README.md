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

## Flags

### `new` / `profile new`

| Flag | Description |
|------|-------------|
| `--yaml` | Output as YAML |
| `--toml` | Output as TOML |
| `--json` | Output as JSON (default) |

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

| Flag | Short | Description |
|------|-------|-------------|
| `--dry-run` | `-n` | Print the generated launch command without executing |

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

## File locations

| Path | Contents |
|------|----------|
| `~/.zeus/profiles/<name>.json` | Saved profile |
| `~/.zeus/running/<name>.pid` | PID file while server is running |
| `<install_dir>/.zeus/<name>/configs/` | Generated `server.cfg` and `basic.cfg` |
| `<install_dir>/.zeus/<name>/mpmissions/` | Mission `.pbo` files |
| `<install_dir>/.zeus/<name>/keys/` | BI signing keys (`.bikey` files) |

## License

MIT — see [LICENSE](LICENSE).
