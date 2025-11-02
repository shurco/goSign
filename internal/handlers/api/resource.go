package api

import (
	"database/sql"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/pkg/utils/webutil"
)

// ResourceHandler generic CRUD handler for resources
type ResourceHandler[T any] struct {
	resourceName string
	repository   ResourceRepository[T]
}

// ResourceRepository interface for working with resources
type ResourceRepository[T any] interface {
	List(page, pageSize int, filters map[string]string) ([]T, int, error)
	Get(id string) (*T, error)
	Create(item *T) error
	Update(id string, item *T) error
	Delete(id string) error
}

// NewResourceHandler creates new generic handler
func NewResourceHandler[T any](resourceName string, repo ResourceRepository[T]) *ResourceHandler[T] {
	return &ResourceHandler[T]{
		resourceName: resourceName,
		repository:   repo,
	}
}

// List returns paginated list of resources
// @Summary List resources
// @Tags {resourceName}
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} map[string]any
// @Router /{resourceName} [get]
func (h *ResourceHandler[T]) List(c *fiber.Ctx) error {
	// Parse pagination parameters
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 20)

	// Limit page size
	if pageSize > 100 {
		pageSize = 100
	}
	if page < 1 {
		page = 1
	}

	// Collect filters
	filters := make(map[string]string)
	c.Context().QueryArgs().VisitAll(func(key, value []byte) {
		keyStr := string(key)
		// Skip pagination parameters
		if keyStr != "page" && keyStr != "page_size" {
			filters[keyStr] = string(value)
		}
	})

	// Get data from repository
	items, total, err := h.repository.List(page, pageSize, filters)
	if err != nil {
		log.Error().Err(err).Str("resource", h.resourceName).Msg("Failed to list resources")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to retrieve "+h.resourceName, nil)
	}

	return webutil.Response(c, fiber.StatusOK, h.resourceName, map[string]any{
		"items":      items,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_pages": (total + pageSize - 1) / pageSize,
	})
}

// Get returns resource by ID
// @Summary Get resource by ID
// @Tags {resourceName}
// @Param id path string true "Resource ID"
// @Success 200 {object} T
// @Failure 404 {object} map[string]any
// @Router /{resourceName}/{id} [get]
func (h *ResourceHandler[T]) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "ID is required", nil)
	}

	item, err := h.repository.Get(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return webutil.Response(c, fiber.StatusNotFound, h.resourceName+" not found", nil)
		}
		log.Error().Err(err).Str("resource", h.resourceName).Str("id", id).Msg("Failed to get resource")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to retrieve "+h.resourceName, nil)
	}

	return webutil.Response(c, fiber.StatusOK, h.resourceName, item)
}

// Create creates new resource
// @Summary Create resource
// @Tags {resourceName}
// @Accept json
// @Param body body T true "Resource data"
// @Success 201 {object} T
// @Failure 400 {object} map[string]any
// @Router /{resourceName} [post]
func (h *ResourceHandler[T]) Create(c *fiber.Ctx) error {
	var item T
	if err := c.BodyParser(&item); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if err := h.repository.Create(&item); err != nil {
		log.Error().Err(err).Str("resource", h.resourceName).Msg("Failed to create resource")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create "+h.resourceName, nil)
	}

	return webutil.Response(c, fiber.StatusCreated, h.resourceName, item)
}

// Update updates resource
// @Summary Update resource
// @Tags {resourceName}
// @Accept json
// @Param id path string true "Resource ID"
// @Param body body T true "Resource data"
// @Success 200 {object} T
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Router /{resourceName}/{id} [put]
func (h *ResourceHandler[T]) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "ID is required", nil)
	}

	var item T
	if err := c.BodyParser(&item); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if err := h.repository.Update(id, &item); err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return webutil.Response(c, fiber.StatusNotFound, h.resourceName+" not found", nil)
		}
		log.Error().Err(err).Str("resource", h.resourceName).Str("id", id).Msg("Failed to update resource")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to update "+h.resourceName, nil)
	}

	return webutil.Response(c, fiber.StatusOK, h.resourceName, item)
}

// Delete deletes resource
// @Summary Delete resource
// @Tags {resourceName}
// @Param id path string true "Resource ID"
// @Success 204 "No content"
// @Failure 404 {object} map[string]any
// @Router /{resourceName}/{id} [delete]
func (h *ResourceHandler[T]) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "ID is required", nil)
	}

	if err := h.repository.Delete(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return webutil.Response(c, fiber.StatusNotFound, h.resourceName+" not found", nil)
		}
		log.Error().Err(err).Str("resource", h.resourceName).Str("id", id).Msg("Failed to delete resource")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to delete "+h.resourceName, nil)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// RegisterRoutes registers all CRUD routes for resource
func (h *ResourceHandler[T]) RegisterRoutes(router fiber.Router) {
	router.Get("/", h.List)
	router.Get("/:id", h.Get)
	router.Post("/", h.Create)
	router.Put("/:id", h.Update)
	router.Delete("/:id", h.Delete)
}

