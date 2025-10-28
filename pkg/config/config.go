package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string

	// Kafka
	KafkaBrokers string
	KafkaGroupID string

	// Redis
	RedisAddr     string
	RedisPassword string

	// JWT
	JWTSecret string

	// Admin
	AdminPassword string

	// Server Ports
	CallCenterPort    string
	CustomerAgentPort string
	DistributorPort   string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		DBHost:     getEnv("DB_HOST", "postgres"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "callcenter"),
		DBPort:     getEnv("DB_PORT", "5432"),

		KafkaBrokers: getEnv("KAFKA_BROKERS", "kafka:9092"),
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "distributor-group"),

		RedisAddr:     getEnv("REDIS_ADDR", "redis:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),

		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),

		CallCenterPort:    getEnv("CALL_CENTER_PORT", "8081"),
		CustomerAgentPort: getEnv("CUSTOMER_AGENT_PORT", "8082"),
		DistributorPort:   getEnv("DISTRIBUTOR_PORT", "8083"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
