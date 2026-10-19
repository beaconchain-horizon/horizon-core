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
	"sync/atomic"
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
	TrialMutex  sync.Mutex
	TrialTotalTxs int64
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

func getRegistryData() (int64, string, bool) {
	out, err := exec.Command("reg", "query", `HKCU\Software\HorizonCore\Trial`, "/v", "InstallTime").Output()
	if err != nil { return 0, "", false }
	parts := strings.Fields(string(out))
	var t int64
	if len(parts) >= 3 { fmt.Sscanf(parts[len(parts)-1], "%d", &t) }
	
	outM, errM := exec.Command("reg", "query", `HKCU\Software\HorizonCore\Trial`, "/v", "MachineID").Output()
	var mid string
	if errM == nil {
		partsM := strings.Fields(string(outM))
		if len(partsM) >= 3 { mid = partsM[len(partsM)-1] }
	}
	if t > 0 && mid != "" { return t, mid, true }
	return 0, "", false
}

func setRegistryData(t int64, mid string) {
	exec.Command("reg", "add", `HKCU\Software\HorizonCore\Trial`, "/v", "InstallTime", "/t", "REG_SZ", "/d", fmt.Sprintf("%d", t), "/f").Run()
	exec.Command("reg", "add", `HKCU\Software\HorizonCore\Trial`, "/v", "MachineID", "/t", "REG_SZ", "/d", mid, "/f").Run()
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
	currentMid := getMachineID()
	path := trialStatePath()
	var state TrialState

	regTime, regMid, regFound := getRegistryData()
	
	if regFound {
		state = TrialState{InstalledAt: regTime, MachineID: regMid}
		if state.MachineID != currentMid {
			return fmt.Errorf("TRIAL: hardware mismatch (Registry)")
		}
	} else {
		if data, err := os.ReadFile(path); err == nil {
			if err := json.Unmarshal(data, &state); err != nil { return fmt.Errorf("invalid trial state") }
			if state.MachineID != currentMid {
				return fmt.Errorf("TRIAL: hardware mismatch (File)")
			}
		} else {
			state = TrialState{InstalledAt: time.Now().Unix(), MachineID: currentMid}
		}
	}
	
	setRegistryData(state.InstalledAt, state.MachineID)
	data, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(path, data, 0644)

	elapsed := time.Now().Unix() - state.InstalledAt
	if int64(trialHours)*3600 - elapsed <= 0 {
		return fmt.Errorf("TRIAL EXPIRED — %d hours", elapsed/3600)
	}
	trialActive = true
	log.Printf("[TRIAL] Active — machine=%s, remaining=%dh", currentMid, (int64(trialHours)*3600 - elapsed)/3600)
	return nil
}

func trialTxLimitReached() (bool, error) {
	if !trialActive {
		return false, nil
	}
	// Use atomic counter to prevent race conditions during drainPending
	currentTotal := atomic.LoadInt64(&TrialTotalTxs)
	if currentTotal >= trialMaxTx {
		return true, nil
	}
	return false, nil
}
