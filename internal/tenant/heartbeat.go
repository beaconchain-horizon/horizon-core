package tenant

import (
	"time"

	"gorm.io/gorm"
)

// Heartbeat ضربان قلب هر agent
type Heartbeat struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TenantID   string    `gorm:"index;not null" json:"tenant_id"`
	AgentURL   string    `json:"agent_url"`
	Version    string    `json:"version"`
	Status     string    `json:"status"` // ok | degraded | down
	Payload    string    `gorm:"type:text" json:"payload"`
	ReceivedAt time.Time `json:"received_at"`
}

// RecordHeartbeat ثبت ضربان + آپدیت LastHeartbeat
func RecordHeartbeat(db *gorm.DB, tenantID, agentURL, version, status, payload string) error {
	hb := Heartbeat{
		TenantID:   tenantID,
		AgentURL:   agentURL,
		Version:    version,
		Status:     status,
		Payload:    payload,
		ReceivedAt: time.Now().UTC(),
	}
	if err := db.Create(&hb).Error; err != nil {
		return err
	}
	now := time.Now().UTC()
	return db.Model(&Tenant{}).
		Where("tenant_id = ?", tenantID).
		Update("last_heartbeat", now).Error
}

// IsStale چک می‌کند agent ساکت شده
func IsStale(t *Tenant, threshold time.Duration) bool {
	if t.LastHeartbeat == nil {
		return true
	}
	return time.Since(*t.LastHeartbeat) > threshold
}
