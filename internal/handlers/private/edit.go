package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/logging"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

func Template(c *fiber.Ctx) error {
	template := "00c95859-98ef-42cd-a801-2023b75a9431"

	response, err := queries.DB.Template(context.Background(), template)
	if err != nil {
		logging.Log.Err(err)
		return err
	}

	return webutil.Response(c, fiber.StatusOK, "Template", response)
}
