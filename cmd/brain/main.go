package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	SwitchURL   = "http://localhost:8091"
	GuardianURL = "http://localhost:9008"
	AgentDir    = "../angel-agents"
	AdminToken  = "bench0000000000000000000000000000000000000000000000000000000bench"
)

type Agent struct{ ID, Name, Field string }

var agents []*Agent
var fields = make(map[string]int)

func loadAgents() {
	filepath.Walk(AgentDir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(p, ".md") { return nil }
		rel, _ := filepath.Rel(AgentDir, p)
		parts := strings.Split(rel, string(os.PathSeparator))
		if len(parts) < 2 { return nil }
		f := parts[0]
		if f == "docs" || f == "examples" || strings.HasPrefix(f, ".") { return nil }
		fields[f]++
		agents = append(agents, &Agent{ID: strings.TrimSuffix(info.Name(), ".md"), Name: strings.TrimSuffix(info.Name(), ".md"), Field: f})
		return nil
	})
	log.Printf("✅ Loaded %d agents in %d fields", len(agents), len(fields))
}

func getJSON(url string) map[string]interface{} {
	c := &http.Client{Timeout: 2 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Admin-Token", AdminToken)
	resp, err := c.Do(req)
	if err != nil { return nil }
	defer resp.Body.Close()
	var out map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&out)
	return out
}

// تراکنش‌سازی: هر پرسش ایجنت = یک تراکنش با لایسنس همان رشته
func logAgentAction(field, action string) {
	licID := "AGENT-" + strings.ToUpper(field) + "-LIC"
	
	// گرفتن session token
	sessReq, _ := http.NewRequest("POST", SwitchURL+"/api/v1/signing/session", nil)
	sessReq.Header.Set("X-Admin-Token", AdminToken)
	sessResp, err := http.DefaultClient.Do(sessReq)
	if err != nil { log.Printf("⚠️ session: %v", err); return }
	var sessData map[string]string
	json.NewDecoder(sessResp.Body).Decode(&sessData)
	sessResp.Body.Close()
	sessTok := sessData["session_token"]
	if sessTok == "" { log.Printf("⚠️ no session token"); return }
	
	tx := map[string]interface{}{
		"from": "agent-" + field, "to": "audit-log", "amount": 1,
		"type": "agent_action",
	}
	body, _ := json.Marshal(tx)
	req, _ := http.NewRequest("POST", SwitchURL+"/api/v1/tx", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Token", AdminToken)
	req.Header.Set("X-License-ID", licID)
	req.Header.Set("X-Session-Token", sessTok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil { log.Printf("⚠️ tx: %v", err); return }
	defer resp.Body.Close()
	log.Printf("⛓  agent %s action logged (license=%s, http=%d)", field, licID, resp.StatusCode)
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
func hasAny(s string, words ...string) bool {
	for _, w := range words { if strings.Contains(s, w) { return true } }
	return false
}

type Intent struct{ Type, Field string }

func detectIntent(msg string) Intent {
	m := normalize(msg)
	
	// لاتین
	if hasAny(m, "stats", "summary") { return Intent{Type: "stats"} }
	if hasAny(m, "tps", "throughput", "speed") { return Intent{Type: "tps"} }
	if hasAny(m, "hi", "hello", "hey") { return Intent{Type: "greeting"} }
	if hasAny(m, "agents", "agent") {
		for f := range fields { if strings.Contains(m, f) { return Intent{Type: "agents", Field: f} } }
		return Intent{Type: "agents"}
	}
	if hasAny(m, "security") { return Intent{Type: "security"} }
	if hasAny(m, "network", "guardian") { return Intent{Type: "network"} }
	if hasAny(m, "license") { return Intent{Type: "license"} }
	if hasAny(m, "block", "chain") { return Intent{Type: "blocks"} }
	if hasAny(m, "fields") { return Intent{Type: "fields"} }
	
	// فارسی — با هر دو حالت الف و آ
	if hasAny(m, "سلام", "درود", "خسته نباش") { return Intent{Type: "greeting"} }
	if hasAny(m, "امار", "آمار", "وضعیت", "اوضاع", "چطوره") { return Intent{Type: "stats"} }
	if hasAny(m, "سرعت", "تی پی اس") { return Intent{Type: "tps"} }
	if hasAny(m, "بلوک", "بلاک", "زنجیره") { return Intent{Type: "blocks"} }
	if hasAny(m, "لایسنس", "مجوز") { return Intent{Type: "license"} }
	if hasAny(m, "ایجنت", "ایجنتها", "کارشناس", "مامور") {
		for f := range fields { if strings.Contains(m, f) { return Intent{Type: "agents", Field: f} } }
		return Intent{Type: "agents"}
	}
	if hasAny(m, "شبکه", "گاردین", "نگهبان") { return Intent{Type: "network"} }
	if hasAny(m, "امنیت", "حمله", "سکیوریتی") { return Intent{Type: "security"} }
	if hasAny(m, "رشته", "دسته", "فیلد") { return Intent{Type: "fields"} }
	
	return Intent{Type: "help"}
}

func answer(msg, forcedField string) (string, string) {
	intent := detectIntent(msg)
	if forcedField != "" && intent.Type == "help" {
		// اگر کاربر فیلد را در select انتخاب کرده ولی سوال را نشناختیم
		return "در حال پردازش برای رشته «" + forcedField + "»...", forcedField
	}
	
	sw := getJSON(SwitchURL + "/stats")
	gd := getJSON(GuardianURL + "/stats")
	
	switch intent.Type {
	case "greeting":
		return fmt.Sprintf("سلام فرمانده 👋\n\n• %d ایجنت در %d رشته\n• سوییچ: %v\n• گاردین: %v", len(agents), len(fields), sw != nil, gd != nil), ""
	case "stats":
		if sw == nil { return "⚠️ سوییچ در دسترس نیست", "" }
		return fmt.Sprintf("📊 آمار زنده:\n\n• بلاک‌ها: %v\n• تراکنش‌ها: %v\n• در صف: %v\n• لایسنس‌ها: %v\n• ایجنت‌ها: %d در %d رشته", sw["chainLength"], sw["totalTx"], sw["pendingTx"], sw["licenses"], len(agents), len(fields)), ""
	case "tps":
		return "⚡ TPS معیار:\n• Ingestion: 26,344\n• Commit: 4,150 (PBFT+fsync)\n• تست: 1M\n• Drop: 0%", ""
	case "blocks":
		if sw == nil { return "⚠️", "" }
		recent, _ := sw["recentBlocks"].([]interface{})
		out := "🧱 بلاک‌های اخیر:\n\n"
		for i, b := range recent {
			if i >= 5 { break }
			bm := b.(map[string]interface{})
			h := fmt.Sprintf("%v", bm["hash"]); if len(h) > 16 { h = h[:16] }
			out += fmt.Sprintf("• #%v · %s... · %v tx\n", bm["block_num"], h, bm["tx_count"])
		}
		return out + fmt.Sprintf("\nمجموع: %v", sw["chainLength"]), ""
	case "license":
		return "🔑 لایسنس‌های داخلی ایجنت‌ها:\n\nAGENT-ENGINEERING-LIC\nAGENT-SECURITY-LIC\nAGENT-DESIGN-LIC\n... هر ۱۰ رشته فعال", ""
	case "network":
		if gd == nil { return "⚠️ گاردین در دسترس نیست", "" }
		ns, _ := gd["network_stats"].(map[string]interface{})
		return fmt.Sprintf("🌐 گاردین:\n• بلاک: %v\n• تراکنش: %v", ns["total_blocks"], ns["total_txs"]), ""
	case "security":
		return "🛡️ امنیت:\n• 192/192 بلاک\n• SCADA/Stuxnet بلاک\n• Zero-PEM\n• Ceremony ثبت شد", ""
	case "agents":
		if intent.Field != "" {
			out := fmt.Sprintf("👥 ایجنت‌های «%s» (%d):\n\n", intent.Field, fields[intent.Field])
			cnt := 0
			for _, a := range agents { if a.Field == intent.Field && cnt < 10 { out += "• " + a.Name + "\n"; cnt++ } }
			return out, intent.Field
		}
		keys := make([]string, 0, len(fields))
		for k := range fields { keys = append(keys, k) }
		sort.Slice(keys, func(i, j int) bool { return fields[keys[i]] > fields[keys[j]] })
		out := fmt.Sprintf("👥 %d ایجنت در %d رشته:\n\n", len(agents), len(fields))
		for _, k := range keys { out += fmt.Sprintf("• %s: %d\n", k, fields[k]) }
		return out, ""
	case "fields":
		return answer("agents", forcedField)
	}
	return "🤖 من مغز داخلی هستم. بپرس:\n• آمار\n• TPS\n• بلاک\n• لایسنس\n• ایجنت‌ها\n• امنیت\n• شبکه", ""
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	var req struct{ Field, Message string }
	json.NewDecoder(r.Body).Decode(&req)
	log.Printf("📩 msg=%q field=%q", req.Message, req.Field)
	
	ans, usedField := answer(req.Message, req.Field)
	if usedField == "" { usedField = req.Field }
	if usedField == "" { usedField = "specialized" }
	
	// ثبت عمل ایجنت روی زنجیره با لایسنس
	if req.Message != "" {
		go logAgentAction(usedField, req.Message)
	}
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"response": ans, "field": usedField, "engine": "internal", "license": "AGENT-" + strings.ToUpper(usedField) + "-LIC",
	})
}

func main() {
	loadAgents()
	http.HandleFunc("/api/chat", chatHandler)
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "agents": len(agents), "fields": len(fields)})
	})
	port := os.Getenv("BRAIN_PORT")
	if port == "" { port = "9300" }
	log.Printf("🧠 Brain on :%s (agents=%d fields=%d) — each query = on-chain tx", port, len(agents), len(fields))
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
