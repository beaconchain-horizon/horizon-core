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
	"sync"
	"time"

	"gorm.io/gorm"
)

const (
	trialHours = 48
	trialMaxTx = 1000000
	trialFile  = "trial-state.json"
)

var (
	trialActive bool
	trialDB     *gorm.DB
	TrialMutex  sync.Mutex // For atomic tx limit check
)

type TrialState struct {
	InstalledAt int64  `json:"installed_at"`
	MachineID   string `json:"machine_id"`
}

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

func getRegistryTime() (int64, bool) {
	out, err := exec.Command("reg", "query", `HKCU\Software\HorizonCore\Trial`, "/v", "InstallTime").Output()
	if err != nil {
		return 0, false
	}
	parts := strings.Fields(string(out))
	if len(parts) >= 3 {
		var t int64
		fmt.Sscanf(parts[len(parts)-1], "%d", &t)
		if t > 0 {
			return t, true
		}
	}
	return 0, false
}

func setRegistryTime(t int64) {
	cmd := exec.Command("reg", "add", `HKCU\Software\HorizonCore\Trial`, "/v", "InstallTime", "/t", "REG_SZ", "/d", fmt.Sprintf("%d", t), "/f")
	_ = cmd.Run()
}

func trialStatePath() string {
	dir, _ := os.UserConfigDir()
	if dir == "" { dir = "." }
	dir = filepath.Join(dir, "HorizonCore")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, trialFile)
}

func checkTrial(db *gorm.DB) error {
	if !isTrialMode() { return nil }

	trialDB = db
	mid := getMachineID()
	path := trialStatePath()

	regTime, regFound := getRegistryTime()
	var state TrialState
	
	if regFound {
		// Registry is the absolute source of truth
		state = TrialState{InstalledAt: regTime, MachineID: mid}
	} else {
		// Registry missing. Try file.
		if data, err := os.ReadFile(path); err == nil {
			if err := json.Unmarshal(data, &state); err != nil {
				return fmt.Errorf("invalid trial state file")
			}
		} else {
			// First run ever
			state = TrialState{InstalledAt: time.Now().Unix(), MachineID: mid}
			log.Printf("[TRIAL] First run — started at %s", time.Unix(state.InstalledAt, 0).Format(time.RFC3339))
		}
	}
	
	// Sync back to both file and registry
	setRegistryTime(state.InstalledAt)
	data, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(path, data, 0644)

	if state.MachineID != mid {
		return fmt.Errorf("TRIAL: hardware mismatch")
	}

	elapsed := time.Now().Unix() - state.InstalledAt
	if int64(trialHours)*3600 - elapsed <= 0 {
		return fmt.Errorf("TRIAL EXPIRED — %d hours since install", elapsed/3600)
	}

	trialActive = true
	log.Printf("[TRIAL] Active — machine=%s, remaining=%dh, max_tx=%d", mid, (int64(trialHours)*3600 - elapsed)/3600, trialMaxTx)
	return nil
}

func trialTxLimitReached() (bool, error) {
	if !trialActive || trialDB == nil { return false, nil }
	var total int64
	err := trialDB.Table("blocks").Select("COALESCE(SUM(tx_count),0)").Scan(&total).Error
	if err != nil { return true, err } // Fail closed on DB error
	if total >= trialMaxTx {
		return true, nil
	}
	return false, nil
}
