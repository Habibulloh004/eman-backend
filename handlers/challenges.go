package handlers

import (
	"eman-backend/database"
	"eman-backend/models"
	"eman-backend/services"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ChallengesHandler struct {
	storage *services.StorageService
}

func NewChallengesHandler(storage *services.StorageService) *ChallengesHandler {
	return &ChallengesHandler{storage: storage}
}

// ============ ADMIN ENDPOINTS ============

// List returns all challenges (admin)
func (h *ChallengesHandler) List(c *fiber.Ctx) error {
	var items []models.Challenge
	query := database.DB.Order("sort_order ASC, created_at DESC")

	if err := query.Find(&items).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to fetch challenges",
		})
	}

	return c.JSON(fiber.Map{
		"items": items,
		"total": len(items),
	})
}

// Get returns a single challenge (admin)
func (h *ChallengesHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	var item models.Challenge
	if err := database.DB.First(&item, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Challenge not found",
		})
	}

	return c.JSON(item)
}

type CreateChallengeRequest struct {
	Title         string `json:"title"`
	TitleUz       string `json:"title_uz"`
	Description   string `json:"description"`
	DescriptionUz string `json:"description_uz"`
	Image         string `json:"image"`
	Prize         string `json:"prize"`
	PrizeUz       string `json:"prize_uz"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	MaxUsers      int    `json:"max_users"`
	SortOrder     int    `json:"sort_order"`
	IsPublished   bool   `json:"is_published"`
}

// Create adds a new challenge (admin)
func (h *ChallengesHandler) Create(c *fiber.Ctx) error {
	var req CreateChallengeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	startDate, _ := time.Parse(time.RFC3339, req.StartDate)
	endDate, _ := time.Parse(time.RFC3339, req.EndDate)

	item := models.Challenge{
		Title:         req.Title,
		TitleUz:       req.TitleUz,
		Description:   req.Description,
		DescriptionUz: req.DescriptionUz,
		Image:         req.Image,
		Prize:         req.Prize,
		PrizeUz:       req.PrizeUz,
		StartDate:     startDate,
		EndDate:       endDate,
		MaxUsers:      req.MaxUsers,
		SortOrder:     req.SortOrder,
		IsPublished:   req.IsPublished,
	}

	if err := database.DB.Create(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to create challenge",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

// Update modifies an existing challenge (admin)
func (h *ChallengesHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	var item models.Challenge
	if err := database.DB.First(&item, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Challenge not found",
		})
	}

	var req CreateChallengeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	startDate, _ := time.Parse(time.RFC3339, req.StartDate)
	endDate, _ := time.Parse(time.RFC3339, req.EndDate)

	item.Title = req.Title
	item.TitleUz = req.TitleUz
	item.Description = req.Description
	item.DescriptionUz = req.DescriptionUz
	item.Image = req.Image
	item.Prize = req.Prize
	item.PrizeUz = req.PrizeUz
	item.StartDate = startDate
	item.EndDate = endDate
	item.MaxUsers = req.MaxUsers
	item.SortOrder = req.SortOrder
	item.IsPublished = req.IsPublished

	if err := database.DB.Save(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to update challenge",
		})
	}

	return c.JSON(item)
}

// Delete removes a challenge (admin)
func (h *ChallengesHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	result := database.DB.Delete(&models.Challenge{}, id)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to delete challenge",
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Challenge not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Challenge deleted",
	})
}

// Upload handles image upload for challenge
func (h *ChallengesHandler) Upload(c *fiber.Ctx) error {
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

// Participants returns all participants of a challenge (admin)
func (h *ChallengesHandler) Participants(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	var participants []models.ChallengeParticipant
	if err := database.DB.Where("challenge_id = ?", id).Order("joined_at DESC").Find(&participants).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to fetch participants",
		})
	}

	return c.JSON(fiber.Map{
		"items": participants,
		"total": len(participants),
	})
}

// ============ PUBLIC ENDPOINTS ============

// ListPublic returns only published, active challenges
func (h *ChallengesHandler) ListPublic(c *fiber.Ctx) error {
	var items []models.Challenge
	now := time.Now()

	query := database.DB.Where("is_published = ? AND end_date > ?", true, now).Order("sort_order ASC, start_date ASC")

	if err := query.Find(&items).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to fetch challenges",
		})
	}

	// Add participant count for each challenge
	type result struct {
		Items []map[string]interface{} `json:"items"`
		Total int                      `json:"total"`
	}

	var response []map[string]interface{}
	for _, item := range items {
		var count int64
		database.DB.Model(&models.ChallengeParticipant{}).Where("challenge_id = ? AND status = ?", item.ID, "active").Count(&count)

		response = append(response, map[string]interface{}{
			"id":             item.ID,
			"title":          item.Title,
			"title_uz":       item.TitleUz,
			"description":    item.Description,
			"description_uz": item.DescriptionUz,
			"image":          item.Image,
			"prize":          item.Prize,
			"prize_uz":       item.PrizeUz,
			"start_date":     item.StartDate,
			"end_date":       item.EndDate,
			"max_users":      item.MaxUsers,
			"participants":   count,
			"created_at":     item.CreatedAt,
		})
	}

	return c.JSON(fiber.Map{
		"items": response,
		"total": len(response),
	})
}

// Join allows a user to join a challenge
func (h *ChallengesHandler) Join(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	var req struct {
		Phone string `json:"phone"`
		Name  string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Phone is required",
		})
	}

	// Check challenge exists and is active
	var challenge models.Challenge
	if err := database.DB.First(&challenge, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Challenge not found",
		})
	}

	now := time.Now()
	if !challenge.IsPublished || now.Before(challenge.StartDate) || now.After(challenge.EndDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Challenge is not active",
		})
	}

	// Check max users limit
	if challenge.MaxUsers > 0 {
		var count int64
		database.DB.Model(&models.ChallengeParticipant{}).Where("challenge_id = ? AND status = ?", challenge.ID, "active").Count(&count)
		if int(count) >= challenge.MaxUsers {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Challenge is full",
			})
		}
	}

	// KEY RULE: Check if user already has an active challenge
	var activeParticipation models.ChallengeParticipant
	err = database.DB.Where("phone = ? AND status = ?", phone, "active").First(&activeParticipation).Error
	if err == nil {
		// User already has an active challenge
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":            true,
			"message":          "You already have an active challenge. Complete or cancel it first.",
			"active_challenge": activeParticipation.ChallengeID,
		})
	}

	// Check if already joined this specific challenge (any status)
	var existingParticipant models.ChallengeParticipant
	err = database.DB.Where("challenge_id = ? AND phone = ?", challenge.ID, phone).First(&existingParticipant).Error
	if err == nil {
		if existingParticipant.Status == "completed" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":   true,
				"message": "You have already completed this challenge",
			})
		}
		if existingParticipant.Status == "cancelled" {
			// Re-join: reactivate
			existingParticipant.Status = "active"
			existingParticipant.JoinedAt = now
			existingParticipant.CompletedAt = nil
			database.DB.Save(&existingParticipant)
			return c.JSON(fiber.Map{
				"success":        true,
				"message":        "Rejoined challenge",
				"participation":  existingParticipant,
			})
		}
	}

	participant := models.ChallengeParticipant{
		ChallengeID: uint(id),
		Phone:       phone,
		Name:        strings.TrimSpace(req.Name),
		Status:      "active",
		JoinedAt:    now,
	}

	if err := database.DB.Create(&participant).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to join challenge",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success":       true,
		"message":       "Joined challenge",
		"participation": participant,
	})
}

// MyChallenge returns the user's active challenge by phone
func (h *ChallengesHandler) MyChallenge(c *fiber.Ctx) error {
	phone := strings.TrimSpace(c.Query("phone"))
	if phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Phone is required",
		})
	}

	var participant models.ChallengeParticipant
	err := database.DB.Preload("Challenge").Where("phone = ? AND status = ?", phone, "active").First(&participant).Error
	if err != nil {
		return c.JSON(fiber.Map{
			"active": false,
		})
	}

	return c.JSON(fiber.Map{
		"active":        true,
		"participation": participant,
	})
}

// Cancel allows a user to cancel their active challenge
func (h *ChallengesHandler) Cancel(c *fiber.Ctx) error {
	var req struct {
		Phone string `json:"phone"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Phone is required",
		})
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	var participant models.ChallengeParticipant
	if err := database.DB.Where("challenge_id = ? AND phone = ? AND status = ?", id, phone, "active").First(&participant).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Active participation not found",
		})
	}

	participant.Status = "cancelled"
	now := time.Now()
	participant.CompletedAt = &now

	if err := database.DB.Save(&participant).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to cancel participation",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Challenge cancelled",
	})
}
