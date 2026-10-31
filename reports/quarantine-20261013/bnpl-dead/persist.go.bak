// Horizon BNPL - Persistence (JSON file)
// Saves all in-memory state to disk on every mutation

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const stateFile = "./chaindata/state.json"

type persistedState struct {
	Txs       map[string]*CreditTx     `json:"txs"`
	Audit     []*AuditLog              `json:"audit"`
	Plans     map[string]*CreditPlan   `json:"plans"`
	Providers map[string]*Provider     `json:"providers"`
	Personnel map[string]*Personnel    `json:"personnel"`
	Orgs      map[string]*Organization `json:"orgs"`
	Blocks    []*MerkleBlock           `json:"blocks"`
}

func saveAll() error {
	mu.RLock()
	st := persistedState{
		Txs:       txs,
		Audit:     auditLog,
		Plans:     engine.plans,
		Providers: engine.providers,
		Personnel: engine.personnel,
		Orgs:      engine.orgs,
		Blocks:    blocks,
	}
	mu.RUnlock()

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(stateFile), 0755)
	tmp := stateFile + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, stateFile)
}

func loadAll() error {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var st persistedState
	if err := json.Unmarshal(data, &st); err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	if st.Txs != nil {
		txs = st.Txs
	}
	if st.Audit != nil {
		auditLog = st.Audit
	}
	if st.Plans != nil {
		engine.plans = st.Plans
	}
	if st.Providers != nil {
		engine.providers = st.Providers
	}
	if st.Personnel != nil {
		engine.personnel = st.Personnel
	}
	if st.Orgs != nil {
		engine.orgs = st.Orgs
	}
	return nil
}
