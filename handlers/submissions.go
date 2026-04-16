package handlers

import (
	"eman-backend/database"
	"eman-backend/models"
	"eman-backend/services"
	ws "eman-backend/websocket"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type SubmissionsHandler struct {
	hub          *ws.Hub
	macroService *services.MacroService
	telegram     *services.TelegramService
}

func NewSubmissionsHandler(macroService *services.MacroService, telegram *services.TelegramService) *SubmissionsHandler {
	return &SubmissionsHandler{
		hub:          ws.GetHub(),
		macroService: macroService,
		telegram:     telegram,
	}
}

func safeLine(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "-"
	}
	return strings.ReplaceAll(trimmed, "\n", " ")
}

func sourceRu(source string) string {
	switch strings.TrimSpace(source) {
	case "catalog_request":
		return "Запрос из каталога"
	case "contact_page":
		return "Контактная страница"
	case "callback":
		return "Обратный звонок"
	case "question":
		return "Вопрос"
	default:
		return source
	}
}

// List returns all submissions (admin)
func (h *SubmissionsHandler) List(c *fiber.Ctx) error {
	var submissions []models.ContactSubmission

	query := database.DB.Order("created_at DESC")

	// Filter by status
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by source
	if source := c.Query("source"); source != "" {
		query = query.Where("source = ?", source)
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset := (page - 1) * limit

	var total int64
	database.DB.Model(&models.ContactSubmission{}).Count(&total)

	if err := query.Limit(limit).Offset(offset).Find(&submissions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to fetch submissions",
		})
	}

	return c.JSON(fiber.Map{
		"items": submissions,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// Get returns a single submission
func (h *SubmissionsHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	var submission models.ContactSubmission
	if err := database.DB.First(&submission, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Submission not found",
		})
	}

	return c.JSON(submission)
}

type CreateSubmissionRequest struct {
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Message     string `json:"message"`
	Source      string `json:"source"`
	EstateID    *int   `json:"estate_id"`
	PaymentPlan string `json:"payment_plan"`
}

// Create adds a new submission (public endpoint)
func (h *SubmissionsHandler) Create(c *fiber.Ctx) error {
	var req CreateSubmissionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Name == "" || req.Phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Name and phone are required",
		})
	}

	// Default source if not provided
	if req.Source == "" {
		req.Source = "contact_page"
	}

	submission := models.ContactSubmission{
		Name:        req.Name,
		Phone:       req.Phone,
		Email:       req.Email,
		Message:     req.Message,
		Source:      req.Source,
		EstateID:    req.EstateID,
		PaymentPlan: req.PaymentPlan,
		Status:      "new",
		IPAddress:   c.IP(),
		UserAgent:   c.Get("User-Agent"),
	}

	if err := database.DB.Create(&submission).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to create submission",
		})
	}

	// Map source to Macro CRM action
	actionMap := map[string]string{
		"contact_page":    "callback",
		"catalog_request": "buy",
		"callback":        "callback",
		"question":        "question",
	}
	action := actionMap[req.Source]
	if action == "" {
		action = "callback"
	}

	// Forward to MacroCRM (non-blocking, don't fail the user request)
	go func() {
		macroResp, err := h.macroService.SendRequest(action, req.Name, req.Phone, req.Email, req.Message, req.EstateID)
		if err != nil {
			log.Printf("[MacroCRM] Failed to send request for submission #%d: %v", submission.ID, err)
			return
		}
		log.Printf("[MacroCRM] Submission #%d sent successfully, macro estate_id: %d", submission.ID, macroResp.EstateID)
	}()

	// Telegram notification (non-blocking)
	go func() {
		if h.telegram == nil || !h.telegram.Enabled() {
			return
		}

		estateID := "-"
		estateDetails := "-"
		if req.EstateID != nil {
			estateID = strconv.Itoa(*req.EstateID)
			if title := h.macroService.GetEstateTitleByID(*req.EstateID); strings.TrimSpace(title) != "" {
				estateDetails = safeLine(title)
			}
		}

		text := fmt.Sprintf(
			"🔔 Новая заявка\nID: %d\nИсточник: %s\nИмя: %s\nТелефон: %s\nID объекта: %s\nОбъект: %s\nПлан оплаты: %s\nСообщение: %s",
			submission.ID,
			safeLine(sourceRu(req.Source)),
			safeLine(req.Name),
			safeLine(req.Phone),
			estateID,
			estateDetails,
			safeLine(req.PaymentPlan),
			safeLine(req.Message),
		)

		if err := h.telegram.SendMessage(text); err != nil {
			log.Printf("[Telegram] Failed to send notification for submission #%d: %v", submission.ID, err)
			return
		}
		log.Printf("[Telegram] Notification sent for submission #%d", submission.ID)
	}()

	// Broadcast new submission to all connected admin clients
	h.hub.Broadcast("new_submission", fiber.Map{
		"id":           submission.ID,
		"name":         submission.Name,
		"phone":        submission.Phone,
		"source":       submission.Source,
		"estate_id":    submission.EstateID,
		"payment_plan": submission.PaymentPlan,
		"message":      submission.Message,
		"createdAt":    submission.CreatedAt,
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Submission received",
		"id":      submission.ID,
	})
}

type UpdateSubmissionRequest struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

// Update modifies submission status/notes (admin)
func (h *SubmissionsHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	var submission models.ContactSubmission
	if err := database.DB.First(&submission, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Submission not found",
		})
	}

	var req UpdateSubmissionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	// Validate status
	validStatuses := map[string]bool{"new": true, "contacted": true, "closed": true}
	if req.Status != "" && !validStatuses[req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid status. Must be: new, contacted, or closed",
		})
	}

	if req.Status != "" {
		submission.Status = req.Status
	}
	submission.Notes = req.Notes

	if err := database.DB.Save(&submission).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to update submission",
		})
	}

	return c.JSON(submission)
}

// Delete removes a submission
func (h *SubmissionsHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid ID",
		})
	}

	result := database.DB.Delete(&models.ContactSubmission{}, id)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to delete submission",
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Submission not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Submission deleted",
	})
}

// Stats returns submission statistics
func (h *SubmissionsHandler) Stats(c *fiber.Ctx) error {
	var totalCount int64
	var newCount int64
	var contactedCount int64
	var closedCount int64

	database.DB.Model(&models.ContactSubmission{}).Count(&totalCount)
	database.DB.Model(&models.ContactSubmission{}).Where("status = ?", "new").Count(&newCount)
	database.DB.Model(&models.ContactSubmission{}).Where("status = ?", "contacted").Count(&contactedCount)
	database.DB.Model(&models.ContactSubmission{}).Where("status = ?", "closed").Count(&closedCount)

	return c.JSON(fiber.Map{
		"total":     totalCount,
		"new":       newCount,
		"contacted": contactedCount,
		"closed":    closedCount,
	})
}
