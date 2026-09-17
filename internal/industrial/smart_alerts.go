package industrial

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/gorm"
)

const (
	rocWindowSize   = 10
	rocThresholdPct = 30.0
	heartbeatSilent = 2 * time.Minute
)

func UpdateSensorState(db *gorm.DB, sensorID string, value float64) {
	var st SensorState
	err := db.First(&st, "sensor_id = ?", sensorID).Error
	if err != nil {
		st = SensorState{SensorID: sensorID, Recent: "[]"}
	}
	st.LastValue = value
	st.LastSeenAt = time.Now().UTC()
	var recent []float64
	json.Unmarshal([]byte(st.Recent), &recent)
	recent = append(recent, value)
	if len(recent) > rocWindowSize {
		recent = recent[len(recent)-rocWindowSize:]
	}
	b, _ := json.Marshal(recent)
	st.Recent = string(b)
	db.Save(&st)
}

func CheckRateOfChange(db *gorm.DB, sensorID string) *Alert {
	var st SensorState
	if err := db.First(&st, "sensor_id = ?", sensorID).Error; err != nil {
		return nil
	}
	var recent []float64
	json.Unmarshal([]byte(st.Recent), &recent)
	if len(recent) < 5 {
		return nil
	}
	first := recent[0]
	last := recent[len(recent)-1]
	if first == 0 {
		return nil
	}
	changePct := (last - first) / first * 100.0
	if changePct > rocThresholdPct || changePct < -rocThresholdPct {
		sev := "warning"
		if changePct > 2*rocThresholdPct || changePct < -2*rocThresholdPct {
			sev = "critical"
		}
		return &Alert{SensorID: sensorID, Value: last, Severity: sev,
			Message: fmt.Sprintf("RAPID-CHANGE %s: %.1f%% over %d samples", sensorID, changePct, len(recent))}
	}
	return nil
}

func StartHeartbeatMonitor(db *gorm.DB, interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			var sensors []Sensor
			db.Where("is_active = ?", true).Find(&sensors)
			now := time.Now().UTC()
			for _, s := range sensors {
				var st SensorState
				if err := db.First(&st, "sensor_id = ?", s.SensorID).Error; err != nil {
					continue
				}
				silent := now.Sub(st.LastSeenAt)
				if silent > heartbeatSilent {
					db.Create(&Alert{SensorID: s.SensorID, Value: st.LastValue, Severity: "warning",
						Message: fmt.Sprintf("SILENT %s: no data for %s", s.SensorID, silent.Round(time.Second))})
					st.LastSeenAt = now
					db.Save(&st)
				}
			}
		}
	}()
	log.Printf("[INDUSTRIAL] heartbeat monitor started (silent-after=%s)", heartbeatSilent)
}

func StartBackgroundJobs(db *gorm.DB) {
	db.AutoMigrate(&SensorState{})
	StartHeartbeatMonitor(db, 1*time.Minute)
	if url := os.Getenv("SYNC_TARGET_URL"); url != "" {
		StartSyncWorker(db, url, os.Getenv("SYNC_API_KEY"), 30*time.Second)
	} else {
		log.Printf("[INDUSTRIAL] SYNC_TARGET_URL not set, sync disabled")
	}
	log.Printf("[INDUSTRIAL] background jobs running")
}
