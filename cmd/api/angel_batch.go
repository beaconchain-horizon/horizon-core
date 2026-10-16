package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"horizon-core/internal/chain"
)

type BatchVote struct {
	Critical       int
	CriticalTotal  int
	Important      int
	ImportantTotal int
	Passed         bool
	Root           string
}

func BatchRoot(txs []chain.PendingTx) string {
	h := sha256.New()
	for _, tx := range txs {
		h.Write([]byte(tx.TxID))
		h.Write([]byte(tx.Payload))
		h.Write([]byte(tx.Signature))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func RunAngelConsensusBatch(txs []chain.PendingTx) BatchVote {
	var result BatchVote
	if len(txs) == 0 {
		result.Passed = true
		return result
	}
	root := BatchRoot(txs)
	result.Root = root

	h := sha256.Sum256([]byte(root))
	sig := hex.EncodeToString(h[:])

	payload := AngelPayload{
		Left:            root,
		Right:           fmt.Sprintf("%d", len(txs)),
		Header:          root,
		Message:         root,
		Signature:       sig,
		MerkleType:      "binary",
		Used:            0,
		Volume:          999999999,
		IssuedAt:        1,
		ExpiresAt:       9999999999,
		UserID:          "batch",
		ProductID:       "batch",
		ExpectedUser:    "batch",
		ExpectedProduct: "batch",
		Mode:            "batch",
	}
	body, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 3 * time.Second}
	votes := make(map[int]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup

	all := make([]int, 0, len(angelCritical)+len(angelImportant))
	all = append(all, angelCritical...)
	all = append(all, angelImportant...)

	for _, port := range all {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			url := fmt.Sprintf("http://localhost:%d/verify", p)
			resp, err := client.Post(url, "application/json", bytes.NewReader(body))
			ok := false
			if err == nil {
				var r AngelResp
				json.NewDecoder(resp.Body).Decode(&r)
				resp.Body.Close()
				ok = r.Vote
			}
			mu.Lock()
			votes[p] = ok
			mu.Unlock()
		}(port)
	}
	wg.Wait()

	for _, p := range angelCritical {
		result.CriticalTotal++
		if votes[p] {
			result.Critical++
		}
	}
	for _, p := range angelImportant {
		result.ImportantTotal++
		if votes[p] {
			result.Important++
		}
	}
	result.Passed = result.Critical == result.CriticalTotal && result.Important >= 2
	return result
}
