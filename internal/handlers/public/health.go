package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/shurco/gosign/pkg/utils/fsutil"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

func Health(c *fiber.Ctx) error {
	return webutil.Response(c, fiber.StatusOK, "Pong", nil)
}

func Upload(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return webutil.StatusBadRequest(c, err.Error())
	}

	validMIMETypes := map[string]bool{
		"application/pdf": true,
		"image/png":       true,
		"image/jpeg":      true,
		"image/tiff":      true,
		//"application/msword":       true,
		//"application/vnd.ms-excel": true,
		//"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		//"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       true,
	}

	for _, fileHeaders := range form.File {
		for _, fileHeader := range fileHeaders {
			if !validMIMETypes[fileHeader.Header.Get("Content-Type")] {
				return webutil.StatusBadRequest(c, "File format not supported")
			}

			fileUUID := uuid.New().String()
			fileExt := fsutil.ExtName(fileHeader.Filename)
			fileName := fmt.Sprintf("%s.%s", fileUUID, fileExt)
			filePath := fmt.Sprintf("./lc_uploads/%s", fileName)

			if err := c.SaveFile(fileHeader, filePath); err != nil {
				return webutil.StatusInternalServerError(c)
			}
		}
	}

	return webutil.Response(c, fiber.StatusOK, "Image added", nil)
}
