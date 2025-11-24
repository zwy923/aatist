package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ObjectStorage defines upload capabilities needed by services.
type ObjectStorage interface {
	Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string, metadata map[string]string) error
	Delete(ctx context.Context, objectName string) error
	BuildPublicURL(objectName string) string
	PresignedPutURL(ctx context.Context, objectName string, expires time.Duration) (string, error)
}

// S3 encapsulates an S3/MinIO client.
type S3 struct {
	client  *minio.Client
	bucket  string
	baseURL string
}

// S3Config holds configuration for S3.
type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
	PublicURL string
	Region    string
}

// NewS3 creates a new S3 client and ensures the bucket exists.
func NewS3(cfg S3Config) (*S3, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("s3 endpoint is required")
	}
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("s3 bucket is required")
	}

	endpoint, overrideSSL := sanitizeEndpoint(cfg.Endpoint)
	if overrideSSL != nil {
		cfg.UseSSL = *overrideSSL
	}

	opts := &minio.Options{
		Secure: cfg.UseSSL,
	}
	if cfg.AccessKey != "" || cfg.SecretKey != "" {
		opts.Creds = credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, "")
	}
	if cfg.Region != "" {
		opts.Region = cfg.Region
	}

	client, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize s3 client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{Region: cfg.Region}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	baseURL := cfg.PublicURL
	if baseURL == "" {
		scheme := "https"
		if !cfg.UseSSL {
			scheme = "http"
		}
		baseURL = fmt.Sprintf("%s://%s/%s", scheme, endpoint, cfg.Bucket)
	} else {
		baseURL = strings.TrimRight(baseURL, "/")
	}

	return &S3{
		client:  client,
		bucket:  cfg.Bucket,
		baseURL: baseURL,
	}, nil
}

// Upload writes an object to the configured bucket and returns its public URL.
func (s *S3) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string, metadata map[string]string) error {
	if objectName == "" {
		return fmt.Errorf("objectName is required")
	}

	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}
	if metadata != nil {
		opts.UserMetadata = metadata
	}

	_, err := s.client.PutObject(ctx, s.bucket, objectName, reader, size, opts)
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}

	return nil
}

func sanitizeEndpoint(endpoint string) (string, *bool) {
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		u, err := url.Parse(endpoint)
		if err == nil && u.Host != "" {
			useSSL := u.Scheme == "https"
			return u.Host, &useSSL
		}
	}
	return endpoint, nil
}

// Delete removes an object from the bucket.
func (s *S3) Delete(ctx context.Context, objectName string) error {
	if objectName == "" {
		return fmt.Errorf("objectName is required")
	}

	err := s.client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// PresignedPutURL generates a pre-signed URL for PUT operations (for direct client uploads).
func (s *S3) PresignedPutURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	if objectName == "" {
		return "", fmt.Errorf("objectName is required")
	}

	url, err := s.client.PresignedPutObject(ctx, s.bucket, objectName, expires)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// BuildPublicURL constructs the public URL for a given object.
func (s *S3) BuildPublicURL(objectName string) string {
	object := strings.TrimLeft(objectName, "/")
	return fmt.Sprintf("%s/%s", s.baseURL, object)
}

// BaseURL exposes the configured base URL for public objects.
func (s *S3) BaseURL() string {
	return s.baseURL
}
