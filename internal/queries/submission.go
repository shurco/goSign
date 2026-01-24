package queries

import (
	"context"
	"encoding/json"

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

// CreateEvent inserts an event into the database with IP address
func (r *SubmissionRepository) CreateEvent(ctx context.Context, event *models.Event) error {
	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO event (id, type, actor_id, resource_type, resource_id, metadata_json, ip, created_at)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6::jsonb, NULLIF($7, '')::inet, $8)
	`, event.ID, event.Type, event.ActorID, event.ResourceType, event.ResourceID, string(metadataJSON), event.IP, event.CreatedAt)

	return err
}

// CreateSubmission is a stub implementation (not used in current flow)
func (r *SubmissionRepository) CreateSubmission(ctx context.Context, submission *models.Submission) error {
	return nil
}

// GetSubmission is a stub implementation (not used in current flow)
func (r *SubmissionRepository) GetSubmission(ctx context.Context, id string) (*models.Submission, error) {
	return nil, nil
}

// UpdateSubmissionState is a stub implementation (not used in current flow)
func (r *SubmissionRepository) UpdateSubmissionState(ctx context.Context, id string, state submission.SubmissionState) error {
	return nil
}

// CreateSubmitter is a stub implementation (not used in current flow)
func (r *SubmissionRepository) CreateSubmitter(ctx context.Context, submitter *models.Submitter) error {
	return nil
}

// GetSubmitters is a stub implementation (not used in current flow)
func (r *SubmissionRepository) GetSubmitters(ctx context.Context, submissionID string) ([]*models.Submitter, error) {
	return nil, nil
}

// GetSubmittersByOrder is a stub implementation (not used in current flow)
func (r *SubmissionRepository) GetSubmittersByOrder(ctx context.Context, submissionID string, order int) ([]*models.Submitter, error) {
	return nil, nil
}

// GetSubmitter is a stub implementation (not used in current flow)
func (r *SubmissionRepository) GetSubmitter(ctx context.Context, id string) (*models.Submitter, error) {
	return nil, nil
}

// UpdateSubmitterStatus is a stub implementation (not used in current flow)
func (r *SubmissionRepository) UpdateSubmitterStatus(ctx context.Context, id string, status models.SubmitterStatus) error {
	return nil
}
