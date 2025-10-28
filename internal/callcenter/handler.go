package callcenter

import (
	"call-center-api/models"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CallCenterHandler struct {
	service CallCenterService
}

func NewCallCenterHandler(service CallCenterService) *CallCenterHandler {
	return &CallCenterHandler{service: service}
}

func (h *CallCenterHandler) CreateCall(c *fiber.Ctx) error {
	var req models.IncomingCall

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// Generate CallID if not provided
	if req.CallID == "" {
		req.CallID = uuid.New().String()
	}

	// Set timestamp if not provided
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	// Validate required fields
	if req.CustomerNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Customer number is required",
		})
	}

	ctx := context.Background()
	if err := h.service.PublishCall(ctx, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to process call",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.Response{
		Success: true,
		Message: "Call processed successfully",
		Data:    req,
	})
}
