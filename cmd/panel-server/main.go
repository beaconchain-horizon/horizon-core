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
	AdminToken  = "bench0000000000000000000000000000000000000000000000000000000bench"
	UploadDir   = "./panel-uploads"
	AgentDir    = "../angel-agents"
	LMStudioURL = "http://localhost:1234/v1/chat/completions"
)

func lmStudioHealth() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:1234/v1/models")
	if err != nil { return false }
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	
	var req struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(400); json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); return
	}

	if !lmStudioHealth() {
		json.NewEncoder(w).Encode(map[string]string{"error": "LM Studio offline. Start the Local Server on port 1234."})
		return
	}

	apiKey := os.Getenv("LM_STUDIO_API_KEY")
	if apiKey == "" {
		json.NewEncoder(w).Encode(map[string]string{"error": "LM_STUDIO_API_KEY env variable not set."})
		return
	}

	systemPrompt := "You are Horizon Core's AI assistant. Reply in Persian if asked in Persian."
	if req.Field != "" {
		agentFiles, _ := filepath.Glob(filepath.Join(AgentDir, req.Field, "*.md"))
		if len(agentFiles) > 0 {
			data, _ := os.ReadFile(agentFiles[0])
			systemPrompt = "You are an expert AI agent in the '" + req.Field + "' field. Base your answer on this persona:\n\n" + string(data)
			if len(systemPrompt) > 8000 { systemPrompt = systemPrompt[:8000] }
		}
	}

	lmReq := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": req.Message},
		},
		"temperature": 0.7,
		"stream":      false,
	}
	
	body, _ := json.Marshal(lmReq)
	client := &http.Client{Timeout: 120 * time.Second}
	httpReq, _ := http.NewRequest("POST", LMStudioURL, bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "LM Studio error: " + err.Error()}); return
	}
	defer resp.Body.Close()
	
	var lmResp struct {
		Choices []struct {
			Message struct { Content string `json:"content"` } `json:"message"`
		} `json:"choices"`
	}
	json.NewDecoder(resp.Body).Decode(&lmResp)

	if len(lmResp.Choices) == 0 {
		json.NewEncoder(w).Encode(map[string]string{"error": "Empty response from LM Studio"}); return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"response": lmResp.Choices[0].Message.Content, "field": req.Field, "model": "LM Studio",
	})
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	var switchStats, guardianStats map[string]interface{}
	respS, errS := http.Get("http://localhost:8091/stats")
	if errS == nil { json.NewDecoder(respS.Body).Decode(&switchStats); respS.Body.Close() }
	respG, errG := http.Get("http://localhost:9008/stats")
	if errG == nil { json.NewDecoder(respG.Body).Decode(&guardianStats); respG.Body.Close() }
	json.NewEncoder(w).Encode(map[string]interface{}{"switch": switchStats, "guardian": guardianStats})
}

func terminalHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("X-Admin-Token") != AdminToken {
		w.WriteHeader(401); json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"}); return
	}
	var req struct{ Command string `json:"command"` }
	json.NewDecoder(r.Body).Decode(&req)
	
	cmd := exec.Command("bash", "-c", req.Command)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout; cmd.Stderr = &stderr
	err := cmd.Run()
	res := map[string]interface{}{"stdout": stdout.String(), "stderr": stderr.String(), "exit": 0}
	if err != nil { res["exit"] = 1 }
	json.NewEncoder(w).Encode(res)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
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
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok", "lm_studio": lmStudioHealth(), "time": time.Now().Format(time.RFC3339),
		})
	})
	http.Handle("/", http.FileServer(http.Dir("./panel")))

	port := os.Getenv("PANEL_PORT")
	if port == "" { port = "9200" }
	log.Printf("🚀 Panel: http://localhost:%s | LM Studio: %v", port, lmStudioHealth())
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
