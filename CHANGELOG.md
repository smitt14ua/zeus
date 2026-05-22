# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.6.3] - 2026-05-22

### Added

- **`onPlayerJoinAttempt` in server config** — new scripting hook called repeatedly while a player is trying to join the server. The SQF expression must return `"ACCEPT"` (let the player in), `"DELAY"` (keep on loading screen), or `"REFUSE"` / `"REFUSE_<message>"` (kick with optional reason). Profile JSON schema and `docs/formats/server_cfg.md` updated.
- **`sendChatMessage` in server config** — new scripting hook that sends a "System" chat message to the specified user. Profile JSON schema and `docs/formats/server_cfg.md` updated.

## [0.6.2] - 2026-05-21

### Added

- **`allowedVoteCmds` and `allowedVotedAdminCmds` in server config** — the two voting control arrays are now supported in profile `config:` sections. `allowedVoteCmds` lets you restrict which vote commands players may initiate and set per-command thresholds; `allowedVotedAdminCmds` controls which admin commands a voted-in admin may use (empty array disables all, omitting the field grants unrestricted access). Both fields are serialised correctly to `server.cfg` with positional tuple syntax. Profile JSON schema and `docs/formats/server_cfg.md` updated with full reference.

## [0.6.1] - 2026-05-18

### Fixed

- **`profile.new` now works in agent mode** — the command was defined in the protocol package and listed in the `manage` scope but had no registered handler in the agent executor. Calling `profile.new` via the panel returned `"unknown command"`. It is now registered and functional.
- **`manage` scope correctly covers `profile.new`** — the scope enforcement map was missing `profile.new`, so agents running with `--allow manage` would reject the command at the scope-check stage even after the handler was added.

### Docs

- `README.md`: added `profile ls --json` to the flags reference, added a **Profile name rules** section documenting the allowlist regex and reserved-name restrictions, added port-conflict detection note to the `start` description, clarified Windows hook temp-file placement.
- `docs/agent-protocol.md`: added `profile.new` to the Command Reference and scope table; expanded `profile.start`, `profile.stop`, and `missions.pull` stream output examples to show hook progress messages; clarified `profile.list` returns compact single-line JSON; expanded Table of Contents with per-command anchors.

## [0.6.0] - 2026-05-18

### Added

- **Hook progress messages for all hook-bearing commands** — `profile add`, `profile stop`, and `missions pull` now print `Running pre-X hooks…` / `Running post-X hooks…` before each hook group runs, matching the existing behavior in `profile start`.
- **Profile name validation** — `profile add` and `profile new` now reject names that could cause path traversal (`../../evil`), hidden-file names (`.hidden`), null bytes, names longer than 64 characters, and Windows reserved device names (`CON`, `NUL`, `COM1`–`COM9`, `LPT1`–`LPT9`). Validation runs at both CLI boundaries and storage entry points.

### Changed

- **`zeus profile ls --json` now emits compact JSON** — output is a single line instead of indented multi-line, making it suitable for piping to `jq` and other tools without a `-c` flag.
- **File permissions hardened** — profile JSON files and generated configs (`server.cfg`, `basic.cfg`, `BEServer.cfg`) are written `0600`; profile and PID directories are created `0700`. Previously `0644`/`0755`.
- **`profile rm` removes all format variants** — if a profile exists as both `.json` and `.yaml` (e.g. after a manual copy), `profile rm` now removes all matching files instead of stopping after the first hit.
- **Mission symlink replacement is atomic on POSIX** — the mpmissions directory symlink is now replaced via a sibling-path create + `rename(2)`, eliminating the window where the directory is temporarily absent. Windows uses a safe remove-then-create fallback (atomic rename is not supported for directory reparse points on Windows).
- **Windows hook `.bat` temp files placed in profile directory** — previously written to the shared system temp (`%TEMP%`), which is writable by all local users. Now written to the profile directory (owned by the operator user).

### Fixed

- **`port: 0` treated as unset** — a profile with `params.port: 0` now falls back to the Arma 3 default (2302) in port-conflict detection (`profile start`) and RCon config generation (`profile add`).
- **RCon port underflow** — `DefaultRCon` no longer wraps around to 65535 when called with `gamePort = 0`.
- **PID ≤ 0 rejected in all process operations** — `Manager.List`, `Kill`, and `WaitReady` now return an error if a `.pid` file contains a non-positive value. On Unix, `kill(0, SIGKILL)` signals the entire process group (including the zeus CLI itself); this fix prevents that.
- **`WaitGone` no longer silently constructs an invalid path** — if the running directory cannot be resolved, it returns `false` immediately instead of joining an empty string with the PID filename.
- **Profile load errors include the filename** — `ProfileRepository.List` now wraps loader errors with the profile filename, making malformed config files easier to identify.
- **TOCTOU race eliminated in profile lookup** — `ProfileRepository.Get` previously did a `stat` then `open`; it now calls `loader.FromFile` directly and handles `os.ErrNotExist` from the single open.
- **Embedded `"` in server.cfg string arrays are escaped** — `motd`, `admins`, and other string-array fields now produce `"say \"hello\""` instead of `"say "hello""` in the generated config.

### Docs

- `docs/ai/common_gotchas.md`: three new entries — profile name validation (#19), Windows symlink non-atomicity (#20), and bat temp file placement (updated #17).

## [0.5.1] - 2026-05-18

### Changed

- **`zeus profile start` now prints progress steps** — the start sequence emits a line at each stage: `Starting profile "…"`, `Launching server process…`, `Process started (launcher PID …), waiting for initialization (timeout …)…`, and `Profile "…" started (PID …)`. When hooks are configured, a `Running pre-start hooks…` / `Running post-start hooks…` line is printed before the relevant hook group runs.

## [0.5.0] - 2026-05-18

### Added

- **`zeus profile start` / `zeus start` rejects port conflicts** — before launching, ZEUS now checks every other running profile's effective game port (from `params.port`, defaulting to `2302`). If any running profile is already bound to the same port, the command exits immediately with a clear error (`profile "X" is already running on port 2302 (PID Y)`) before hooks are run or the process is spawned.
- **`arma.DefaultPort` constant (`2302`)** — the Arma 3 default game port is now exported from the `arma` package.

## [0.4.1] - 2026-05-18

### Fixed

- **Agent no longer restarts after `update` when already on the latest version** — `execUpdate` now returns a boolean indicating whether a new binary was actually written. The restart only fires when `true`; an "already up to date" response sends `success: true` and leaves the agent running.

### Docs

- `docs/agent-protocol.md`: documented `update` command behaviour in agent mode (behaviour table, expanded prose, two new sequence diagrams for update+restart and update+no-restart); added connection lifecycle note that the heartbeat goroutine is per-connection scoped; added two rows to the error handling table.

## [0.4.0] - 2026-05-18

### Added

- **Agent auto-restarts after a successful `update` command** — when the panel sends `update` and the binary is replaced, the agent now spawns a new process from the updated executable with the original arguments and exits cleanly. The panel sees the agent reconnect at the new version without any manual intervention. The update output stream ends with `"Restarting agent..."` before the result frame is sent.

### Fixed

- **Agent now reconnects after the server temporarily goes away** — the heartbeat goroutine was started with the root context, so when the WebSocket read loop returned on a dropped connection the deferred `<-hbDone` blocked forever and the reconnect loop in `Run` never fired. Fixed by giving the heartbeat goroutine a per-connection context; `cancelConn()` is deferred before `<-hbDone` so the goroutine exits as soon as the connection is lost and `connect` returns promptly for the next retry.
- **`zeus profile add` finds `.bikey` keys in any immediate mod subdirectory** — keys were previously only collected from a directory named exactly `keys` (case-insensitive). The scanner now checks every first-level subdirectory of the mod folder (`bikeys/`, `key/`, etc.), matching how some mods ship their keys. Subdirectories two levels deep (e.g. `@mod/optionals/@sub/keys/`) are still attributed to their respective submod only.

## [0.3.2] - 2026-05-16

### Fixed

- **Symlinked mod paths now work correctly during `zeus profile add`** — `e.IsDir()` on a directory entry does not follow symlinks on any platform, so symlinked `addons/`, `keys/`, and `optionals/` subdirectories inside a mod folder were silently skipped, and symlinked submod directories inside `optionals/` were never loaded. Fixed by using `os.Stat` (which follows symlinks) for all directory checks in `LoadMod` and `loadSubmods`. The mod path itself is also resolved via `filepath.EvalSymlinks` at load time, handling Windows junctions and Linux/macOS symlinks uniformly.

## [0.3.1] - 2026-05-16

### Changed

- **`zeus start` / `zeus profile start` now waits for the server to be confirmed running** instead of returning immediately after spawning the process.
  - Polls for the Arma 3 PID file (written by the server via `-pid=`); uses the direct OS launch PID to fail fast if the process exits before the file appears.
  - Holds a 3-second stability window after the PID file appears to verify the server hasn't crashed on startup.
  - Returns an error if the PID file does not appear within the timeout or the process exits during the stability window.
  - New `--start-timeout` flag controls the wait limit (default `1m0s`).
  - Post-start hooks (`post_profile_run`, `post_profile_start`) now fire only after startup is confirmed.
- **Agent default heartbeat interval reduced from 30 s to 1 s** (`--heartbeat` flag).
- **Agent sends a heartbeat immediately after `hello`** (previously the first heartbeat was delayed by one full interval).
- **Agent sends a heartbeat after every command result** so the panel receives updated profile/PID state immediately without waiting for the next periodic tick.

### Docs

- `docs/agent-protocol.md`: updated heartbeat timing (immediate + post-command), corrected default interval, added `profile.start` blocking behaviour description.
- `README.md`: added `zeus agent` command section, `--start-timeout` flag, updated hook lifecycle descriptions.

## [0.3.0] - 2026-05-16

### Added

- **`zeus agent`** — new command that connects outbound to a web panel server over WebSocket (WSS), enabling a browser-based panel to list profiles, start/stop servers, stream output, and pull missions without opening inbound ports on the Arma 3 machine.
  - Reconnects automatically after disconnect (configurable delay, default 5 s).
  - Sends a `hello` on connect and periodic `heartbeat` carrying profile status and PID.
  - Executes commands concurrently; correlates `stream` / `result` messages back to the originating command via a UUID.
  - Supported remote commands: `profile.list`, `profile.info`, `profile.add`, `profile.start`, `profile.stop`, `profile.rm`, `missions.pull`, `update`.
- **Scope-based command allowlist (`--allow`)** — restrict which commands the agent accepts. Scopes: `view` (list/info), `control` (start/stop), `manage` (add/rm/missions), `update`, `all` (default). Disallowed commands are rejected immediately with a descriptive error; the agent reports its allowed scopes in every `hello` and `heartbeat` so the panel UI can hide unavailable actions.
- **JSON Schema for profile** (`docs/schema/profile.json`) — draft-07 schema covering the full profile data model (`StartupParams`, `ServerConfig`, `BasicServerConfig`, `RCon`, `MissionSource`, `ProfileHooks` and all nested enums). Panel editors can reference it for validation and autocompletion.

### Docs

- `docs/agent-protocol.md` — full agent protocol specification: transport, authentication, message envelope, connection lifecycle, command reference, scope system, sequence diagrams (Mermaid), error handling, implementation checklist, and code examples in Node.js, Python, and Go.

## [0.2.0] - 2026-05-16

### Added

- **`optionalkeys` directory** — created alongside `keys` when a profile is added or updated; files placed there by the user are never touched by ZEUS. The folder is automatically appended to `-keysFolder` at server launch so Arma 3 recognises keys for optional mods.
- **Lifecycle hooks** — profiles can now declare shell commands under a `hooks:` key that execute at nine lifecycle points:
  - `post_profile_add` — after `zeus profile add` completes
  - `pre_profile_start` / `post_profile_start` — around `zeus start`
  - `pre_profile_run` / `post_profile_run` — immediately before/after the server process is spawned
  - `pre_profile_stop` / `post_profile_stop` — around `zeus stop`
  - `pre_pull_missions` / `post_pull_missions` — around `zeus missions pull`

  Each command receives `ZEUS_PROFILE`, `ZEUS_PROFILE_DIR`, and `ZEUS_PROFILE_INSTALL_DIR` as environment variables and runs with the profile directory as its working directory. Execution stops on the first non-zero exit.

### Fixed

- Hook commands on Windows are now written to a temporary `.bat` file before execution. The previous `cmd /C "..."` approach caused Go's argument escaping to mangle internal double quotes, breaking commands such as `curl -H "Content-Type: ..."`.

## [0.1.7] - 2026-05-13

### CI / Security

- Branch protection on `main`: force pushes and deletion blocked; `test` and `analyze` (CodeQL) required to pass on PRs
- Dependabot enabled for Go modules and GitHub Actions (weekly, Monday)
- CodeQL static analysis on every push/PR to `main` and weekly schedule
- Secret scanning and push protection enabled
- All GitHub Actions updated to Node 24 runtime: `actions/checkout@v6`, `actions/setup-go@v6`, `github/codeql-action@v4`, `softprops/action-gh-release@v3`
- Added issue templates (bug report, feature request) and PR template
- Added `SECURITY.md` with private vulnerability reporting instructions
- GitHub Discussions enabled

## [0.1.6] - 2026-05-13

### Added
- `install.sh` — POSIX installer for Linux and macOS (x86_64 and arm64); installs to `/usr/local/bin`
- `install.ps1` — PowerShell installer for Windows; installs to `%LOCALAPPDATA%\Programs\Zeus` and adds it to user PATH
- Release workflow now builds and publishes binaries for all five platforms: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`
- Raw installer-friendly binary assets (`zeus-linux-amd64`, `zeus-darwin-arm64`, etc.) published alongside the existing versioned archives

## [0.1.5] - 2026-05-13

### Added
- `zeus update` command — checks GitHub releases and self-updates the binary in place; dev builds are rejected with a helpful install message

## [0.1.4] - 2026-05-13

### Fixed
- Relative mod paths in `params.mod` and `params.server_mod` now work — ZEUS resolves them to `<install_dir>/<path>` before scanning for keys (`profile add`) and before building the `-mod=` launch argument (`profile start`)

### Docs
- `docs/ai/conventions.md`: updated Mod Paths section — paths can now be absolute or relative to `install_dir`
- `README.md`: added mod path examples showing both absolute and relative forms

## [0.1.3] - 2026-05-13

### Fixed
- Removed automatic `-bePath` injection from server launch — Arma 3 sets the BattlEye path correctly on its own; the forced override was causing startup problems
- `be_path` can still be set manually in the profile's `params:` section if a non-default path is needed

### Docs
- Added `CLAUDE.md` (Claude Code session context), `AGENTS.md` (other AI agents), and `CONTRIBUTING.md` (build, test, PR process)
- Updated S3 prefix scoping documentation in `README.md` and `docs/ai/common_gotchas.md`
- Added `internal/missions` to architecture rules package graph

## [0.1.2] - 2026-05-13

### Fixed
- S3 driver now ignores `.pbo` objects stored in subdirectories when no prefix is set — only flat objects at the configured prefix level are downloaded
- S3 driver uses `ListObjectsV2` prefix scoping correctly: setting `prefix: submissions` downloads only files directly under `submissions/`, root-level files are excluded

### Added
- Integration tests: `TestS3Driver_SubdirIgnoredAtRootPrefix` and `TestS3Driver_PrefixScopesDownload` covering the subdirectory filtering and prefix scoping behaviour

## [0.1.1] - 2026-05-13

### Added
- Tests for `new` / `profile new` command verifying TOML, JSON, and YAML output each contain the `[rcon]` section with password and port

### Fixed
- Loader test fixtures for TOML and JSON now include an `rcon` section, covering the round-trip parse path that was previously untested

### Docs
- Note that the `-mpmissions` Arma 3 startup parameter bug is fixed as of server revision 153745

## [0.1.0] - 2026-05-13

### Added
- Initial release: Arma 3 server manager CLI (`zeus`)
- `profile new` / `new` — generate profile templates in YAML, TOML, or JSON
- `profile add` — add a profile from file or stdin
- `profile run` / `start` — launch an Arma 3 dedicated server
- `stop` — stop a running server by profile name
- `missions` / `add` — mission source management
- BattlEye config generation (`BEServer.cfg`, `BEServer_x64.cfg`)
- RCon defaults: deterministic MD5 password per profile name, port = gamePort − 1
- TOML custom scalar types (`DataSize`, `DataTransferRate`, `Time`) accepting both integer and string forms

[Unreleased]: https://github.com/smitt14ua/zeus/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/smitt14ua/zeus/compare/v0.4.1...v0.5.0
[0.4.1]: https://github.com/smitt14ua/zeus/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/smitt14ua/zeus/compare/v0.3.2...v0.4.0
[0.3.2]: https://github.com/smitt14ua/zeus/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/smitt14ua/zeus/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/smitt14ua/zeus/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/smitt14ua/zeus/compare/v0.1.7...v0.2.0
[0.1.7]: https://github.com/smitt14ua/zeus/compare/v0.1.6...v0.1.7
[0.1.6]: https://github.com/smitt14ua/zeus/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/smitt14ua/zeus/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/smitt14ua/zeus/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/smitt14ua/zeus/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/smitt14ua/zeus/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/smitt14ua/zeus/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/smitt14ua/zeus/releases/tag/v0.1.0
