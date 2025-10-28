package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shurco/gosign/pkg/utils/fsutil"
)

// LocalStorage implements storage on local file system
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local storage
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base path: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
	}, nil
}

// getFullPath returns the full path to file
func (s *LocalStorage) getFullPath(key string) string {
	// Clean key from potentially dangerous characters
	cleanKey := filepath.Clean(key)
	cleanKey = strings.TrimPrefix(cleanKey, "/")
	return filepath.Join(s.basePath, cleanKey)
}

// Upload uploads a file
func (s *LocalStorage) Upload(ctx context.Context, key string, reader io.Reader, metadata *BlobMetadata) error {
	fullPath := s.getFullPath(key)

	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data
	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Download downloads a file
func (s *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := s.getFullPath(key)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// Delete deletes a file
func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	fullPath := s.getFullPath(key)

	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetURL returns local path to file
func (s *LocalStorage) GetURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	fullPath := s.getFullPath(key)

	// Check file exists
	if !fsutil.IsFile(fullPath) {
		return "", fmt.Errorf("file not found: %s", key)
	}

	// For local storage return relative path
	return fmt.Sprintf("/uploads/%s", key), nil
}

// List returns list of files with prefix
func (s *LocalStorage) List(ctx context.Context, prefix string) ([]string, error) {
	prefixPath := s.getFullPath(prefix)
	var files []string

	err := filepath.Walk(prefixPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // ignore read errors
		}

		if !info.IsDir() {
			// Get relative path from basePath
			relPath, err := filepath.Rel(s.basePath, path)
			if err != nil {
				return nil
			}
			files = append(files, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}

// Exists checks if file exists
func (s *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	fullPath := s.getFullPath(key)
	return fsutil.IsFile(fullPath), nil
}

// GetMetadata returns file metadata
func (s *LocalStorage) GetMetadata(ctx context.Context, key string) (*BlobMetadata, error) {
	fullPath := s.getFullPath(key)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &BlobMetadata{
		Size:     info.Size(),
		Modified: info.ModTime(),
	}, nil
}

