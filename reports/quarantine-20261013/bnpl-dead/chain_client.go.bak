// Horizon BNPL - Client for connecting to Horizon Core switch

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	switchURL        = os.Getenv("SWITCH_URL")
	switchAdminToken = os.Getenv("SWITCH_ADMIN_TOKEN")
)

// sendAnchorToSwitch - best-effort push to Horizon Core
func sendAnchorToSwitch(blk *MerkleBlock) error {
	if switchURL == "" {
		log.Printf("[chain] SWITCH_URL not set, skipping push")
		return nil
	}
	if switchAdminToken == "" {
		log.Printf("[chain] SWITCH_ADMIN_TOKEN not set, skipping push")
		return nil
	}

	payload := map[string]interface{}{
		"block_height": blk.Height,
		"block_hash":   blk.Hash,
		"merkle_root":  blk.MerkleRoot,
		"fanout":       blk.Fanout,
		"tx_ids":       blk.TxIDs,
		"timestamp":    blk.Timestamp,
		"signature":    blk.Signature,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := switchURL + "/api/v1/bnpl/anchor"
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Token", switchAdminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("switch request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var out map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&out)
		return errors.New("switch returned " + resp.Status + ": " + fmt.Sprint(out))
	}

	log.Printf("[chain] pushed block %d to Horizon Core", blk.Height)
	return nil
}
