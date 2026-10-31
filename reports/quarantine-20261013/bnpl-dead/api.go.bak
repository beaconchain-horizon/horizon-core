// Horizon BNPL - HTTP API + Chain Integration

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	engine   *LOSEngine
	txs      = make(map[string]*CreditTx)
	auditLog []*AuditLog
	blocks   []*MerkleBlock
	mu       sync.RWMutex
)

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func hashStr(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func appendAudit(eventType, actorID, entityType, entityID string, payload interface{}) {
	mu.Lock()
	defer mu.Unlock()
	pb, _ := json.Marshal(payload)
	prevHash := "0"
	if len(auditLog) > 0 {
		prevHash = auditLog[len(auditLog)-1].CurrHash
	}
	entry := &AuditLog{
		ID:         int64(len(auditLog) + 1),
		EventType:  eventType,
		ActorID:    actorID,
		EntityType: entityType,
		EntityID:   entityID,
		Payload:    string(pb),
		PrevHash:   prevHash,
		CreatedAt:  time.Now(),
	}
	entry.CurrHash = hashStr(prevHash + entry.EventType + entry.EntityID + entry.Payload)
	auditLog = append(auditLog, entry)
}

type errString string

func (e errString) Error() string { return string(e) }

func validatePlan(p *CreditPlan) error {
	if p.Name == "" {
		return errString("name required")
	}
	if p.ProviderID == "" {
		return errString("provider_id required")
	}
	if p.MinAmount <= 0 || p.MaxAmount <= 0 {
		return errString("min_amount and max_amount required")
	}
	if p.MinAmount >= p.MaxAmount {
		return errString("min_amount must be less than max_amount")
	}
	if p.Installments <= 0 {
		return errString("installments must be positive")
	}
	now := time.Now()
	if p.StartDate.IsZero() {
		p.StartDate = now
	}
	if p.EndDate.IsZero() {
		p.EndDate = now.AddDate(1, 0, 0)
	}
	if p.EndDate.Before(p.StartDate) {
		return errString("end_date must be after start_date")
	}
	return nil
}

func enrichTxFromPlan(tx *CreditTx, plan *CreditPlan) error {
	if plan == nil {
		return errString("plan not found")
	}
	if tx.Amount < plan.MinAmount || tx.Amount > plan.MaxAmount {
		return errString("amount outside plan range")
	}
	if tx.ProviderID == "" {
		tx.ProviderID = plan.ProviderID
	}
	if tx.Installments == 0 {
		tx.Installments = plan.Installments
	}
	if tx.InterestRate == 0 {
		tx.InterestRate = plan.InterestRate
	}
	total := float64(tx.Amount) * (1 + tx.InterestRate)
	tx.InstallAmount = int64(total / float64(tx.Installments))
	return nil
}

// POST /api/v1/plans
func handleCreatePlan(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJSON(w, 405, map[string]string{"error": "POST only"})
		return
	}
	var p CreditPlan
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	if p.PlanID == "" {
		p.PlanID = newID("plan")
	}
	if err := validatePlan(&p); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	p.CreatedAt = time.Now()
	if p.Status == "" {
		p.Status = "active"
	}
	engine.AddPlan(&p)
	appendAudit("plan_created", "admin", "plan", p.PlanID, p)
	saveAll()
	writeJSON(w, 200, map[string]interface{}{"ok": true, "plan": p})
}

// POST /api/v1/los/apply
func handleApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJSON(w, 405, map[string]string{"error": "POST only"})
		return
	}
	var req LOSApplication
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	if err := engine.Submit(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	appendAudit("los_applied", req.PersonnelID, "application", req.AppID, req)
	saveAll()
	writeJSON(w, 200, map[string]interface{}{"ok": true, "application": req})
}

// GET /api/v1/los/app/{id}
func handleGetApp(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/los/app/")
	a, ok := engine.GetApp(id)
	if !ok {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, 200, a)
}

// POST /api/v1/tx/create
func handleCreateTx(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJSON(w, 405, map[string]string{"error": "POST only"})
		return
	}
	var tx CreditTx
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	if tx.TxID == "" {
		tx.TxID = newID("tx")
	}
	if tx.Nonce == "" {
		tx.Nonce = newID("n")
	}

	plan, ok := engine.GetPlan(tx.PlanID)
	if !ok {
		writeJSON(w, 400, map[string]string{"error": "plan not found"})
		return
	}
	if err := enrichTxFromPlan(&tx, plan); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	tx.CreatedAt = time.Now()
	if tx.Status == "" {
		tx.Status = "pending"
	}

	sig, err := signData(txSignablePayload(&tx))
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "sign failed: " + err.Error()})
		return
	}
	tx.Signature = sig
	tx.PubKey = issuerPubPEM
	tx.Address = issuerAddress

	mu.Lock()
	txs[tx.TxID] = &tx
	mu.Unlock()

	appendAudit("tx_created", tx.PersonnelID, "tx", tx.TxID, tx)
	saveAll()
	writeJSON(w, 200, map[string]interface{}{"ok": true, "tx": tx})
}

// GET /api/v1/tx/{id}
func handleGetTx(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/tx/")
	mu.RLock()
	tx, ok := txs[id]
	mu.RUnlock()
	if !ok {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, 200, tx)
}

// GET /api/v1/tx/verify/{id}
func handleVerifyTx(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/tx/verify/")
	mu.RLock()
	tx, ok := txs[id]
	mu.RUnlock()
	if !ok {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	valid := verifyData(tx.PubKey, txSignablePayload(tx), tx.Signature)
	writeJSON(w, 200, map[string]interface{}{
		"tx_id":     tx.TxID,
		"valid":     valid,
		"signature": tx.Signature,
	})
}

// GET /api/v1/issuer/pubkey
func handleIssuerPubKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(issuerPubPEM))
}

// GET /api/v1/audit/log?limit=20
func handleAuditLog(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	start := len(auditLog) - limit
	if start < 0 {
		start = 0
	}
	writeJSON(w, 200, auditLog[start:])
}

// GET /api/v1/audit/verify/{id}
func handleAuditVerify(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/audit/verify/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	mu.RLock()
	defer mu.RUnlock()
	if id < 1 || id > int64(len(auditLog)) {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	entry := auditLog[id-1]
	expected := hashStr(entry.PrevHash + entry.EventType + entry.EntityID + entry.Payload)
	writeJSON(w, 200, map[string]interface{}{
		"id":        entry.ID,
		"integrity": expected == entry.CurrHash,
		"prev_hash": entry.PrevHash,
		"curr_hash": entry.CurrHash,
	})
}

// GET /api/v1/stats
func handleStats(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	var total int64
	for _, t := range txs {
		total += t.Amount
	}
	writeJSON(w, 200, map[string]interface{}{
		"plans":     len(engine.plans),
		"personnel": len(engine.personnel),
		"txs":       len(txs),
		"audit":     len(auditLog),
		"volume":    total,
	})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func seedDemo() {
	engine.AddProvider(&Provider{
		ProviderID: "prov_1", Name: "Bank X", Type: "bank",
		CreditLimit: 100_000_000_000, Status: "active",
	})
	engine.AddPlan(&CreditPlan{
		PlanID: "plan_home", Name: "Home Appliance", Type: "public",
		ProviderID: "prov_1", MinAmount: 5_000_000, MaxAmount: 500_000_000,
		Installments: 12, InterestRate: 0.18,
		StartDate: time.Now().Add(-30 * 24 * time.Hour),
		EndDate:   time.Now().Add(180 * 24 * time.Hour),
		Status:    "active", CreatedAt: time.Now(),
	})
	engine.AddPersonnel(&Personnel{
		PersonnelID: "p_1", NationalID: "001", FullName: "Ali",
		CreditLimit: 200_000_000, CreditScore: 720, Status: "active",
	})
}

func main() {
	if err := loadKeyFromEnvOrFile("./chaindata/issuer.pem"); err != nil {
		log.Fatal("key init failed: ", err)
	}
	log.Printf("Issuer key loaded: %s", issuerAddress)

	engine = NewLOSEngine()
	if err := loadAll(); err != nil {
		log.Printf("load state failed: %v", err)
	}

	// اگر خالی بود، seed کن
	mu.RLock()
	empty := len(engine.plans) == 0 && len(engine.personnel) == 0
	mu.RUnlock()
	if empty {
		seedDemo()
		saveAll()
		log.Printf("seeded demo data")
	} else {
		log.Printf("loaded state: %d plans, %d personnel, %d txs, %d audit",
			len(engine.plans), len(engine.personnel), len(txs), len(auditLog))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/plans", handleCreatePlan)
	mux.HandleFunc("/api/v1/los/apply", handleApply)
	mux.HandleFunc("/api/v1/los/app/", handleGetApp)
	mux.HandleFunc("/api/v1/tx/create", handleCreateTx)
	mux.HandleFunc("/api/v1/tx/verify/", handleVerifyTx)
	mux.HandleFunc("/api/v1/tx/", handleGetTx)
	mux.HandleFunc("/api/v1/issuer/pubkey", handleIssuerPubKey)
	mux.HandleFunc("/api/v1/issuer/address", handleIssuerAddress)
	mux.HandleFunc("/api/v1/audit/log", handleAuditLog)
	mux.HandleFunc("/api/v1/audit/verify/", handleAuditVerify)
	mux.HandleFunc("/api/v1/stats", handleStats)
	mux.HandleFunc("/api/v1/anchor", handleAnchor)
	mux.HandleFunc("/api/v1/blocks", handleBlocks)
	mux.HandleFunc("/api/v1/proof/", handleProof)

	port := "8090"
	log.Printf("Horizon BNPL API on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, cors(mux)))
}

// POST /api/v1/anchor?fanout=2
// Anchors all pending txs into one block
func handleAnchor(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJSON(w, 405, map[string]string{"error": "POST only"})
		return
	}
	fanout := 2
	if v := r.URL.Query().Get("fanout"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 2 && n <= 3 {
			fanout = n
		}
	}

	mu.RLock()
	var pending []string
	for id, tx := range txs {
		if tx.BlockHash == "" {
			pending = append(pending, id)
		}
	}
	mu.RUnlock()

	if len(pending) == 0 {
		writeJSON(w, 400, map[string]string{"error": "no pending txs"})
		return
	}

	blk, err := anchorTxs(pending, fanout)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	appendAudit("block_anchored", "system", "block", blk.Hash, blk)
	saveAll()
	go sendAnchorToSwitch(blk)
	writeJSON(w, 200, map[string]interface{}{"ok": true, "block": blk})
}

// GET /api/v1/blocks?limit=20
func handleBlocks(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	start := len(blocks) - limit
	if start < 0 {
		start = 0
	}
	out := make([]*MerkleBlock, 0, len(blocks)-start)
	for i := len(blocks) - 1; i >= start; i-- {
		out = append(out, blocks[i])
	}
	writeJSON(w, 200, out)
}

// GET /api/v1/proof/{tx_id}?fanout=2
func handleProof(w http.ResponseWriter, r *http.Request) {
	txID := strings.TrimPrefix(r.URL.Path, "/api/v1/proof/")
	fanout := 2
	if v := r.URL.Query().Get("fanout"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 2 && n <= 3 {
			fanout = n
		}
	}

	mu.RLock()
	tx, ok := txs[txID]
	if !ok {
		mu.RUnlock()
		writeJSON(w, 404, map[string]string{"error": "tx not found"})
		return
	}
	var blk *MerkleBlock
	for _, b := range blocks {
		if b.Hash == tx.BlockHash {
			blk = b
			break
		}
	}
	mu.RUnlock()

	if blk == nil {
		writeJSON(w, 400, map[string]string{"error": "tx not anchored yet"})
		return
	}

	proof, err := merkleProof(blk.TxIDs, txID, fanout)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, 200, map[string]interface{}{
		"tx_id":       txID,
		"block_hash":  blk.Hash,
		"merkle_root": blk.MerkleRoot,
		"fanout":      fanout,
		"proof":       proof,
		"leaf_hash":   leafHash(txID),
	})
}

// GET /api/v1/issuer/address
func handleIssuerAddress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(issuerAddress))
}
