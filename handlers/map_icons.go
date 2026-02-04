package handlers

import (
	"eman-backend/database"
	"eman-backend/models"
	"eman-backend/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type MapIconTypeHandler struct {
	storage *services.StorageService
}

func NewMapIconTypeHandler(storage *services.StorageService) *MapIconTypeHandler {
	return &MapIconTypeHandler{storage: storage}
}

type MapIconHandler struct{}

func NewMapIconHandler() *MapIconHandler {
	return &MapIconHandler{}
}

func hydrateMapIconTypeNames(item *models.MapIconType) {
	if item == nil {
		return
	}
	if item.NameRu == "" && item.Name != "" {
		item.NameRu = item.Name
	}
	if item.NameUz == "" && item.Name != "" {
		item.NameUz = item.Name
	}
}

func hydrateMapIconNames(item *models.MapIcon) {
	if item == nil {
		return
	}
	if item.NameRu == "" && item.Name != "" {
		item.NameRu = item.Name
	}
	if item.NameUz == "" && item.Name != "" {
		item.NameUz = item.Name
	}
	hydrateMapIconTypeNames(&item.Type)
}

// ===== Map Icon Types (admin) =====

func (h *MapIconTypeHandler) List(c *fiber.Ctx) error {
	var items []models.MapIconType
	if err := database.DB.Order("created_at DESC").Find(&items).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to fetch map icon types",
		})
	}

	for i := range items {
		hydrateMapIconTypeNames(&items[i])
	}

	return c.JSON(fiber.Map{
		"items": items,
		"total": len(items),
	})
}

type CreateMapIconTypeRequest struct {
	Name   string `json:"name"`
	NameRu string `json:"name_ru"`
	NameUz string `json:"name_uz"`
	Icon   string `json:"icon"`
}

func (h *MapIconTypeHandler) Create(c *fiber.Ctx) error {
	var req CreateMapIconTypeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	if req.NameRu == "" && req.NameUz == "" && req.Name != "" {
		req.NameRu = req.Name
		req.NameUz = req.Name
	}

	if req.NameRu == "" || req.NameUz == "" || req.Icon == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Name (RU), name (UZ) and icon are required",
		})
	}

	item := models.MapIconType{
		Name:   req.Name,
		NameRu: req.NameRu,
		NameUz: req.NameUz,
		Icon:   req.Icon,
	}

	if item.Name == "" {
		item.Name = item.NameRu
	}

	if err := database.DB.Create(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to create map icon type",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *MapIconTypeHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	var item models.MapIconType
	if err := database.DB.First(&item, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Map icon type not found",
		})
	}

	var req CreateMapIconTypeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	if req.NameRu == "" && req.NameUz == "" && req.Name != "" {
		req.NameRu = req.Name
		req.NameUz = req.Name
	}

	if req.NameRu == "" || req.NameUz == "" || req.Icon == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Name (RU), name (UZ) and icon are required",
		})
	}

	item.Name = req.Name
	item.NameRu = req.NameRu
	item.NameUz = req.NameUz
	item.Icon = req.Icon

	if item.Name == "" {
		item.Name = item.NameRu
	}

	if err := database.DB.Save(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to update map icon type",
		})
	}

	return c.JSON(item)
}

func (h *MapIconTypeHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	var count int64
	if err := database.DB.Model(&models.MapIcon{}).Where("type_id = ?", id).Count(&count).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to check map icon usage",
		})
	}

	if count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Map icon type is used by existing markers",
		})
	}

	result := database.DB.Delete(&models.MapIconType{}, id)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to delete map icon type",
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Map icon type not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Map icon type deleted",
	})
}

// Upload handles icon upload
func (h *MapIconTypeHandler) Upload(c *fiber.Ctx) error {
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

// ===== Map Icons (markers) =====

func (h *MapIconHandler) List(c *fiber.Ctx) error {
	var items []models.MapIcon
	if err := database.DB.Preload("Type").Order("created_at DESC").Find(&items).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to fetch map icons",
		})
	}

	for i := range items {
		hydrateMapIconNames(&items[i])
	}

	return c.JSON(fiber.Map{
		"items": items,
		"total": len(items),
	})
}

func (h *MapIconHandler) ListPublic(c *fiber.Ctx) error {
	return h.List(c)
}

type CreateMapIconRequest struct {
	Name      string  `json:"name"`
	NameRu    string  `json:"name_ru"`
	NameUz    string  `json:"name_uz"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
	TypeID    uint    `json:"type_id"`
}

func (h *MapIconHandler) Create(c *fiber.Ctx) error {
	var req CreateMapIconRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	if req.NameRu == "" && req.NameUz == "" && req.Name != "" {
		req.NameRu = req.Name
		req.NameUz = req.Name
	}

	if req.NameRu == "" || req.NameUz == "" || req.TypeID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Name (RU), name (UZ), type, latitude and longitude are required",
		})
	}

	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid coordinates",
		})
	}

	item := models.MapIcon{
		Name:      req.Name,
		NameRu:    req.NameRu,
		NameUz:    req.NameUz,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		TypeID:    req.TypeID,
	}

	if item.Name == "" {
		item.Name = item.NameRu
	}

	if err := database.DB.Create(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to create map icon",
		})
	}

	if err := database.DB.Preload("Type").First(&item, item.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to load map icon",
		})
	}

	hydrateMapIconNames(&item)

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *MapIconHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	var item models.MapIcon
	if err := database.DB.First(&item, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Map icon not found",
		})
	}

	var req CreateMapIconRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	if req.NameRu == "" && req.NameUz == "" && req.Name != "" {
		req.NameRu = req.Name
		req.NameUz = req.Name
	}

	if req.NameRu == "" || req.NameUz == "" || req.TypeID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Name (RU), name (UZ), type, latitude and longitude are required",
		})
	}

	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid coordinates",
		})
	}

	item.Name = req.Name
	item.NameRu = req.NameRu
	item.NameUz = req.NameUz
	item.Latitude = req.Latitude
	item.Longitude = req.Longitude
	item.TypeID = req.TypeID

	if item.Name == "" {
		item.Name = item.NameRu
	}

	if err := database.DB.Save(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to update map icon",
		})
	}

	if err := database.DB.Preload("Type").First(&item, item.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to load map icon",
		})
	}

	hydrateMapIconNames(&item)

	return c.JSON(item)
}

func (h *MapIconHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	result := database.DB.Delete(&models.MapIcon{}, id)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to delete map icon",
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Map icon not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Map icon deleted",
	})
}
