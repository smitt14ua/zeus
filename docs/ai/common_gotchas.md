# Common Gotchas

## Purpose

Hidden edge cases, dangerous assumptions, and non-obvious behaviors in the ZEUS
codebase and Arma 3 server configuration.

---

## 1. Arma 3 Config Defaults Are Not Zero Values

**Trap:** Assuming unset fields produce correct output.

Several Arma 3 defaults are non-zero:

| Field | Go zero | Arma 3 default |
|-------|---------|----------------|
| `MaxPing` | `0` | `-1` (unlimited) |
| `MaxPacketLoss` | `0` | `-1` (unlimited) |
| `MaxDesync` | `0` | `-1` (unlimited) |
| `DisconnectTimeout` | `0` | `15` |
| `VerifySignatures` | `0` | `2` |
| `BattlEye` | `false` | `true` |
| `VoteThreshold` | `0.0` | `0.5` |
| `MaxPlayers` | `0` | `64` |
| `VoteMissionPlayers` | `0` | `1` |

If `Profile.Config` is zero-initialized before YAML unmarshal, these fields will be
written to `server.cfg` with wrong values (e.g. `MaxPing = 0` kicks everyone immediately).

**Fix:** Always use `arma.NewDefaultServerConfig()` and `arma.NewDefaultBasicServerConfig()`
when creating a `Profile` before unmarshaling.

---

## 2. Explicit Zero in YAML Overrides Default

**Trap:** Thinking zeros in YAML are "unset".

The YAML library uses presence-tracking. A field explicitly written as `0` in YAML
is distinct from an absent field and correctly overrides the pre-populated default.

```yaml
config:
  max_ping: 0   # overrides default -1 → written to server.cfg as: maxPing = 0;
```

This is tested in `TestProfileLoader_ExplicitZeroOverridesDefault`.

---

## 3. ProfileRepository.Get Does Not Pre-populate Defaults

**Trap:** Expecting `repo.Get()` to return a Profile with Arma 3 defaults.

`ProfileRepository.Get` and `List` use `ProfileLoader.FromFile` in a mode that does
**not** call `NewDefaultServerConfig()`. This is intentional — see [Architecture Rules](architecture_rules.md).

Do not add default pre-population to `Get`/`List`. Profiles stored on disk already
carry the values they need.

---

## 4. Pipe vs Terminal stdin Conflict

**Trap:** Using `bufio.Scanner(os.Stdin)` when stdin carries piped JSON.

`profile add` reads JSON from stdin when piped. If interactive prompts also try to
read from `os.Stdin`, the pipe content is consumed by the scanner instead of the user's
keystrokes, producing garbled output and empty answers.

**Fix:** Check `stdinIsPipe()` before creating a scanner. When piping, `force = true`
automatically, which skips all interactive prompts (scanner is never created).

---

## 5. Relative Paths in Runner Are Intentional

**Trap:** "Fixing" the relative paths in `Runner.prepareParams`.

```go
zeusDir := filepath.Join(".zeus", p.Name)  // NOT: filepath.Join(installDir, ".zeus", p.Name)
```

These are relative paths passed as startup parameters to Arma 3. They are resolved
relative to `install_dir` because `cmd.Dir = prepared.InstallDir` in `Runner.Run`.
Making them absolute would break Arma 3's path resolution.

The PID file IS absolute (it's a ZEUS internal path, not passed to Arma 3).

---

## 6. YAML Package Is Not gopkg.in/yaml.v3

**Trap:** Importing the wrong YAML package.

ZEUS uses `go.yaml.in/yaml/v4` (not `gopkg.in/yaml.v3`). The `gopkg.in/yaml.v3`
in `go.mod` is an indirect dependency of `testify`.

```go
import "go.yaml.in/yaml/v4"   // ✓ correct
import "gopkg.in/yaml.v3"      // ✗ wrong
```

---

## 7. MaxPing/MaxPacketLoss/MaxDesync Default Is -1, Not 0

**Trap:** Setting these to `0` thinking it's "no limit".

`-1` means unlimited (no kicking). `0` means kick immediately on any ping/packet loss.
This is a common misconfiguration that kicks all players instantly.

```yaml
config:
  max_ping: 200        # kick players above 200ms
  # max_ping: 0        # ← kicks everyone immediately
  # (omit entirely)    # ← uses default -1, never kicks
```

---

## 8. PID File Survives Crashes

**Trap:** Assuming the PID file is always cleaned up.

Arma 3 deletes its own PID file on clean exit. On crash, the file remains.
`Manager.WaitGone` polls both `processExists(pid)` AND PID file absence.

If a stale PID file exists from a previous crash, `Manager.List` will report that
profile as "running" even though the process is dead.

**Workaround:** Manual deletion of `~/.zeus/running/<name>.pid` clears the stale state.

---

## 9. Mission Rotation Difficulty vs forcedDifficulty

**Trap:** Setting `forcedDifficulty` in server.cfg and expecting it to always apply.

If a mission in the rotation has a `difficulty` field set, it overrides `forcedDifficulty`.
`forcedDifficulty` only applies when the mission does not specify difficulty.

---

## 10. autoInit Breaks Mission Parameters

**Trap:** Using `-autoInit` with missions that have configurable parameters.

`-autoInit` causes the server to initialize the mission before players connect.
Mission parameter prompts never display, and all parameters return default values only.

Do not combine `-autoInit` with missions that rely on `class Params`.

---

## 11. verifySignatures = 1 Is Deprecated

**Trap:** Using `verifySignatures = 1`.

Value `1` was deprecated and now falls back to `2`. Effective values are `0` (disabled)
and `2` (v2 signatures only). The default is `2`.

---

## 12. keysFolder Excludes Default Keys Folder If Specified

**Trap:** Adding `-keysFolder` and losing the base game keys.

By default, the Arma 3 `keys/` folder in the install directory is always included.
When `-keysFolder` is specified, the default folder is **still included** unless
explicitly excluded with `!keys`:

```
-keysFolder=!keys;@mymod/keys   # excludes base keys, only uses mod keys
-keysFolder=@mymod/keys         # includes both base keys AND mod keys
```

ZEUS appends to `KeysFolder` in `prepareParams`, which generates a cumulative list.

---

## 13. TOML Custom Scalar Types Accept Both Integer and String

**Trap:** Assuming TOML `max_bandwidth = 750000000` (integer) won't work.

`DataSize`, `DataTransferRate`, and `Time` have `UnmarshalTOML` implementations that
handle both TOML integers and TOML strings. Both of these are valid:

```toml
max_bandwidth = 750000000      # integer — raw bits per second
max_bandwidth = "750Mbps"      # string — parsed by ParseDataTransferRate
```

In YAML, both are also valid (`max_bandwidth: 750000000` or `max_bandwidth: "750Mbps"`).
Do not remove the integer case from `UnmarshalTOML` — it is needed for round-trip
compatibility if a TOML file is machine-generated.

---

## 14. Non-JSON Input Is Converted to JSON on Save

**Trap:** Expecting a YAML- or TOML-sourced profile to be stored in its original format.

`ProfileRepository.Save` always writes `.json`, regardless of input format. A profile
added from `server.yaml` or `server.toml` is stored as `~/.zeus/profiles/<name>.json`.
The original file is not kept. `profile info --toml` re-serializes from the stored JSON.

---

## 15. S3 Driver Only Syncs Flat Objects at Prefix Level

**Trap:** Expecting S3 objects in subdirectories to be downloaded.

`listPBOs` strips the configured prefix from each key and skips any result that still
contains `/`. This means only objects stored directly at the prefix level are synced.

```
bucket: missions  (prefix: "")
  mission1.VR.pbo           → downloaded ✓
  submissions/mission3.pbo  → ignored (contains "/" after strip) ✗
```

With `prefix: submissions`, `ListObjectsV2` only returns objects whose keys start with
`submissions/`. After stripping, `mission3.Stratis.pbo` has no `/` → downloaded.
Root-level files are excluded by the S3 listing itself.

**Fix:** Set `prefix: <folder>` in `mission_source` to scope the pull to that folder.
Objects not under the prefix are never returned by S3, and objects deeper than one level
below the prefix are filtered by the `/` check.

---

## 16. S3 Objects Use Full Key for Download, Flat Name Locally

**Trap:** Assuming the local filename and the S3 key are the same string.

`S3Driver.listPBOs` returns `map[string]s3Object` where the map key is the flat local
filename and `s3Object.key` is the full S3 object key (used for `GetObject`). These are
different when a prefix is set.

```
S3 key:     "submissions/mission3.Stratis.pbo"
local name: "mission3.Stratis.pbo"
```

Do not reconstruct the S3 key from the local name — use `obj.key` from the struct.

---

## Related

- [Conventions](conventions.md)
- [Architecture Rules](architecture_rules.md)
