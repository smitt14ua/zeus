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

func (d S3Driver) pull(source arma.MissionSource, targetDir string, dryRun, serverRunning bool) (Result, error) {
	if source.Bucket == "" {
		return Result{}, fmt.Errorf("s3 driver requires bucket")
	}

	client, err := d.newClient(source)
	if err != nil {
		return Result{}, err
	}

	prefix := source.Prefix
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	s3Files, err := d.listPBOs(client, source.Bucket, prefix)
	if err != nil {
		return Result{}, err
	}

	existingETags, err := loadETags(targetDir)
	if err != nil {
		return Result{}, err
	}

	var result Result

	// Files in S3 not in sidecar, or ETag changed, or local file missing → add/update
	for name, etag := range s3Files {
		localPath := filepath.Join(targetDir, name)
		knownETag, known := existingETags[name]
		localExists := fileExists(localPath)

		switch {
		case !known || !localExists:
			result.Added = append(result.Added, name)
			if !dryRun {
				if err := os.MkdirAll(targetDir, 0755); err != nil {
					return Result{}, err
				}
				key := prefix + name
				if err := d.download(client, source.Bucket, key, localPath); err != nil {
					return Result{}, err
				}
				existingETags[name] = etag
			}
		case knownETag != etag:
			if serverRunning {
				result.Locked = append(result.Locked, name)
			} else {
				result.Updated = append(result.Updated, name)
				if !dryRun {
					key := prefix + name
					if err := d.download(client, source.Bucket, key, localPath); err != nil {
						return Result{}, err
					}
					existingETags[name] = etag
				}
			}
		default:
			result.Skipped = append(result.Skipped, name)
		}
	}

	// Files in sidecar not in S3 → remove (skip if server is running)
	for name := range existingETags {
		if _, ok := s3Files[name]; !ok {
			if serverRunning {
				result.Locked = append(result.Locked, name)
			} else {
				result.Removed = append(result.Removed, name)
				if !dryRun {
					localPath := filepath.Join(targetDir, name)
					if err := os.Remove(localPath); err != nil && !os.IsNotExist(err) {
						return Result{}, err
					}
					delete(existingETags, name)
				}
			}
		}
	}

	if !dryRun {
		if err := saveETags(targetDir, existingETags); err != nil {
			return Result{}, err
		}
	}

	return result, nil
}

func (d S3Driver) newClient(source arma.MissionSource) (*s3.Client, error) {
	var optFns []func(*config.LoadOptions) error

	if source.Region != "" {
		optFns = append(optFns, config.WithRegion(source.Region))
	} else {
		optFns = append(optFns, config.WithRegion("us-east-1"))
	}

	if source.AccessKeyID != "" && source.SecretAccessKey != "" {
		optFns = append(optFns, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(source.AccessKeyID, source.SecretAccessKey, ""),
		))
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), optFns...)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	var s3Opts []func(*s3.Options)
	s3Opts = append(s3Opts, func(o *s3.Options) {
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	})
	if source.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(source.Endpoint)
			o.UsePathStyle = true
		})
	}

	return s3.NewFromConfig(cfg, s3Opts...), nil
}

func (d S3Driver) listPBOs(client *s3.Client, bucket, prefix string) (map[string]string, error) {
	result := make(map[string]string)
	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("listing S3 objects: %w", err)
		}
		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)
			if !strings.HasSuffix(key, ".pbo") {
				continue
			}
			name := strings.TrimPrefix(key, prefix)
			etag := strings.Trim(aws.ToString(obj.ETag), `"`)
			result[name] = etag
		}
	}
	return result, nil
}

func (d S3Driver) download(client *s3.Client, bucket, key, dst string) error {
	resp, err := client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("downloading s3://%s/%s: %w", bucket, key, err)
	}
	defer resp.Body.Close()

	tmp := dst + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, dst); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
