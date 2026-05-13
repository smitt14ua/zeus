# Missions S3 Driver — Design Spec

**Date:** 2026-05-08
**Status:** Approved

## Summary

Add an S3 driver to `zeus missions pull` that syncs `.pbo` mission files from an S3-compatible bucket (AWS S3, MinIO, etc.) into the profile's local mpmissions directory.

Uses `aws-sdk-go-v2` for the S3 client. Change detection is based on ETag comparison via a local sidecar file.

---

## 1. Profile Schema

New optional fields on `MissionSource` (flat, all `omitempty`):

```yaml
mission_source:
  driver: s3
  bucket: mpmissions
  prefix: ""                        # optional — S3 key prefix, e.g. "servers/prod/"
  endpoint: "http://localhost:9000" # optional — omit for real AWS
  region: us-east-1                 # optional — defaults to "us-east-1" when empty
  access_key_id: minioadmin         # optional — see credential resolution below
  secret_access_key: minioadmin     # optional — see credential resolution below
```

```toml
[mission_source]
driver            = "s3"
bucket            = "mpmissions"
prefix            = ""
endpoint          = "http://localhost:9000"
region            = "us-east-1"
access_key_id     = "minioadmin"
secret_access_key = "minioadmin"
```

### Credential resolution order

1. Inline `access_key_id` + `secret_access_key` fields (both must be non-empty)
2. `AWS_ACCESS_KEY_ID` + `AWS_SECRET_ACCESS_KEY` environment variables
3. Standard AWS SDK credential chain (shared credentials file, IAM instance role, etc.)

### Struct additions to `internal/arma/mission_source.go`

```go
type MissionSource struct {
    Driver string `json:"driver"           yaml:"driver"           toml:"driver"`
    Path   string `json:"path,omitempty"   yaml:"path,omitempty"   toml:"path,omitempty"`
    Mode   string `json:"mode,omitempty"   yaml:"mode,omitempty"   toml:"mode,omitempty"`

    // S3 driver fields
    Bucket          string `json:"bucket,omitempty"          yaml:"bucket,omitempty"          toml:"bucket,omitempty"`
    Prefix          string `json:"prefix,omitempty"          yaml:"prefix,omitempty"          toml:"prefix,omitempty"`
    Endpoint        string `json:"endpoint,omitempty"        yaml:"endpoint,omitempty"        toml:"endpoint,omitempty"`
    Region          string `json:"region,omitempty"          yaml:"region,omitempty"          toml:"region,omitempty"`
    AccessKeyID     string `json:"accessKeyId,omitempty"     yaml:"access_key_id,omitempty"   toml:"access_key_id,omitempty"`
    SecretAccessKey string `json:"secretAccessKey,omitempty" yaml:"secret_access_key,omitempty" toml:"secret_access_key,omitempty"`
}
```

---

## 2. `internal/missions` Package Changes

### New files

```
internal/missions/
  s3_driver.go      — S3Driver: credential setup, list, diff, download
  s3_etags.go       — loadETags / saveETags helpers
  s3_driver_test.go — integration tests (build tag: integration)
```

### S3Driver

```go
type S3Driver struct{}

func (d S3Driver) pull(source arma.MissionSource, targetDir string, dryRun bool) (Result, error)
```

Dispatched from `Puller.Pull` when `source.Driver == "s3"`.

### Sync algorithm

1. Validate `source.Bucket` non-empty → error `"s3 driver requires bucket"`
2. Build `aws.Config` with credential provider chain (inline → env → SDK default)
3. If `source.Endpoint` non-empty, set `BaseEndpoint` + `UsePathStyle: true` (required for MinIO)
4. List all objects at `s3://bucket/prefix` with paginated `ListObjectsV2`; filter to keys ending in `.pbo`
5. Strip prefix from each key to get bare filename (e.g. `"servers/prod/op.pbo"` → `"op.pbo"`)
6. Load ETag sidecar: `loadETags(targetDir)` → `map[string]string` (filename → ETag)
7. Diff:
   - Filename in S3, not in sidecar → **Added** (download)
   - Filename in S3, ETag matches sidecar, local file exists → **Skipped**
   - Filename in S3, ETag matches sidecar, local file missing → re-download → **Added**
   - Filename in S3, ETag differs from sidecar → **Updated** (download)
   - Filename in sidecar, not in S3 → **Removed** (delete local file)
8. If not `dryRun`: apply downloads (atomic: temp file + rename) and deletes
9. If not `dryRun`: save updated sidecar

No symlink mode — S3 driver is copy-only.

### ETag sidecar

File: `<targetDir>/.zeus-s3-etags`

Format: JSON object mapping basename → ETag string.

```json
{
  "op_cobra.pbo": "\"d41d8cd98f00b204e9800998ecf8427e\"",
  "op_patrol_v2.pbo": "\"abc123...\""
}
```

Not a `.pbo` file — `scanPBOs` already ignores it.

#### `s3_etags.go`

```go
func loadETags(dir string) (map[string]string, error)
// Returns empty map (not error) if sidecar does not exist.

func saveETags(dir string, etags map[string]string) error
// Writes atomically (temp + rename).
```

### `puller.go` change

```go
case "s3":
    return S3Driver{}.pull(source, targetDir, dryRun)
```

---

## 3. Dependencies

Add via `go get`:

```
github.com/aws/aws-sdk-go-v2
github.com/aws/aws-sdk-go-v2/config
github.com/aws/aws-sdk-go-v2/credentials
github.com/aws/aws-sdk-go-v2/service/s3
```

---

## 4. Testing

Integration tests in `internal/missions/s3_driver_test.go` behind build tag:

```go
//go:build integration
```

Run with: `go test -tags integration ./internal/missions/...`

Tests use MinIO at `localhost:9000`, bucket `mpmissions`, credentials `minioadmin`/`minioadmin`.

Test cases:
- `TestS3Driver_AddNew` — upload `.pbo` to MinIO, pull, assert Added + file exists
- `TestS3Driver_SkipUnchanged` — pull twice, second run asserts all Skipped
- `TestS3Driver_UpdateChanged` — re-upload with new content, pull, assert Updated
- `TestS3Driver_RemoveDeleted` — delete from S3, pull, assert Removed + local file gone
- `TestS3Driver_DryRun` — pull with dryRun=true, assert no local changes
- `TestS3Driver_MissingBucket` — empty bucket field → error
- `TestS3Driver_LocalFileMissing` — sidecar exists but local file deleted → re-download → Added

Each test uses a unique key prefix (`t.Name()`) to avoid cross-test interference.

---

## 5. Error Cases

| Situation | Behavior |
|-----------|----------|
| `bucket` field empty | fatal: `"s3 driver requires bucket"` |
| Bucket does not exist | fatal with S3 error |
| Credential resolution fails | fatal with AWS error |
| Object download fails | fatal with S3 error |
| Sidecar write fails | fatal with os error |
| `dryRun = true` | diff computed, no S3 calls beyond ListObjects, no local changes |

---

## 6. Files Changed / Created

| File | Change |
|------|--------|
| `internal/arma/mission_source.go` | Add 6 S3 fields |
| `internal/missions/s3_driver.go` | New — `S3Driver` |
| `internal/missions/s3_etags.go` | New — `loadETags`, `saveETags` |
| `internal/missions/s3_driver_test.go` | New — integration tests |
| `internal/missions/puller.go` | Add `"s3"` dispatch case |
| `conf/example.yaml` | Add S3 `mission_source` example block |
| `conf/example.toml` | Add S3 `[mission_source]` example block |
| `go.mod` / `go.sum` | Add `aws-sdk-go-v2` deps |
