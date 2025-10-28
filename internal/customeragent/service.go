package customeragent

import (
	"call-center-api/models"
	"call-center-api/pkg/config"
	"call-center-api/pkg/database"
	"call-center-api/pkg/middleware"
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AgentService interface {
	RegisterAgent(name, password string, isAdmin bool) (*models.Agent, error)
	Login(agentID, password string) (string, error)
	GenerateAdminToken(username string) (string, error)
	GetAssignedCalls(agentID string) ([]models.AssignedCall, error)
	CompleteCall(callID, agentID, notes, status string) (*models.AssignedCall, error)
	GetAgentStats() ([]map[string]interface{}, error)
	GetKafkaConsumer(brokers []string, topic, groupID string) (*database.KafkaConsumer, error)
	ConsumeAssignedCalls(ctx context.Context, consumer *database.KafkaConsumer, handler func(models.AssignedCall) error) error
}

type agentService struct {
	db *gorm.DB
}

func NewAgentService(db *gorm.DB) AgentService {
	return &agentService{db: db}
}

func (s *agentService) RegisterAgent(name, password string, isAdmin bool) (*models.Agent, error) {
	// Generate 6-digit agent ID
	agentID := generateAgentID()

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	agent := &models.Agent{
		ID:        agentID,
		Name:      name,
		Password:  string(hashedPassword),
		IsAdmin:   isAdmin,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.db.Create(agent).Error; err != nil {
		return nil, err
	}

	return agent, nil
}

func (s *agentService) Login(agentID, password string) (string, error) {
	var agent models.Agent
	if err := s.db.Where("id = ? AND is_active = ?", agentID, true).First(&agent).Error; err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(agent.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token
	cfg := config.Load()
	claims := middleware.Claims{
		AgentID: agent.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *agentService) GenerateAdminToken(username string) (string, error) {
	// Generate JWT token for admin with special agent_id "admin"
	cfg := config.Load()
	claims := middleware.Claims{
		AgentID: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *agentService) GetAssignedCalls(agentID string) ([]models.AssignedCall, error) {
	var calls []models.AssignedCall
	if err := s.db.Where("assigned_agent_id = ?", agentID).Find(&calls).Error; err != nil {
		return nil, err
	}
	return calls, nil
}

func (s *agentService) GetAgentStats() ([]map[string]interface{}, error) {
	var agents []models.Agent
	if err := s.db.Find(&agents).Error; err != nil {
		return nil, err
	}

	var stats []map[string]interface{}
	for _, agent := range agents {
		var totalCalls int64
		var completedCalls int64

		s.db.Model(&models.AssignedCall{}).Where("assigned_agent_id = ?", agent.ID).Count(&totalCalls)
		s.db.Model(&models.AssignedCall{}).Where("assigned_agent_id = ? AND status = ?", agent.ID, "completed").Count(&completedCalls)

		status := "inactive"
		if agent.IsActive {
			status = "active"
		}

		stats = append(stats, map[string]interface{}{
			"agent_id":        agent.ID,
			"agent_name":      agent.Name,
			"status":          status,
			"total_calls":     totalCalls,
			"completed_calls": completedCalls,
		})
	}

	return stats, nil
}

func (s *agentService) CompleteCall(callID, agentID, notes, status string) (*models.AssignedCall, error) {
	var call models.AssignedCall
	if err := s.db.Where("call_id = ?", callID).First(&call).Error; err != nil {
		return nil, errors.New("call not found")
	}

	// Verify agent owns this call
	if call.AssignedAgentID != agentID {
		return nil, errors.New("unauthorized: call not assigned to you")
	}

	// Update call
	call.Status = status
	if notes != "" {
		// You could add a Notes field to the model
	}
	call.Timestamp = time.Now()

	if err := s.db.Save(&call).Error; err != nil {
		return nil, err
	}

	return &call, nil
}

func (s *agentService) GetKafkaConsumer(brokers []string, topic, groupID string) (*database.KafkaConsumer, error) {
	return database.NewKafkaConsumer(brokers, topic, groupID)
}

func (s *agentService) ConsumeAssignedCalls(ctx context.Context, consumer *database.KafkaConsumer, handler func(models.AssignedCall) error) error {
	return consumer.ConsumeAssignedCalls(ctx, handler)
}

func generateAgentID() string {
	// Generate a 6-digit ID
	id := uuid.New().String()
	return id[:6]
}
