package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/logging"
	"github.com/shurco/gosign/pkg/storage/postgres"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

func Template(c *fiber.Ctx) error {
	template := "00c95859-98ef-42cd-a801-2023b75a9431"
	response := &models.Template{}

	// template info
	err := postgres.Pool.QueryRow(context.Background(), `
		SELECT
			template.id,
			template.slug,
			template.name,
			template.folder_id,
			template.submitters,
			template.fields,
			template.schema,
			template.created_at,
			template.updated_at
		FROM
			"template"
		WHERE
			"template"."id" = $1
	`, template).Scan(
		&response.ID,
		&response.Slug,
		&response.Name,
		&response.FolderID,
		&response.Submitters,
		&response.Fields,
		&response.Schema,
		&response.CreatedAt,
		&response.UpdatedAt,
	)
	if err != nil {
		logging.Log.Err(err)
		return err
	}

	// documents list
	rows, err := postgres.Pool.Query(context.Background(), `
		SELECT
			storage_attachment.id,
			storage_attachment.service_name,
			jsonb_agg(jsonb_build_object (
				'id', storage_blob.id,
				'metadata', storage_blob.metadata,
				'filename', storage_blob.filename
			)) AS preview_images,
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
	`, template)
	if err != nil {
		logging.Log.Err(err)
		return err
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
			return err
		}

		if serviceName == "disk" {
			document.URL = "http://localhost:8088/drive/pages"
		}

		if document.Metadata.Pdf.NumberOfPages == 0 {
			document.Metadata.Pdf.NumberOfPages = 1
		}
		document.PreviewImages = previewImages

		response.Documents = append(response.Documents, *document)
	}

	return webutil.Response(c, fiber.StatusOK, "Template", response)
}
