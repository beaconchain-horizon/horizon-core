package industrial

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type syncBatch struct {
	Readings []Reading `json:"readings"`
}

func StartSyncWorker(db *gorm.DB, targetURL, apiKey string, interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			syncOnce(db, targetURL, apiKey)
		}
	}()
	log.Printf("[INDUSTRIAL] sync worker → %s every %s", targetURL, interval)
}

func syncOnce(db *gorm.DB, targetURL, apiKey string) {
	var readings []Reading
	db.Where("synced IS NULL").Limit(200).Find(&readings)
	if len(readings) == 0 {
		return
	}
	body, _ := json.Marshal(syncBatch{Readings: readings})
	req, err := http.NewRequest("POST", targetURL+"/api/v1/industrial/reading/batch", bytes.NewReader(body))
	if err != nil {
		log.Printf("[SYNC] req build error: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-Admin-Key", apiKey)
	}
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		log.Printf("[SYNC] error: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		now := time.Now().UTC()
		for _, r := range readings {
			db.Model(&Reading{}).Where("id = ?", r.ID).Update("synced", now)
		}
		log.Printf("[SYNC] pushed %d readings", len(readings))
	} else {
		log.Printf("[SYNC] target returned %d", resp.StatusCode)
	}
}
