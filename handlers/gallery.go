package handlers

import (
	"eman-backend/database"
	"eman-backend/models"
	"eman-backend/services"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type GalleryHandler struct {
	storage *services.StorageService
}

func NewGalleryHandler(storage *services.StorageService) *GalleryHandler {
	return &GalleryHandler{storage: storage}
}

// List returns all gallery items (admin)
func (h *GalleryHandler) List(c *fiber.Ctx) error {
	var items []models.GalleryItem

	query := database.DB.Order("sort_order ASC, created_at DESC")

	// Filter by category
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	// Filter by type
	if itemType := c.Query("type"); itemType != "" {
		query = query.Where("type = ?", itemType)
	}

	if err := query.Find(&items).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to fetch gallery items",
		})
	}

	return c.JSON(fiber.Map{
		"items": items,
		"total": len(items),
	})
}

// ListPublic returns only published gallery items (public)
func (h *GalleryHandler) ListPublic(c *fiber.Ctx) error {
	var items []models.GalleryItem

	query := database.DB.Where("is_published = ?", true).Order("sort_order ASC, created_at DESC")

	// Filter by category
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	// Filter by type
	if itemType := c.Query("type"); itemType != "" {
		query = query.Where("type = ?", itemType)
	}

	if err := query.Find(&items).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to fetch gallery items",
		})
	}

	return c.JSON(fiber.Map{
		"items": items,
		"total": len(items),
	})
}

// Get returns a single gallery item
func (h *GalleryHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	var item models.GalleryItem
	if err := database.DB.First(&item, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Gallery item not found",
		})
	}

	return c.JSON(item)
}

type CreateGalleryRequest struct {
	Title         string `json:"title"`
	TitleUz       string `json:"title_uz"`
	Description   string `json:"description"`
	DescriptionUz string `json:"description_uz"`
	Type          string `json:"type"`
	URL           string `json:"url"`
	RedirectURL   string `json:"redirect_url"`
	Thumbnail     string `json:"thumbnail"`
	Category      string `json:"category"`
	SortOrder     int    `json:"sort_order"`
	IsPublished   bool   `json:"is_published"`
}

// Create adds a new gallery item
func (h *GalleryHandler) Create(c *fiber.Ctx) error {
	var req CreateGalleryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	item := models.GalleryItem{
		Title:         req.Title,
		TitleUz:       req.TitleUz,
		Description:   req.Description,
		DescriptionUz: req.DescriptionUz,
		Type:          req.Type,
		URL:           req.URL,
		RedirectURL:   req.RedirectURL,
		Thumbnail:     req.Thumbnail,
		Category:      req.Category,
		SortOrder:     req.SortOrder,
		IsPublished:   req.IsPublished,
	}

	if err := database.DB.Create(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to create gallery item",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

// Update modifies an existing gallery item
func (h *GalleryHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	var item models.GalleryItem
	if err := database.DB.First(&item, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Gallery item not found",
		})
	}

	var req CreateGalleryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	item.Title = req.Title
	item.TitleUz = req.TitleUz
	item.Description = req.Description
	item.DescriptionUz = req.DescriptionUz
	item.Type = req.Type
	item.URL = req.URL
	item.RedirectURL = req.RedirectURL
	item.Thumbnail = req.Thumbnail
	item.Category = req.Category
	item.SortOrder = req.SortOrder
	item.IsPublished = req.IsPublished

	if err := database.DB.Save(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to update gallery item",
		})
	}

	return c.JSON(item)
}

// Delete removes a gallery item
func (h *GalleryHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	result := database.DB.Delete(&models.GalleryItem{}, id)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to delete gallery item",
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Gallery item not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Gallery item deleted",
	})
}

type ReorderRequest struct {
	Items []struct {
		ID        uint `json:"id"`
		SortOrder int  `json:"sort_order"`
	} `json:"items"`
}

// Reorder updates sort order of multiple items
func (h *GalleryHandler) Reorder(c *fiber.Ctx) error {
	var req ReorderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	for _, item := range req.Items {
		if err := database.DB.Model(&models.GalleryItem{}).Where("id = ?", item.ID).Update("sort_order", item.SortOrder).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to reorder items",
			})
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Items reordered",
	})
}

// Upload handles file upload
func (h *GalleryHandler) Upload(c *fiber.Ctx) error {
	relativePath, err := uploadFromRequest(c, h.storage)
	if err != nil {
		message := err.Error()
		if errors.Is(err, errNoFileUploaded) {
			message = "No file uploaded"
		}

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": message,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"url":     "/uploads/" + relativePath,
		"path":    relativePath,
	})
}
