# Missions S3 Driver Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an S3 driver to `zeus missions pull` that syncs `.pbo` files from an S3-compatible bucket (AWS S3 or MinIO) into the profile's local mpmissions directory using ETag-based change detection.

**Architecture:** New `S3Driver` in `internal/missions/s3_driver.go` dispatched from the existing `Puller.Pull` switch. Change detection uses a `.zeus-s3-etags` JSON sidecar in `targetDir` to avoid re-downloading unchanged files. Downloads use atomic temp-file-then-rename. Credentials resolve: inline profile fields → env vars → standard AWS SDK chain.

**Tech Stack:** `aws-sdk-go-v2` (S3 client, credential chain, paginator), Go stdlib for sidecar I/O. Integration tests require MinIO running at `localhost:9000` with bucket `mpmissions`, credentials `minioadmin`/`minioadmin`. Run them with `go test -tags integration ./internal/missions/...`.

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `internal/arma/mission_source.go` | Modify | Add 6 S3 fields: Bucket, Prefix, Endpoint, Region, AccessKeyID, SecretAccessKey |
| `internal/profile/loader_test.go` | Modify | Add YAML + TOML round-trip tests for S3 fields |
| `internal/missions/s3_etags.go` | Create | `loadETags` / `saveETags` — JSON sidecar for per-file ETag tracking |
| `internal/missions/s3_etags_test.go` | Create | Unit tests for sidecar helpers (no build tag, no network) |
| `internal/missions/s3_driver.go` | Create | `S3Driver` — credential setup, ListObjectsV2, diff, GetObject download |
| `internal/missions/s3_driver_test.go` | Create | Integration tests (`//go:build integration`) against MinIO |
| `internal/missions/puller.go` | Modify | Add `"s3"` case to driver dispatch switch |
| `conf/example.yaml` | Modify | Add commented S3 `mission_source` example block |
| `conf/example.toml` | Modify | Add commented S3 `[mission_source]` example block |

---

## Task 1: Add `aws-sdk-go-v2` dependencies

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Fetch required SDK packages**

```
go get github.com/aws/aws-sdk-go-v2/config@latest
go get github.com/aws/aws-sdk-go-v2/credentials@latest
go get github.com/aws/aws-sdk-go-v2/service/s3@latest
go mod tidy
```

- [ ] **Step 2: Verify build still passes**

```
go build ./...
```

Expected: no output (success).

- [ ] **Step 3: Commit**

```
git add go.mod go.sum
git commit -m "chore: add aws-sdk-go-v2 dependencies"
```

---

## Task 2: S3 fields on `MissionSource` + profile loader tests

**Files:**
- Modify: `internal/arma/mission_source.go`
- Modify: `internal/profile/loader_test.go`

- [ ] **Step 1: Write failing profile loader tests for S3 fields**

Append to `internal/profile/loader_test.go` (after `TestProfileLoader_MissionSource_Absent`):

```go
func TestProfileLoader_MissionSource_S3_YAML(t *testing.T) {
	loader := ProfileLoader{}
	const yaml = `
name: srv
install_dir: /opt/arma3
mission_source:
  driver: s3
  bucket: mpmissions
  prefix: servers/prod/
  endpoint: "http://localhost:9000"
  region: us-east-1
  access_key_id: minioadmin
  secret_access_key: minioadmin
`
	p, err := loader.FromBytes([]byte(yaml))
	require.NoError(t, err)
	require.NotNil(t, p.MissionSource)
	assert.Equal(t, "s3", p.MissionSource.Driver)
	assert.Equal(t, "mpmissions", p.MissionSource.Bucket)
	assert.Equal(t, "servers/prod/", p.MissionSource.Prefix)
	assert.Equal(t, "http://localhost:9000", p.MissionSource.Endpoint)
	assert.Equal(t, "us-east-1", p.MissionSource.Region)
	assert.Equal(t, "minioadmin", p.MissionSource.AccessKeyID)
	assert.Equal(t, "minioadmin", p.MissionSource.SecretAccessKey)
}

func TestProfileLoader_MissionSource_S3_TOML(t *testing.T) {
	loader := ProfileLoader{}

	path := filepath.Join(t.TempDir(), "s3.toml")
	const toml = `
name = "srv"
install_dir = "/opt/arma3"

[mission_source]
driver            = "s3"
bucket            = "mpmissions"
prefix            = "servers/prod/"
endpoint          = "http://localhost:9000"
region            = "us-east-1"
access_key_id     = "minioadmin"
secret_access_key = "minioadmin"
`
	require.NoError(t, os.WriteFile(path, []byte(toml), 0600))

	p, err := loader.FromFile(path)
	require.NoError(t, err)
	require.NotNil(t, p.MissionSource)
	assert.Equal(t, "s3", p.MissionSource.Driver)
	assert.Equal(t, "mpmissions", p.MissionSource.Bucket)
	assert.Equal(t, "servers/prod/", p.MissionSource.Prefix)
	assert.Equal(t, "http://localhost:9000", p.MissionSource.Endpoint)
	assert.Equal(t, "us-east-1", p.MissionSource.Region)
	assert.Equal(t, "minioadmin", p.MissionSource.AccessKeyID)
	assert.Equal(t, "minioadmin", p.MissionSource.SecretAccessKey)
}
```

- [ ] **Step 2: Run to confirm failure**

```
go test ./internal/profile/... -run "TestProfileLoader_MissionSource_S3" -v
```

Expected: FAIL — fields do not exist yet.

- [ ] **Step 3: Add S3 fields to `MissionSource`**

Replace the full content of `internal/arma/mission_source.go`:

```go
package arma

// MissionSource describes where mission .pbo files are fetched from.
type MissionSource struct {
	Driver string `json:"driver"         yaml:"driver"         toml:"driver"`
	Path   string `json:"path,omitempty" yaml:"path,omitempty" toml:"path,omitempty"`
	// Mode: "copy" (default) or "symlink". symlink replaces the mpmissions dir with a link to Path.
	Mode string `json:"mode,omitempty" yaml:"mode,omitempty" toml:"mode,omitempty"`

	// S3 driver fields
	Bucket          string `json:"bucket,omitempty"          yaml:"bucket,omitempty"          toml:"bucket,omitempty"`
	Prefix          string `json:"prefix,omitempty"          yaml:"prefix,omitempty"          toml:"prefix,omitempty"`
	Endpoint        string `json:"endpoint,omitempty"        yaml:"endpoint,omitempty"        toml:"endpoint,omitempty"`
	Region          string `json:"region,omitempty"          yaml:"region,omitempty"          toml:"region,omitempty"`
	AccessKeyID     string `json:"accessKeyId,omitempty"     yaml:"access_key_id,omitempty"  toml:"access_key_id,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty" yaml:"secret_access_key,omitempty" toml:"secret_access_key,omitempty"`
}
```

- [ ] **Step 4: Run tests to confirm pass**

```
go test ./internal/profile/... -run "TestProfileLoader_MissionSource_S3" -v
```

Expected: PASS.

- [ ] **Step 5: Run full suite to confirm no regressions**

```
go test ./...
```

Expected: all packages `ok`.

- [ ] **Step 6: Commit**

```
git add internal/arma/mission_source.go internal/profile/loader_test.go
git commit -m "feat: add S3 fields to MissionSource"
```

---

## Task 3: ETag sidecar helpers

**Files:**
- Create: `internal/missions/s3_etags.go`
- Create: `internal/missions/s3_etags_test.go`

- [ ] **Step 1: Write failing unit tests**

Create `internal/missions/s3_etags_test.go`:

```go
package missions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadETags_NonExistent(t *testing.T) {
	etags, err := loadETags(t.TempDir())
	require.NoError(t, err)
	assert.Empty(t, etags)
}

func TestSaveAndLoadETags_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := map[string]string{
		"op_cobra.pbo":    "abc123def456",
		"op_patrol_v2.pbo": "deadbeef0000",
	}
	require.NoError(t, saveETags(dir, want))

	got, err := loadETags(dir)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestSaveETags_Overwrites(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, saveETags(dir, map[string]string{"a.pbo": "v1"}))
	require.NoError(t, saveETags(dir, map[string]string{"b.pbo": "v2"}))

	got, err := loadETags(dir)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"b.pbo": "v2"}, got)
}

func TestSaveETags_SidecarNotPBO(t *testing.T) {
	// Sidecar file must not be named .pbo so scanPBOs ignores it.
	dir := t.TempDir()
	require.NoError(t, saveETags(dir, map[string]string{"x.pbo": "hash"}))
	sidecarPath := filepath.Join(dir, etagSidecar)
	assert.True(t, filepath.Ext(sidecarPath) != ".pbo")
}

func TestLoadETags_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, etagSidecar), []byte("not json"), 0644))
	_, err := loadETags(dir)
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run to confirm failure**

```
go test ./internal/missions/... -run "TestLoadETags|TestSaveETags" -v
```

Expected: FAIL — `loadETags`, `saveETags`, `etagSidecar` not defined.

- [ ] **Step 3: Create `internal/missions/s3_etags.go`**

```go
package missions

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const etagSidecar = ".zeus-s3-etags"

// loadETags reads the ETag sidecar from dir.
// Returns an empty map (not an error) if the sidecar does not exist yet.
func loadETags(dir string) (map[string]string, error) {
	data, err := os.ReadFile(filepath.Join(dir, etagSidecar))
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	var etags map[string]string
	if err := json.Unmarshal(data, &etags); err != nil {
		return nil, err
	}
	return etags, nil
}

// saveETags writes the ETag sidecar atomically (temp + rename).
func saveETags(dir string, etags map[string]string) error {
	data, err := json.MarshalIndent(etags, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".zeus-s3-etags-tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, filepath.Join(dir, etagSidecar))
}
```

- [ ] **Step 4: Run tests to confirm pass**

```
go test ./internal/missions/... -run "TestLoadETags|TestSaveETags" -v
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```
git add internal/missions/s3_etags.go internal/missions/s3_etags_test.go
git commit -m "feat: add ETag sidecar helpers for S3 change detection"
```

---

## Task 4: S3Driver + integration tests

**Files:**
- Create: `internal/missions/s3_driver.go`
- Create: `internal/missions/s3_driver_test.go`

- [ ] **Step 1: Write failing integration tests**

Create `internal/missions/s3_driver_test.go`:

```go
//go:build integration

package missions

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smitt14ua/zeus/internal/arma"
)

const (
	testEndpoint  = "http://localhost:9000"
	testBucket    = "mpmissions"
	testAccessKey = "minioadmin"
	testSecretKey = "minioadmin"
	testRegion    = "us-east-1"
)

func newTestS3Client(t *testing.T) *s3.Client {
	t.Helper()
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(testRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(testAccessKey, testSecretKey, ""),
		),
	)
	require.NoError(t, err)
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(testEndpoint)
		o.UsePathStyle = true
	})
}

func s3Put(t *testing.T, client *s3.Client, key, content string) {
	t.Helper()
	_, err := client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(testBucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(content),
	})
	require.NoError(t, err)
}

func s3Delete(t *testing.T, client *s3.Client, key string) {
	t.Helper()
	_, err := client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(testBucket),
		Key:    aws.String(key),
	})
	require.NoError(t, err)
}

func testSource(prefix string) arma.MissionSource {
	return arma.MissionSource{
		Driver:          "s3",
		Bucket:          testBucket,
		Prefix:          prefix,
		Endpoint:        testEndpoint,
		Region:          testRegion,
		AccessKeyID:     testAccessKey,
		SecretAccessKey: testSecretKey,
	}
}

func TestS3Driver_AddNew(t *testing.T) {
	client := newTestS3Client(t)
	prefix := "TestS3Driver_AddNew/"
	key := prefix + "op_cobra.pbo"
	s3Put(t, client, key, "pbodata")
	t.Cleanup(func() { s3Delete(t, client, key) })

	dst := t.TempDir()
	result, err := Puller{}.Pull(testSource(prefix), dst, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"op_cobra.pbo"}, result.Added)
	assert.Empty(t, result.Updated)
	assert.Empty(t, result.Removed)
	assert.FileExists(t, filepath.Join(dst, "op_cobra.pbo"))
}

func TestS3Driver_SkipUnchanged(t *testing.T) {
	client := newTestS3Client(t)
	prefix := "TestS3Driver_SkipUnchanged/"
	key := prefix + "op.pbo"
	s3Put(t, client, key, "data")
	t.Cleanup(func() { s3Delete(t, client, key) })

	dst := t.TempDir()
	_, err := Puller{}.Pull(testSource(prefix), dst, false)
	require.NoError(t, err)

	result, err := Puller{}.Pull(testSource(prefix), dst, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"op.pbo"}, result.Skipped)
	assert.Empty(t, result.Added)
	assert.Empty(t, result.Updated)
}

func TestS3Driver_UpdateChanged(t *testing.T) {
	client := newTestS3Client(t)
	prefix := "TestS3Driver_UpdateChanged/"
	key := prefix + "op.pbo"
	s3Put(t, client, key, "original content")
	t.Cleanup(func() { s3Delete(t, client, key) })

	dst := t.TempDir()
	_, err := Puller{}.Pull(testSource(prefix), dst, false)
	require.NoError(t, err)

	s3Put(t, client, key, "updated content longer")

	result, err := Puller{}.Pull(testSource(prefix), dst, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"op.pbo"}, result.Updated)
	assert.Empty(t, result.Skipped)
}

func TestS3Driver_RemoveDeleted(t *testing.T) {
	client := newTestS3Client(t)
	prefix := "TestS3Driver_RemoveDeleted/"
	key := prefix + "op.pbo"
	s3Put(t, client, key, "data")

	dst := t.TempDir()
	_, err := Puller{}.Pull(testSource(prefix), dst, false)
	require.NoError(t, err)

	s3Delete(t, client, key)

	result, err := Puller{}.Pull(testSource(prefix), dst, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"op.pbo"}, result.Removed)
	assert.NoFileExists(t, filepath.Join(dst, "op.pbo"))
}

func TestS3Driver_DryRun(t *testing.T) {
	client := newTestS3Client(t)
	prefix := "TestS3Driver_DryRun/"
	key := prefix + "op.pbo"
	s3Put(t, client, key, "data")
	t.Cleanup(func() { s3Delete(t, client, key) })

	dst := t.TempDir()
	result, err := Puller{}.Pull(testSource(prefix), dst, true)
	require.NoError(t, err)
	assert.Equal(t, []string{"op.pbo"}, result.Added)
	assert.NoFileExists(t, filepath.Join(dst, "op.pbo"))
	_, statErr := os.Stat(filepath.Join(dst, etagSidecar))
	assert.True(t, os.IsNotExist(statErr)) // sidecar not written in dryRun
}

func TestS3Driver_MissingBucket(t *testing.T) {
	_, err := Puller{}.Pull(arma.MissionSource{Driver: "s3", Bucket: ""}, t.TempDir(), false)
	assert.ErrorContains(t, err, "bucket")
}

func TestS3Driver_LocalFileMissing(t *testing.T) {
	client := newTestS3Client(t)
	prefix := "TestS3Driver_LocalFileMissing/"
	key := prefix + "op.pbo"
	s3Put(t, client, key, "data")
	t.Cleanup(func() { s3Delete(t, client, key) })

	dst := t.TempDir()
	_, err := Puller{}.Pull(testSource(prefix), dst, false)
	require.NoError(t, err)

	require.NoError(t, os.Remove(filepath.Join(dst, "op.pbo")))

	result, err := Puller{}.Pull(testSource(prefix), dst, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"op.pbo"}, result.Added)
	assert.FileExists(t, filepath.Join(dst, "op.pbo"))
}

func TestS3Driver_IgnoresNonPBO(t *testing.T) {
	client := newTestS3Client(t)
	prefix := "TestS3Driver_IgnoresNonPBO/"
	pboKey := prefix + "op.pbo"
	txtKey := prefix + "readme.txt"
	s3Put(t, client, pboKey, "pbo data")
	s3Put(t, client, txtKey, "text")
	t.Cleanup(func() {
		s3Delete(t, client, pboKey)
		s3Delete(t, client, txtKey)
	})

	dst := t.TempDir()
	result, err := Puller{}.Pull(testSource(prefix), dst, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"op.pbo"}, result.Added)
	assert.NoFileExists(t, filepath.Join(dst, "readme.txt"))
}
```

- [ ] **Step 2: Run to confirm failure**

```
go test -tags integration ./internal/missions/... -run "TestS3Driver" -v
```

Expected: FAIL — `S3Driver` not defined / `"s3"` not in dispatch.

- [ ] **Step 3: Create `internal/missions/s3_driver.go`**

```go
package missions

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/smitt14ua/zeus/internal/arma"
)

// S3Driver syncs missions from an S3-compatible bucket.
type S3Driver struct{}

func (d S3Driver) pull(source arma.MissionSource, targetDir string, dryRun bool) (Result, error) {
	if source.Bucket == "" {
		return Result{}, fmt.Errorf("s3 driver requires bucket")
	}

	// Normalize prefix: ensure trailing slash when non-empty.
	prefix := source.Prefix
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	if !dryRun {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return Result{}, fmt.Errorf("creating target directory: %w", err)
		}
	}

	client, err := d.newClient(source)
	if err != nil {
		return Result{}, fmt.Errorf("building S3 client: %w", err)
	}

	s3Objects, err := d.listPBOs(client, source.Bucket, prefix)
	if err != nil {
		return Result{}, fmt.Errorf("listing S3 objects: %w", err)
	}

	knownETags, err := loadETags(targetDir)
	if err != nil {
		return Result{}, fmt.Errorf("loading etag sidecar: %w", err)
	}

	var result Result
	newETags := make(map[string]string)

	for name, etag := range s3Objects {
		localPath := filepath.Join(targetDir, name)
		knownETag, inSidecar := knownETags[name]

		if inSidecar && knownETag == etag && fileExists(localPath) {
			result.Skipped = append(result.Skipped, name)
			newETags[name] = etag
			continue
		}

		if !inSidecar || !fileExists(localPath) {
			result.Added = append(result.Added, name)
		} else {
			result.Updated = append(result.Updated, name)
		}

		newETags[name] = etag
		if !dryRun {
			if err := d.download(client, source.Bucket, prefix+name, localPath); err != nil {
				return Result{}, fmt.Errorf("downloading %s: %w", name, err)
			}
		}
	}

	for name := range knownETags {
		if _, ok := s3Objects[name]; !ok {
			result.Removed = append(result.Removed, name)
			if !dryRun {
				localPath := filepath.Join(targetDir, name)
				if err := os.Remove(localPath); err != nil && !os.IsNotExist(err) {
					return Result{}, fmt.Errorf("removing %s: %w", name, err)
				}
			}
		}
	}

	if !dryRun {
		if err := saveETags(targetDir, newETags); err != nil {
			return Result{}, fmt.Errorf("saving etag sidecar: %w", err)
		}
	}

	return result, nil
}

func (d S3Driver) newClient(source arma.MissionSource) (*s3.Client, error) {
	region := source.Region
	if region == "" {
		region = "us-east-1"
	}

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}
	if source.AccessKeyID != "" && source.SecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(source.AccessKeyID, source.SecretAccessKey, ""),
		))
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	s3Opts := []func(*s3.Options){}
	if source.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(source.Endpoint)
			o.UsePathStyle = true // required for MinIO and path-style S3 endpoints
		})
	}

	return s3.NewFromConfig(cfg, s3Opts...), nil
}

// listPBOs returns a map of bare filename → stripped ETag for all .pbo objects under prefix.
// Only direct children of the prefix are included (no sub-prefix recursion).
func (d S3Driver) listPBOs(client *s3.Client, bucket, prefix string) (map[string]string, error) {
	result := make(map[string]string)
	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)
			if !strings.HasSuffix(strings.ToLower(key), ".pbo") {
				continue
			}
			name := strings.TrimPrefix(key, prefix)
			if name == "" || strings.Contains(name, "/") {
				continue // skip exact prefix match or nested keys
			}
			// Strip surrounding quotes from ETag (S3 returns `"hash"` with quotes).
			etag := strings.Trim(aws.ToString(obj.ETag), `"`)
			result[name] = etag
		}
	}
	return result, nil
}

// download fetches an S3 object and writes it atomically to dst.
func (d S3Driver) download(client *s3.Client, bucket, key, dst string) error {
	out, err := client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	defer out.Body.Close()

	tmp, err := os.CreateTemp(filepath.Dir(dst), ".pbo-tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := io.Copy(tmp, out.Body); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing download: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Rename(tmpName, dst); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming to target: %w", err)
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
```

- [ ] **Step 4: Run integration tests**

Ensure MinIO is running at `localhost:9000` with bucket `mpmissions` created and credentials `minioadmin`/`minioadmin`.

```
go test -tags integration ./internal/missions/... -run "TestS3Driver" -v
```

Expected: all PASS.

- [ ] **Step 5: Run non-integration suite to confirm no regressions**

```
go test ./internal/missions/...
```

Expected: `ok` (integration tests skipped, 9 copy tests + 5 unit sidecar tests pass).

- [ ] **Step 6: Commit**

```
git add internal/missions/s3_driver.go internal/missions/s3_driver_test.go
git commit -m "feat: add S3Driver with ETag-based change detection"
```

---

## Task 5: Wire S3 driver into `Puller`

**Files:**
- Modify: `internal/missions/puller.go`

- [ ] **Step 1: Add `"s3"` case to the dispatch switch**

In `internal/missions/puller.go`, replace the switch block:

```go
func (p Puller) Pull(source arma.MissionSource, targetDir string, dryRun bool) (Result, error) {
	switch source.Driver {
	case "path":
		return PathDriver{}.pull(source, targetDir, dryRun)
	case "s3":
		return S3Driver{}.pull(source, targetDir, dryRun)
	default:
		return Result{}, fmt.Errorf("unsupported driver %q", source.Driver)
	}
}
```

- [ ] **Step 2: Build check**

```
go build ./...
```

Expected: no output.

- [ ] **Step 3: Run full non-integration suite**

```
go test ./...
```

Expected: all packages `ok`.

- [ ] **Step 4: Run integration suite**

```
go test -tags integration ./internal/missions/... -v
```

Expected: all `TestS3Driver_*` PASS.

- [ ] **Step 5: Commit**

```
git add internal/missions/puller.go
git commit -m "feat: register S3 driver in Puller dispatch"
```

---

## Task 6: Update example configs

**Files:**
- Modify: `conf/example.yaml`
- Modify: `conf/example.toml`

- [ ] **Step 1: Add S3 example to `conf/example.yaml`**

Append after the existing `mission_source:` block:

```yaml

# S3 driver example (AWS S3 or MinIO):
# mission_source:
#   driver: s3
#   bucket: mpmissions
#   prefix: ""                          # optional — key prefix, e.g. "servers/prod/"
#   endpoint: "http://localhost:9000"   # optional — omit for real AWS
#   region: us-east-1                   # optional — defaults to us-east-1
#   access_key_id: ""                   # optional — falls back to AWS_ACCESS_KEY_ID env var
#   secret_access_key: ""               # optional — falls back to AWS_SECRET_ACCESS_KEY env var
```

- [ ] **Step 2: Add S3 example to `conf/example.toml`**

Append after the existing `[mission_source]` block:

```toml

# S3 driver example (AWS S3 or MinIO):
# [mission_source]
# driver            = "s3"
# bucket            = "mpmissions"
# prefix            = ""                         # optional — key prefix, e.g. "servers/prod/"
# endpoint          = "http://localhost:9000"    # optional — omit for real AWS
# region            = "us-east-1"               # optional — defaults to us-east-1
# access_key_id     = ""                         # optional — falls back to AWS_ACCESS_KEY_ID
# secret_access_key = ""                         # optional — falls back to AWS_SECRET_ACCESS_KEY
```

- [ ] **Step 3: Final full test run**

```
go test ./...
```

Expected: all packages `ok`.

- [ ] **Step 4: Commit**

```
git add conf/example.yaml conf/example.toml
git commit -m "docs: add S3 driver example to profile configs"
```

---

## Self-Review

**Spec coverage:**
- ✅ MissionSource S3 fields (Bucket, Prefix, Endpoint, Region, AccessKeyID, SecretAccessKey) — Task 2
- ✅ Credential resolution: inline → env → SDK chain — Task 4 (`newClient`)
- ✅ Custom endpoint + `UsePathStyle` for MinIO — Task 4
- ✅ List .pbo only, skip non-.pbo, skip nested keys — Task 4 (`listPBOs`)
- ✅ ETag sidecar for change detection — Task 3
- ✅ Add / Update / Removed / Skipped result categories — Task 4
- ✅ Atomic download (temp + rename) — Task 4 (`download`)
- ✅ dryRun: no filesystem changes, no sidecar write — Task 4
- ✅ Error: empty bucket → fatal — Task 4
- ✅ Local file missing despite sidecar → re-download → Added — Task 4
- ✅ Puller dispatch wired — Task 5
- ✅ Integration tests for all 8 scenarios — Task 4
- ✅ Example configs updated — Task 6

**Type consistency:**
- `S3Driver.pull` signature matches `PathDriver.pull` — `(source arma.MissionSource, targetDir string, dryRun bool) (Result, error)` ✅
- `etagSidecar` constant defined in `s3_etags.go`, referenced in `s3_etags_test.go` ✅
- `fileExists` defined in `s3_driver.go`, used only there ✅
- `loadETags` / `saveETags` defined in `s3_etags.go`, called from `s3_driver.go` ✅
