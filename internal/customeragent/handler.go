package customeragent

import (
	"call-center-api/models"
	"call-center-api/pkg/config"
	"call-center-api/pkg/database"
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
)

type AgentHandler struct {
	service       AgentService
	db            *gorm.DB
	kafkaProducer *database.KafkaProducer
}

func NewAgentHandler(service AgentService, db *gorm.DB, kafkaProducer *database.KafkaProducer) *AgentHandler {
	return &AgentHandler{
		service:       service,
		db:            db,
		kafkaProducer: kafkaProducer,
	}
}

func (h *AgentHandler) RegisterAdmin(c *fiber.Ctx) error {
	var req models.RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	// Check admin password
	cfg := config.Load()
	adminPassword := c.Get("X-Admin-Password")
	if adminPassword != cfg.AdminPassword {
		return c.Status(401).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid admin credentials",
		})
	}

	agent, err := h.service.RegisterAgent(req.Name, req.Password, false)
	if err != nil {
		return c.Status(500).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to register agent",
			Error:   err.Error(),
		})
	}

	return c.Status(201).JSON(models.Response{
		Success: true,
		Message: "Agent registered successfully",
		Data:    agent,
	})
}

func (h *AgentHandler) AdminLogin(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	// Check admin credentials
	cfg := config.Load()
	// Default admin username is "admin"
	if req.Username != "admin" || req.Password != cfg.AdminPassword {
		return c.Status(401).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid admin credentials",
		})
	}

	// Generate admin token
	token, err := h.service.GenerateAdminToken(req.Username)
	if err != nil {
		return c.Status(500).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to generate token",
			Error:   err.Error(),
		})
	}

	return c.JSON(models.Response{
		Success: true,
		Message: "Admin login successful",
		Data: models.LoginResponse{
			Token: token,
		},
	})
}

func (h *AgentHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	token, err := h.service.Login(req.AgentID, req.Password)
	if err != nil {
		return c.Status(401).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid credentials",
			Error:   err.Error(),
		})
	}

	return c.JSON(models.Response{
		Success: true,
		Message: "Login successful",
		Data: models.LoginResponse{
			Token: token,
		},
	})
}

func (h *AgentHandler) GetCalls(c *fiber.Ctx) error {
	agentID := c.Locals("agent_id").(string)

	// Skip if admin is accessing
	if agentID == "admin" {
		return c.JSON(models.Response{
			Success: true,
			Data:    []models.AssignedCall{},
		})
	}

	calls, err := h.service.GetAssignedCalls(agentID)
	if err != nil {
		return c.Status(500).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to fetch calls",
			Error:   err.Error(),
		})
	}

	return c.JSON(models.Response{
		Success: true,
		Data:    calls,
	})
}

func (h *AgentHandler) CompleteCall(c *fiber.Ctx) error {
	agentID := c.Locals("agent_id").(string)
	callID := c.Params("id")

	var req struct {
		Notes  string `json:"notes"`
		Status string `json:"status"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	call, err := h.service.CompleteCall(callID, agentID, req.Notes, req.Status)
	if err != nil {
		return c.Status(500).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to complete call",
			Error:   err.Error(),
		})
	}

	return c.JSON(models.Response{
		Success: true,
		Message: "Call completed successfully",
		Data:    call,
	})
}

func (h *AgentHandler) CreateAgent(c *fiber.Ctx) error {
	var req struct {
		AgentID   string `json:"agent_id"`
		AgentName string `json:"agent_name"`
		Password  string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	agent, err := h.service.RegisterAgent(req.AgentName, req.Password, false)
	if err != nil {
		return c.Status(500).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to create agent",
			Error:   err.Error(),
		})
	}

	// Publish agent creation event to Kafka
	if h.kafkaProducer != nil {
		agentData, _ := json.Marshal(agent)
		key := fmt.Sprintf("create_agent:%s", agent.ID)
		if err := h.kafkaProducer.PublishMessage(context.Background(), key, agentData); err != nil {
			fmt.Printf("Warning: Failed to publish agent creation event: %v\n", err)
			// Don't fail the request if Kafka publish fails
		} else {
			fmt.Printf("Published agent creation event: %s\n", key)
		}
	}

	return c.Status(201).JSON(models.Response{
		Success: true,
		Message: "Agent created successfully",
		Data:    agent,
	})
}

func (h *AgentHandler) GetAgentStats(c *fiber.Ctx) error {
	stats, err := h.service.GetAgentStats()
	if err != nil {
		return c.Status(500).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to fetch agent stats",
			Error:   err.Error(),
		})
	}

	return c.JSON(models.Response{
		Success: true,
		Data:    stats,
	})
}

func (h *AgentHandler) DeleteAgent(c *fiber.Ctx) error {
	agentID := c.Params("id")

	if agentID == "" {
		return c.Status(400).JSON(models.ErrorResponse{
			Success: false,
			Message: "Agent ID is required",
		})
	}

	// Find the agent first
	var agent models.Agent
	if err := h.db.Where("id = ?", agentID).First(&agent).Error; err != nil {
		return c.Status(404).JSON(models.ErrorResponse{
			Success: false,
			Message: "Agent not found",
		})
	}

	// Soft delete by marking as inactive
	if err := h.db.Model(&agent).Update("is_active", false).Error; err != nil {
		return c.Status(500).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to delete agent",
			Error:   err.Error(),
		})
	}

	// Publish agent deletion event to Kafka
	if h.kafkaProducer != nil {
		agentData, _ := json.Marshal(agent)
		key := fmt.Sprintf("delete_agent:%s", agent.ID)
		if err := h.kafkaProducer.PublishMessage(context.Background(), key, agentData); err != nil {
			fmt.Printf("Warning: Failed to publish agent deletion event: %v\n", err)
		} else {
			fmt.Printf("Published agent deletion event: %s\n", key)
		}
	}

	return c.JSON(models.Response{
		Success: true,
		Message: "Agent deleted successfully",
		Data:    agent,
	})
}

func (h *AgentHandler) WebSocketHandler(c *websocket.Conn, agentID string) {
	defer func() {
		c.Close()
		fmt.Printf("WebSocket closed for agent: %s\n", agentID)
	}()

	fmt.Printf("WebSocket connected for agent: %s\n", agentID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create Kafka consumer for this agent's assigned_calls
	cfg := config.Load()
	brokers := []string{cfg.KafkaBrokers}

	consumer, err := h.service.GetKafkaConsumer(brokers, "assigned_calls", "ws-"+agentID)
	if err != nil {
		fmt.Printf("Error creating Kafka consumer: %v\n", err)
		c.WriteJSON(fiber.Map{"error": "Failed to connect to message stream"})
		return
	}
	defer consumer.Close()

	// Send initial connection success message
	err = c.WriteJSON(fiber.Map{
		"type":    "connected",
		"message": fmt.Sprintf("Connected as agent %s", agentID),
	})
	if err != nil {
		fmt.Printf("Error sending connected message: %v\n", err)
		return
	}

	// Channel to signal when WebSocket closes
	wsClosed := make(chan struct{})

	// Start consuming messages in background
	go func() {
		fmt.Printf("Starting Kafka consumer for agent %s on topic assigned_calls with group ws-%s\n", agentID, agentID)
		err := h.service.ConsumeAssignedCalls(ctx, consumer, func(call models.AssignedCall) error {
			fmt.Printf("Received call from Kafka: %s for agent %s (current agent: %s)\n", call.CallID, call.AssignedAgentID, agentID)
			// Only send if assigned to this agent
			if call.AssignedAgentID == agentID {
				fmt.Printf("Sending call %s to agent %s via WebSocket\n", call.CallID, agentID)
				data := fiber.Map{
					"type": "new_call",
					"data": call,
				}
				if err := c.WriteJSON(data); err != nil {
					fmt.Printf("Error sending message to WebSocket: %v\n", err)
					// Don't return error - just log it and continue consuming
					return nil
				}
				fmt.Printf("Successfully sent call %s to agent %s\n", call.CallID, agentID)
			} else {
				fmt.Printf("Skipping call %s (assigned to %s, not %s)\n", call.CallID, call.AssignedAgentID, agentID)
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Kafka consumer stopped for agent %s: %v\n", agentID, err)
		}
		fmt.Printf("Kafka consumer goroutine exiting for agent %s\n", agentID)
	}()

	// Keep connection alive - blocking read loop
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				fmt.Printf("WebSocket error for agent %s: %v\n", agentID, err)
			} else {
				fmt.Printf("WebSocket closed normally for agent %s\n", agentID)
			}
			close(wsClosed)
			break
		}
		// Handle ping/pong or other client messages
		fmt.Printf("Received message from agent %s: %s\n", agentID, string(message))
	}

	// WebSocket closed, cancel Kafka consumer context
	cancel()
	fmt.Printf("WebSocket disconnected, stopping Kafka consumer for agent: %s\n", agentID)
}
