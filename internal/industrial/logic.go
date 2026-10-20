package industrial

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"horizon-core/internal/crypto"
)

type Site struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SiteID    string    `gorm:"uniqueIndex" json:"site_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Location  string    `json:"location"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type Sensor struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SensorID  string    `gorm:"uniqueIndex" json:"sensor_id"`
	SiteID    string    `gorm:"index" json:"site_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Unit      string    `json:"unit"`
	MinValue  float64   `json:"min_value"`
	MaxValue  float64   `json:"max_value"`
	PublicKey string    `gorm:"type:text" json:"public_key"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type Reading struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	SensorID   string     `gorm:"index" json:"sensor_id"`
	Value      float64    `json:"value"`
	Nonce      string     `gorm:"uniqueIndex" json:"nonce"`
	Timestamp  int64      `json:"timestamp"`
	Signature  string     `gorm:"type:text" json:"signature"`
	Verified   bool       `json:"verified"`
	RecordedAt time.Time  `json:"recorded_at"`
	Synced     *time.Time `json:"synced_at,omitempty"`
}

type Alert struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SensorID  string    `gorm:"index" json:"sensor_id"`
	Value     float64   `json:"value"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Ack       bool      `json:"ack"`
	CreatedAt time.Time `json:"created_at"`
}

type TamperEvent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SensorID  string    `gorm:"index" json:"sensor_id"`
	Reason    string    `json:"reason"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

type ControlCommand struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SensorID   string    `gorm:"index" json:"sensor_id"`
	Action     string    `json:"action"`
	Value      float64   `json:"value"`
	OperatorID string    `json:"operator_id"`
	LicenseID  string    `json:"license_id"`
	Signature  string    `gorm:"type:text" json:"signature"`
	Executed   bool      `json:"executed"`
	CreatedAt  time.Time `json:"created_at"`
}

type SensorState struct {
	SensorID   string    `gorm:"primaryKey" json:"sensor_id"`
	LastValue  float64   `json:"last_value"`
	LastSeenAt time.Time `json:"last_seen_at"`
	Recent     string    `gorm:"type:text" json:"recent"`
}

type SignableReading struct {
	SensorID  string  `json:"sensor_id"`
	Value     float64 `json:"value"`
	Nonce     string  `json:"nonce"`
	Timestamp int64   `json:"timestamp"`
}

func (r SignableReading) Canonical() []byte {
	b, _ := json.Marshal(r)
	return b
}

func HashReading(r SignableReading) string {
	h := sha256.Sum256(r.Canonical())
	return hex.EncodeToString(h[:])
}

func VerifyReading(pubPEM string, r SignableReading, sigHex string) bool {
	pub, err := crypto.PEMToPublicKey(pubPEM)
	if err != nil {
		return false
	}
	return crypto.VerifySignature(pub, r.Canonical(), sigHex)
}

var (
	seenNonces  = make(map[string]int64)
	nonceMu     sync.Mutex
	nonceWindow = int64(300)
)

func NonceFresh(nonce string, ts int64) bool {
	if nonce == "" {
		return false
	}
	now := time.Now().Unix()
	if ts < now-300 || ts > now+120 {
		return false
	}
	nonceMu.Lock()
	defer nonceMu.Unlock()
	for k, v := range seenNonces {
		if now-v > nonceWindow {
			delete(seenNonces, k)
		}
	}
	if _, exists := seenNonces[nonce]; exists {
		return false
	}
	seenNonces[nonce] = now
	return true
}

func EvaluateReading(s Sensor, value float64) *Alert {
	if value > s.MaxValue {
		return &Alert{SensorID: s.SensorID, Value: value, Severity: "critical",
			Message: fmt.Sprintf("HIGH %s=%.3f>%.3f", s.SensorID, value, s.MaxValue)}
	}
	if value < s.MinValue {
		return &Alert{SensorID: s.SensorID, Value: value, Severity: "critical",
			Message: fmt.Sprintf("LOW %s=%.3f<%.3f", s.SensorID, value, s.MinValue)}
	}
	span := s.MaxValue - s.MinValue
	if span > 0 {
		m := span * 0.1
		if value > s.MaxValue-m {
			return &Alert{SensorID: s.SensorID, Value: value, Severity: "warning",
				Message: fmt.Sprintf("NEAR-HIGH %s=%.3f", s.SensorID, value)}
		}
		if value < s.MinValue+m {
			return &Alert{SensorID: s.SensorID, Value: value, Severity: "warning",
				Message: fmt.Sprintf("NEAR-LOW %s=%.3f", s.SensorID, value)}
		}
	}
	return nil
}

func ValidateCommand(cmd ControlCommand, licActive bool) error {
	if !licActive {
		return fmt.Errorf("license not active")
	}
	if cmd.SensorID == "" {
		return fmt.Errorf("sensor_id required")
	}
	if cmd.Action == "" {
		return fmt.Errorf("action required")
	}
	switch cmd.Action {
	case "open", "close", "set", "reset":
		return nil
	}
	return fmt.Errorf("invalid action: %s", cmd.Action)
}
