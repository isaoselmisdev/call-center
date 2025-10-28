package models

import (
	"time"

	"gorm.io/gorm"
)

// Agent represents an agent in the system
type Agent struct {
	ID        string         `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	Password  string         `gorm:"not null" json:"-"`
	IsAdmin   bool           `gorm:"default:false" json:"is_admin"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// LoginRequest represents agent login request
type LoginRequest struct {
	AgentID  string `json:"agent_id" validate:"required,len=6"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest represents agent registration request
type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token string `json:"token"`
}
