package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/models"
)

// APIKeyRepository defines interface for API key storage operations
type APIKeyRepository interface {
	GetByKeyHash(keyHash string) (*models.APIKey, error)
	UpdateLastUsed(keyID string, lastUsed time.Time) error
	Create(apiKey *models.APIKey) error
	Update(apiKey *models.APIKey) error
	Delete(keyID string) error
	ListByAccount(accountID string) ([]*models.APIKey, error)
}

// APIKeyService handles API key operations
type APIKeyService struct {
	repo APIKeyRepository
}

// NewAPIKeyService creates new API key service
func NewAPIKeyService(repo APIKeyRepository) *APIKeyService {
	return &APIKeyService{repo: repo}
}

// ValidateAPIKey validates API key and returns key model
func (s *APIKeyService) ValidateAPIKey(keyHash string) (*models.APIKey, error) {
	return s.repo.GetByKeyHash(keyHash)
}

// UpdateLastUsed updates last used timestamp for API key
func (s *APIKeyService) UpdateLastUsed(keyID string) error {
	return s.repo.UpdateLastUsed(keyID, time.Now())
}

// GenerateKey generates new random API key (32 bytes = 43 chars base64)
func GenerateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CreateAPIKey creates new API key for account
func (s *APIKeyService) CreateAPIKey(accountID, name string, expiresAt *time.Time) (string, *models.APIKey, error) {
	// Generate random key
	key, err := GenerateAPIKey()
	if err != nil {
		return "", nil, err
	}

	// Hash the key for storage
	keyHash := middleware.HashAPIKey(key)

	apiKey := &models.APIKey{
		Name:      name,
		KeyHash:   keyHash,
		AccountID: accountID,
		Enabled:   true,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(apiKey); err != nil {
		return "", nil, err
	}

	// Return plain key only once (won't be retrievable later)
	return key, apiKey, nil
}

// EnableKey enables API key
func (s *APIKeyService) EnableKey(keyID string) error {
	apiKey, err := s.repo.GetByKeyHash(keyID)
	if err != nil {
		return err
	}
	if apiKey == nil {
		return errors.New("API key not found")
	}

	apiKey.Enabled = true
	apiKey.UpdatedAt = time.Now()
	return s.repo.Update(apiKey)
}

// DisableKey disables API key
func (s *APIKeyService) DisableKey(keyID string) error {
	apiKey, err := s.repo.GetByKeyHash(keyID)
	if err != nil {
		return err
	}
	if apiKey == nil {
		return errors.New("API key not found")
	}

	apiKey.Enabled = false
	apiKey.UpdatedAt = time.Now()
	return s.repo.Update(apiKey)
}

// DeleteKey deletes API key
func (s *APIKeyService) DeleteKey(keyID string) error {
	return s.repo.Delete(keyID)
}

// ListAccountKeys lists all API keys for account
func (s *APIKeyService) ListAccountKeys(accountID string) ([]*models.APIKey, error) {
	return s.repo.ListByAccount(accountID)
}

