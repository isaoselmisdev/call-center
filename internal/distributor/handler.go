package distributor

import (
	"github.com/gofiber/fiber/v2"
)

type DistributorHandler struct {
	service DistributorService
}

func NewDistributorHandler(service DistributorService) *DistributorHandler {
	return &DistributorHandler{service: service}
}

func (h *DistributorHandler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}
