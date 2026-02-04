package handlers

import (
	"eman-backend/database"
	"eman-backend/models"
	"eman-backend/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ProjectsHandler struct {
	storage *services.StorageService
}

func NewProjectsHandler(storage *services.StorageService) *ProjectsHandler {
	return &ProjectsHandler{storage: storage}
}

// List returns all projects (admin)
func (h *ProjectsHandler) List(c *fiber.Ctx) error {
	var items []models.Project

	query := database.DB.Order("sort_order ASC, created_at DESC")

	if err := query.Find(&items).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to fetch projects",
		})
	}

	return c.JSON(fiber.Map{
		"items": items,
		"total": len(items),
	})
}

// ListPublic returns only published projects (public)
func (h *ProjectsHandler) ListPublic(c *fiber.Ctx) error {
	var items []models.Project

	query := database.DB.Where("is_published = ?", true).Order("sort_order ASC, created_at DESC")

	if err := query.Find(&items).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to fetch projects",
		})
	}

	return c.JSON(fiber.Map{
		"items": items,
		"total": len(items),
	})
}

// Get returns a single project
func (h *ProjectsHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	var item models.Project
	if err := database.DB.First(&item, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Project not found",
		})
	}

	return c.JSON(item)
}

type CreateProjectRequest struct {
	TypeRu        string `json:"type_ru"`
	TypeUz        string `json:"type_uz"`
	AreaRu        string `json:"area_ru"`
	AreaUz        string `json:"area_uz"`
	DescriptionRu string `json:"description_ru"`
	DescriptionUz string `json:"description_uz"`
	Image         string `json:"image"`
	SortOrder     int    `json:"sort_order"`
	IsPublished   bool   `json:"is_published"`
}

// Create adds a new project
func (h *ProjectsHandler) Create(c *fiber.Ctx) error {
	var req CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	// Auto-increment sort_order
	var maxSortOrder int
	database.DB.Model(&models.Project{}).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxSortOrder)

	item := models.Project{
		TypeRu:        req.TypeRu,
		TypeUz:        req.TypeUz,
		AreaRu:        req.AreaRu,
		AreaUz:        req.AreaUz,
		DescriptionRu: req.DescriptionRu,
		DescriptionUz: req.DescriptionUz,
		Image:         req.Image,
		SortOrder:     maxSortOrder + 1,
		IsPublished:   req.IsPublished,
	}

	if err := database.DB.Create(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to create project",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

// Update modifies an existing project
func (h *ProjectsHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	var item models.Project
	if err := database.DB.First(&item, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Project not found",
		})
	}

	var req CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	item.TypeRu = req.TypeRu
	item.TypeUz = req.TypeUz
	item.AreaRu = req.AreaRu
	item.AreaUz = req.AreaUz
	item.DescriptionRu = req.DescriptionRu
	item.DescriptionUz = req.DescriptionUz
	item.Image = req.Image
	item.SortOrder = req.SortOrder
	item.IsPublished = req.IsPublished

	if err := database.DB.Save(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to update project",
		})
	}

	return c.JSON(item)
}

// Delete removes a project
func (h *ProjectsHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	result := database.DB.Delete(&models.Project{}, id)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to delete project",
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Project not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Project deleted",
	})
}

// Upload handles file upload for project images
func (h *ProjectsHandler) Upload(c *fiber.Ctx) error {
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
