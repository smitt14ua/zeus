# Architecture Rules

## Purpose

Package boundaries, invariants, and design decisions that must not be violated.
For LLMs and developers working on the ZEUS codebase.

## Package Dependency Graph

```
cmd/
 ├── internal/profile
 ├── internal/storage
 ├── internal/process
 ├── internal/missions
 └── internal/arma

internal/storage
 └── internal/profile

internal/process
 └── internal/profile

internal/profile
 └── internal/arma

internal/missions
 └── internal/arma

internal/arma
 └── (no internal deps)
```

**Rules:**
- `internal/arma` has no internal imports. It is a pure domain library.
- `internal/profile` imports `internal/arma` only.
- `internal/storage` and `internal/process` import `internal/profile` only.
- `internal/missions` imports `internal/arma` only.
- `cmd/` is the only package that imports across multiple internal packages.
- No circular imports. No internal package imports `cmd/`.

## Package Responsibilities

| Package | Owns |
|---------|------|
| `internal/arma` | Arma 3 config types, dumper, startup params, mod scanning, custom scalar types |
| `internal/profile` | Profile struct, YAML/TOML loading (with default pre-population), config file writing |
| `internal/storage` | Profile persistence (YAML files in `~/.zeus/profiles/`; reads both `.yaml` and `.toml`) |
| `internal/process` | PID file management, process existence checks, server launch |
| `internal/missions` | `.pbo` sync drivers (path copy/symlink, S3); ETag sidecar; `Puller` dispatcher |
| `cmd/` | CLI surface, user prompts, flag parsing |

## Profile Lifecycle

```
profile add (cmd)
  → ProfileLoader.FromFile         — YAML or TOML → Profile (with defaults; format detected by extension)
  → processProfileMods             — interactive mod/key management
  → ProfileRepository.Save         — write profile.yaml to ~/.zeus/profiles/ (always YAML)
  → ProfileWriter.Write            — write server.cfg + basic.cfg to install_dir

profile start (cmd)
  → ProfileRepository.Get          — load profile.yaml or .toml via loader.FromFile (no defaults)
  → Runner.prepareParams           — inject managed paths into Params
  → Runner.Run                     — exec.Start(arma3server_x64...)
  → creates ~/.zeus/running/<name>.pid (written by Arma 3 via -pid flag)

profile stop (cmd)
  → Manager.Kill                   — read PID file, os.FindProcess, proc.Kill, remove PID file
  → Manager.WaitGone               — poll until process gone AND PID file gone
```

## Config Generation Invariant

`ProfileWriter.Write` is called at `add` time, not at `start` time.
The generated `.cfg` files in `<install_dir>/.zeus/<name>/configs/` are created when
a profile is added or updated, then read by Arma 3 at start time.

If the profile YAML is modified externally (e.g. manually), `profile add` must be
re-run to regenerate the `.cfg` files.

## Defaults Pre-population Invariant

`ProfileRepository.Get` and `List` use `ProfileLoader.FromFile` without default
pre-population — they do **not** call `NewDefaultServerConfig()`. This is correct because:

1. Profiles saved by `ProfileRepository.Save` already contain all explicit values
   (defaults were merged at `add` time and round-trip through YAML marshal).
2. The `.cfg` files were already generated at `add` time.
3. `Runner.prepareParams` only modifies `Params` paths, not `Config`/`Basic`.

Do **not** add default pre-population to `ProfileRepository.Get`. It is intentionally
absent.

## Relative vs Absolute Paths in Runner

`Runner.prepareParams` sets config/mpmissions/keys/profiles paths as **relative** to
`install_dir`:

```go
zeusDir := filepath.Join(".zeus", p.Name)
p.Params.Config = &config   // ".zeus/<name>/configs/server.cfg"
```

This is intentional: `cmd.Dir = prepared.InstallDir` in `Runner.Run`, so Arma 3
resolves these relative to its installation directory.

The PID file path is **absolute** (`filepath.Join(home, ".zeus", "running", ...)`).
It is managed by ZEUS, not passed to Arma 3 as a relative path.

## PID File Ownership

| Creator | ZEUS injects path via `-pid` startup parameter |
|---------|------------------------------------------------|
| Writer | Arma 3 server process |
| Cleaner (clean exit) | Arma 3 server process |
| Cleaner (kill) | `Manager.Kill` removes PID file after sending kill signal |
| Cleaner (crash) | Not cleaned — `Manager.WaitGone` detects stale PID files |

ZEUS does **not** write the PID file itself. It only reads and deletes it.

## Platform-Specific Process Checking

```
internal/process/exists_windows.go  — OpenProcess + GetExitCodeProcess (STILL_ACTIVE=259)
internal/process/exists_unix.go     — syscall.Kill(pid, 0)
```

Both files implement the same `processExists(pid int) bool` signature. Selected by
build tags (`//go:build windows` / `//go:build !windows`). No external dependencies.

## Reflection-Based Argument Builder

`process/runner.go:buildArgs` iterates `StartupParams` fields using reflection, reading
`arg` struct tags. It handles three cases:

- `*bool` pointer (non-nil + true): append `-argName` (flag, no value)
- `*T` pointer (non-nil, non-bool): append `-argName=value`
- `[]string` slice (non-empty): append `-argName=v1;v2;v3`

Adding a new startup parameter requires only adding a field with an `arg` tag to
`StartupParams`. No changes to `buildArgs` are needed.

## Testability Rules

- Never call `os.UserHomeDir()` directly in testable code. Use the `HomeDir string`
  field pattern.
- Tests use `t.TempDir()` for isolation.
- The `internal/profile` package is tested with `package profile` (white-box).
- The `internal/storage` and `internal/process` packages are tested with `package storage`
  / `package process` (white-box).

## Related

- [Conventions](conventions.md)
- [Common Gotchas](common_gotchas.md)
