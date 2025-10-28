package queries

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/shurco/gosign/internal/models"
)

// NotificationRepository manages notifications in the database
type NotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create creates a new notification
func (r *NotificationRepository) Create(notification *models.Notification) error {
	query := `
		INSERT INTO notification (
			id, type, recipient, template, context, 
			status, scheduled_at, related_id, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)`

	contextJSON, err := json.Marshal(notification.Context)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(query,
		notification.ID,
		notification.Type,
		notification.Recipient,
		notification.Template,
		contextJSON,
		notification.Status,
		notification.ScheduledAt,
		notification.RelatedID,
		time.Now(),
	)
	return err
}

// GetScheduledReady retrieves all notifications ready to be sent
func (r *NotificationRepository) GetScheduledReady() ([]*models.Notification, error) {
	query := `
		SELECT id, type, recipient, template, context, status, 
		       scheduled_at, sent_at, related_id, created_at
		FROM notification
		WHERE status = $1 
		  AND scheduled_at IS NOT NULL 
		  AND scheduled_at <= $2
		ORDER BY scheduled_at ASC
		LIMIT 100`

	rows, err := r.db.Query(query, models.NotificationStatusPending, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*models.Notification
	for rows.Next() {
		var n models.Notification
		var contextJSON []byte

		err := rows.Scan(
			&n.ID,
			&n.Type,
			&n.Recipient,
			&n.Template,
			&contextJSON,
			&n.Status,
			&n.ScheduledAt,
			&n.SentAt,
			&n.RelatedID,
			&n.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(contextJSON) > 0 {
			if err := json.Unmarshal(contextJSON, &n.Context); err != nil {
				return nil, err
			}
		}

		notifications = append(notifications, &n)
	}

	return notifications, rows.Err()
}

// UpdateStatus updates notification status
func (r *NotificationRepository) UpdateStatus(id string, status models.NotificationStatus) error {
	query := `
		UPDATE notification 
		SET status = $1, sent_at = $2
		WHERE id = $3`

	var sentAt *time.Time
	if status == models.NotificationStatusSent || status == models.NotificationStatusFailed {
		now := time.Now()
		sentAt = &now
	}

	_, err := r.db.Exec(query, status, sentAt, id)
	return err
}

// CancelByRelatedID cancels all pending notifications for related_id
func (r *NotificationRepository) CancelByRelatedID(relatedID string) error {
	query := `
		UPDATE notification
		SET status = $1
		WHERE related_id = $2 
		  AND status = $3`

	_, err := r.db.Exec(query, models.NotificationStatusCancelled, relatedID, models.NotificationStatusPending)
	return err
}

// GetByID retrieves notification by ID
func (r *NotificationRepository) GetByID(id string) (*models.Notification, error) {
	query := `
		SELECT id, type, recipient, template, context, status, 
		       scheduled_at, sent_at, related_id, created_at
		FROM notification 
		WHERE id = $1`

	var n models.Notification
	var contextJSON []byte

	err := r.db.QueryRow(query, id).Scan(
		&n.ID,
		&n.Type,
		&n.Recipient,
		&n.Template,
		&contextJSON,
		&n.Status,
		&n.ScheduledAt,
		&n.SentAt,
		&n.RelatedID,
		&n.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(contextJSON) > 0 {
		if err := json.Unmarshal(contextJSON, &n.Context); err != nil {
			return nil, err
		}
	}

	return &n, nil
}

// List retrieves a filtered list of notifications
func (r *NotificationRepository) List(limit, offset int, status *models.NotificationStatus) ([]*models.Notification, error) {
	query := `
		SELECT id, type, recipient, template, context, status, 
		       scheduled_at, sent_at, related_id, created_at
		FROM notification
		WHERE ($1::text IS NULL OR status = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*models.Notification
	for rows.Next() {
		var n models.Notification
		var contextJSON []byte

		err := rows.Scan(
			&n.ID,
			&n.Type,
			&n.Recipient,
			&n.Template,
			&contextJSON,
			&n.Status,
			&n.ScheduledAt,
			&n.SentAt,
			&n.RelatedID,
			&n.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(contextJSON) > 0 {
			if err := json.Unmarshal(contextJSON, &n.Context); err != nil {
				return nil, err
			}
		}

		notifications = append(notifications, &n)
	}

	return notifications, rows.Err()
}

