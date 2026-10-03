package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"sync"
	"time"

	"horizon-core/internal/industrial"

	"gorm.io/gorm"
)

const (
	AnchorInterval = 30 * time.Second
	AnchorMinBatch = 100
)

type AnchorState struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	LastAnchoredID uint      `gorm:"index" json:"last_anchored_id"`
	LastBlockNum   int64     `json:"last_block_num"`
	TotalAnchored  uint64    `json:"total_anchored"`
	UpdatedAt      time.Time `json:"updated_at"`
}

var (
	anchorMutex   sync.Mutex
	anchorStarted bool
)

func startIndustrialAnchorWorker() {
	anchorMutex.Lock()
	if anchorStarted {
		anchorMutex.Unlock()
		return
	}
	anchorStarted = true
	anchorMutex.Unlock()

	if err := db.AutoMigrate(&AnchorState{}); err != nil {
		log.Printf("⚠️  anchor: migrate failed: %v", err)
		return
	}

	go func() {
		ticker := time.NewTicker(AnchorInterval)
		defer ticker.Stop()

		log.Printf("⛓  industrial anchor worker started (every %s)", AnchorInterval)

		for range ticker.C {
			if err := anchorIndustrialOnce(db); err != nil {
				log.Printf("⚠️  anchor: %v", err)
			}
		}
	}()
}

func anchorIndustrialOnce(db *gorm.DB) error {
	var state AnchorState
	if err := db.First(&state).Error; err != nil {
		state = AnchorState{LastAnchoredID: 0, TotalAnchored: 0}
	}

	var readings []industrial.Reading
	if err := db.
		Where("id > ?", state.LastAnchoredID).
		Order("id ASC").
		Limit(50000).
		Find(&readings).Error; err != nil {
		return fmt.Errorf("read readings: %w", err)
	}

	if len(readings) < AnchorMinBatch {
		return nil
	}

	merkleRoot, err := computeReadingsMerkleRoot(readings)
	if err != nil {
		return fmt.Errorf("merkle: %w", err)
	}

	stateMutex.RLock()
	key := unlockedKey
	stateMutex.RUnlock()
	if key == nil {
		return fmt.Errorf("key locked")
	}

	firstID := readings[0].ID
	lastID := readings[len(readings)-1].ID

	payload := map[string]interface{}{
		"merkle_root":   merkleRoot,
		"reading_count": len(readings),
		"from_id":       firstID,
		"to_id":         lastID,
		"first_sensor":  readings[0].SensorID,
		"last_sensor":   readings[len(readings)-1].SensorID,
		"anchored_at":   time.Now().Unix(),
	}

	// Submit to batcher (returns string, ignore)
	if b := GetBatcher(); b != nil {
		_ = b.Submit(TxIndustrialAnchor, payload)
	} else {
		if _, _, err := appendChainTx(db, key, TxIndustrialAnchor, payload); err != nil {
			return fmt.Errorf("chain append: %w", err)
		}
	}

	state.LastAnchoredID = lastID
	state.TotalAnchored += uint64(len(readings))
	state.UpdatedAt = time.Now()

	if state.ID == 0 {
		db.Create(&state)
	} else {
		db.Save(&state)
	}

	prefix := merkleRoot
	if len(prefix) > 16 {
		prefix = prefix[:16]
	}
	log.Printf("⛓  anchored %d readings (ID %d→%d, root=%s)",
		len(readings), firstID, lastID, prefix)

	return nil
}

func computeReadingsMerkleRoot(readings []industrial.Reading) (string, error) {
	if len(readings) == 0 {
		return "", fmt.Errorf("no readings")
	}

	leaves := make([][]byte, len(readings))
	for i, r := range readings {
		canonical := fmt.Sprintf("reading|%d|%s|%v|%s|%d",
			r.ID, r.SensorID, r.Value, r.Nonce, r.Timestamp)
		h := sha256.Sum256([]byte(canonical))
		leaves[i] = h[:]
	}

	return computeBinaryMerkleRoot(leaves), nil
}

func computeBinaryMerkleRoot(leaves [][]byte) string {
	if len(leaves) == 0 {
		return ""
	}

	current := make([][]byte, len(leaves))
	for i, l := range leaves {
		c := make([]byte, len(l))
		copy(c, l)
		current[i] = c
	}

	for len(current) > 1 {
		var next [][]byte
		for i := 0; i < len(current); i += 2 {
			left := current[i]
			var right []byte
			if i+1 < len(current) {
				right = current[i+1]
			} else {
				right = current[i]
			}
			h := sha256.New()
			h.Write(left)
			h.Write(right)
			next = append(next, h.Sum(nil))
		}
		current = next
	}

	root := current[0]
	const hexdigits = "0123456789abcdef"
	out := make([]byte, len(root)*2)
	for i, b := range root {
		out[i*2] = hexdigits[b>>4]
		out[i*2+1] = hexdigits[b&0x0f]
	}
	return string(out)
}
