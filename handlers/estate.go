package handlers

import (
	"eman-backend/services"
	"net/url"

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
	params := make(url.Values)

	// Поддерживаемые query параметры для фильтрации
	queryParams := []string{
		"type",       // living, commercial, parking
		"activity",   // sell, rent
		"category",   // flat, house, etc
		"limit",      // количество записей
		"offset",     // смещение
		"rooms",      // количество комнат
		"floor",      // конкретные этажи (повторяемый параметр)
		"price_from", // цена от
		"price_to",   // цена до
		"area_from",  // площадь от
		"area_to",    // площадь до
		"floor_from", // этаж от
		"floor_to",   // этаж до
	}

	// Preserve repeated params such as ?rooms=2&rooms=3, ?floor=4&floor=5
	c.Context().QueryArgs().VisitAll(func(key, value []byte) {
		k := string(key)
		for _, supported := range queryParams {
			if k == supported {
				params[k] = append(params[k], string(value))
				break
			}
		}
	})

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
