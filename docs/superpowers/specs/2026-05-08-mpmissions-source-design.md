# mpmissions Source — Design Spec

**Date:** 2026-05-08
**Status:** Approved

## Summary

Add a `mission_source` field to the profile that describes where `.pbo` mission files come from, and a `zeus missions pull <profile>` command that syncs those files into the profile's mpmissions directory.

Scope: v1 ships `path` driver only. `ftp`, `s3`, and `git` drivers are out of scope and reserved for future work.

---

## 1. Profile Schema

New optional top-level field on `Profile`:

```yaml
mission_source:
  driver: path
  path: /opt/missions
  mode: copy      # or: symlink — default: copy
```

```toml
[mission_source]
driver = "path"
path = "/opt/missions"
mode = "copy"
```

`MissionSource` is a pointer (`*arma.MissionSource`); absent means no source configured.

### Struct

```go
// internal/arma/mission_source.go
type MissionSource struct {
    Driver string `json:"driver"           yaml:"driver"           toml:"driver"`
    Path   string `json:"path,omitempty"   yaml:"path,omitempty"   toml:"path,omitempty"`
    Mode   string `json:"mode,omitempty"   yaml:"mode,omitempty"   toml:"mode,omitempty"`
}
```

`Mode` accepted values: `"copy"` (default when empty), `"symlink"`.

### Profile struct change

```go
// internal/profile/profile.go
MissionSource *arma.MissionSource `json:"missionSource,omitempty" yaml:"mission_source,omitempty" toml:"mission_source,omitempty"`
```

---

## 2. `internal/missions` Package

New package. No imports from `cmd/`. Depends only on `internal/arma`.

```
internal/missions/
  puller.go       — Puller entry point, driver dispatch
  path_driver.go  — PathDriver: copy and symlink implementations
  sync.go         — shared .pbo diff/sync helpers
```

### Puller

```go
type Puller struct{}

func (p Puller) Pull(source arma.MissionSource, targetDir string, dryRun bool) (Result, error)
```

Dispatches by `source.Driver`. Returns `Result` (counts for summary output).

### Result

```go
type Result struct {
    Added    []string
    Updated  []string
    Removed  []string
    Skipped  []string
    Symlinked bool   // true when symlink mode completed
}
```

### PathDriver — copy mode

Target dir is a real directory managed by ZEUS.

Sync algorithm:
1. Scan `source.Path` → set of `.pbo` filenames
2. Scan `targetDir` → set of `.pbo` filenames (non-`.pbo` files left untouched)
3. Delete `.pbo` files in target not present in source → `Removed`
4. For each `.pbo` in source:
   - Not in target → copy → `Added`
   - In target, same size + mtime → `Skipped`
   - In target, different → overwrite copy → `Updated`

In `dryRun` mode: compute diff, return `Result`, make no filesystem changes.

### PathDriver — symlink mode

Target mpmissions path is replaced with a single directory symlink to `source.Path`.

Algorithm:
1. Stat `targetDir`:
   - Already a symlink pointing to `source.Path` → no-op, `Result{Symlinked: false}`
   - Already a symlink pointing elsewhere → remove + create new symlink
   - Is a real directory → **error**: "target is a real directory; remove it manually before switching to symlink mode"
   - Does not exist → create symlink
2. In `dryRun` mode: report what would happen, make no changes.

---

## 3. Command

```
zeus missions pull <profile> [--dry-run]
```

New file: `cmd/missions.go`

Registers `missionsCmd` (command group) and `missionsPullCmd` under it.

### Behavior

1. Load profile: `ProfileRepository.Get(name)`
2. Check `profile.MissionSource != nil` → fatal `"profile %q has no mission_source configured"`
3. Resolve target: `filepath.Join(profile.InstallDir, ".zeus", profile.Name, "mpmissions")`
4. Ensure target dir exists (create if absent) for copy mode; skip for symlink mode
5. Call `missions.Puller{}.Pull(*profile.MissionSource, targetDir, dryRun)`
6. Print summary (see below)

### Output — copy mode

```
Pulling missions for "my-server" (path, copy)
  + op_cobra.pbo
  ~ op_patrol_v2.pbo
  - old_mission.pbo
  = op_escort.pbo
Done. 1 added, 1 updated, 1 removed, 1 unchanged.
```

Dry-run prepends `[dry-run] ` to header; no `Done.` line, prints `Dry run complete. No changes made.`

### Output — symlink mode

```
Pulling missions for "my-server" (path, symlink)
  → /opt/missions
Done.
```

Or if already correct:
```
Pulling missions for "my-server" (path, symlink)
  = /opt/missions (symlink unchanged)
Done.
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | false | Show what would change without making changes |

---

## 4. Error Cases

| Situation | Behavior |
|-----------|----------|
| `mission_source` absent | fatal: "no mission_source configured" |
| Unknown driver | fatal: "unsupported driver %q" |
| `source.Path` does not exist | fatal: "source path %q does not exist" |
| Target is real dir in symlink mode | fatal: "target is a real directory; remove it manually" |
| File copy fails | fatal with os error |
| Mode field empty | treated as `"copy"` |

---

## 5. Out of Scope (v1)

- `ftp`, `s3`, `git` drivers
- Automatic pull on `zeus start`
- `zeus missions list` (show current mpmissions contents)
- Signature file (`.pbo.bisign`) syncing — source dir is authoritative for `.pbo` only

---

## 6. Files Changed / Created

| File | Change |
|------|--------|
| `internal/arma/mission_source.go` | New — `MissionSource` struct |
| `internal/profile/profile.go` | Add `MissionSource` field |
| `internal/missions/puller.go` | New — `Puller`, `Result` |
| `internal/missions/path_driver.go` | New — `PathDriver` (copy + symlink) |
| `internal/missions/sync.go` | New — `.pbo` diff helpers |
| `internal/missions/puller_test.go` | New — tests |
| `cmd/missions.go` | New — `missionsCmd`, `missionsPullCmd` |
| `cmd/root.go` | Register `missionsCmd` |
| `conf/example.yaml` | Add `mission_source` example block |
| `conf/example.toml` | Add `[mission_source]` example block |
