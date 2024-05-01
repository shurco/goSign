package queries

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shurco/gosign/internal/models"
)

// TrustQueries is ...
type TrustQueries struct {
	*pgxpool.Pool
}

// CheckAKI is ...
func (q *TrustQueries) CheckAKI(ctx context.Context, aki string) (*models.TrustCert, error) {
	if aki == "" {
		return nil, nil
	}

	trustCert := &models.TrustCert{}
	query := `
		SELECT
			"list",
			"name",
			"aki",
			"ski"
		FROM
			"trust_list"
		WHERE
			"ski" = $1
		LIMIT
			1
	`
	err := q.QueryRow(ctx, query,
		aki,
	).Scan(
		&trustCert.List,
		&trustCert.Name,
		&trustCert.AKI,
		&trustCert.SKI,
	)
	if err != nil {
		return nil, err
	}

	return trustCert, nil
}

// AddAdobeTL is ...
func (q *TrustQueries) AddAdobeTL(ctx context.Context, trust models.TrustCerts) error {
	var certsToInsert []models.TrustCert

	for _, v := range trust.Certs {
		var aki string
		query := `
			SELECT
				"aki"
			FROM
				"trust_list"
			WHERE
				"aki" = $1
			LIMIT
				1
		`
		err := q.QueryRow(ctx, query, v.AKI).Scan(&aki)
		if err != nil && err != pgx.ErrNoRows {
			return err
		}
		if err == pgx.ErrNoRows {
			certsToInsert = append(certsToInsert, v)
		}
	}

	if len(certsToInsert) > 0 {
		batch := &pgx.Batch{}
		insertQuery := `
			INSERT INTO
				"trust_list" ("list", "name", "aki", "ski")
			VALUES
				($1, $2, $3, $4)
		`
		for _, cert := range certsToInsert {
			batch.Queue(insertQuery, cert.List, cert.Name, cert.AKI, cert.SKI)
		}

		br := q.SendBatch(ctx, batch)
		defer br.Close()

		for range certsToInsert {
			_, err := br.Exec()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// TimeUpdateAdobeTL is ...
func (q *TrustQueries) TimeUpdateAdobeTL(ctx context.Context) (*time.Time, error) {
	var date sql.NullTime
	query := `
		SELECT
			MIN("created_at") AS "earliest_record"
		FROM
			"trust_list"
	`
	if err := q.QueryRow(ctx, query).Scan(&date); err != nil {
		return nil, err
	}

	if date.Valid {
		return &date.Time, nil
	}

	return nil, nil
}

// DeleteAdobeTL is ...
func (q *TrustQueries) DeleteAdobeTL(ctx context.Context, list string) error {
	query := `
		DELETE FROM "trust_list"
		WHERE
			"list" = $1
	`
	_, err := q.Exec(ctx, query, list)
	return err
}

// ClearAdobeTL is ...
func (q *TrustQueries) ClearAdobeTL(ctx context.Context) error {
	query := `
		DELETE FROM "trust_list"
		WHERE
			"list" != 'gosing'
	`
	_, err := q.Exec(ctx, query)
	return err
}
