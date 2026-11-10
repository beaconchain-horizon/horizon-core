package main

import (
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

const AdminToken = "bench0000000000000000000000000000000000000000000000000000000bench"
const UploadDir = "./panel-uploads"

// ═══════════════ SECURITY: command whitelist ═══════════════
var allowedPrefixes = []string{
	"git ", "go ", "ls", "cat ", "grep ", "tail ", "head ", "ps ", "tasklist",
	"echo ", "pwd", "whoami", "date", "df ", "free ", "netstat",
	"curl ", "wget ", "./horizon-", "taskkill", "go run",
}
var blockedPatterns = []string{"rm -rf /", "format ", ":(){", "del /f", "shutdown", "reboot", "mkfs", "dd if=", "> /dev/sd", "chmod 777 /"}

func isAllowed(cmd string) (bool, string) {
	for _, b := range blockedPatterns {
		if strings.Contains(strings.ToLower(cmd), b) { return false, "blocked pattern: " + b }
	}
	for _, p := range allowedPrefixes {
		if strings.HasPrefix(strings.TrimSpace(cmd), p) { return true, "" }
	}
	return false, "command not whitelisted. Allowed prefixes: " + strings.Join(allowedPrefixes, ", ")
}

// ═══════════════ CORS ═══════════════
func cors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Admin-Token")
}

// ═══════════════ HANDLERS ═══════════════
func healthHandler(w http.ResponseWriter, r *http.Request) {
	cors(w); w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "panel-server", "time": time.Now().Format(time.RFC3339)})
}

func authCheck(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("X-Admin-Token") != AdminToken {
		cors(w); w.WriteHeader(401)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return false
	}
	return true
}

func terminalHandler(w http.ResponseWriter, r *http.Request) {
	cors(w); w.Header().Set("Content-Type", "application/json")
	if !authCheck(w, r) { return }
	var req struct{ Command string `json:"command"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(400); json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); return
	}
	ok, reason := isAllowed(req.Command)
	if !ok {
		json.NewEncoder(w).Encode(map[string]string{"stderr": "⛔ " + reason}); return
	}
	// Execute via bash
	cmd := exec.Command("bash", "-c", req.Command)
	cmd.Dir = "."
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout; cmd.Stderr = &stderr
	err := cmd.Run()
	resp := map[string]interface{}{
		"stdout": stdout.String(), "stderr": stderr.String(),
		"exit":   0,
	}
	if err != nil { resp["exit"] = 1; resp["stderr"] = stderr.String() + "\n" + err.Error() }
	json.NewEncoder(w).Encode(resp)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	cors(w); w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		w.WriteHeader(405); json.NewEncoder(w).Encode(map[string]string{"error": "POST only"}); return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 100<<20) // 100MB max
	file, header, err := r.FormFile("file")
	if err != nil { w.WriteHeader(400); json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); return }
	defer file.Close()
	os.MkdirAll(UploadDir, 0755)
	dst := filepath.Join(UploadDir, fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(header.Filename)))
	out, _ := os.Create(dst)
	defer out.Close()
	n, _ := io.Copy(out, file)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "uploaded", "file": filepath.Base(dst), "size": n,
	})
}

func generateImageHandler(w http.ResponseWriter, r *http.Request) {
	cors(w); w.Header().Set("Content-Type", "application/json")
	var req struct{ Prompt string `json:"prompt"` }
	json.NewDecoder(r.Body).Decode(&req)
	// Local: placeholder (could plug in Stable Diffusion / local model later)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "queued",
		"prompt":  req.Prompt,
		"url":     "", // will be filled by local image model
		"message": "Image generation requires a local model (Stable Diffusion / Ollama). Configure in panel-server.",
	})
}

func main() {
	os.MkdirAll(UploadDir, 0755)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/upload", uploadHandler)
	mux.HandleFunc("/terminal/exec", terminalHandler)
	mux.HandleFunc("/generate-image", generateImageHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve static panel
		http.FileServer(http.Dir("./panel")).ServeHTTP(w, r)
	})
	port := os.Getenv("PANEL_PORT")
	if port == "" { port = "9200" }
	log.Printf("🚀 Panel Server on :%s — serving ./panel + terminal API", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
