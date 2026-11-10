package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const GuardianURL = "http://localhost:9008"
const AgentDir = "../angel-agents"

type Agent struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Field string `json:"field"`
	File  string `json:"file"`
	Goal  string `json:"goal"`
}

var agents []*Agent
var fields = make(map[string]int)

func loadAgents() {
	skipDirs := map[string]bool{".git": true, ".github": true, "docs": true, "examples": true}
	skipFiles := map[string]bool{"README.md": true, "CONTRIBUTING.md": true, "LICENSE": true, "HORIZON_ANGEL.md": true, "CODE_OF_CONDUCT.md": true, "SECURITY.md": true, "AGENT_REGISTRY.md": true, "PULL_REQUEST_TEMPLATE.md": true}

	filepath.Walk(AgentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil { return nil }
		if info.IsDir() {
			if skipDirs[info.Name()] { return filepath.SkipDir }
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".md") || skipFiles[info.Name()] { return nil }

		rel, _ := filepath.Rel(AgentDir, path)
		parts := strings.Split(rel, string(os.PathSeparator))
		if len(parts) < 2 { return nil }
		field := parts[0]

		data, _ := ioutil.ReadFile(path)
		lines := strings.Split(string(data), "\n")
		name, goal := strings.TrimSuffix(info.Name(), ".md"), ""
		for i, l := range lines {
			if strings.HasPrefix(l, "# ") {
				name = strings.TrimPrefix(l, "# ")
				for j := i + 1; j < len(lines); j++ {
					t := strings.TrimSpace(lines[j])
					if t != "" && !strings.HasPrefix(t, "#") {
						goal = t
						break
					}
				}
				break
			}
		}

		agents = append(agents, &Agent{
			ID: strings.TrimSuffix(info.Name(), ".md"), Name: name, Field: field, File: rel, Goal: goal,
		})
		fields[field]++
		return nil
	})
	log.Printf("✅ Loaded %d agents across %d fields.", len(agents), len(fields))
}

func guardianOnline() bool {
	c := &http.Client{Timeout: 2 * time.Second}
	r, err := c.Get(GuardianURL + "/health")
	if err != nil { return false }
	defer r.Body.Close()
	return r.StatusCode == 200
}

func reportToGuardian(action, field, status string) {
	payload := map[string]interface{}{"agent_id": "runner", "status": status, "tasks": 1, "violations": 0}
	b, _ := json.Marshal(payload)
	c := &http.Client{Timeout: 2 * time.Second}
	c.Post(GuardianURL+"/inspect", "application/json", bytes.NewBuffer(b))
}

func processCommand(field, command string) (string, []*Agent) {
	var matches []*Agent
	for _, a := range agents {
		if strings.EqualFold(a.Field, field) { matches = append(matches, a) }
	}
	if len(matches) == 0 {
		return fmt.Sprintf("No agents found for field '%s'.", field), nil
	}
	return fmt.Sprintf("Command assigned to %d agents in field '%s'.", len(matches), field), matches
}

func watchGitDirectives(dir string) {
	log.Printf("👁 Watching Git directives in: %s", dir)
	os.MkdirAll(filepath.Join(dir, "responses"), 0755)
	os.MkdirAll(filepath.Join(dir, "processed"), 0755)

	for {
		files, _ := filepath.Glob(filepath.Join(dir, "*.json"))
		for _, f := range files {
			if !guardianOnline() { continue }
			data, _ := ioutil.ReadFile(f)
			var cmd struct { Field string `json:"field"`; Command string `json:"command"` }
			if err := json.Unmarshal(data, &cmd); err != nil { continue }

			msg, matches := processCommand(cmd.Field, cmd.Command)
			reportToGuardian("GIT_COMMAND", cmd.Field, msg)

			resp := map[string]interface{}{
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"command": cmd.Command, "field": cmd.Field, "result": msg, "agents": matches,
			}
			respBytes, _ := json.MarshalIndent(resp, "", "  ")
			respPath := filepath.Join(dir, "responses", fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(f)))
			ioutil.WriteFile(respPath, respBytes, 0644)
			os.Rename(f, filepath.Join(dir, "processed", filepath.Base(f)))
		}
		time.Sleep(3 * time.Second)
	}
}

func callHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !guardianOnline() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "Guardian offline. Fail-closed."})
		return
	}
	var req struct { Field string `json:"field"`; Command string `json:"command"` }
	json.NewDecoder(r.Body).Decode(&req)
	msg, matches := processCommand(req.Field, req.Command)
	reportToGuardian("TERMINAL_COMMAND", req.Field, msg)
	json.NewEncoder(w).Encode(map[string]interface{}{"result": msg, "agents": matches})
}

func main() {
	loadAgents()
	go watchGitDirectives(filepath.Join(AgentDir, "directives"))

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok", "agents": len(agents), "fields": len(fields), "guardian_online": guardianOnline(),
		})
	})
	http.HandleFunc("/call", callHandler)
	http.HandleFunc("/fields", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fields)
	})

	port := os.Getenv("AGENT_RUNNER_PORT")
	if port == "" { port = "9101" }
	log.Printf("🚀 Agent Runner on :%s — supervised by Guardian :9008", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
