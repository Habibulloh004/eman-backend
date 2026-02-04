package handlers

import (
	"eman-backend/services"

	"github.com/gofiber/fiber/v2"
)

type UploadHandler struct {
	storage *services.StorageService
}

func NewUploadHandler(storage *services.StorageService) *UploadHandler {
	return &UploadHandler{storage: storage}
}

// Upload handles single file upload
func (h *UploadHandler) Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "No file uploaded",
		})
	}

	relativePath, err := h.storage.UploadFile(file)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"url":     "/uploads/" + relativePath,
		"path":    relativePath,
	})
}

// UploadMultiple handles multiple files upload
func (h *UploadHandler) UploadMultiple(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid form data",
		})
	}

	files := form.File["files"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "No files uploaded",
		})
	}

	var uploaded []fiber.Map
	for _, file := range files {
		relativePath, err := h.storage.UploadFile(file)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
				"file":    file.Filename,
			})
		}

		uploaded = append(uploaded, fiber.Map{
			"url":           "/uploads/" + relativePath,
			"path":          relativePath,
			"original_name": file.Filename,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"files":   uploaded,
		"count":   len(uploaded),
	})
}
