package storage

import (
	"context"
	"fmt"

	"github.com/shurco/gosign/pkg/appdir"
)

// NewStorage creates storage according to configuration
func NewStorage(ctx context.Context, cfg Config) (BlobStorage, error) {
	switch cfg.Provider {
	case "local":
		// Local storage path is next to the executable, not configurable
		return NewLocalStorage(appdir.LcUploads())

	case "s3":
		s3cfg := S3Config{
			AccessKeyID:     cfg.Options["access_key_id"],
			SecretAccessKey: cfg.Options["secret_access_key"],
			Region:          cfg.Region,
			Bucket:          cfg.Bucket,
			Endpoint:        cfg.Endpoint,
		}
		return NewS3Storage(ctx, s3cfg)

	// case "gcs":
	// 	// TODO: implement GCS storage
	// 	return nil, fmt.Errorf("GCS storage not implemented yet")

	// case "azure":
	// 	// TODO: implement Azure storage
	// 	return nil, fmt.Errorf("Azure storage not implemented yet")

	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", cfg.Provider)
	}
}

