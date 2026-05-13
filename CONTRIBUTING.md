# Contributing to ZEUS

## Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| Go | 1.22+ | see `go.mod` for exact minimum |
| MinIO | any | integration tests only — `localhost:9000` |

## Build

```bash
git clone https://github.com/smitt14ua/zeus
cd zeus
go build -o zeus .
```

## Tests

```bash
# Unit tests (no external dependencies)
go test ./...

# Unit + integration tests (requires MinIO)
go test -tags=integration ./...
```

Integration tests connect to MinIO at `http://localhost:9000` with credentials
`minioadmin` / `minioadmin` and bucket `mpmissions`. Start MinIO locally before
running them:

```bash
# Docker
docker run -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data --console-address ":9001"

# Then create the bucket
mc alias set local http://localhost:9000 minioadmin minioadmin
mc mb local/mpmissions
```

## Code Conventions

Read these before making changes — they cover naming, struct tags, file paths, and
invariants that are easy to violate without realising:

- [Conventions](docs/ai/conventions.md)
- [Architecture Rules](docs/ai/architecture_rules.md)
- [Common Gotchas](docs/ai/common_gotchas.md)

Key points:

- `yaml:` and `toml:` tags always use `snake_case`; `json:` tags use `camelCase`
- Always add all three tags (`yaml`, `toml`, `json`) when adding a new struct field
- Never add default pre-population to `ProfileRepository.Get` — see architecture rules
- `ProfileWriter.Write` is called at `add` time, not `start` time
- S3 driver: flat objects only at prefix level; subdirectory objects are ignored

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add X
fix: correct Y
test: add coverage for Z
docs: update prefix behaviour in README
refactor: extract helper from X
chore: bump dependency
```

Keep the subject line under 72 characters. Add a body when the motivation is
non-obvious.

## Cross-Platform Builds

All five targets are built in CI via `GOOS`/`GOARCH` with `CGO_ENABLED=0`.
To reproduce locally:

```bash
GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -o zeus-linux-arm64 .
GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -o zeus-darwin-arm64 .
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o zeus-windows-amd64.exe .
```

## Testing the Installers

**install.sh** (requires a release with raw binary assets):
```bash
# Dry-run against the real GitHub release
sh install.sh
```

**install.ps1**:
```powershell
# From repo root
.\install.ps1
```

## Pull Requests

A PR template will guide you. Short version:

1. Branch from `main`
2. Run `go test ./...` (unit) and `go test -tags=integration ./...` if touching `internal/missions`
3. Update `CHANGELOG.md` under `## [Unreleased]`
4. Open a PR — title follows the same Conventional Commits format

## Changelog

Add entries to `CHANGELOG.md` under `## [Unreleased]` using these sections:

- `### Added` — new features
- `### Fixed` — bug fixes
- `### Changed` — behaviour changes to existing features
- `### Docs` — documentation-only changes

Releases are cut by maintainers: bump version, move Unreleased entries, tag, push.
