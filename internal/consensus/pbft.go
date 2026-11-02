package consensus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type BlockProposal struct {
	BlockNumber uint64 `json:"block_number"`
	BlockHash   string `json:"block_hash"`
	ProposerID  string `json:"proposer_id"`
}

type Vote struct {
	VoterID   string `json:"voter_id"`
	BlockHash string `json:"block_hash"`
	Signature string `json:"signature"`
}

var (
	VotePool = make(map[string][]Vote)
	Mutex    sync.Mutex
)

func BroadcastBlock(proposal BlockProposal, peers []string) {
	payload, _ := json.Marshal(proposal)
	for _, peer := range peers {
		go func(p string) {
			url := fmt.Sprintf("http://%s/api/v1/consensus/receive-block", p)
			req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-License-ID", "INTERNAL-CONSENSUS")
			client := &http.Client{Timeout: 2 * time.Second}
			resp, err := client.Do(req)
			if err == nil { resp.Body.Close() }
		}(peer)
	}
}

func SendVote(vote Vote, proposerURL string) {
	payload, _ := json.Marshal(vote)
	url := fmt.Sprintf("http://%s/api/v1/consensus/receive-vote", proposerURL)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-License-ID", "INTERNAL-CONSENSUS")
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err == nil { resp.Body.Close() }
}

func AddVote(blockHash string, vote Vote) (bool, int) {
	Mutex.Lock()
	defer Mutex.Unlock()
	VotePool[blockHash] = append(VotePool[blockHash], vote)
	count := len(VotePool[blockHash])
	if count >= 5 {
		log.Printf("🎉 BLOCK %s FINALIZED with %d votes (5/7 Quorum reached)!", blockHash[:8], count)
		return true, count
	}
	return false, count
}
