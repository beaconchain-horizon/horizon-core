package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	AdminToken   = "bench0000000000000000000000000000000000000000000000000000000bench"
	UploadDir    = "./panel-uploads"
	OllamaURL    = "http://localhost:11434"
	OllamaModel  = "qwen2.5:7b" // مدل پیش‌فرض
	AgentDir     = "../angel-agents"
)

// لیست سفید برای امنیت ترمینال
var allowedPrefixes = []string{"git ", "go ", "ls", "cat ", "grep ", "tail ", "head ", "ps ", "tasklist", "echo ", "pwd", "whoami", "date", "df ", "free ", "netstat", "curl ", "wget ", "./horizon-", "taskkill", "go run"}
var blockedPatterns = []string{"rm -rf /", "format ", ":(){", "del /f", "shutdown", "reboot", "mkfs", "dd if=", "> /dev/sd", "chmod 777 /"}

func isAllowed(cmd string) (bool, string) {
	for _, b := range blockedPatterns {
		if strings.Contains(strings.ToLower(cmd), b) { return false, "blocked: " + b }
	}
	for _, p := range allowedPrefixes {
		if strings.HasPrefix(strings.TrimSpace(cmd), p) { return true, "" }
	}
	return false, "not whitelisted"
}

func cors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Admin-Token")
}

func ollamaHealth() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(OllamaURL + "/api/tags")
	if err != nil { return false }
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// هندلر چت (اتصال به ایجنت و اولاما)
func chatHandler(w http.ResponseWriter, r *http.Request) {
	cors(w); w.Header().Set("Content-Type", "application/json")
	var req struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(400); json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); return
	}

	if !ollamaHealth() {
		json.NewEncoder(w).Encode(map[string]string{"error": "Ollama offline. Run 'ollama serve' in another terminal."})
		return
	}

	// خواندن شخصیت ایجنت از فایل .md
	systemPrompt := "You are Horizon Core's AI assistant. Reply in Persian if asked in Persian. Keep answers concise."
	if req.Field != "" {
		agentFiles, _ := filepath.Glob(filepath.Join(AgentDir, req.Field, "*.md"))
		if len(agentFiles) > 0 {
			data, _ := os.ReadFile(agentFiles[0])
			systemPrompt = "You are an expert AI agent in the '" + req.Field + "' field. Base your answer on this persona:\n\n" + string(data)
			if len(systemPrompt) > 8000 { systemPrompt = systemPrompt[:8000] }
		}
	}

	ollamaReq := map[string]interface{}{
		"model": OllamaModel, "prompt": req.Message, "system": systemPrompt, "stream": false,
		"options": map[string]interface{}{"temperature": 0.7},
	}
	body, _ := json.Marshal(ollamaReq)
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Post(OllamaURL+"/api/generate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Ollama error: " + err.Error()}); return
	}
	defer resp.Body.Close()
	
	var ollamaResp struct { Response string `json:"response"` }
	json.NewDecoder(resp.Body).Decode(&ollamaResp)

	json.NewEncoder(w).Encode(map[string]string{
		"response": ollamaResp.Response, "field": req.Field, "model": OllamaModel,
	})
}

// هندلر آمار (اتصال به سوییچ و نگهبان)
func statsHandler(w http.ResponseWriter, r *http.Request) {
	cors(w); w.Header().Set("Content-Type", "application/json")
	
	// آمار سوییچ
	var switchStats map[string]interface{}
	respS, errS := http.Get("http://localhost:8091/stats")
	if errS == nil {
		json.NewDecoder(respS.Body).Decode(&switchStats)
		respS.Body.Close()
	}

	// آمار نگهبان
	var guardianStats map[string]interface{}
	respG, errG := http.Get("http://localhost:9008/stats")
	if errG == nil {
		json.NewDecoder(respG.Body).Decode(&guardianStats)
		respG.Body.Close()
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"switch": switchStats, "guardian": guardianStats,
	})
}

func terminalHandler(w http.ResponseWriter, r *http.Request) {
	cors(w); w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("X-Admin-Token") != AdminToken {
		w.WriteHeader(401); json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"}); return
	}
	var req struct{ Command string `json:"command"` }
	json.NewDecoder(r.Body).Decode(&req)
	
	if ok, reason := isAllowed(req.Command); !ok {
		json.NewEncoder(w).Encode(map[string]string{"stderr": "⛔ " + reason}); return
	}
	
	cmd := exec.Command("bash", "-c", req.Command)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout; cmd.Stderr = &stderr
	err := cmd.Run()
	res := map[string]interface{}{"stdout": stdout.String(), "stderr": stderr.String(), "exit": 0}
	if err != nil { res["exit"] = 1 }
	json.NewEncoder(w).Encode(res)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	cors(w); w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 200<<20)
	file, header, err := r.FormFile("file")
	if err != nil { w.WriteHeader(400); json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); return }
	defer file.Close()
	
	os.MkdirAll(UploadDir, 0755)
	dst := filepath.Join(UploadDir, fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(header.Filename)))
	out, _ := os.Create(dst); defer out.Close()
	io.Copy(out, file)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "uploaded", "file": filepath.Base(dst)})
}

func main() {
	os.MkdirAll(UploadDir, 0755)
	http.HandleFunc("/api/chat", chatHandler)
	http.HandleFunc("/api/stats", statsHandler)
	http.HandleFunc("/api/terminal", terminalHandler)
	http.HandleFunc("/api/upload", uploadHandler)
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		cors(w); w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok", "ollama": ollamaHealth(), "time": time.Now().Format(time.RFC3339),
		})
	})
	
	// سرو کردن فایل HTML
	http.Handle("/", http.FileServer(http.Dir("./panel")))

	port := os.Getenv("PANEL_PORT")
	if port == "" { port = "9200" }
	log.Printf("🚀 Super App Panel: http://localhost:%s | Ollama: %v", port, ollamaHealth())
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
