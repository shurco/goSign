package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"

	"github.com/shurco/gosign/internal/models"
)

type fakeSubmitterRepo struct {
	bySlug map[string]*models.Submitter
}

func (f *fakeSubmitterRepo) GetSubmitterBySlug(_ context.Context, slug string) (*models.Submitter, error) {
	submitter, ok := f.bySlug[slug]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return submitter, nil
}

func TestEmbedHandler_GetEmbedConfig(t *testing.T) {
	slug := "test-slug"
	repo := &fakeSubmitterRepo{
		bySlug: map[string]*models.Submitter{
			slug: {
				ID:     "submitter-1",
				Slug:   slug,
				Status: models.SubmitterStatusPending,
			},
		},
	}

	app := fiber.New()
	handler := NewEmbedHandler(repo)
	handler.RegisterRoutes(app)

	req := httptest.NewRequest(http.MethodGet, "/embed/"+slug+"/config", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error: %v", err)
	}

	var envelope struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if envelope.Data["slug"] != slug {
		t.Fatalf("slug = %v, want %s", envelope.Data["slug"], slug)
	}
	if envelope.Data["embed_url"] != "/embed/"+slug {
		t.Fatalf("embed_url = %v", envelope.Data["embed_url"])
	}
}

func TestEmbedHandler_GetEmbedPage_NotFound(t *testing.T) {
	repo := &fakeSubmitterRepo{bySlug: map[string]*models.Submitter{}}

	app := fiber.New()
	handler := NewEmbedHandler(repo)
	handler.RegisterRoutes(app)

	req := httptest.NewRequest(http.MethodGet, "/embed/missing-slug", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestEmbedHandler_GetEmbedPage_Completed(t *testing.T) {
	slug := "done-slug"
	repo := &fakeSubmitterRepo{
		bySlug: map[string]*models.Submitter{
			slug: {
				ID:     "submitter-2",
				Slug:   slug,
				Status: models.SubmitterStatusCompleted,
			},
		},
	}

	app := fiber.New()
	handler := NewEmbedHandler(repo)
	handler.RegisterRoutes(app)

	req := httptest.NewRequest(http.MethodGet, "/embed/"+slug, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusGone {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusGone)
	}
}
