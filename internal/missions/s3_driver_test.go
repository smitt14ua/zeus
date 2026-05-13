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
	"github.com/smitt14ua/zeus/internal/arma"
)

const (
	testEndpoint  = "http://localhost:9000"
	testBucket    = "mpmissions"
	testAccessKey = "minioadmin"
	testSecretKey = "minioadmin"
	testRegion    = "us-east-1"
)

func testSource(t *testing.T) arma.MissionSource {
	t.Helper()
	return arma.MissionSource{
		Driver:          "s3",
		Bucket:          testBucket,
		Prefix:          t.Name() + "/",
		Endpoint:        testEndpoint,
		Region:          testRegion,
		AccessKeyID:     testAccessKey,
		SecretAccessKey: testSecretKey,
	}
}

func testClient(t *testing.T) *s3.Client {
	t.Helper()
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(testRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(testAccessKey, testSecretKey, "")),
	)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(testEndpoint)
		o.UsePathStyle = true
	})
}

func uploadPBO(t *testing.T, client *s3.Client, prefix, name, content string) {
	t.Helper()
	key := prefix + name
	_, err := client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(testBucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(content),
	})
	if err != nil {
		t.Fatalf("upload %q: %v", key, err)
	}
}

func deletePBO(t *testing.T, client *s3.Client, prefix, name string) {
	t.Helper()
	key := prefix + name
	_, err := client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(testBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		t.Fatalf("delete %q: %v", key, err)
	}
}

func TestS3Driver_AddNew(t *testing.T) {
	client := testClient(t)
	source := testSource(t)
	dir := t.TempDir()

	uploadPBO(t, client, source.Prefix, "op_cobra.pbo", "content1")
	t.Cleanup(func() { deletePBO(t, client, source.Prefix, "op_cobra.pbo") })

	drv := S3Driver{}
	result, err := drv.pull(source, dir, false, false)
	if err != nil {
		t.Fatalf("pull: %v", err)
	}
	if len(result.Added) != 1 || result.Added[0] != "op_cobra.pbo" {
		t.Errorf("expected Added=[op_cobra.pbo], got %v", result.Added)
	}
	if _, err := os.Stat(filepath.Join(dir, "op_cobra.pbo")); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestS3Driver_SkipUnchanged(t *testing.T) {
	client := testClient(t)
	source := testSource(t)
	dir := t.TempDir()

	uploadPBO(t, client, source.Prefix, "op_patrol.pbo", "content1")
	t.Cleanup(func() { deletePBO(t, client, source.Prefix, "op_patrol.pbo") })

	drv := S3Driver{}
	_, err := drv.pull(source, dir, false, false)
	if err != nil {
		t.Fatalf("first pull: %v", err)
	}

	result, err := drv.pull(source, dir, false, false)
	if err != nil {
		t.Fatalf("second pull: %v", err)
	}
	if len(result.Skipped) != 1 || result.Skipped[0] != "op_patrol.pbo" {
		t.Errorf("expected Skipped=[op_patrol.pbo], got %v", result.Skipped)
	}
	if len(result.Added)+len(result.Updated) != 0 {
		t.Errorf("expected no add/update, got added=%v updated=%v", result.Added, result.Updated)
	}
}

func TestS3Driver_UpdateChanged(t *testing.T) {
	client := testClient(t)
	source := testSource(t)
	dir := t.TempDir()

	uploadPBO(t, client, source.Prefix, "op_delta.pbo", "original")
	t.Cleanup(func() { deletePBO(t, client, source.Prefix, "op_delta.pbo") })

	drv := S3Driver{}
	_, err := drv.pull(source, dir, false, false)
	if err != nil {
		t.Fatalf("first pull: %v", err)
	}

	uploadPBO(t, client, source.Prefix, "op_delta.pbo", "updated content")

	result, err := drv.pull(source, dir, false, false)
	if err != nil {
		t.Fatalf("second pull: %v", err)
	}
	if len(result.Updated) != 1 || result.Updated[0] != "op_delta.pbo" {
		t.Errorf("expected Updated=[op_delta.pbo], got %v", result.Updated)
	}
}

func TestS3Driver_RemoveDeleted(t *testing.T) {
	client := testClient(t)
	source := testSource(t)
	dir := t.TempDir()

	uploadPBO(t, client, source.Prefix, "op_gone.pbo", "content")

	drv := S3Driver{}
	_, err := drv.pull(source, dir, false, false)
	if err != nil {
		t.Fatalf("first pull: %v", err)
	}

	deletePBO(t, client, source.Prefix, "op_gone.pbo")

	result, err := drv.pull(source, dir, false, false)
	if err != nil {
		t.Fatalf("second pull: %v", err)
	}
	if len(result.Removed) != 1 || result.Removed[0] != "op_gone.pbo" {
		t.Errorf("expected Removed=[op_gone.pbo], got %v", result.Removed)
	}
	if _, err := os.Stat(filepath.Join(dir, "op_gone.pbo")); !os.IsNotExist(err) {
		t.Errorf("file should be deleted")
	}
}

func TestS3Driver_DryRun(t *testing.T) {
	client := testClient(t)
	source := testSource(t)
	dir := t.TempDir()

	uploadPBO(t, client, source.Prefix, "op_dry.pbo", "content")
	t.Cleanup(func() { deletePBO(t, client, source.Prefix, "op_dry.pbo") })

	drv := S3Driver{}
	result, err := drv.pull(source, dir, true, false)
	if err != nil {
		t.Fatalf("pull: %v", err)
	}
	if len(result.Added) != 1 || result.Added[0] != "op_dry.pbo" {
		t.Errorf("expected Added=[op_dry.pbo], got %v", result.Added)
	}
	if _, err := os.Stat(filepath.Join(dir, "op_dry.pbo")); !os.IsNotExist(err) {
		t.Errorf("dry run should not create file")
	}
}

func TestS3Driver_MissingBucket(t *testing.T) {
	source := arma.MissionSource{
		Driver:          "s3",
		Endpoint:        testEndpoint,
		Region:          testRegion,
		AccessKeyID:     testAccessKey,
		SecretAccessKey: testSecretKey,
	}
	drv := S3Driver{}
	_, err := drv.pull(source, t.TempDir(), false, false)
	if err == nil || !strings.Contains(err.Error(), "bucket") {
		t.Errorf("expected bucket error, got %v", err)
	}
}

// TestS3Driver_SubdirIgnoredAtRootPrefix verifies that objects stored under a
// subdirectory in the bucket are ignored when no prefix is configured.
// Only flat objects at the prefix level are synced.
func TestS3Driver_SubdirIgnoredAtRootPrefix(t *testing.T) {
	client := testClient(t)
	source := testSource(t)
	dir := t.TempDir()

	uploadPBO(t, client, source.Prefix, "root_mission.Altis.pbo", "content")
	uploadPBO(t, client, source.Prefix, "subdir/nested.Stratis.pbo", "content")
	t.Cleanup(func() {
		deletePBO(t, client, source.Prefix, "root_mission.Altis.pbo")
		deletePBO(t, client, source.Prefix, "subdir/nested.Stratis.pbo")
	})

	drv := S3Driver{}
	result, err := drv.pull(source, dir, false, false)
	if err != nil {
		t.Fatalf("pull: %v", err)
	}
	if len(result.Added) != 1 || result.Added[0] != "root_mission.Altis.pbo" {
		t.Errorf("expected Added=[root_mission.Altis.pbo] only, got %v", result.Added)
	}
	if _, err := os.Stat(filepath.Join(dir, "root_mission.Altis.pbo")); err != nil {
		t.Errorf("root file not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "nested.Stratis.pbo")); err == nil {
		t.Error("subdirectory file must not be downloaded")
	}
}

// TestS3Driver_PrefixScopesDownload verifies that only objects directly under
// the configured prefix are synced, and objects outside it are ignored.
func TestS3Driver_PrefixScopesDownload(t *testing.T) {
	client := testClient(t)
	source := testSource(t)
	dir := t.TempDir()

	// Files at root level of test prefix.
	uploadPBO(t, client, source.Prefix, "root.Altis.pbo", "content")
	// Files in a "submissions" subfolder.
	uploadPBO(t, client, source.Prefix, "submissions/sub.Stratis.pbo", "content")
	t.Cleanup(func() {
		deletePBO(t, client, source.Prefix, "root.Altis.pbo")
		deletePBO(t, client, source.Prefix, "submissions/sub.Stratis.pbo")
	})

	// Pull with prefix pointing at submissions/ — only sub.Stratis.pbo expected.
	subSource := source
	subSource.Prefix = source.Prefix + "submissions"
	drv := S3Driver{}
	result, err := drv.pull(subSource, dir, false, false)
	if err != nil {
		t.Fatalf("pull: %v", err)
	}
	if len(result.Added) != 1 || result.Added[0] != "sub.Stratis.pbo" {
		t.Errorf("expected Added=[sub.Stratis.pbo], got %v", result.Added)
	}
	if _, err := os.Stat(filepath.Join(dir, "sub.Stratis.pbo")); err != nil {
		t.Errorf("sub file not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "root.Altis.pbo")); err == nil {
		t.Error("root file must not be downloaded when prefix scopes to submissions/")
	}
}

func TestS3Driver_LocalFileMissing(t *testing.T) {
	client := testClient(t)
	source := testSource(t)
	dir := t.TempDir()

	uploadPBO(t, client, source.Prefix, "op_missing.pbo", "content")
	t.Cleanup(func() { deletePBO(t, client, source.Prefix, "op_missing.pbo") })

	drv := S3Driver{}
	_, err := drv.pull(source, dir, false, false)
	if err != nil {
		t.Fatalf("first pull: %v", err)
	}

	// Delete local file, keep sidecar
	if err := os.Remove(filepath.Join(dir, "op_missing.pbo")); err != nil {
		t.Fatal(err)
	}

	result, err := drv.pull(source, dir, false, false)
	if err != nil {
		t.Fatalf("second pull: %v", err)
	}
	if len(result.Added) != 1 || result.Added[0] != "op_missing.pbo" {
		t.Errorf("expected Added=[op_missing.pbo] (re-download), got %v", result.Added)
	}
	if _, err := os.Stat(filepath.Join(dir, "op_missing.pbo")); err != nil {
		t.Errorf("file should exist after re-download: %v", err)
	}
}
