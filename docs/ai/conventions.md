# Conventions

## Purpose

Naming, structural, and behavioral conventions used throughout the ZEUS codebase.
For use by LLMs, AI agents, and developers unfamiliar with the project.

## Profile Key Naming

Profile files (YAML and TOML) use `snake_case` for all keys:

```yaml
install_dir: /opt/arma3
params:
  limit_fps: 100
  max_mem: 4096
  server_mod: [...]
config:
  max_players: 32
  password_admin: secret
  forced_difficulty: veteran
  verify_signatures: 2
basic:
  max_msg_send: 128
  max_size_guaranteed: 512
```

The same key names apply in TOML:

```toml
install_dir = "/opt/arma3"
[params]
limit_fps = 100
[config]
max_players = 32
password_admin = "secret"
[basic]
max_msg_send = 128
```

Arma 3 native config files use `camelCase` or `PascalCase` (`MaxMsgSend`, `maxPlayers`).
The mapping is handled by the struct field names and YAML/TOML tags.

## Struct Tags

Four tag kinds are used:

| Tag | Purpose | Example |
|-----|---------|---------|
| `yaml:"..."` | YAML key — `snake_case` | `yaml:"max_players"` |
| `toml:"..."` | TOML key — `snake_case` | `toml:"max_players"` |
| `json:"..."` | JSON key — `camelCase` | `json:"maxPlayers"` |
| `arg:"..."` | Arma 3 CLI argument name | `arg:"port"` |

`yaml:` and `toml:` tags always use `snake_case` and must match each other.
`json:` tags use `camelCase` (matching Arma 3 native field names where applicable).
When adding a new field, add all three tags.

`arg` tags drive `process/runner.go`'s reflection-based argument builder. Fields without
an `arg` tag are never included in the generated command.

## Pointer Fields in StartupParams

`StartupParams` uses `*T` pointer fields (not `T`) for optional values. A `nil` pointer
means "not set" — the argument is omitted from the generated command. This is the
canonical way to represent optional CLI flags.

```go
Port    *uint16 `yaml:"port,omitempty"    arg:"port"`
Server  *bool   `yaml:"server,omitempty"  arg:"server"`
```

Bool pointers: if the pointer is non-nil and `true`, the flag is added as `-argName`
(no value). If `false`, it is omitted entirely.

## Config Dumper: Write-Only-Non-Default Pattern

`arma/config_dumper.go` never writes a field whose value equals the Arma 3 default.
This keeps generated `.cfg` files minimal and readable.

Consequence: `Profile.Config` and `Profile.Basic` **must** be pre-populated with Arma 3
defaults before YAML unmarshaling, or unset fields will appear as Go zero values and
be written incorrectly to configs.

```go
// Correct — in ProfileLoader.fromYAML / fromTOML
p := Profile{
    Config: arma.NewDefaultServerConfig(),
    Basic:  arma.NewDefaultBasicServerConfig(),
}
yaml.Unmarshal(data, &p)   // or toml.Decode for TOML
```

## File Paths

| Path | Description |
|------|-------------|
| `~/.zeus/profiles/<name>.yaml` | Saved profile |
| `~/.zeus/running/<name>.pid` | PID file (absolute, OS home dir) |
| `<install_dir>/.zeus/<name>/configs/server.cfg` | Generated server config |
| `<install_dir>/.zeus/<name>/configs/basic.cfg` | Generated basic config |
| `<install_dir>/.zeus/<name>/mpmissions/` | Mission files |
| `<install_dir>/.zeus/<name>/keys/` | BI signing keys (`.bikey` files) |

The configs/mpmissions/keys paths are **relative to `install_dir`** when passed to Arma 3,
because the server process runs with `cmd.Dir = install_dir`. The PID file path is
**absolute** because it belongs to ZEUS's process tracking, not Arma 3.

## HomeDir Field Pattern

`Manager`, `Runner`, and `ProfileRepository` all have a `HomeDir string` field:

```go
type Manager struct { HomeDir string }
```

When `HomeDir` is empty, the real `os.UserHomeDir()` is used. In tests, `t.TempDir()`
is assigned to keep tests hermetic. Never mock `os.UserHomeDir` — use this field instead.

## Error Handling in cmd/

All command-level errors use the `fatal`/`fatalf` helpers:

```go
fatal(err)                          // prints err to stderr, exits 1
fatalf("profile %q not found", name) // formats + exits 1
```

Warnings that do not abort execution use `fmt.Fprintf(os.Stderr, "warning: ...")`.

## Mod Paths

Mod paths in the `params.mod` YAML list should be absolute paths to mod directories.
Relative paths are resolved relative to `install_dir` by Arma 3 (not by ZEUS).

Keys (`.bikey` files) are identified by lowercase `.bikey` extension check.
Submods are directories inside a mod folder prefixed with `@`.

## Related

- [Architecture Rules](architecture_rules.md)
- [Common Gotchas](common_gotchas.md)
