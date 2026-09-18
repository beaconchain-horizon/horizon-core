// ============================================================
//  HORIZON CLIENT AGENT
//  روی سرور مشتری نصب می‌شود و به پنل مرکزی وصل می‌شود
// ============================================================
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	TenantID     string `json:"tenant_id"`
	CentralURL   string `json:"central_url"`
	AgentToken   string `json:"agent_token"`
	SwitchURL    string `json:"switch_url"` // آدرس سوئیچ محلی
	HeartbeatSec int    `json:"heartbeat_sec"`
}

type HeartbeatPayload struct {
	TenantID string `json:"tenant_id"`
	AgentURL string `json:"agent_url"`
	Version  string `json:"version"`
	Status   string `json:"status"`
	Payload  string `json:"payload"`
}

const Version = "1.0.0"

func main() {
	configPath := flag.String("config", "./agent.json", "path to agent config")
	flag.Parse()

	log.Printf("🛰️  Horizon Client Agent v%s", Version)

	data, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("read config: %v", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("parse config: %v", err)
	}
	if cfg.HeartbeatSec == 0 {
		cfg.HeartbeatSec = 60
	}

	log.Printf("📡 Tenant: %s", cfg.TenantID)
	log.Printf("🌐 Central: %s", cfg.CentralURL)
	log.Printf("🔧 Local Switch: %s", cfg.SwitchURL)

	// Heartbeat loop
	ticker := time.NewTicker(time.Duration(cfg.HeartbeatSec) * time.Second)
	defer ticker.Stop()

	// اولین heartbeat فوری
	sendHeartbeat(cfg)

	// Graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			sendHeartbeat(cfg)
		case <-sig:
			log.Printf("👋 shutting down")
			return
		}
	}
}

func sendHeartbeat(cfg Config) {
	hostname, _ := os.Hostname()
	payload := HeartbeatPayload{
		TenantID: cfg.TenantID,
		AgentURL: hostname,
		Version:  Version,
		Status:   "ok",
		Payload:  "{}",
	}
	body, _ := json.Marshal(payload)

	url := cfg.CentralURL + "/api/v1/tenant/heartbeat"
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		log.Printf("❌ build request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Token", cfg.AgentToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ heartbeat: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Printf("✅ heartbeat sent")
	} else {
		log.Printf("⚠️  heartbeat status %d", resp.StatusCode)
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = fmt.Sprint
}
