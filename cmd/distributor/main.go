package main

import (
	"call-center-api/internal/distributor"
	"call-center-api/models"
	"call-center-api/pkg/config"
	"call-center-api/pkg/database"
	"call-center-api/pkg/logger"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func main() {
	logger.Init()

	cfg := config.Load()

	// Initialize Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.ErrorLogger.Fatalf("Failed to connect to Redis: %v", err)
	}
	logger.InfoLogger.Println("Connected to Redis")

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

	// Sync active agents from PostgreSQL to Redis
	if err := syncAgentsToRedis(ctx, db, rdb); err != nil {
		logger.ErrorLogger.Fatalf("Failed to sync agents to Redis: %v", err)
	}
	logger.InfoLogger.Println("Synced agents to Redis")

	// Initialize Kafka consumer for incoming_calls
	brokers := strings.Split(cfg.KafkaBrokers, ",")
	kafkaConsumer, err := database.NewKafkaConsumer(brokers, "incoming_calls", cfg.KafkaGroupID)
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	// Initialize Kafka consumer for agent_changes
	agentChangeConsumer, err := database.NewKafkaConsumer(brokers, "agent_changes", "distributor-agent-sync")
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to create agent change consumer: %v", err)
	}

	// Initialize Kafka producer for assigned_calls
	kafkaProducer, err := database.NewKafkaProducer(brokers, "assigned_calls")
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to create Kafka producer: %v", err)
	}

	// Initialize service
	service := distributor.NewDistributorService(kafkaConsumer, kafkaProducer, rdb, db)

	// Set agent change consumer
	service.SetAgentChangeConsumer(agentChangeConsumer)

	// Initialize handler
	handler := distributor.NewDistributorHandler(service)

	// Create Fiber app for health checks
	app := fiber.New()
	app.Get("/health", handler.HealthCheck)

	// Start distributor for incoming calls in background
	go func() {
		if err := service.Start(ctx); err != nil {
			logger.ErrorLogger.Fatalf("Distributor stopped: %v", err)
		}
	}()

	// Start agent change consumer in background
	go func() {
		if err := service.StartAgentChangeConsumer(ctx); err != nil {
			logger.ErrorLogger.Printf("Agent change consumer stopped: %v", err)
		}
	}()

	// Start HTTP server
	go func() {
		port := fmt.Sprintf(":%s", cfg.DistributorPort)
		if err := app.Listen(port); err != nil {
			logger.ErrorLogger.Fatalf("Failed to start server: %v", err)
		}
	}()

	logger.InfoLogger.Printf("Distributor started on port %s", cfg.DistributorPort)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.InfoLogger.Println("Shutting down Distributor...")
	kafkaConsumer.Close()
	agentChangeConsumer.Close()
	kafkaProducer.Close()
	rdb.Close()
	app.Shutdown()
}

// syncAgentsToRedis reads all active agents from PostgreSQL and populates Redis
func syncAgentsToRedis(ctx context.Context, db *gorm.DB, rdb *redis.Client) error {
	// Fetch all active agents from database
	var agents []models.Agent
	if err := db.Where("is_active = ?", true).Find(&agents).Error; err != nil {
		return fmt.Errorf("failed to fetch agents: %w", err)
	}

	if len(agents) == 0 {
		logger.InfoLogger.Println("No active agents found in database")
		return nil
	}

	// Clear existing Redis list
	if err := rdb.Del(ctx, "available_agents").Err(); err != nil {
		return fmt.Errorf("failed to clear Redis list: %w", err)
	}

	// Add all active agents to Redis
	agentIDs := make([]interface{}, len(agents))
	for i, agent := range agents {
		agentIDs[i] = agent.ID
	}

	if err := rdb.RPush(ctx, "available_agents", agentIDs...).Err(); err != nil {
		return fmt.Errorf("failed to push agents to Redis: %w", err)
	}

	logger.InfoLogger.Printf("Synced %d active agents to Redis: %v", len(agents), agentIDs)
	return nil
}
