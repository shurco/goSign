package api

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/shurco/gosign/internal/services/submission"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// BulkHandler handles bulk operations
type BulkHandler struct {
	submissionSvc *submission.Service
}

// NewBulkHandler creates a new bulk handler
func NewBulkHandler(submissionSvc *submission.Service) *BulkHandler {
	return &BulkHandler{
		submissionSvc: submissionSvc,
	}
}

// BulkResult represents the result of a bulk operation
type BulkResult struct {
	Total     int               `json:"total"`
	Success   int               `json:"success"`
	Failed    int               `json:"failed"`
	Results   []BulkItemResult  `json:"results"`
}

// BulkItemResult represents the result for a single record
type BulkItemResult struct {
	Row         int    `json:"row"`
	SubmissionID string `json:"submission_id,omitempty"`
	Status      string `json:"status"` // "success" or "failed"
	Error       string `json:"error,omitempty"`
}

// BulkCreateSubmissions bulk creates submissions from CSV/XLSX
// @Summary Bulk create submissions
// @Description Create multiple submissions from CSV file
// @Tags Submissions
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV file with submissions"
// @Param template_id formData string true "Template ID"
// @Param send_immediately formData boolean false "Send immediately after creation"
// @Success 200 {object} map[string]any "Success response with results"
// @Failure 400 {object} map[string]any "Bad request"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /api/v1/submissions/bulk [post]
func (h *BulkHandler) BulkCreateSubmissions(c *fiber.Ctx) error {
	// Get template_id
	templateID := c.FormValue("template_id")
	if templateID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "template_id is required", nil)
	}

	sendImmediately := c.FormValue("send_immediately") == "true"

	// Get file
	file, err := c.FormFile("file")
	if err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "file is required", nil)
	}

	// Open file
	fileData, err := file.Open()
	if err != nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to open file", map[string]any{
			"error": err.Error(),
		})
	}
	defer fileData.Close()

	// Parse CSV
	records, err := h.parseCSV(fileData)
	if err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Failed to parse CSV", map[string]any{
			"error": err.Error(),
		})
	}

	// Process records
	result := h.processBulkSubmissions(c.Context(), templateID, records, sendImmediately)

	return webutil.Response(c, fiber.StatusOK, "Bulk submissions processed", result)
}

// parseCSV parses a CSV file
func (h *BulkHandler) parseCSV(file io.Reader) ([]map[string]string, error) {
	reader := csv.NewReader(file)
	
	// Read headers
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}

	// Normalize headers (remove spaces, lowercase)
	for i, h := range headers {
		headers[i] = strings.TrimSpace(strings.ToLower(h))
	}

	var records []map[string]string

	// Read records
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		// Create map from headers and values
		record := make(map[string]string)
		for i, value := range row {
			if i < len(headers) {
				record[headers[i]] = strings.TrimSpace(value)
			}
		}

		records = append(records, record)
	}

	return records, nil
}

// processBulkSubmissions processes bulk submission creation
func (h *BulkHandler) processBulkSubmissions(ctx context.Context, templateID string, records []map[string]string, sendImmediately bool) *BulkResult {
	result := &BulkResult{
		Total:   len(records),
		Results: make([]BulkItemResult, 0, len(records)),
	}

	for i, record := range records {
		rowNum := i + 2 // +2 because row 1 is headers, numbering starts from 1

		// Validate required fields
		title, ok := record["title"]
		if !ok || title == "" {
			result.Failed++
			result.Results = append(result.Results, BulkItemResult{
				Row:    rowNum,
				Status: "failed",
				Error:  "missing required field: title",
			})
			continue
		}

		email, ok := record["email"]
		if !ok || email == "" {
			result.Failed++
			result.Results = append(result.Results, BulkItemResult{
				Row:    rowNum,
				Status: "failed",
				Error:  "missing required field: email",
			})
			continue
		}

		name := record["name"]
		if name == "" {
			name = email // Use email as name by default
		}

		// Create submission using existing service
		input := submission.CreateSubmissionInput{
			TemplateID:  templateID,
			CreatedByID: "", // TODO: Get from context/auth
			Submitters: []submission.SubmitterInput{
				{
					Name:  name,
					Email: email,
				},
			},
		}

		sub, err := h.submissionSvc.Create(ctx, input)
		if err != nil {
			result.Failed++
			result.Results = append(result.Results, BulkItemResult{
				Row:    rowNum,
				Status: "failed",
				Error:  fmt.Sprintf("failed to create submission: %v", err),
			})
			continue
		}

		// Send immediately if requested
		if sendImmediately {
			if err := h.submissionSvc.Send(ctx, sub.ID); err != nil {
				// Submission created but not sent - still consider it success
				result.Success++
				result.Results = append(result.Results, BulkItemResult{
					Row:          rowNum,
					SubmissionID: sub.ID,
					Status:       "success",
					Error:        fmt.Sprintf("created but failed to send: %v", err),
				})
				continue
			}
		}

		result.Success++
		result.Results = append(result.Results, BulkItemResult{
			Row:          rowNum,
			SubmissionID: sub.ID,
			Status:       "success",
		})
	}

	return result
}

// RegisterRoutes registers bulk operations routes
func (h *BulkHandler) RegisterRoutes(router fiber.Router) {
	router.Post("/submissions/bulk", h.BulkCreateSubmissions)
}

