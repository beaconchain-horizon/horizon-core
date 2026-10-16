package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"horizon-core/internal/chain"
)

const (
	trialHours = 48
	trialMaxTx = 1000000
	trialFile  = "trial-state.json"
)

type TrialState struct {
	InstalledAt int64  `json:"installed_at"`
	MachineID   string `json:"machine_id"`
}

var (
	trialActive bool
	trialDB     *gorm.DB
)

func isTrialMode() bool {
	v := strings.ToLower(os.Getenv("HORIZON_TRIAL_MODE"))
	return v == "true" || v == "1" || v == "yes"
}

func getMachineID() string {
	out, err := exec.Command("cmd", "/c", "wmic csproduct get uuid").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			uuid := strings.TrimSpace(lines[1])
			if uuid != "" && !strings.Contains(uuid, "UUID") {
				h := sha256.Sum256([]byte(uuid))
				return hex.EncodeToString(h[:8])
			}
		}
	}
	hostname, _ := os.Hostname()
	h := sha256.Sum256([]byte(hostname))
	return hex.EncodeToString(h[:8])
}

func trialStatePath() string {
	dir := os.Getenv("HORIZON_TRIAL_DIR")
	if dir == "" {
		dir = "."
	}
	return filepath.Join(dir, trialFile)
}

func checkTrial(db *gorm.DB) error {
	if !isTrialMode() {
		return nil
	}

	trialDB = db
	mid := getMachineID()
	path := trialStatePath()

	var state TrialState
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &state); err != nil {
			return fmt.Errorf("invalid trial state file")
		}
	} else {
		state = TrialState{
			InstalledAt: time.Now().Unix(),
			MachineID:   mid,
		}
		data, _ := json.MarshalIndent(state, "", "  ")
		if err := os.WriteFile(path, data, 0644); err != nil {
			return fmt.Errorf("cannot create trial state: %w", err)
		}
		log.Printf("[TRIAL] First run — started at %s", time.Unix(state.InstalledAt, 0).Format(time.RFC3339))
	}

	if state.MachineID != mid {
		return fmt.Errorf("TRIAL: hardware mismatch")
	}

	elapsed := time.Now().Unix() - state.InstalledAt
	remaining := int64(trialHours)*3600 - elapsed
	if remaining <= 0 {
		return fmt.Errorf("TRIAL EXPIRED — %d hours since install", elapsed/3600)
	}

	trialActive = true
	log.Printf("[TRIAL] Active — machine=%s, remaining=%dh, max_tx=%d", mid, remaining/3600, trialMaxTx)
	return nil
}

func trialTxLimitReached() bool {
	if !trialActive || trialDB == nil {
		return false
	}
	var total int64
	trialDB.Model(&chain.Block{}).Select("COALESCE(SUM(tx_count),0)").Scan(&total)
	if total >= trialMaxTx {
		log.Printf("[TRIAL] Transaction limit reached: %d / %d", total, trialMaxTx)
		return true
	}
	return false
}
