# ZEUS — AI Agent Context

> This file is read by AI coding agents (GitHub Copilot, Codex, Gemini, etc.).
> Claude Code users: see [CLAUDE.md](CLAUDE.md) instead.

## Project Summary

ZEUS is a Go CLI for managing Arma 3 dedicated server profiles on Windows and Linux.
A profile is a YAML/TOML/JSON file that bundles server config, startup parameters,
mod list, and RCon settings. ZEUS generates native Arma 3 `.cfg` files, launches
the server process, and tracks running instances via PID files.

## Required Reading Before Modifying Code

| Document | What It Covers |
|----------|---------------|
| [docs/ai/architecture_rules.md](docs/ai/architecture_rules.md) | Package boundaries, profile lifecycle, invariants |
| [docs/ai/conventions.md](docs/ai/conventions.md) | Struct tags, naming, file paths, pointer patterns |
| [docs/ai/common_gotchas.md](docs/ai/common_gotchas.md) | Edge cases that cause subtle bugs |

## Repository Layout

```
cmd/                    CLI commands (cobra) — no business logic
internal/
  arma/                 Arma 3 types, config dumper, startup params (no internal deps)
  profile/              Profile struct, loader, writer
  storage/              Profile persistence (~/.zeus/profiles/)
  process/              PID tracking, server launch
  missions/             .pbo sync — path driver + S3 driver
docs/
  ai/                   Architecture, conventions, gotchas for AI agents
  formats/              server.cfg / basic.cfg field reference
  protocols/            Arma 3 startup parameters
  glossary.md           Domain terms
CLAUDE.md               Extended context for Claude Code
CONTRIBUTING.md         Build, test, PR process
CHANGELOG.md            Release history
```

## Critical Rules

1. **Never add `omitempty` to `Profile.RCon`** — the rcon section must always appear
   in serialised output.
2. **Always use `arma.NewDefaultServerConfig()` and `arma.NewDefaultBasicServerConfig()`**
   when creating a Profile before unmarshaling user input.
3. **`ProfileRepository.Get` must not pre-populate defaults.** Profiles on disk carry
   all values; injecting defaults here would override explicit user settings.
4. **`ProfileWriter.Write` runs at `add` time, not `start` time.** `.cfg` files are
   pre-generated and read by Arma 3 at launch.
5. **Relative paths in `Runner.prepareParams` are intentional.** `cmd.Dir = install_dir`
   — Arma 3 resolves them relative to its installation directory.
6. **S3 driver: prefix scopes to flat level only.** Objects one or more `/` below the
   prefix are ignored. Use `obj.key` from `s3Object` for `GetObject`, not a
   reconstructed path.

## Struct Tag Convention

```go
// Always add all three tags. yaml/toml use snake_case; json uses camelCase.
Field Type `json:"camelCase,omitempty" yaml:"snake_case,omitempty" toml:"snake_case,omitempty"`
```

## Running Tests

```bash
go test ./...                          # unit tests
go test -tags=integration ./...        # + S3 integration (MinIO localhost:9000)
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for MinIO setup.
