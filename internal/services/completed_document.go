package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/pdf"
)

// CompletedDocumentBuilder builds and caches a final, completed PDF for a submission.
//
// It is intentionally simple and filesystem-backed:
// - source pages are in PagesDir: lc_pages/{attachment_id}/0.pdf
// - completed PDFs are written to SignedDir: lc_signed/submission_{submission_id}.pdf
type CompletedDocumentBuilder struct {
	Pool           *pgxpool.Pool
	TemplateQueries *queries.TemplateQueries
	PagesDir       string
	SignedDir      string
	AssetsDir      string
}

func (b *CompletedDocumentBuilder) CompletedPDFPath(submissionID string) string {
	// Versioned filename to avoid serving older cached files after new append steps
	// (e.g. certificate/audit pages) are introduced.
	return filepath.Join(b.SignedDir, fmt.Sprintf("submission_%s_completed_v2.pdf", submissionID))
}

func (b *CompletedDocumentBuilder) CertificatePDFPath(submissionID string) string {
	return filepath.Join(b.SignedDir, fmt.Sprintf("submission_%s_certificate_v1.pdf", submissionID))
}

func firstNonNilTime(ts ...*time.Time) *time.Time {
	for _, t := range ts {
		if t == nil || t.IsZero() {
			continue
		}
		tt := *t
		return &tt
	}
	return nil
}

type loadedSubmitter struct {
	name              string
	email             string
	slug              string
	ip                string
	location          string
	sentAt            *time.Time
	openedAt          *time.Time
	completedAt       *time.Time
	createdAt         time.Time
	updatedAt         time.Time
	templateSubmitter string
	fields            map[string]any
}

type submissionData struct {
	tpl           *models.Template
	values        map[string]any
	submitters    []loadedSubmitter
	completedAtMax *time.Time
	publicBaseURL string
}

func (b *CompletedDocumentBuilder) loadSubmissionData(ctx context.Context, submissionID string) (*submissionData, error) {
	// Load template id + public_base_url (for QR).
	var templateID string
	var publicBaseURL string
	if err := b.Pool.QueryRow(ctx, `
		SELECT template_id, COALESCE(preferences->>'public_base_url', '')
		FROM submission
		WHERE id = $1
	`, submissionID).Scan(&templateID, &publicBaseURL); err != nil {
		return nil, fmt.Errorf("failed to load submission: %w", err)
	}

	tpl, err := b.TemplateQueries.Template(ctx, templateID)
	if err != nil || tpl == nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	// Load all submitters (metadata + timestamps + identity) in one go.
	rows, err := b.Pool.Query(ctx, `
		SELECT
			COALESCE(name, '') AS name,
			COALESCE(email, '') AS email,
			COALESCE(slug, '') AS slug,
			COALESCE(host(ip)::text, '') AS ip,
			sented_at,
			opened_at,
			completed_at,
			created_at,
			updated_at,
			COALESCE(metadata, '{}'::jsonb)::text AS metadata_json
		FROM submitter
		WHERE submission_id = $1
		ORDER BY created_at ASC
	`, submissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load submitters: %w", err)
	}
	defer rows.Close()

	values := make(map[string]any)
	submitters := make([]loadedSubmitter, 0, 4)
	var completedAtMax *time.Time

	for rows.Next() {
		var (
			name        string
			email       string
			slug        string
			ip          string
			sentAt      *time.Time
			openedAt    *time.Time
			completedAt *time.Time
			createdAt   time.Time
			updatedAt   time.Time
			metaJSON    string
		)
		if err := rows.Scan(&name, &email, &slug, &ip, &sentAt, &openedAt, &completedAt, &createdAt, &updatedAt, &metaJSON); err != nil {
			return nil, fmt.Errorf("failed to scan submitter: %w", err)
		}

		var meta map[string]any
		if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
			// If metadata is invalid JSON, create empty map
			meta = map[string]any{}
		}
		templateSubmitterID, _ := meta["template_submitter_id"].(string)
		
		// Extract location from metadata (can be string for backward compatibility or map with city/country/full)
		// Location is stored in submitter.metadata.location when completing signing
		var location string
		if locationAny, ok := meta["location"]; ok && locationAny != nil {
			if locationStr, ok := locationAny.(string); ok && locationStr != "" {
				// Backward compatibility: old format (string)
				location = locationStr
			} else if locationMap, ok := locationAny.(map[string]any); ok {
				// New format: map with city, country, full
				// Prefer "full" field if available, otherwise build from city and country
				if full, ok := locationMap["full"].(string); ok && full != "" {
					location = full
				} else {
					// Build from city and country if full is not available
					var parts []string
					if city, ok := locationMap["city"].(string); ok && city != "" {
						parts = append(parts, city)
					}
					if country, ok := locationMap["country"].(string); ok && country != "" {
						parts = append(parts, country)
					}
					if len(parts) > 0 {
						location = strings.Join(parts, ", ")
					}
				}
			}
		}

		fieldsAny, _ := meta["fields"]
		fieldsMap, ok := fieldsAny.(map[string]any)
		if !ok {
			fieldsMap = map[string]any{}
		}
		for k, v := range fieldsMap {
			values[k] = v
		}

		submitters = append(submitters, loadedSubmitter{
			name:              name,
			email:             email,
			slug:              slug,
			ip:                ip,
			location:          location,
			sentAt:            sentAt,
			openedAt:          openedAt,
			completedAt:       completedAt,
			createdAt:         createdAt,
			updatedAt:         updatedAt,
			templateSubmitter: templateSubmitterID,
			fields:            fieldsMap,
		})

		// Track the overall completion time as max(signed_at).
		if completedAt != nil && !completedAt.IsZero() {
			if completedAtMax == nil || completedAt.After(*completedAtMax) {
				t := *completedAt
				completedAtMax = &t
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate submitters: %w", err)
	}

	return &submissionData{
		tpl:            tpl,
		values:         values,
		submitters:     submitters,
		completedAtMax: completedAtMax,
		publicBaseURL:  strings.TrimRight(publicBaseURL, "/"),
	}, nil
}

// IsSubmissionFullyCompleted returns true if ALL submitters for the submission are completed.
func (b *CompletedDocumentBuilder) IsSubmissionFullyCompleted(ctx context.Context, submissionID string) (bool, error) {
	if b.Pool == nil {
		return false, fmt.Errorf("db pool not configured")
	}
	var ok bool
	err := b.Pool.QueryRow(ctx, `
		SELECT (count(*) > 0) AND bool_and(COALESCE(status, 'pending') = 'completed')
		FROM submitter
		WHERE submission_id = $1
	`, submissionID).Scan(&ok)
	if err != nil {
		return false, err
	}
	return ok, nil
}

// EnsureCompletedPDF generates the completed PDF (if missing) and returns its path.
// It does NOT check completion; caller must ensure submission is completed.
func (b *CompletedDocumentBuilder) EnsureCompletedPDF(ctx context.Context, submissionID string) (string, error) {
	if b.Pool == nil {
		return "", fmt.Errorf("db pool not configured")
	}
	if b.TemplateQueries == nil {
		return "", fmt.Errorf("template queries not configured")
	}
	if b.PagesDir == "" || b.SignedDir == "" {
		return "", fmt.Errorf("pages/signed dirs not configured")
	}

	outPath := b.CompletedPDFPath(submissionID)
	if _, err := os.Stat(outPath); err == nil {
		return outPath, nil
	}

	data, err := b.loadSubmissionData(ctx, submissionID)
	if err != nil {
		return "", err
	}
	tpl := data.tpl

	// 3) Render base completed PDF.
	outBytes, err := pdf.RenderCompletedTemplatePDF(pdf.RenderCompletedTemplatePDFInput{
		PagesDir: b.PagesDir,
		Schema:   tpl.Schema,
		Fields:   tpl.Fields,
		Values:   data.values,
	})
	if err != nil {
		return "", err
	}

	// 4) Append signature certificate page(s) when the submission is fully completed.
	// Pick one "example signature" per signer based on the template fields that belong to them.
	certSigners := make([]pdf.SignatureCertificateSigner, 0, len(data.submitters))
	qrSlug := ""
	completedAtMax := data.completedAtMax
	for _, s := range data.submitters {
		var sigVal any
		var sigID string

		// Prefer fields assigned to this signer (multi-signer templates).
		if s.templateSubmitter != "" {
			for i := range tpl.Fields {
				f := tpl.Fields[i]
				if f.SubmitterID != s.templateSubmitter {
					continue
				}
				switch f.Type {
				case "signature", "initials", "stamp", "image":
					if v, ok := s.fields[f.ID]; ok && strings.TrimSpace(fmt.Sprint(v)) != "" {
						sigVal = v
						if id, _ := s.fields[f.ID+"_signature_id"].(string); id != "" {
							sigID = strings.TrimSpace(id)
						}
					}
				}
				if sigVal != nil {
					break
				}
			}
		}

		// Fallback: first available signature-ish field value for this signer.
		if sigVal == nil {
			for i := range tpl.Fields {
				f := tpl.Fields[i]
				switch f.Type {
				case "signature", "initials", "stamp", "image":
					if v, ok := s.fields[f.ID]; ok && strings.TrimSpace(fmt.Sprint(v)) != "" {
						sigVal = v
						if id, _ := s.fields[f.ID+"_signature_id"].(string); id != "" {
							sigID = strings.TrimSpace(id)
						}
					}
				}
				if sigVal != nil {
					break
				}
			}
		}

		// Use actual completed_at (no fallback to updated_at for accuracy)
		if s.completedAt != nil && !s.completedAt.IsZero() {
			if completedAtMax == nil || s.completedAt.After(*completedAtMax) {
				t := *s.completedAt
				completedAtMax = &t
			}
		}

		if qrSlug == "" && strings.TrimSpace(s.slug) != "" {
			qrSlug = s.slug
		}

		certSigners = append(certSigners, pdf.SignatureCertificateSigner{
			Name:           s.name,
			Email:          s.email,
			IP:             s.ip,
			SentAt:         s.sentAt,
			OpenedAt:       s.openedAt,
			CompletedAt:    s.completedAt,
			Location:       s.location,
			SignatureValue: sigVal,
			SignatureID:    sigID,
		})
	}

	if strings.TrimSpace(b.AssetsDir) == "" {
		return "", fmt.Errorf("assets dir not configured")
	}
	if strings.TrimSpace(data.publicBaseURL) == "" {
		return "", fmt.Errorf("public_base_url is not set for submission")
	}
	if strings.TrimSpace(qrSlug) == "" {
		return "", fmt.Errorf("missing submitter slug for certificate QR url")
	}
	qrURL := fmt.Sprintf("%s/public/sign/%s/certificate", strings.TrimRight(data.publicBaseURL, "/"), qrSlug)

	certBytes, err := pdf.GenerateSignatureCertificatePDF(pdf.SignatureCertificateInput{
		DocumentName: tpl.Name,
		Reference:    submissionID,
		CompletedAt:  completedAtMax,
		AssetsDir:    b.AssetsDir,
		QRURL:        qrURL,
		Signers:      certSigners,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate signature certificate: %w", err)
	}
	outBytes, err = pdf.AppendSignatureCertificate(outBytes, certBytes)
	if err != nil {
		return "", fmt.Errorf("failed to append signature certificate: %w", err)
	}

	if err := os.MkdirAll(b.SignedDir, 0755); err != nil {
		return "", fmt.Errorf("failed to ensure signed dir: %w", err)
	}
	if err := os.WriteFile(outPath, outBytes, 0644); err != nil {
		return "", fmt.Errorf("failed to write completed PDF: %w", err)
	}
	return outPath, nil
}

// EnsureCertificatePDF generates the certificate-only PDF (if missing) and returns its path.
// It does NOT check completion; caller must ensure submission is completed.
func (b *CompletedDocumentBuilder) EnsureCertificatePDF(ctx context.Context, submissionID string) (string, error) {
	if b.Pool == nil {
		return "", fmt.Errorf("db pool not configured")
	}
	if b.TemplateQueries == nil {
		return "", fmt.Errorf("template queries not configured")
	}
	if b.SignedDir == "" {
		return "", fmt.Errorf("signed dir not configured")
	}

	outPath := b.CertificatePDFPath(submissionID)
	if _, err := os.Stat(outPath); err == nil {
		return outPath, nil
	}

	data, err := b.loadSubmissionData(ctx, submissionID)
	if err != nil {
		return "", err
	}
	tpl := data.tpl

	certSigners := make([]pdf.SignatureCertificateSigner, 0, len(data.submitters))
	qrSlug := ""
	completedAtMax := data.completedAtMax
	for _, s := range data.submitters {
		var sigVal any
		var sigID string

		if s.templateSubmitter != "" {
			for i := range tpl.Fields {
				f := tpl.Fields[i]
				if f.SubmitterID != s.templateSubmitter {
					continue
				}
				switch f.Type {
				case "signature", "initials", "stamp", "image":
					if v, ok := s.fields[f.ID]; ok && strings.TrimSpace(fmt.Sprint(v)) != "" {
						sigVal = v
						if id, _ := s.fields[f.ID+"_signature_id"].(string); id != "" {
							sigID = strings.TrimSpace(id)
						}
					}
				}
				if sigVal != nil {
					break
				}
			}
		}
		if sigVal == nil {
			for i := range tpl.Fields {
				f := tpl.Fields[i]
				switch f.Type {
				case "signature", "initials", "stamp", "image":
					if v, ok := s.fields[f.ID]; ok && strings.TrimSpace(fmt.Sprint(v)) != "" {
						sigVal = v
						if id, _ := s.fields[f.ID+"_signature_id"].(string); id != "" {
							sigID = strings.TrimSpace(id)
						}
					}
				}
				if sigVal != nil {
					break
				}
			}
		}

		// Use actual completed_at (no fallback to updated_at for accuracy)
		if s.completedAt != nil && !s.completedAt.IsZero() {
			if completedAtMax == nil || s.completedAt.After(*completedAtMax) {
				t := *s.completedAt
				completedAtMax = &t
			}
		}
		if qrSlug == "" && strings.TrimSpace(s.slug) != "" {
			qrSlug = s.slug
		}

		certSigners = append(certSigners, pdf.SignatureCertificateSigner{
			Name:           s.name,
			Email:          s.email,
			IP:             s.ip,
			SentAt:         s.sentAt, // Use actual sent_at, not fallback to created_at
			OpenedAt:       s.openedAt,
			CompletedAt:    s.completedAt, // Use actual completed_at, not fallback to updated_at
			Location:       s.location,
			SignatureValue: sigVal,
			SignatureID:    sigID,
		})
	}

	if strings.TrimSpace(b.AssetsDir) == "" {
		return "", fmt.Errorf("assets dir not configured")
	}
	if strings.TrimSpace(data.publicBaseURL) == "" {
		return "", fmt.Errorf("public_base_url is not set for submission")
	}
	if strings.TrimSpace(qrSlug) == "" {
		return "", fmt.Errorf("missing submitter slug for certificate QR url")
	}
	qrURL := fmt.Sprintf("%s/public/sign/%s/certificate", strings.TrimRight(data.publicBaseURL, "/"), qrSlug)

	certBytes, err := pdf.GenerateSignatureCertificatePDF(pdf.SignatureCertificateInput{
		DocumentName: tpl.Name,
		Reference:    submissionID,
		CompletedAt:  completedAtMax,
		AssetsDir:    b.AssetsDir,
		QRURL:        qrURL,
		Signers:      certSigners,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate signature certificate: %w", err)
	}

	if err := os.MkdirAll(b.SignedDir, 0755); err != nil {
		return "", fmt.Errorf("failed to ensure signed dir: %w", err)
	}
	if err := os.WriteFile(outPath, certBytes, 0644); err != nil {
		return "", fmt.Errorf("failed to write certificate PDF: %w", err)
	}
	return outPath, nil
}

