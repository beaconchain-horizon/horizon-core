// ============================================================
//  HORIZON CLIENT AGENT
//  نسخه توکن‌محور — بدون URL hardcode
//  
//  منطق:
//   ۱. با liara_api_token از api.liara.ir آدرس مرکز را می‌گیرد
//   ۲. subdomain را استخراج می‌کند
//   ۳. به https://<subdomain>.liara.run heartbeat می‌فرستد
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
	"strings"
	"syscall"
	"time"
)

type Config struct {
	TenantID       string `json:"tenant_id"`
	LiaraRegion    string `json:"liara_region"`     // default: api.liara.ir
	LiaraProjectID string `json:"liara_project_id"` // default: horizon-switch
	LiaraAPIToken  string `json:"liara_api_token"`
	AgentToken     string `json:"agent_token"`
	HeartbeatSec   int    `json:"heartbeat_sec"`
}

type HeartbeatPayload struct {
	TenantID string `json:"tenant_id"`
	AgentURL string `json:"agent_url"`
	Version  string `json:"version"`
	Status   string `json:"status"`
	Payload  string `json:"payload"`
}

const Version = "1.0.0"

var (
	cfg        Config
	centralURL string
)

func main() {
	configPath := flag.String("config", "./agent.json", "path to agent config")
	flag.Parse()

	log.Printf("🛰️  Horizon Client Agent v%s", Version)

	// ─── بارگذاری config ───
	data, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("read config: %v", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("parse config: %v", err)
	}

	// ─── مقادیر پیش‌فرض ───
	if cfg.LiaraRegion == "" {
		cfg.LiaraRegion = "api.liara.ir"
	}
	if cfg.LiaraProjectID == "" {
		cfg.LiaraProjectID = "horizon-switch"
	}
	if cfg.HeartbeatSec == 0 {
		cfg.HeartbeatSec = 60
	}

	if cfg.LiaraAPIToken == "" {
		log.Fatalf("❌ liara_api_token در config نیست")
	}
	if cfg.AgentToken == "" {
		log.Fatalf("❌ agent_token در config نیست")
	}

	log.Printf("📡 Tenant:        %s", cfg.TenantID)
	log.Printf("🌐 Liara Region:  %s", cfg.LiaraRegion)
	log.Printf("📦 Project:       %s", cfg.LiaraProjectID)

	// ─── کشف آدرس مرکز ───
	if err := discoverCentralURL(); err != nil {
		log.Printf("⚠️  کشف آدرس با خطا: %v", err)
		log.Printf("   تلاش می‌کنم هر ۶۰ ثانیه یکبار...")
	}

	// ─── حلقه heartbeat ───
	ticker := time.NewTicker(time.Duration(cfg.HeartbeatSec) * time.Second)
	defer ticker.Stop()

	// اولین heartbeat فوری
	if centralURL != "" {
		sendHeartbeat()
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			if centralURL == "" {
				// تلاش برای کشف مجدد
				if err := discoverCentralURL(); err != nil {
					log.Printf("❌ کشف آدرس: %v", err)
					continue
				}
			}
			sendHeartbeat()
		case <-sig:
			log.Printf("👋 shutting down")
			return
		}
	}
}

// ═══════════════════════════════════════════════════════════════
//  discoverCentralURL — از Liara API آدرس مرکز را می‌گیرد
// ═══════════════════════════════════════════════════════════════
func discoverCentralURL() error {
	// ─── اول تلاش کن subdomain را فعال کن (idempotent) ───
	enableURL := fmt.Sprintf(
		"https://%s/v1/projects/%s/default-subdomain/enable",
		cfg.LiaraRegion, cfg.LiaraProjectID,
	)

	req, _ := http.NewRequest("POST", enableURL, nil)
	req.Header.Set("Authorization", "Bearer "+cfg.LiaraAPIToken)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err == nil {
		resp.Body.Close()
		// اگه 200 یا 500 داد، مهم نیست — فقط تلاش
	}

	// ─── حالا اطلاعات پروژه را بگیر ───
	infoURL := fmt.Sprintf(
		"https://%s/v1/projects/%s",
		cfg.LiaraRegion, cfg.LiaraProjectID,
	)

	req, err = http.NewRequest("GET", infoURL, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.LiaraAPIToken)

	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("call liara api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("liara api status: %d", resp.StatusCode)
	}

	// ─── پارس پاسخ ───
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	// subdomain را از فیلدهای مختلف استخراج کن
	subdomain := extractSubdomain(result)
	if subdomain == "" {
		return fmt.Errorf("subdomain در پاسخ نیست")
	}

	centralURL = "https://" + subdomain
	log.Printf("✅ آدرس مرکز: %s", centralURL)
	return nil
}

func extractSubdomain(data map[string]interface{}) string {
	// تلاش مسیرهای مختلف برای پیدا کردن subdomain
	keys := []string{"subdomain", "defaultSubdomain", "domain", "url"}
	for _, k := range keys {
		if v, ok := data[k].(string); ok && v != "" {
			// اگه URL کامل بود
			if strings.HasPrefix(v, "http") {
				v = strings.TrimPrefix(v, "https://")
				v = strings.TrimPrefix(v, "http://")
				v = strings.TrimSuffix(v, "/")
				return v
			}
			// اگه فقط subdomain بود
			if !strings.Contains(v, ".") {
				return v + ".liara.run"
			}
			return v
		}
	}

	// nested
	if d, ok := data["domain"].(map[string]interface{}); ok {
		if v, ok := d["subdomain"].(string); ok {
			return v + ".liara.run"
		}
	}

	return ""
}

// ═══════════════════════════════════════════════════════════════
//  sendHeartbeat — ارسال ضربان به مرکز
// ═══════════════════════════════════════════════════════════════
func sendHeartbeat() {
	if centralURL == "" {
		return
	}

	hostname, _ := os.Hostname()
	payload := HeartbeatPayload{
		TenantID: cfg.TenantID,
		AgentURL: hostname,
		Version:  Version,
		Status:   "ok",
		Payload:  "{}",
	}
	body, _ := json.Marshal(payload)

	url := centralURL + "/api/v1/tenant/heartbeat"
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
}
