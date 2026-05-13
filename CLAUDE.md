# ZEUS — Claude Code Context

CLI tool for managing Arma 3 dedicated server profiles on Windows and Linux.
A **profile** bundles server config, startup params, mod list, and RCon settings
into one YAML/TOML/JSON file. ZEUS generates `.cfg` files, launches the server,
and tracks running instances.

## Essential Reading

Read these before touching code:

1. [Architecture Rules](docs/ai/architecture_rules.md) — package boundaries, lifecycle, invariants
2. [Conventions](docs/ai/conventions.md) — naming, struct tags, file paths
3. [Common Gotchas](docs/ai/common_gotchas.md) — traps that cause subtle bugs

## Package Map

| Package | Responsibility |
|---------|---------------|
| `internal/arma` | Arma 3 config types, dumper, startup params, mod scanning. **No internal deps.** |
| `internal/profile` | Profile struct, YAML/TOML/JSON load, config file write |
| `internal/storage` | Profile persistence — `~/.zeus/profiles/*.json` |
| `internal/process` | PID tracking, process existence, server launch |
| `internal/missions` | `.pbo` sync — path (copy/symlink) and S3 drivers |
| `cmd/` | CLI surface only — no business logic here |

## Key Invariants

- **`ProfileWriter.Write` runs at `add` time**, not `start` time. `.cfg` files are
  generated when a profile is registered, then read by Arma 3 at launch.
- **`ProfileRepository.Get` does not pre-populate defaults.** Profiles on disk already
  carry all values. Do not add default injection there.
- **Profiles always stored as `.json`** regardless of input format (YAML/TOML inputs
  are converted on save).
- **Relative paths in Runner are intentional.** `cmd.Dir = install_dir`, so Arma 3
  resolves `.zeus/<name>/configs/server.cfg` relative to its installation directory.
- **S3 driver syncs flat objects only.** After stripping the prefix, any key that
  still contains `/` is a subdirectory object and is skipped. Use `prefix: folder/`
  to scope to a subfolder; objects outside the prefix are excluded by the S3 listing.

## Struct Tags

Always add all three when adding a field:

```go
Field Type `json:"camelCase,omitempty" yaml:"snake_case,omitempty" toml:"snake_case,omitempty"`
```

`yaml:` and `toml:` must match. `json:` uses camelCase.

## Testing

```bash
go test ./...                          # unit tests (no deps)
go test -tags=integration ./...        # + S3 integration (MinIO at localhost:9000)
```

Integration tests require MinIO: bucket `mpmissions`, credentials `minioadmin`/`minioadmin`.

## Release Artifacts

Each tag produces two asset types per platform:

| Type | Example | Used by |
|------|---------|---------|
| Versioned archive | `zeus_0.1.6_linux_amd64.tar.gz` | `zeus update` (go-selfupdate) |
| Raw binary | `zeus-linux-amd64` | `install.sh` / `install.ps1` |

Platforms: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`.
All built from `ubuntu-latest` with `CGO_ENABLED=0`.

## Docs Index

- [`docs/ai/`](docs/ai/) — architecture, conventions, gotchas (start here)
- [`docs/formats/`](docs/formats/) — server.cfg / basic.cfg field reference
- [`docs/protocols/startup_params.md`](docs/protocols/startup_params.md) — Arma 3 startup params
- [`docs/glossary.md`](docs/glossary.md) — domain terms and abbreviations
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — build, test, PR process
- [`CHANGELOG.md`](CHANGELOG.md) — release history
- [`.github/ISSUE_TEMPLATE/`](.github/ISSUE_TEMPLATE/) — bug + feature request forms
- [`.github/PULL_REQUEST_TEMPLATE.md`](.github/PULL_REQUEST_TEMPLATE.md) — PR checklist
