package tenant

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
)

// Tenant مشتری/سازمان
type Tenant struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	TenantID       string     `gorm:"uniqueIndex;not null" json:"tenant_id"`
	Name           string     `gorm:"not null" json:"name"`
	Type           string     `gorm:"not null" json:"type"` // bank | refinery | powerplant | gas | other
	ContactEmail   string     `json:"contact_email"`
	ContactPhone   string     `json:"contact_phone"`
	DeploymentMode string     `gorm:"default:'cloud'" json:"deployment_mode"` // cloud | on-premise | hybrid
	AgentURL       string     `json:"agent_url"`
	AgentToken     string     `gorm:"type:text" json:"-"`
	LastHeartbeat  *time.Time `json:"last_heartbeat"`
	IsActive       bool       `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// AutoMigrate
func Init(db *gorm.DB) error {
	return db.AutoMigrate(&Tenant{}, &Heartbeat{})
}

// GenerateToken توکن امن برای agent
func GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// FindByID
func FindByID(db *gorm.DB, tenantID string) (*Tenant, error) {
	var t Tenant
	if err := db.Where("tenant_id = ? AND is_active = ?", tenantID, true).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// FindByToken
func FindByToken(db *gorm.DB, token string) (*Tenant, error) {
	var t Tenant
	if err := db.Where("agent_token = ? AND is_active = ?", token, true).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}
