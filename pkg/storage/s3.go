package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Storage implements AWS S3 storage
type S3Storage struct {
	client *minio.Client
	bucket string
}

// S3Config contains S3 configuration
type S3Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Bucket          string
	Endpoint        string // for MinIO and other S3-compatible storages
	UseSSL          bool   // use SSL/TLS for connection
}

// NewS3Storage creates new S3 storage
func NewS3Storage(ctx context.Context, cfg S3Config) (*S3Storage, error) {
	// Determine endpoint
	endpoint := cfg.Endpoint
	if endpoint == "" {
		// Default AWS S3 endpoint based on region
		if cfg.Region == "" {
			cfg.Region = "us-east-1"
		}
		endpoint = fmt.Sprintf("s3.%s.amazonaws.com", cfg.Region)
	}

	// Determine SSL usage - default to true for AWS S3, false for custom endpoints (like MinIO)
	useSSL := cfg.UseSSL
	if !cfg.UseSSL && cfg.Endpoint == "" {
		// Default to SSL for AWS S3 when endpoint is not specified
		useSSL = true
	}

	// Create MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: useSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	return &S3Storage{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// Upload uploads file to S3
func (s *S3Storage) Upload(ctx context.Context, key string, reader io.Reader, metadata *BlobMetadata) error {
	opts := minio.PutObjectOptions{}

	if metadata != nil && metadata.ContentType != "" {
		opts.ContentType = metadata.ContentType
	}

	_, err := s.client.PutObject(ctx, s.bucket, key, reader, -1, opts)
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

// Download downloads file from S3
func (s *S3Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	// MinIO GetObject returns object immediately, errors occur on first read
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}

	return obj, nil
}

// Delete deletes file from S3
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

// GetURL returns presigned URL for file access
func (s *S3Storage) GetURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucket, key, expiration, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// List returns list of files with prefix
func (s *S3Storage) List(ctx context.Context, prefix string) ([]string, error) {
	var files []string

	objectCh := s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for obj := range objectCh {
		if obj.Err != nil {
			return nil, fmt.Errorf("failed to list S3 objects: %w", obj.Err)
		}
		files = append(files, obj.Key)
	}

	return files, nil
}

// Exists checks if file exists
func (s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		// Check if error is "not found"
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" || errResp.Code == "NotFound" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check S3 object: %w", err)
	}

	return true, nil
}

// GetMetadata returns file metadata
func (s *S3Storage) GetMetadata(ctx context.Context, key string) (*BlobMetadata, error) {
	objInfo, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 metadata: %w", err)
	}

	metadata := &BlobMetadata{
		Size:        objInfo.Size,
		ContentType: objInfo.ContentType,
		Modified:    objInfo.LastModified,
		ETag:        objInfo.ETag,
	}

	return metadata, nil
}

