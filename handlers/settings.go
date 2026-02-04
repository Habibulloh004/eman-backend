package handlers

import (
	"eman-backend/database"
	"eman-backend/models"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type SettingsHandler struct {
	cache     map[string][]models.SiteSetting
	cacheMu   sync.RWMutex
	cacheTime time.Time
	cacheTTL  time.Duration
}

func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{
		cache:    make(map[string][]models.SiteSetting),
		cacheTTL: 5 * time.Minute, // Cache for 5 minutes
	}
}

// clearCache invalidates the cache
func (h *SettingsHandler) clearCache() {
	h.cacheMu.Lock()
	defer h.cacheMu.Unlock()
	h.cache = make(map[string][]models.SiteSetting)
	h.cacheTime = time.Time{}
}

// GetPublic returns all settings grouped by category (public endpoint)
func (h *SettingsHandler) GetPublic(c *fiber.Ctx) error {
	h.cacheMu.RLock()
	if len(h.cache) > 0 && time.Since(h.cacheTime) < h.cacheTTL {
		result := h.cache
		h.cacheMu.RUnlock()
		return c.JSON(result)
	}
	h.cacheMu.RUnlock()

	var settings []models.SiteSetting
	if err := database.DB.Find(&settings).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to fetch settings",
		})
	}

	// Group by category
	grouped := make(map[string][]models.SiteSetting)
	for _, s := range settings {
		grouped[s.Category] = append(grouped[s.Category], s)
	}

	// Update cache
	h.cacheMu.Lock()
	h.cache = grouped
	h.cacheTime = time.Now()
	h.cacheMu.Unlock()

	return c.JSON(grouped)
}

// GetByCategory returns settings for a specific category (public endpoint)
func (h *SettingsHandler) GetByCategory(c *fiber.Ctx) error {
	category := c.Params("category")

	var settings []models.SiteSetting
	if err := database.DB.Where("category = ?", category).Find(&settings).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to fetch settings",
		})
	}

	// Convert to key-value map for easier frontend usage
	result := make(map[string]interface{})
	for _, s := range settings {
		result[s.Key] = s.Value
	}

	return c.JSON(result)
}

// List returns all settings (admin endpoint)
func (h *SettingsHandler) List(c *fiber.Ctx) error {
	var settings []models.SiteSetting

	query := database.DB.Order("category, key")

	// Filter by category if provided
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Find(&settings).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to fetch settings",
		})
	}

	return c.JSON(settings)
}

// Get returns a single setting by key (admin endpoint)
func (h *SettingsHandler) Get(c *fiber.Ctx) error {
	key := c.Params("key")

	var setting models.SiteSetting
	if err := database.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Setting not found",
		})
	}

	return c.JSON(setting)
}

type UpdateSettingRequest struct {
	Value string `json:"value"`
}

// Update modifies a setting value (admin endpoint)
func (h *SettingsHandler) Update(c *fiber.Ctx) error {
	key := c.Params("key")

	var setting models.SiteSetting
	if err := database.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Setting not found",
		})
	}

	var req UpdateSettingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	setting.Value = req.Value
	if err := database.DB.Save(&setting).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to update setting",
		})
	}

	// Clear cache
	h.clearCache()

	return c.JSON(setting)
}

type BulkUpdateRequest struct {
	Settings []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"settings"`
}

// BulkUpdate modifies multiple settings at once (admin endpoint)
func (h *SettingsHandler) BulkUpdate(c *fiber.Ctx) error {
	var req BulkUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	tx := database.DB.Begin()

	for _, item := range req.Settings {
		result := tx.Model(&models.SiteSetting{}).
			Where("key = ?", item.Key).
			Update("value", item.Value)

		if result.Error != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to update settings",
			})
		}
	}

	tx.Commit()

	// Clear cache
	h.clearCache()

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Settings updated",
		"count":   len(req.Settings),
	})
}

// Seed resets settings to defaults (admin endpoint)
func (h *SettingsHandler) Seed(c *fiber.Ctx) error {
	// Delete all existing settings
	if err := database.DB.Exec("DELETE FROM site_settings").Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to clear settings",
		})
	}

	// Re-seed defaults
	defaults := models.DefaultSettings()
	for _, setting := range defaults {
		if err := database.DB.Create(&setting).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to seed settings",
			})
		}
	}

	// Clear cache
	h.clearCache()

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Settings reset to defaults",
		"count":   len(defaults),
	})
}

// GetCategories returns all available setting categories
func (h *SettingsHandler) GetCategories(c *fiber.Ctx) error {
	categories := []fiber.Map{
		{"key": models.CategoryContact, "label": "Контакты", "label_uz": "Kontaktlar"},
		{"key": models.CategorySocial, "label": "Соц. сети", "label_uz": "Ijtimoiy tarmoqlar"},
		{"key": models.CategoryPricing, "label": "Цены", "label_uz": "Narxlar"},
		{"key": models.CategoryFAQ, "label": "FAQ", "label_uz": "FAQ"},
		{"key": models.CategoryFeatures, "label": "Особенности", "label_uz": "Xususiyatlar"},
		{"key": models.CategoryContent, "label": "Контент", "label_uz": "Kontent"},
	}

	return c.JSON(categories)
}
