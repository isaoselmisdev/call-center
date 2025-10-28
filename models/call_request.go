package models

import (
	"time"

	"gorm.io/gorm"
)

// IncomingCall represents a call received by the call center
type IncomingCall struct {
	CallID         string    `json:"call_id"`
	CustomerNumber string    `json:"customer_number"`
	Timestamp      time.Time `json:"timestamp"`
}

// AssignedCall represents a call assigned to an agent
type AssignedCall struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	CallID          string         `gorm:"uniqueIndex;not null" json:"call_id"`
	CustomerNumber  string         `gorm:"not null" json:"customer_number"`
	Timestamp       time.Time      `json:"timestamp"`
	AssignedAgentID string         `gorm:"not null" json:"assigned_agent_id"`
	Status          string         `json:"status"`
	Notes           string         `json:"notes"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
