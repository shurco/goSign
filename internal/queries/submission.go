package queries

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/services/submission"
)

// SubmissionRepository implements submission.Repository interface
type SubmissionRepository struct {
	pool *pgxpool.Pool
}

// NewSubmissionRepository creates a new submission repository
func NewSubmissionRepository(pool *pgxpool.Pool) *SubmissionRepository {
	return &SubmissionRepository{pool: pool}
}

const submitterSelectSQL = `
	SELECT
		s.id,
		s.submission_id,
		COALESCE(s.name, ''),
		COALESCE(s.email, ''),
		COALESCE(s.phone, ''),
		s.slug,
		COALESCE(s.status, 'pending'),
		s.sented_at,
		s.opened_at,
		s.completed_at,
		s.declined_at,
		COALESCE(s.metadata, '{}'::jsonb),
		s.created_at,
		s.updated_at
	FROM submitter s`

func scanSubmitterRow(row pgx.Row) (*models.Submitter, error) {
	var (
		submitter   models.Submitter
		status      string
		metadataRaw []byte
	)
	if err := row.Scan(
		&submitter.ID,
		&submitter.SubmissionID,
		&submitter.Name,
		&submitter.Email,
		&submitter.Phone,
		&submitter.Slug,
		&status,
		&submitter.SentAt,
		&submitter.OpenedAt,
		&submitter.CompletedAt,
		&submitter.DeclinedAt,
		&metadataRaw,
		&submitter.CreatedAt,
		&submitter.UpdatedAt,
	); err != nil {
		return nil, err
	}

	submitter.Status = models.SubmitterStatus(status)
	if len(metadataRaw) > 0 {
		if err := json.Unmarshal(metadataRaw, &submitter.Metadata); err != nil {
			return nil, fmt.Errorf("queries: unmarshal submitter metadata: %w", err)
		}
	}
	submitter.Order = submitterOrderFromMetadata(submitter.Metadata)
	return &submitter, nil
}

func submitterOrderFromMetadata(metadata map[string]any) int {
	if metadata == nil {
		return 0
	}
	switch v := metadata["order"].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	default:
		return 0
	}
}

// CreateEvent inserts an event into the database with IP address
func (r *SubmissionRepository) CreateEvent(ctx context.Context, event *models.Event) error {
	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("queries: create event: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO event (id, type, actor_id, resource_type, resource_id, metadata_json, ip, created_at)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6::jsonb, NULLIF($7, '')::inet, $8)
	`, event.ID, event.Type, event.ActorID, event.ResourceType, event.ResourceID, string(metadataJSON), event.IP, event.CreatedAt)
	if err != nil {
		return fmt.Errorf("queries: create event: %w", err)
	}

	return nil
}

// CreateSubmission inserts a new submission row
func (r *SubmissionRepository) CreateSubmission(ctx context.Context, sub *models.Submission) error {
	signingMode := sub.SigningMode
	if signingMode == "" {
		signingMode = models.SigningModeSequential
	}

	preferencesJSON, err := json.Marshal(map[string]string{
		"signing_mode": string(signingMode),
	})
	if err != nil {
		return fmt.Errorf("queries: create submission: %w", err)
	}

	slug := uuid.New().String()

	_, err = r.pool.Exec(ctx, `
		INSERT INTO submission (
			id, template_id, created_by_user_id, slug, source, submitters_order,
			locale, preferences, created_at, updated_at
		)
		VALUES (
			$1, $2, NULLIF($3, '')::uuid, $4, 'api', '0',
			NULLIF($5, ''), $6::jsonb, $7, $8
		)
	`, sub.ID, sub.TemplateID, sub.CreatedByID, slug, sub.Locale, string(preferencesJSON), sub.CreatedAt, sub.UpdatedAt)
	if err != nil {
		return fmt.Errorf("queries: create submission: %w", err)
	}

	return nil
}

// GetSubmission returns a submission with derived status and signing mode.
// AccountID is resolved through the creating user or the template folder.
func (r *SubmissionRepository) GetSubmission(ctx context.Context, id string) (*models.Submission, error) {
	var (
		sub         models.Submission
		status      string
		signingMode string
	)
	err := r.pool.QueryRow(ctx, `
		SELECT
			sub.id,
			sub.template_id,
			COALESCE(sub.created_by_user_id::text, ''),
			COALESCE(u.account_id::text, tf.account_id::text, ''),
			COALESCE(sub.locale, ''),
			COALESCE(sub.preferences->>'signing_mode', 'sequential'),
			sub.created_at,
			sub.updated_at,
			CASE
				WHEN sub.archived_at IS NOT NULL THEN COALESCE(sub.preferences->>'state', 'cancelled')
				WHEN count(s.id) = 0 THEN 'draft'
				WHEN bool_and(COALESCE(s.status, 'pending') = 'completed') THEN 'completed'
				WHEN bool_or(COALESCE(s.status, 'pending') = 'declined') THEN 'cancelled'
				WHEN bool_or(COALESCE(s.status, 'pending') = 'opened')
					OR sum(CASE WHEN COALESCE(s.status, 'pending') = 'completed' THEN 1 ELSE 0 END) > 0
					THEN 'in_progress'
				ELSE 'pending'
			END AS status
		FROM submission sub
		LEFT JOIN "user" u ON u.id = sub.created_by_user_id
		LEFT JOIN template t ON t.id = sub.template_id
		LEFT JOIN template_folder tf ON tf.id = t.folder_id
		LEFT JOIN submitter s ON s.submission_id = sub.id
		WHERE sub.id = $1
		GROUP BY sub.id, sub.template_id, sub.created_by_user_id, u.account_id, tf.account_id,
			sub.locale, sub.preferences, sub.created_at, sub.updated_at, sub.archived_at
	`, id).Scan(
		&sub.ID,
		&sub.TemplateID,
		&sub.CreatedByID,
		&sub.AccountID,
		&sub.Locale,
		&signingMode,
		&sub.CreatedAt,
		&sub.UpdatedAt,
		&status,
	)
	if err != nil {
		return nil, fmt.Errorf("queries: get submission: %w", err)
	}

	sub.Status = models.SubmissionStatus(status)
	sub.SigningMode = models.SigningMode(signingMode)
	return &sub, nil
}

// UpdateSubmissionState stores state in preferences and archives when terminal
func (r *SubmissionRepository) UpdateSubmissionState(ctx context.Context, id string, state submission.SubmissionState) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE submission
		SET
			preferences = preferences || jsonb_build_object('state', $2::text),
			archived_at = CASE
				WHEN $2::text IN ('expired', 'cancelled') THEN NOW()
				ELSE archived_at
			END,
			updated_at = NOW()
		WHERE id = $1
	`, id, string(state))
	if err != nil {
		return fmt.Errorf("queries: update submission state: %w", err)
	}

	return nil
}

// CreateSubmitter inserts a submitter row
func (r *SubmissionRepository) CreateSubmitter(ctx context.Context, submitter *models.Submitter) error {
	meta := submitter.Metadata
	if meta == nil {
		meta = make(map[string]any)
	}
	meta["order"] = submitter.Order

	metadataJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("queries: create submitter: %w", err)
	}

	status := submitter.Status
	if status == "" {
		status = models.SubmitterStatusPending
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO submitter (
			id, submission_id, name, email, phone, slug, status, metadata, created_at, updated_at
		)
		VALUES (
			$1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''),
			$6, $7, $8::jsonb, $9, $10
		)
	`, submitter.ID, submitter.SubmissionID, submitter.Name, submitter.Email, submitter.Phone,
		submitter.Slug, string(status), string(metadataJSON), submitter.CreatedAt, submitter.UpdatedAt)
	if err != nil {
		return fmt.Errorf("queries: create submitter: %w", err)
	}

	return nil
}

// GetSubmitters returns submitters for a submission ordered by signing order
func (r *SubmissionRepository) GetSubmitters(ctx context.Context, submissionID string) ([]*models.Submitter, error) {
	rows, err := r.pool.Query(ctx, submitterSelectSQL+`
		WHERE s.submission_id = $1
		ORDER BY (s.metadata->>'order')::int NULLS LAST, s.created_at
	`, submissionID)
	if err != nil {
		return nil, fmt.Errorf("queries: get submitters: %w", err)
	}
	defer rows.Close()

	return collectSubmitters(rows, "get submitters")
}

// GetSubmittersByOrder returns submitters matching a signing order
func (r *SubmissionRepository) GetSubmittersByOrder(ctx context.Context, submissionID string, order int) ([]*models.Submitter, error) {
	rows, err := r.pool.Query(ctx, submitterSelectSQL+`
		WHERE s.submission_id = $1
		  AND COALESCE((s.metadata->>'order')::int, 0) = $2
		ORDER BY s.created_at
	`, submissionID, order)
	if err != nil {
		return nil, fmt.Errorf("queries: get submitters by order: %w", err)
	}
	defer rows.Close()

	return collectSubmitters(rows, "get submitters by order")
}

// GetSubmitter returns a submitter by id
func (r *SubmissionRepository) GetSubmitter(ctx context.Context, id string) (*models.Submitter, error) {
	row := r.pool.QueryRow(ctx, submitterSelectSQL+` WHERE s.id = $1`, id)
	submitter, err := scanSubmitterRow(row)
	if err != nil {
		return nil, fmt.Errorf("queries: get submitter: %w", err)
	}
	return submitter, nil
}

// GetSubmitterBySlug returns a submitter by signing slug
func (r *SubmissionRepository) GetSubmitterBySlug(ctx context.Context, slug string) (*models.Submitter, error) {
	row := r.pool.QueryRow(ctx, submitterSelectSQL+` WHERE s.slug = $1`, slug)
	submitter, err := scanSubmitterRow(row)
	if err != nil {
		return nil, fmt.Errorf("queries: get submitter by slug: %w", err)
	}
	return submitter, nil
}

// UpdateSubmitterStatus updates submitter status and related timestamps
func (r *SubmissionRepository) UpdateSubmitterStatus(ctx context.Context, id string, status models.SubmitterStatus) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE submitter
		SET
			status = $2,
			completed_at = CASE WHEN $2 = 'completed' THEN NOW() ELSE completed_at END,
			declined_at = CASE WHEN $2 = 'declined' THEN NOW() ELSE declined_at END,
			opened_at = CASE
				WHEN $2 = 'opened' THEN COALESCE(opened_at, NOW())
				ELSE opened_at
			END,
			sented_at = CASE
				WHEN $2 = 'opened' THEN COALESCE(sented_at, NOW())
				ELSE sented_at
			END,
			updated_at = NOW()
		WHERE id = $1
	`, id, string(status))
	if err != nil {
		return fmt.Errorf("queries: update submitter status: %w", err)
	}

	return nil
}

func collectSubmitters(rows pgx.Rows, op string) ([]*models.Submitter, error) {
	var submitters []*models.Submitter
	for rows.Next() {
		submitter, err := scanSubmitterRow(rows)
		if err != nil {
			return nil, fmt.Errorf("queries: %s: %w", op, err)
		}
		submitters = append(submitters, submitter)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("queries: %s: %w", op, err)
	}
	return submitters, nil
}
