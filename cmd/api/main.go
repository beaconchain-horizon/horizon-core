package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"time"

	"horizon-core/internal/auth"
	"horizon-core/internal/consensus"
	"horizon-core/internal/db"
	"github.com/gin-gonic/gin"
)

func main() {
	db.Init()
	port := os.Getenv("PORT")
	if port == "" { port = "9001" }
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" { nodeID = "node-1" }

	r := gin.Default()
	r.Use(auth.LicenseAuthMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "online", "node": nodeID})
	})

	r.POST("/api/v1/consensus/propose", func(c *gin.Context) {
		blockHash := fmt.Sprintf("%x", sha256.Sum256([]byte(time.Now().String())))
		proposal := consensus.BlockProposal{
			BlockNumber: uint64(time.Now().Unix()),
			BlockHash:   blockHash,
			ProposerID:  nodeID,
		}
		consensus.AddVote(blockHash, consensus.Vote{VoterID: nodeID, BlockHash: blockHash})
		peers := []string{"127.0.0.1:9001", "127.0.0.1:9002", "127.0.0.1:9003", "127.0.0.1:9004", "127.0.0.1:9005", "127.0.0.1:9006", "127.0.0.1:9007"}
		consensus.BroadcastBlock(proposal, peers)
		c.JSON(200, gin.H{"status": "proposed", "hash": blockHash})
	})

	r.POST("/api/v1/consensus/receive-block", func(c *gin.Context) {
		log.Printf("📥 [ReceiveBlock] Got request")
		var proposal consensus.BlockProposal
		if err := c.ShouldBindJSON(&proposal); err != nil {
			c.JSON(400, gin.H{"error": "Invalid block"})
			return
		}
		vote := consensus.Vote{VoterID: nodeID, BlockHash: proposal.BlockHash}
		consensus.SendVote(vote, "127.0.0.1:9001")
		c.JSON(200, gin.H{"status": "voted"})
	})

	r.POST("/api/v1/consensus/receive-vote", func(c *gin.Context) {
		var vote consensus.Vote
		if err := c.ShouldBindJSON(&vote); err != nil {
			c.JSON(400, gin.H{"error": "Invalid vote"})
			return
		}
		isFinalized, count := consensus.AddVote(vote.BlockHash, vote)
		if isFinalized {
			log.Printf("🎉 BLOCK %s FINALIZED!", vote.BlockHash[:8])
		}
		c.JSON(200, gin.H{"status": "vote_added", "total_votes": count})
	})

	log.Printf("✅ %s listening on :%s", nodeID, port)
	r.Run(":" + port)
}
