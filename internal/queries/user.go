package queries

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserQueries handles user-related database operations
type UserQueries struct {
	pool *pgxpool.Pool
}

// NewUserQueries creates a new UserQueries instance
func NewUserQueries(pool *pgxpool.Pool) *UserQueries {
	return &UserQueries{pool: pool}
}

// UserRecord represents a user record from database
type UserRecord struct {
	ID              string
	FirstName       string
	LastName        string
	Email           string
	Password        string
	AccountID       string
	Role            int
	EmailVerified   bool
	EmailVerifiedAt *time.Time
	OTPSecret       string
	OTPEnabled      bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// AccountRecord represents an account record from database
type AccountRecord struct {
	ID       string
	Name     string
	Timezone string
	Locale   string
}

// CreateUser creates a new user with account
func (q *UserQueries) CreateUser(ctx context.Context, email, password, firstName, lastName string) (*UserRecord, error) {
	tx, err := q.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create account first
	accountID := uuid.New().String()
	accountName := fmt.Sprintf("%s %s", firstName, lastName)
	
	_, err = tx.Exec(ctx, `
		INSERT INTO account (id, name, timezone, locale, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, accountID, accountName, "UTC", "en")
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	// Create user
	userID := uuid.New().String()
	_, err = tx.Exec(ctx, `
		INSERT INTO "user" (id, first_name, last_name, email, password, account_id, role, email_verified, otp_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
	`, userID, firstName, lastName, email, password, accountID, 1, false, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &UserRecord{
		ID:            userID,
		FirstName:     firstName,
		LastName:      lastName,
		Email:         email,
		Password:      password,
		AccountID:     accountID,
		Role:          1,
		EmailVerified: false,
		OTPEnabled:    false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}

// GetUserByEmail retrieves user by email
func (q *UserQueries) GetUserByEmail(ctx context.Context, email string) (*UserRecord, error) {
	var user UserRecord
	err := q.pool.QueryRow(ctx, `
		SELECT id, first_name, last_name, email, password, account_id, role, 
		       email_verified, email_verified_at, COALESCE((otp_secret->>'secret')::text, '') as otp_secret, 
		       otp_enabled, created_at, updated_at
		FROM "user"
		WHERE email = $1 AND archived_at IS NULL
	`, email).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Password,
		&user.AccountID, &user.Role, &user.EmailVerified, &user.EmailVerifiedAt,
		&user.OTPSecret, &user.OTPEnabled, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

// GetUserByID retrieves user by ID
func (q *UserQueries) GetUserByID(ctx context.Context, userID string) (*UserRecord, error) {
	var user UserRecord
	err := q.pool.QueryRow(ctx, `
		SELECT id, first_name, last_name, email, password, account_id, role, 
		       email_verified, email_verified_at, COALESCE((otp_secret->>'secret')::text, '') as otp_secret,
		       otp_enabled, created_at, updated_at
		FROM "user"
		WHERE id = $1 AND archived_at IS NULL
	`, userID).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Password,
		&user.AccountID, &user.Role, &user.EmailVerified, &user.EmailVerifiedAt,
		&user.OTPSecret, &user.OTPEnabled, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

// UpdatePassword updates user password
func (q *UserQueries) UpdatePassword(ctx context.Context, userID, hashedPassword string) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE "user"
		SET password = $1, updated_at = NOW()
		WHERE id = $2
	`, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

// MarkEmailAsVerified marks user email as verified
func (q *UserQueries) MarkEmailAsVerified(ctx context.Context, userID string) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE "user"
		SET email_verified = true, email_verified_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to mark email as verified: %w", err)
	}
	return nil
}

// CreatePasswordResetToken creates a password reset token
func (q *UserQueries) CreatePasswordResetToken(ctx context.Context, userID string, expiresIn time.Duration) (string, error) {
	token, err := generateSecureToken(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	expiresAt := time.Now().Add(expiresIn)
	_, err = q.pool.Exec(ctx, `
		INSERT INTO password_reset_token (user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`, userID, token, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create password reset token: %w", err)
	}

	return token, nil
}

// ValidatePasswordResetToken validates and returns userID for a password reset token
func (q *UserQueries) ValidatePasswordResetToken(ctx context.Context, token string) (string, error) {
	var userID string
	var expiresAt time.Time
	var usedAt *time.Time

	err := q.pool.QueryRow(ctx, `
		SELECT user_id, expires_at, used_at
		FROM password_reset_token
		WHERE token = $1
	`, token).Scan(&userID, &expiresAt, &usedAt)
	if err != nil {
		return "", fmt.Errorf("invalid or expired token: %w", err)
	}

	if usedAt != nil {
		return "", fmt.Errorf("token already used")
	}

	if time.Now().After(expiresAt) {
		return "", fmt.Errorf("token expired")
	}

	return userID, nil
}

// MarkPasswordResetTokenAsUsed marks token as used
func (q *UserQueries) MarkPasswordResetTokenAsUsed(ctx context.Context, token string) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE password_reset_token
		SET used_at = NOW()
		WHERE token = $1
	`, token)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}
	return nil
}

// CreateEmailVerificationToken creates an email verification token
func (q *UserQueries) CreateEmailVerificationToken(ctx context.Context, userID string, expiresIn time.Duration) (string, error) {
	token, err := generateSecureToken(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	expiresAt := time.Now().Add(expiresIn)
	_, err = q.pool.Exec(ctx, `
		INSERT INTO email_verification_token (user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`, userID, token, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create email verification token: %w", err)
	}

	return token, nil
}

// ValidateEmailVerificationToken validates and returns userID for an email verification token
func (q *UserQueries) ValidateEmailVerificationToken(ctx context.Context, token string) (string, error) {
	var userID string
	var expiresAt time.Time
	var usedAt *time.Time

	err := q.pool.QueryRow(ctx, `
		SELECT user_id, expires_at, used_at
		FROM email_verification_token
		WHERE token = $1
	`, token).Scan(&userID, &expiresAt, &usedAt)
	if err != nil {
		return "", fmt.Errorf("invalid or expired token: %w", err)
	}

	if usedAt != nil {
		return "", fmt.Errorf("token already used")
	}

	if time.Now().After(expiresAt) {
		return "", fmt.Errorf("token expired")
	}

	return userID, nil
}

// MarkEmailVerificationTokenAsUsed marks token as used
func (q *UserQueries) MarkEmailVerificationTokenAsUsed(ctx context.Context, token string) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE email_verification_token
		SET used_at = NOW()
		WHERE token = $1
	`, token)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}
	return nil
}

// Enable2FA enables 2FA for user
func (q *UserQueries) Enable2FA(ctx context.Context, userID, secret string) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE "user"
		SET otp_secret = jsonb_build_object('secret', $1), 
		    otp_enabled = true, 
		    updated_at = NOW()
		WHERE id = $2
	`, secret, userID)
	if err != nil {
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}
	return nil
}

// Disable2FA disables 2FA for user
func (q *UserQueries) Disable2FA(ctx context.Context, userID string) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE "user"
		SET otp_secret = '{}'::jsonb, 
		    otp_enabled = false, 
		    consumed_timestep = NULL,
		    updated_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}
	return nil
}

// UpdateLoginInfo updates user login information
func (q *UserQueries) UpdateLoginInfo(ctx context.Context, userID, ipAddress string) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE "user"
		SET last_sign_in_at = current_sign_in_at,
		    last_sign_in_ip = current_sign_in_ip,
		    current_sign_in_at = NOW(),
		    current_sign_in_ip = $2::inet,
		    sign_in_count = NOT sign_in_count,
		    updated_at = NOW()
		WHERE id = $1
	`, userID, ipAddress)
	if err != nil {
		return fmt.Errorf("failed to update login info: %w", err)
	}
	return nil
}

// CreateOrUpdateOAuthAccount creates or updates OAuth account link
func (q *UserQueries) CreateOrUpdateOAuthAccount(ctx context.Context, userID, provider, providerUserID, accessToken, refreshToken string, expiresAt *time.Time) error {
	_, err := q.pool.Exec(ctx, `
		INSERT INTO oauth_account (user_id, provider, provider_user_id, access_token, refresh_token, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (provider, provider_user_id)
		DO UPDATE SET
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			expires_at = EXCLUDED.expires_at,
			updated_at = NOW()
	`, userID, provider, providerUserID, accessToken, refreshToken, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create/update OAuth account: %w", err)
	}
	return nil
}

// GetUserByOAuthProvider retrieves user by OAuth provider and provider user ID
func (q *UserQueries) GetUserByOAuthProvider(ctx context.Context, provider, providerUserID string) (*UserRecord, error) {
	var user UserRecord
	err := q.pool.QueryRow(ctx, `
		SELECT u.id, u.first_name, u.last_name, u.email, u.password, u.account_id, u.role,
		       u.email_verified, u.email_verified_at, COALESCE((u.otp_secret->>'secret')::text, '') as otp_secret,
		       u.otp_enabled, u.created_at, u.updated_at
		FROM "user" u
		JOIN oauth_account oa ON oa.user_id = u.id
		WHERE oa.provider = $1 AND oa.provider_user_id = $2 AND u.archived_at IS NULL
	`, provider, providerUserID).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Password,
		&user.AccountID, &user.Role, &user.EmailVerified, &user.EmailVerifiedAt,
		&user.OTPSecret, &user.OTPEnabled, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by OAuth provider: %w", err)
	}
	return &user, nil
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

