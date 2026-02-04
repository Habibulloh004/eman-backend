package handlers

import (
	"eman-backend/services"

	"github.com/gofiber/fiber/v2"
)

type EstateHandler struct {
	macroService *services.MacroService
}

func NewEstateHandler(macroService *services.MacroService) *EstateHandler {
	return &EstateHandler{macroService: macroService}
}

// GetComplexes возвращает список жилых комплексов
func (h *EstateHandler) GetComplexes(c *fiber.Ctx) error {
	data, err := h.macroService.GetComplexes()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	c.Set("Content-Type", "application/json")
	return c.Send(data)
}

// GetEstates возвращает список квартир с фильтрацией
func (h *EstateHandler) GetEstates(c *fiber.Ctx) error {
	params := make(map[string]string)

	// Поддерживаемые query параметры для фильтрации
	queryParams := []string{
		"type",       // living, commercial, parking
		"activity",   // sell, rent
		"category",   // flat, house, etc
		"limit",      // количество записей
		"offset",     // смещение
		"rooms",      // количество комнат
		"price_from", // цена от
		"price_to",   // цена до
		"area_from",  // площадь от
		"area_to",    // площадь до
		"floor_from", // этаж от
		"floor_to",   // этаж до
	}

	for _, param := range queryParams {
		if value := c.Query(param); value != "" {
			params[param] = value
		}
	}

	data, err := h.macroService.GetEstates(params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	c.Set("Content-Type", "application/json")
	return c.Send(data)
}
