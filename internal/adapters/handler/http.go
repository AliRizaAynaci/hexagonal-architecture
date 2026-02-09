package handler

import (
	"hexagonal/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

// HTTPHandler - Web isteklerini karşılayan adaptör.
// Service katmanını çağırır, JSON verisini parse eder.
type HTTPHandler struct {
	svc ports.ConcertService
}

func NewHTTPHandler(svc ports.ConcertService) *HTTPHandler {
	return &HTTPHandler{
		svc: svc,
	}
}

func (h *HTTPHandler) CreateConcert(c *fiber.Ctx) error {
	// İstekten gelen JSON gövdesi için bir DTO (Data Transfer Object) tanımladık.
	// Neden: Domain nesnesini doğrudan dışarı açmak istemeyiz (Security & Decoupling).
	var body struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Capacity int    `json:"capacity"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid bodu"})
	}

	// Service katmanini cagir
	concert, err := h.svc.CreateConcert(body.ID, body.Name, body.Capacity)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(200).JSON(concert)
}

func (h *HTTPHandler) BuyTicket(c *fiber.Ctx) error {
	var body struct {
		ConcertID string `json:"concert_id"`
		Quantity  int    `json:"quantity"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	err := h.svc.BuyTicket(body.ConcertID, body.Quantity)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(200).JSON(fiber.Map{"message": "ticket bought successfully"})
}
