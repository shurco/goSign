package app

import (
	"context"
	"fmt"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/internal/services/submission"
)

type simpleTemplateRepository struct {
	templateQueries *queries.TemplateQueries
}

func (r *simpleTemplateRepository) List(page, pageSize int, filters map[string]string) ([]models.Template, int, error) {
	if r.templateQueries == nil {
		return []models.Template{}, 0, nil
	}
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	req := queries.TemplateSearchRequest{
		Limit:  pageSize,
		Offset: offset,
		Query:  filters["query"],
	}
	result, err := r.templateQueries.SearchTemplates(context.Background(), req)
	if err != nil {
		return nil, 0, err
	}
	return result.Templates, result.Total, nil
}

func (r *simpleTemplateRepository) Get(id string) (*models.Template, error) {
	if r.templateQueries == nil {
		return nil, fmt.Errorf("template queries not initialized")
	}
	return r.templateQueries.Template(context.Background(), id)
}

func (r *simpleTemplateRepository) Create(item *models.Template) error {
	return fmt.Errorf("not implemented")
}

func (r *simpleTemplateRepository) Update(id string, item *models.Template) error {
	if r.templateQueries == nil {
		return fmt.Errorf("template queries not initialized")
	}
	return r.templateQueries.UpdateTemplate(context.Background(), id, item)
}

func (r *simpleTemplateRepository) Delete(id string) error {
	return fmt.Errorf("not implemented")
}

type simpleSubmissionRepository struct {
	submissionRepo submission.Repository
}

func (r *simpleSubmissionRepository) List(page, pageSize int, filters map[string]string) ([]models.Submission, int, error) {
	return []models.Submission{}, 0, nil
}

func (r *simpleSubmissionRepository) Get(id string) (*models.Submission, error) {
	if r.submissionRepo == nil {
		return nil, fmt.Errorf("submission repository not initialized")
	}
	return r.submissionRepo.GetSubmission(context.Background(), id)
}

func (r *simpleSubmissionRepository) Create(item *models.Submission) error {
	if r.submissionRepo == nil {
		return fmt.Errorf("submission repository not initialized")
	}
	return r.submissionRepo.CreateSubmission(context.Background(), item)
}

func (r *simpleSubmissionRepository) Update(id string, item *models.Submission) error {
	return fmt.Errorf("not implemented")
}

func (r *simpleSubmissionRepository) Delete(id string) error {
	return fmt.Errorf("not implemented")
}

type simpleWebhookRepository struct{}

func (r *simpleWebhookRepository) List(page, pageSize int, filters map[string]string) ([]models.Webhook, int, error) {
	return []models.Webhook{}, 0, nil
}

func (r *simpleWebhookRepository) Get(id string) (*models.Webhook, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *simpleWebhookRepository) Create(item *models.Webhook) error {
	return fmt.Errorf("not implemented")
}

func (r *simpleWebhookRepository) Update(id string, item *models.Webhook) error {
	return fmt.Errorf("not implemented")
}

func (r *simpleWebhookRepository) Delete(id string) error {
	return fmt.Errorf("not implemented")
}
