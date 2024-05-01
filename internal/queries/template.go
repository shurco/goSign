package queries

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/logging"
)

// TemplateQueries is ...
type TemplateQueries struct {
	*pgxpool.Pool
}

// Template is ...
func (q *TemplateQueries) Template(ctx context.Context, id string) (*models.Template, error) {
	template := &models.Template{}

	// template info
	query := `
		SELECT
			"template"."id",
			"template"."slug",
			"template"."name",
			"template"."folder_id",
			"template"."submitters",
			"template"."fields",
			"template"."schema",
			"template"."created_at",
			"template"."updated_at"
		FROM
			"template"
		WHERE
			"template"."id" = $1
	`
	err := q.QueryRow(ctx, query, id).Scan(
		&template.ID,
		&template.Slug,
		&template.Name,
		&template.FolderID,
		&template.Submitters,
		&template.Fields,
		&template.Schema,
		&template.CreatedAt,
		&template.UpdatedAt,
	)
	if err != nil {
		logging.Log.Err(err)
		return nil, err
	}

	// documents list
	query = `
		SELECT
			"storage_attachment"."id",
			"storage_attachment"."service_name",
			jsonb_agg(
				jsonb_build_object(
					'id',
					storage_blob.id,
					'metadata',
					storage_blob.metadata,
					'filename',
					storage_blob.filename
				)
			) AS preview_images,
			storage_attachment.created_at
		FROM
			"storage_attachment"
			LEFT JOIN "storage_blob" ON "storage_attachment"."blob_id" = "storage_blob"."id"
		WHERE
			"storage_attachment"."record_type" = 'Template'
			AND "storage_attachment"."name" = 'documents'
			AND "storage_attachment"."record_id" = $1
		GROUP BY
			storage_attachment.id
	`
	rows, err := q.Query(ctx, query, id)
	if err != nil {
		logging.Log.Err(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var serviceName string
		document := &models.Document{}
		previewImages := []models.PreviewImages{}
		err := rows.Scan(
			&document.ID,
			&serviceName,
			&previewImages,
			&document.CreatedAt,
		)
		if err != nil {
			logging.Log.Err(err)
			return nil, err
		}

		if serviceName == "disk" {
			document.URL = "http://localhost:8088/drive/pages"
		}

		if document.Metadata.Pdf.NumberOfPages == 0 {
			document.Metadata.Pdf.NumberOfPages = 1
		}
		document.PreviewImages = previewImages

		template.Documents = append(template.Documents, *document)
	}

	return template, nil
}
