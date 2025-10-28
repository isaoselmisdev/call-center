package distributor

import (
	"call-center-api/models"
	"call-center-api/pkg/database"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type DistributorService interface {
	Start(ctx context.Context) error
	SetAgentChangeConsumer(consumer *database.KafkaConsumer)
	StartAgentChangeConsumer(ctx context.Context) error
}

type distributorService struct {
	kafkaConsumer            *database.KafkaConsumer
	agentChangeKafkaConsumer *database.KafkaConsumer
	kafkaProducer            *database.KafkaProducer
	redis                    *redis.Client
	db                       *gorm.DB
}

func NewDistributorService(
	kafkaConsumer *database.KafkaConsumer,
	kafkaProducer *database.KafkaProducer,
	redis *redis.Client,
	db *gorm.DB,
) DistributorService {
	return &distributorService{
		kafkaConsumer: kafkaConsumer,
		kafkaProducer: kafkaProducer,
		redis:         redis,
		db:            db,
	}
}

func (s *distributorService) SetAgentChangeConsumer(consumer *database.KafkaConsumer) {
	s.agentChangeKafkaConsumer = consumer
}

func (s *distributorService) Start(ctx context.Context) error {
	return s.kafkaConsumer.ConsumeMessages(ctx, s.processIncomingCall)
}

func (s *distributorService) StartAgentChangeConsumer(ctx context.Context) error {
	if s.agentChangeKafkaConsumer == nil {
		return fmt.Errorf("agent change consumer not initialized")
	}
	return s.consumeAgentChanges(ctx)
}

func (s *distributorService) processIncomingCall(call models.IncomingCall) error {
	fmt.Printf("Processing incoming call: %s\n", call.CallID)

	// Assign agent using round-robin
	agentID := s.assignAgent()
	if agentID == "" {
		fmt.Printf("No available agent for call: %s\n", call.CallID)
		return nil // Skip if no agents available
	}

	// Create assigned call
	assignedCall := models.AssignedCall{
		CallID:          call.CallID,
		CustomerNumber:  call.CustomerNumber,
		Timestamp:       time.Now(),
		AssignedAgentID: agentID,
		Status:          "assigned",
	}

	// Publish to assigned_calls topic
	if err := s.kafkaProducer.PublishAssignedCall(context.Background(), assignedCall); err != nil {
		return err
	}

	// Save to database
	if err := s.db.Create(&assignedCall).Error; err != nil {
		fmt.Printf("Error saving to database: %v\n", err)
	}

	fmt.Printf("Call %s assigned to agent %s\n", call.CallID, agentID)
	return nil
}

func (s *distributorService) assignAgent() string {
	ctx := context.Background()

	// Get available agents
	agents, err := s.redis.LRange(ctx, "available_agents", 0, -1).Result()
	if err != nil || len(agents) == 0 {
		return ""
	}
	fmt.Printf("Available agents: %v\n", agents)

	// Round-robin: pop first agent from left, push to right (end)
	agent, err := s.redis.LPop(ctx, "available_agents").Result()
	if err != nil {
		return ""
	}
	s.redis.RPush(ctx, "available_agents", agent)

	fmt.Printf("Assigned to agent: %s\n", agent)
	return agent
}

// consumeAgentChanges listens to agent_changes topic and syncs Redis
func (s *distributorService) consumeAgentChanges(ctx context.Context) error {
	fmt.Println("Starting agent changes consumer...")

	handler := &agentChangeHandler{
		handler: s.agentChangeKafkaConsumer,
		processAgentChange: func(key string, value []byte) error {
			return s.handleAgentChange(ctx, key, value)
		},
	}

	// Keep consuming until context is canceled
	for {
		topics := []string{"agent_changes"}
		if err := s.agentChangeKafkaConsumer.ConsumeRawMessages(ctx, topics, handler); err != nil {
			if ctx.Err() != nil {
				fmt.Printf("Agent change consumer context canceled\n")
				return ctx.Err()
			}
			fmt.Printf("Agent change consumer error: %v\n", err)
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// handleAgentChange processes agent creation/deletion events
func (s *distributorService) handleAgentChange(ctx context.Context, key string, value []byte) error {
	fmt.Printf("Received agent change event - Key: %s\n", key)

	// Parse the key to determine action type
	parts := strings.SplitN(key, ":", 2)
	if len(parts) != 2 {
		fmt.Printf("Invalid key format: %s\n", key)
		return nil
	}

	action := parts[0]
	// agentID := parts[1] - available in parts[1] if needed

	// Unmarshal agent data
	var agent models.Agent
	if err := json.Unmarshal(value, &agent); err != nil {
		fmt.Printf("Error unmarshaling agent data: %v\n", err)
		return nil
	}

	switch action {
	case "create_agent":
		return s.handleAgentCreation(ctx, agent)
	case "delete_agent":
		return s.handleAgentDeletion(ctx, agent)
	default:
		fmt.Printf("Unknown action: %s\n", action)
	}

	return nil
}

// handleAgentCreation adds the new agent to Redis
func (s *distributorService) handleAgentCreation(ctx context.Context, agent models.Agent) error {
	fmt.Printf("Adding agent %s (%s) to Redis\n", agent.ID, agent.Name)

	// Check if agent is active
	if !agent.IsActive {
		fmt.Printf("Agent %s is not active, skipping\n", agent.ID)
		return nil
	}

	// Check if agent already exists in Redis
	agents, err := s.redis.LRange(ctx, "available_agents", 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to read from Redis: %w", err)
	}

	for _, existingAgent := range agents {
		if existingAgent == agent.ID {
			fmt.Printf("Agent %s already exists in Redis\n", agent.ID)
			return nil
		}
	}

	// Add agent to Redis
	if err := s.redis.RPush(ctx, "available_agents", agent.ID).Err(); err != nil {
		return fmt.Errorf("failed to add agent to Redis: %w", err)
	}

	fmt.Printf("✓ Successfully added agent %s to Redis\n", agent.ID)

	// Log current state
	updatedAgents, _ := s.redis.LRange(ctx, "available_agents", 0, -1).Result()
	fmt.Printf("Current available agents in Redis: %v\n", updatedAgents)

	return nil
}

// handleAgentDeletion removes the agent from Redis
func (s *distributorService) handleAgentDeletion(ctx context.Context, agent models.Agent) error {
	fmt.Printf("Removing agent %s (%s) from Redis\n", agent.ID, agent.Name)

	// Remove agent from Redis list
	removed, err := s.redis.LRem(ctx, "available_agents", 0, agent.ID).Result()
	if err != nil {
		return fmt.Errorf("failed to remove agent from Redis: %w", err)
	}

	if removed > 0 {
		fmt.Printf("✓ Successfully removed agent %s from Redis (removed %d occurrences)\n", agent.ID, removed)
	} else {
		fmt.Printf("Agent %s was not found in Redis\n", agent.ID)
	}

	// Log current state
	updatedAgents, _ := s.redis.LRange(ctx, "available_agents", 0, -1).Result()
	fmt.Printf("Current available agents in Redis: %v\n", updatedAgents)

	return nil
}

// agentChangeHandler implements sarama.ConsumerGroupHandler for raw agent change messages
type agentChangeHandler struct {
	handler            *database.KafkaConsumer
	processAgentChange func(key string, value []byte) error
}

func (h *agentChangeHandler) Setup(session sarama.ConsumerGroupSession) error {
	fmt.Printf("Agent change consumer group handler setup - MemberID: %s, GenerationID: %d\n", session.MemberID(), session.GenerationID())
	return nil
}

func (h *agentChangeHandler) Cleanup(sarama.ConsumerGroupSession) error {
	fmt.Printf("Agent change consumer group handler cleanup\n")
	return nil
}

func (h *agentChangeHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	fmt.Printf("Starting ConsumeClaim for agent changes topic %s partition %d at offset %d\n", claim.Topic(), claim.Partition(), claim.InitialOffset())

	for message := range claim.Messages() {
		fmt.Printf("Received agent change message from topic %s partition %d offset %d, key: %s\n",
			message.Topic, message.Partition, message.Offset, string(message.Key))

		if h.processAgentChange != nil {
			if err := h.processAgentChange(string(message.Key), message.Value); err != nil {
				fmt.Printf("Error processing agent change: %v\n", err)
			}
		}
		session.MarkMessage(message, "")
	}

	fmt.Printf("ConsumeClaim loop exited for topic %s partition %d\n", claim.Topic(), claim.Partition())
	return nil
}
