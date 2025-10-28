package main

import (
	"call-center-api/internal/callcenter"
	"call-center-api/pkg/config"
	"call-center-api/pkg/database"
	"call-center-api/pkg/logger"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
)

func main() {
	logger.Init()

	cfg := config.Load()

	// Initialize Kafka producer
	brokers := strings.Split(cfg.KafkaBrokers, ",")
	kafkaProducer, err := database.NewKafkaProducer(brokers, "incoming_calls")
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to create Kafka producer: %v", err)
	}

	// Initialize service
	service := callcenter.NewCallCenterService(kafkaProducer)

	// Initialize handler
	handler := callcenter.NewCallCenterHandler(service)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Call Center API",
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

	// Start server
	go func() {
		port := fmt.Sprintf(":%s", cfg.CallCenterPort)
		if err := app.Listen(port); err != nil {
			logger.ErrorLogger.Fatalf("Failed to start server: %v", err)
		}
	}()

	logger.InfoLogger.Printf("Call Center API started on port %s", cfg.CallCenterPort)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.InfoLogger.Println("Shutting down Call Center API...")
	kafkaProducer.Close()
	app.Shutdown()
}

func setupRoutes(app *fiber.App, handler *callcenter.CallCenterHandler) {
	v1 := app.Group("/api/v1")

	v1.Post("/calls", handler.CreateCall)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
}
