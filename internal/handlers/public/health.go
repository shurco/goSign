package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"

	workerv1 "github.com/shurco/gosign/internal/rpc/workerv1"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// workerStatusTimeout keeps /health fast even when the worker is unreachable.
const workerStatusTimeout = 500 * time.Millisecond

// HealthHandler reports API liveness and, best-effort, worker status via gRPC.
type HealthHandler struct {
	worker workerv1.WorkerServiceClient
}

// NewHealthHandler creates a health handler; worker may be nil (no client).
func NewHealthHandler(worker workerv1.WorkerServiceClient) *HealthHandler {
	return &HealthHandler{worker: worker}
}

// Health responds 200 while the API is alive; worker outage does not fail it.
func (h *HealthHandler) Health(c fiber.Ctx) error {
	worker := "unavailable"
	if h.worker != nil {
		ctx, cancel := context.WithTimeout(c.Context(), workerStatusTimeout)
		defer cancel()
		if _, err := h.worker.Status(ctx, &workerv1.StatusRequest{}); err == nil {
			worker = "ok"
		}
	}
	return webutil.Response(c, fiber.StatusOK, "Pong", fiber.Map{"worker": worker})
}
