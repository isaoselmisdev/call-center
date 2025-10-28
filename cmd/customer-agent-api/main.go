package main

import (
	"call-center-api/internal/customeragent"
	"call-center-api/pkg/config"
	"call-center-api/pkg/database"
	"call-center-api/pkg/logger"
	"call-center-api/pkg/middleware"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v5"
)

func main() {
	logger.Init()

	cfg := config.Load()

	// Initialize database
	db, err := database.NewPostgres(
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
	)
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to connect to database: %v", err)
	}
	logger.InfoLogger.Println("Connected to PostgreSQL")

	// Initialize Kafka producer for agent changes
	brokers := []string{cfg.KafkaBrokers}
	kafkaProducer, err := database.NewKafkaProducer(brokers, "agent_changes")
	if err != nil {
		logger.ErrorLogger.Printf("Warning: Failed to create Kafka producer: %v", err)
		kafkaProducer = nil // Continue without Kafka
	} else {
		logger.InfoLogger.Println("Connected to Kafka producer")
	}

	// Initialize service
	service := customeragent.NewAgentService(db)

	// Initialize handler
	handler := customeragent.NewAgentHandler(service, db, kafkaProducer)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Customer Agent API",
	})

	// CORS middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Method() == "OPTIONS" {
			return c.SendStatus(200)
		}
		return c.Next()
	})

	// Setup routes
	setupRoutes(app, handler)

	go func() {
		port := fmt.Sprintf(":%s", cfg.CustomerAgentPort)
		if err := app.Listen(port); err != nil {
			logger.ErrorLogger.Fatalf("Failed to start server: %v", err)
		}
	}()

	logger.InfoLogger.Printf("Customer Agent API started on port %s", cfg.CustomerAgentPort)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.InfoLogger.Println("Shutting down Customer Agent API...")
	if kafkaProducer != nil {
		kafkaProducer.Close()
	}
	app.Shutdown()
}

func setupRoutes(app *fiber.App, handler *customeragent.AgentHandler) {
	// Public routes
	app.Post("/api/v1/admin/register", handler.RegisterAdmin)
	app.Post("/api/v1/admin/login", handler.AdminLogin)
	app.Post("/api/v1/auth/login", handler.Login)

	// Protected routes (for agents)
	v1 := app.Group("/api/v1", middleware.AuthMiddleware())
	{
		v1.Get("/calls", handler.GetCalls)
		v1.Post("/calls/:id/complete", handler.CompleteCall)
	}

	// Admin routes (protected)
	admin := app.Group("/api/v1", middleware.AuthMiddleware())
	{
		admin.Post("/agents", handler.CreateAgent)
		admin.Get("/agents/stats", handler.GetAgentStats)
		admin.Delete("/agents/:id", handler.DeleteAgent)
	}

	// WebSocket route - needs special handling for auth
	app.Get("/ws/assigned", func(c *fiber.Ctx) error {
		// Check if this is a WebSocket upgrade request
		if !websocket.IsWebSocketUpgrade(c) {
			return c.Status(fiber.StatusUpgradeRequired).SendString("WebSocket upgrade required")
		}

		// Extract token from query parameter
		token := c.Query("token")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token required",
			})
		}

		// Validate token and get agent ID
		cfg := config.Load()
		parsedToken, err := jwt.ParseWithClaims(token, &middleware.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !parsedToken.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		claims, ok := parsedToken.Claims.(*middleware.Claims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		// Upgrade to WebSocket
		return websocket.New(func(conn *websocket.Conn) {
			handler.WebSocketHandler(conn, claims.AgentID)
		})(c)
	})

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
}
