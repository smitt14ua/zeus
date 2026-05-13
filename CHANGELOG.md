# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/smitt14ua/zeus/compare/v0.1.6...HEAD
[0.1.6]: https://github.com/smitt14ua/zeus/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/smitt14ua/zeus/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/smitt14ua/zeus/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/smitt14ua/zeus/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/smitt14ua/zeus/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/smitt14ua/zeus/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/smitt14ua/zeus/releases/tag/v0.1.0
