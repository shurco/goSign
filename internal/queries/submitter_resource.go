package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shurco/gosign/internal/models"
)

// SubmitterResourceRepository implements api.ResourceRepository for submitters
type SubmitterResourceRepository struct {
	pool *pgxpool.Pool
}

// NewSubmitterResourceRepository creates a new submitter resource repository
func NewSubmitterResourceRepository(pool *pgxpool.Pool) *SubmitterResourceRepository {
	return &SubmitterResourceRepository{pool: pool}
}

// List returns paginated submitters with optional filters
func (r *SubmitterResourceRepository) List(page, pageSize int, filters map[string]string) ([]models.Submitter, int, error) {
	ctx := context.Background()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var (
		conds  []string
		args   []any
		argPos = 1
	)

	if v := filters["submission_id"]; v != "" {
		conds = append(conds, fmt.Sprintf("s.submission_id = $%d", argPos))
		args = append(args, v)
		argPos++
	}
	if v := filters["status"]; v != "" {
		conds = append(conds, fmt.Sprintf("s.status = $%d", argPos))
		args = append(args, v)
		argPos++
	}
	if v := filters["email"]; v != "" {
		conds = append(conds, fmt.Sprintf("s.email = $%d", argPos))
		args = append(args, v)
		argPos++
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	query := submitterSelectSQL + `,
		COUNT(*) OVER() AS total_count
	` + where + `
		ORDER BY s.created_at DESC
		LIMIT $` + fmt.Sprint(argPos) + ` OFFSET $` + fmt.Sprint(argPos+1)

	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("queries: list submitters: %w", err)
	}
	defer rows.Close()

	var (
		items []models.Submitter
		total int
	)
	for rows.Next() {
		submitter, rowTotal, err := scanSubmitterRowWithTotal(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("queries: list submitters: %w", err)
		}
		total = rowTotal
		items = append(items, *submitter)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("queries: list submitters: %w", err)
	}

	return items, total, nil
}

func scanSubmitterRowWithTotal(row pgx.Row) (*models.Submitter, int, error) {
	var (
		submitter   models.Submitter
		status      string
		metadataRaw []byte
		total       int
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
		&total,
	); err != nil {
		return nil, 0, err
	}

	submitter.Status = models.SubmitterStatus(status)
	if len(metadataRaw) > 0 {
		if err := json.Unmarshal(metadataRaw, &submitter.Metadata); err != nil {
			return nil, 0, fmt.Errorf("queries: unmarshal submitter metadata: %w", err)
		}
	}
	submitter.Order = submitterOrderFromMetadata(submitter.Metadata)
	return &submitter, total, nil
}

// Get returns a submitter by id
func (r *SubmitterResourceRepository) Get(id string) (*models.Submitter, error) {
	ctx := context.Background()
	row := r.pool.QueryRow(ctx, submitterSelectSQL+` WHERE s.id = $1`, id)
	submitter, err := scanSubmitterRow(row)
	if err != nil {
		return nil, fmt.Errorf("queries: get submitter resource: %w", err)
	}
	return submitter, nil
}

// Create inserts a new submitter
func (r *SubmitterResourceRepository) Create(item *models.Submitter) error {
	return NewSubmissionRepository(r.pool).CreateSubmitter(context.Background(), item)
}

// Update updates submitter fields
func (r *SubmitterResourceRepository) Update(id string, item *models.Submitter) error {
	ctx := context.Background()

	meta := item.Metadata
	if meta == nil {
		meta = make(map[string]any)
	}
	if item.Order != 0 {
		meta["order"] = item.Order
	}

	metadataJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("queries: update submitter resource: %w", err)
	}

	status := item.Status
	if status == "" {
		status = models.SubmitterStatusPending
	}

	tag, err := r.pool.Exec(ctx, `
		UPDATE submitter
		SET
			name = NULLIF($2, ''),
			email = NULLIF($3, ''),
			phone = NULLIF($4, ''),
			status = $5,
			metadata = $6::jsonb,
			updated_at = NOW()
		WHERE id = $1
	`, id, item.Name, item.Email, item.Phone, string(status), string(metadataJSON))
	if err != nil {
		return fmt.Errorf("queries: update submitter resource: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// Delete removes a submitter by id
func (r *SubmitterResourceRepository) Delete(id string) error {
	ctx := context.Background()

	tag, err := r.pool.Exec(ctx, `DELETE FROM submitter WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("queries: delete submitter resource: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
