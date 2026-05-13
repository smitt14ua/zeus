## What

<!-- One paragraph: what changed and why. -->

## Type

- [ ] Bug fix
- [ ] New feature
- [ ] Refactor / cleanup
- [ ] Docs only
- [ ] Chore (deps, CI, tooling)

## Checklist

- [ ] `go test ./...` passes
- [ ] `go test -tags=integration ./...` passes (if touching `internal/missions`)
- [ ] New fields have all three struct tags (`yaml`, `toml`, `json`) — see [Conventions](docs/ai/conventions.md)
- [ ] `CHANGELOG.md` updated under `## [Unreleased]`
- [ ] PR title follows Conventional Commits (`feat:`, `fix:`, `docs:`, etc.)

## Testing

<!-- Describe how you verified this works. Commands run, edge cases checked. -->

## Notes for reviewer

<!-- Anything non-obvious: tradeoffs made, things explicitly not handled, follow-up work. -->
