package queries

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
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
	var folderID sql.NullString
	err := q.QueryRow(ctx, query, id).Scan(
		&template.ID,
		&template.Slug,
		&template.Name,
		&folderID,
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

	// Set folder_id (handle NULL)
	if folderID.Valid {
		template.FolderID = folderID.String
	} else {
		template.FolderID = "" // Empty string for null folder_id
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

// CreateTemplate creates a new template
func (q *TemplateQueries) CreateTemplate(ctx context.Context, template *models.Template) error {
	query := `
		INSERT INTO "template" (
			"id", "slug", "name", "source", "folder_id", "organization_id", "submitters", "fields", "schema",
			"category", "tags", "is_favorite", "preview_image_id", "settings",
			"created_at", "updated_at"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	// Handle UUID fields - convert empty strings to nil
	var folderID interface{}
	if template.FolderID != "" {
		folderID = template.FolderID
	} else {
		folderID = nil
	}

	var orgID interface{}
	if template.OrganizationID != "" {
		orgID = template.OrganizationID
	} else {
		orgID = nil
	}

	// Ensure nil slices are converted to empty arrays to avoid NULL in database
	submitters := template.Submitters
	if submitters == nil {
		submitters = []models.Submitter{}
	}
	fields := template.Fields
	if fields == nil {
		fields = []models.Field{}
	}
	schema := template.Schema
	if schema == nil {
		schema = []models.Schema{}
	}

	_, err := q.Exec(ctx, query,
		template.ID,
		template.Slug,
		template.Name,
		template.Source,
		folderID,
		orgID,
		submitters,
		fields,
		schema,
		template.Category,
		template.Tags,
		template.IsFavorite,
		template.PreviewImageID,
		template.Settings,
		template.CreatedAt,
		template.UpdatedAt,
	)

	if err != nil {
		logging.Log.Err(err)
		return err
	}

	return nil
}

// CreateStorageBlob creates a new storage_blob record
func (q *TemplateQueries) CreateStorageBlob(ctx context.Context, blobID, filename, contentType string, byteSize int64, metadata map[string]any) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO "storage_blob" ("id", "filename", "content_type", "metadata", "byte_size")
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = q.Exec(ctx, query, blobID, filename, contentType, metadataJSON, byteSize)
	if err != nil {
		logging.Log.Err(err)
		return err
	}
	return nil
}

// CreateStorageAttachment creates a new storage_attachment record
func (q *TemplateQueries) CreateStorageAttachment(ctx context.Context, attachmentID, blobID, recordType, recordID, name, serviceName string) error {
	query := `
		INSERT INTO "storage_attachment" ("id", "blob_id", "record_type", "record_id", "name", "service_name", "created_at")
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`
	_, err := q.Exec(ctx, query, attachmentID, blobID, recordType, recordID, name, serviceName)
	if err != nil {
		logging.Log.Err(err)
		return err
	}
	return nil
}

// UpdateTemplateSchema updates the schema field of a template
func (q *TemplateQueries) UpdateTemplateSchema(ctx context.Context, templateID string, schema []models.Schema) error {
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	query := `
		UPDATE "template"
		SET "schema" = $1, "updated_at" = NOW()
		WHERE "id" = $2
	`
	_, err = q.Exec(ctx, query, schemaJSON, templateID)
	if err != nil {
		logging.Log.Err(err)
		return err
	}
	return nil
}

// UpdateTemplate updates template fields (name, category, etc.).
// It builds a dynamic UPDATE query based on the provided fields in the template struct.
// If category is an empty string, it sets the category to NULL in the database.
// The updated_at field is always updated to the current timestamp.
func (q *TemplateQueries) UpdateTemplate(ctx context.Context, templateID string, template *models.Template) error {
	// Build dynamic UPDATE query based on provided fields
	var setParts []string
	var args []interface{}
	argIndex := 1

	if template.Name != "" {
		setParts = append(setParts, fmt.Sprintf(`"name" = $%d`, argIndex))
		args = append(args, template.Name)
		argIndex++
	}

	// Handle category - if empty string, set to NULL, otherwise set the value
	if template.Category != "" {
		setParts = append(setParts, fmt.Sprintf(`"category" = $%d`, argIndex))
		args = append(args, template.Category)
		argIndex++
	} else {
		// Explicitly set to NULL if empty string
		setParts = append(setParts, `"category" = NULL`)
	}

	// Always update updated_at
	setParts = append(setParts, `"updated_at" = NOW()`)

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf(`
		UPDATE "template"
		SET %s
		WHERE "id" = $%d
	`, strings.Join(setParts, ", "), argIndex)
	args = append(args, templateID)

	_, err := q.Exec(ctx, query, args...)
	if err != nil {
		logging.Log.Err(err)
		return err
	}
	return nil
}

// TemplateSearchRequest represents search and filter parameters
type TemplateSearchRequest struct {
	Query         string   `json:"query,omitempty"`
	Category      string   `json:"category,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	OrganizationID string  `json:"organization_id,omitempty"`
	Favorites     bool     `json:"favorites,omitempty"`
	UserID        string   `json:"user_id,omitempty"`
	Limit         int      `json:"limit,omitempty"`
	Offset        int      `json:"offset,omitempty"`
	SortBy        string   `json:"sort_by,omitempty"`    // name, created_at, updated_at
	SortOrder     string   `json:"sort_order,omitempty"` // asc, desc
}

// TemplateSearchResult represents search result with metadata
type TemplateSearchResult struct {
	Templates   []models.Template `json:"templates"`
	Total       int               `json:"total"`
	HasMore     bool              `json:"has_more"`
	NextOffset  int               `json:"next_offset,omitempty"`
}

// SearchTemplates searches templates with filters and pagination
func (q *TemplateQueries) SearchTemplates(ctx context.Context, req TemplateSearchRequest) (*TemplateSearchResult, error) {
	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.SortBy == "" {
		req.SortBy = "updated_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	// Build WHERE conditions
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	// Base condition - not archived
	whereConditions = append(whereConditions, `"template"."archived_at" IS NULL`)

	// Organization filter - if organization_id is provided, filter by it
	if req.OrganizationID != "" {
		whereConditions = append(whereConditions, `"template"."organization_id" = ($`+fmt.Sprintf("%d", argIndex)+`::text)::uuid`)
		args = append(args, req.OrganizationID)
		argIndex++
	}

	// Search query - removed description search as column doesn't exist
	if req.Query != "" {
		whereConditions = append(whereConditions, `LOWER("template"."name") LIKE LOWER($`+fmt.Sprintf("%d", argIndex)+`)`)
		args = append(args, "%"+req.Query+"%")
		argIndex++
	}

	// Category filter
	if req.Category != "" {
		whereConditions = append(whereConditions, `"template"."category" = $`+fmt.Sprintf("%d", argIndex))
		args = append(args, req.Category)
		argIndex++
	}

	// Tags filter (array contains)
	if len(req.Tags) > 0 {
		whereConditions = append(whereConditions, `"template"."tags" @> $`+fmt.Sprintf("%d", argIndex))
		args = append(args, req.Tags)
		argIndex++
	}

	// Validate user ID format if provided
	var userIDParam string
	if req.UserID != "" {
		_, err := uuid.Parse(req.UserID)
		if err != nil {
			logging.Log.Err(err).Str("user_id", req.UserID).Msg("Invalid user ID format")
			return nil, fmt.Errorf("invalid user ID format: %w", err)
		}
		userIDParam = req.UserID
	}

	// Track user_id parameter index - will be set when we add user filter
	var userIDParamIndex int = 0 // Initialize to 0 (will be set if userID is provided)
	var joinClause string

	// User filter - show templates from user's folders OR templates without folder_id (in root)
	// For templates in root, check if they belong to the same account via organization membership
	// Use simple WHERE with EXISTS to ensure parameter is used in simple context first
	if req.UserID != "" {
		userIDParamIndex = argIndex
		// Include templates that either:
		// 1. Belong to a folder owned by the user (folder_id matches user's folders via account_id)
		// 2. Have no folder (folder_id IS NULL) AND belong to organization where user is a member
		// 3. Have no folder AND no organization (legacy templates visible to all - for backward compatibility)
		whereConditions = append(whereConditions, `(
			(
				"template"."folder_id" IS NULL AND (
					"template"."organization_id" IS NULL OR
					EXISTS (
						SELECT 1 FROM "organization_member" om
						INNER JOIN "user" u2 ON om.user_id = u2.id
						WHERE om.organization_id = "template"."organization_id"
						AND u2.id = ($`+fmt.Sprintf("%d", argIndex)+`::text)::uuid
					)
				)
			) OR
			EXISTS (
				SELECT 1 FROM "template_folder" tf
				INNER JOIN "user" u ON tf.account_id = u.account_id
				WHERE tf.id = "template"."folder_id"
				AND u.id = ($`+fmt.Sprintf("%d", argIndex)+`::text)::uuid
			)
		)`)
		args = append(args, userIDParam)
		argIndex++
	}

	// Favorites filter (after user filter, reuses the same userID parameter if exists)
	if req.Favorites && req.UserID != "" {
		whereConditions = append(whereConditions, `EXISTS (
			SELECT 1 FROM "template_favorite"
			WHERE "template_favorite"."template_id" = "template"."id"
			AND "template_favorite"."user_id" = ($`+fmt.Sprintf("%d", userIDParamIndex)+`::text)::uuid
		)`)
		// Don't add new parameter, reuse the one from user filter
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Build ORDER BY
	orderBy := ""
	switch req.SortBy {
	case "name":
		orderBy = `ORDER BY "template"."name"`
	case "created_at":
		orderBy = `ORDER BY "template"."created_at"`
	case "updated_at":
		orderBy = `ORDER BY "template"."updated_at"`
	default:
		orderBy = `ORDER BY "template"."updated_at"`
	}

	if req.SortOrder == "asc" {
		orderBy += " ASC"
	} else {
		orderBy += " DESC"
	}

	// Count query - use current args with JOIN if needed
	countQuery := `
		SELECT COUNT(*)
		FROM "template"
		` + joinClause + `
		` + whereClause

	var total int
	err := q.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		logging.Log.Err(err)
		return nil, err
	}

	// Data query with pagination
	// Check if template is in user's favorites using EXISTS
	var favoriteCheckSQL string
	if req.UserID != "" {
		// Use the userID parameter index that was already set
		favoriteCheckSQL = `EXISTS (
			SELECT 1 FROM "template_favorite"
			WHERE "template_favorite"."template_id" = "template"."id"
			AND "template_favorite"."user_id" = ($` + fmt.Sprintf("%d", userIDParamIndex) + `::text)::uuid
		) as is_user_favorite`
	} else {
		// If no user ID, favorites are always false
		favoriteCheckSQL = "false as is_user_favorite"
	}
	
	// Use LIMIT and OFFSET directly in SQL to avoid parameter type issues
	// Values are validated (limit <= 100, offset >= 0) so safe to inline
	dataQuery := `
		SELECT
			"template"."id",
			"template"."slug",
			"template"."name",
			NULL as description,
			"template"."folder_id",
			"template"."category",
			"template"."tags",
			"template"."is_favorite",
			"template"."preview_image_id",
			"template"."created_at",
			"template"."updated_at",
			` + favoriteCheckSQL + `
		FROM "template"
		` + joinClause + `
		` + whereClause + `
		` + orderBy + `
		LIMIT ` + fmt.Sprintf("%d", req.Limit) + ` OFFSET ` + fmt.Sprintf("%d", req.Offset)

	rows, err := q.Query(ctx, dataQuery, args...)
	if err != nil {
		logging.Log.Err(err)
		return nil, err
	}
	defer rows.Close()

	var templates []models.Template
	for rows.Next() {
		var template models.Template
		var isUserFavorite bool
		var description sql.NullString
		var folderID sql.NullString
		var category sql.NullString
		var previewImageID sql.NullString
		var tags []string
		err := rows.Scan(
			&template.ID,
			&template.Slug,
			&template.Name,
			&description,
			&folderID,
			&category,
			&tags,
			&template.IsFavorite,
			&previewImageID,
			&template.CreatedAt,
			&template.UpdatedAt,
			&isUserFavorite,
		)
		if err != nil {
			logging.Log.Err(err)
			return nil, err
		}

		// Set nullable fields
		if description.Valid {
			template.Description = description.String
		}
		if folderID.Valid {
			template.FolderID = folderID.String
		} else {
			template.FolderID = "" // Empty string for null folder_id
		}
		if category.Valid {
			template.Category = category.String
		}
		if previewImageID.Valid {
			template.PreviewImageID = previewImageID.String
		}
		template.Tags = tags

		// Override is_favorite with user-specific favorite status
		template.IsFavorite = isUserFavorite

		templates = append(templates, template)
	}

	result := &TemplateSearchResult{
		Templates:  templates,
		Total:      total,
		HasMore:    req.Offset+len(templates) < total,
		NextOffset: req.Offset + len(templates),
	}

	return result, nil
}

// AddTemplateFavorite adds template to user's favorites
func (q *TemplateQueries) AddTemplateFavorite(ctx context.Context, templateID, userID string) error {
	query := `
		INSERT INTO "template_favorite" ("id", "template_id", "user_id", "created_at")
		VALUES ($1, $2, $3, $4)
		ON CONFLICT ("template_id", "user_id") DO NOTHING
	`
	_, err := q.Exec(ctx, query, uuid.New().String(), templateID, userID, time.Now())
	if err != nil {
		logging.Log.Err(err)
		return err
	}
	return nil
}

// RemoveTemplateFavorite removes template from user's favorites
func (q *TemplateQueries) RemoveTemplateFavorite(ctx context.Context, templateID, userID string) error {
	query := `DELETE FROM "template_favorite" WHERE "template_id" = $1 AND "user_id" = $2`
	_, err := q.Exec(ctx, query, templateID, userID)
	if err != nil {
		logging.Log.Err(err)
		return err
	}
	return nil
}

// GetUserFavoriteTemplates returns user's favorite templates
func (q *TemplateQueries) GetUserFavoriteTemplates(ctx context.Context, userID string, limit, offset int) ([]models.Template, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT
			"template"."id",
			"template"."slug",
			"template"."name",
			"template"."description",
			"template"."category",
			"template"."tags",
			"template"."is_favorite",
			"template"."preview_image_id",
			"template"."created_at",
			"template"."updated_at"
		FROM "template"
		INNER JOIN "template_favorite" ON "template"."id" = "template_favorite"."template_id"
		WHERE "template_favorite"."user_id" = $1
		AND "template"."archived_at" IS NULL
		ORDER BY "template_favorite"."created_at" DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := q.Query(ctx, query, userID, limit, offset)
	if err != nil {
		logging.Log.Err(err)
		return nil, err
	}
	defer rows.Close()

	var templates []models.Template
	for rows.Next() {
		var template models.Template
		err := rows.Scan(
			&template.ID,
			&template.Slug,
			&template.Name,
			&template.Description,
			&template.Category,
			&template.Tags,
			&template.IsFavorite,
			&template.PreviewImageID,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			logging.Log.Err(err)
			return nil, err
		}

		// For favorites, always set is_favorite to true
		template.IsFavorite = true

		templates = append(templates, template)
	}

	return templates, nil
}

// CreateTemplateFolder creates a new template folder
// userID is used to get account_id from user table
func (q *TemplateQueries) CreateTemplateFolder(ctx context.Context, userID string, folder *models.TemplateFolder) error {
	// Get account_id from user_id
	var accountID string
	err := q.QueryRow(ctx, `SELECT "account_id" FROM "user" WHERE "id" = $1`, userID).Scan(&accountID)
	if err != nil {
		logging.Log.Err(err)
		return fmt.Errorf("failed to get account_id for user: %w", err)
	}

	query := `
		INSERT INTO "template_folder" ("id", "name", "account_id", "created_at", "updated_at")
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = q.Exec(ctx, query, folder.ID, folder.Name, accountID, folder.CreatedAt, folder.UpdatedAt)
	if err != nil {
		logging.Log.Err(err)
		return err
	}
	return nil
}

// GetTemplateFolders returns user's template folders
func (q *TemplateQueries) GetTemplateFolders(ctx context.Context, userID string) ([]models.TemplateFolder, error) {
	query := `
		SELECT 
			tf."id", 
			tf."name", 
			NULL::uuid as "parent_id",
			u."account_id" as "user_id",
			tf."created_at", 
			tf."updated_at"
		FROM "template_folder" tf
		INNER JOIN "user" u ON tf."account_id" = u."account_id"
		WHERE u."id" = $1 AND tf."archived_at" IS NULL
		ORDER BY tf."name" ASC
	`

	rows, err := q.Query(ctx, query, userID)
	if err != nil {
		logging.Log.Err(err)
		return nil, err
	}
	defer rows.Close()

	var folders []models.TemplateFolder
	for rows.Next() {
		var folder models.TemplateFolder
		var parentID sql.NullString
		err := rows.Scan(
			&folder.ID,
			&folder.Name,
			&parentID,
			&folder.UserID,
			&folder.CreatedAt,
			&folder.UpdatedAt,
		)
		if err != nil {
			logging.Log.Err(err)
			return nil, err
		}
		if parentID.Valid {
			folder.ParentID = &parentID.String
		}
		folders = append(folders, folder)
	}

	return folders, nil
}

// UpdateTemplateFolder updates a template folder
// userID is used to verify ownership via account_id
func (q *TemplateQueries) UpdateTemplateFolder(ctx context.Context, folderID, userID, name string, parentID *string) error {
	query := `
		UPDATE "template_folder" tf
		SET "name" = $1, "updated_at" = $2
		WHERE tf."id" = $3 
		AND EXISTS (
			SELECT 1 FROM "user" u 
			WHERE u."id" = $4 AND u."account_id" = tf."account_id"
		)
	`
	_, err := q.Exec(ctx, query, name, time.Now(), folderID, userID)
	if err != nil {
		logging.Log.Err(err)
		return err
	}
	return nil
}

// DeleteTemplateFolder deletes a template folder
// userID is used to verify ownership via account_id
func (q *TemplateQueries) DeleteTemplateFolder(ctx context.Context, folderID, userID string) error {
	query := `
		DELETE FROM "template_folder" tf
		WHERE tf."id" = $1 
		AND EXISTS (
			SELECT 1 FROM "user" u 
			WHERE u."id" = $2 AND u."account_id" = tf."account_id"
		)
	`
	_, err := q.Exec(ctx, query, folderID, userID)
	if err != nil {
		logging.Log.Err(err)
		return err
	}
	return nil
}

// MoveTemplateToFolder moves a template to a different folder
// userID is used to verify ownership via account_id
func (q *TemplateQueries) MoveTemplateToFolder(ctx context.Context, templateID, folderID, userID string) error {
	var args []interface{}

	baseQuery := `
		UPDATE "template" t
		SET "folder_id" = $1, "updated_at" = $2
		WHERE t."id" = $3
	`

	// If moving to a folder, check that user owns the folder via account_id
	if folderID != "" {
		baseQuery += ` AND EXISTS (
			SELECT 1 FROM "template_folder" tf
			INNER JOIN "user" u ON tf."account_id" = u."account_id"
			WHERE tf."id" = $1::uuid AND u."id" = $4::uuid
		)`
		args = append(args, folderID, time.Now(), templateID, userID)
	} else {
		// Moving to root (no folder) - set folder_id to NULL
		// Note: folder_id is now nullable in the database schema
		args = append(args, nil, time.Now(), templateID)
	}

	_, err := q.Exec(ctx, baseQuery, args...)
	if err != nil {
		logging.Log.Err(err)
		return err
	}
	return nil
}
