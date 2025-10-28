package storage

import (
	"context"
	"io"
	"time"
)

// BlobMetadata contains blob object metadata
type BlobMetadata struct {
	Size        int64
	ContentType string
	Modified    time.Time
	ETag        string
	Custom      map[string]string
}

// BlobStorage represents interface for working with storage
type BlobStorage interface {
	// Upload uploads file to storage
	Upload(ctx context.Context, key string, reader io.Reader, metadata *BlobMetadata) error

	// Download downloads file from storage
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete deletes file from storage
	Delete(ctx context.Context, key string) error

	// GetURL returns URL for file access
	GetURL(ctx context.Context, key string, expiration time.Duration) (string, error)

	// List returns list of files with specified prefix
	List(ctx context.Context, prefix string) ([]string, error)

	// Exists checks if file exists
	Exists(ctx context.Context, key string) (bool, error)

	// GetMetadata returns file metadata
	GetMetadata(ctx context.Context, key string) (*BlobMetadata, error)
}

// Config contains storage configuration
type Config struct {
	Provider string            // local, s3, gcs, azure
	Bucket   string            // bucket/container name
	Region   string            // region (for cloud storage)
	Endpoint string            // custom endpoint (for MinIO, etc.)
	BasePath string            // base path (for local)
	Options  map[string]string // additional options
}

