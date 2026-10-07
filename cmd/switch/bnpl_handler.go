// Horizon Switch - BNPL Anchor Handler
// Receives BNPL Merkle blocks and appends them to the Horizon Core chain

package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const TxBNPLAnchor = "bnpl.anchor"

// POST /api/v1/bnpl/anchor
func bnplAnchorHandler(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	stateMutex.RLock()
	key := unlockedKey
	stateMutex.RUnlock()

	if key == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "signer key locked"})
		return
	}

	tx, block, err := appendChainTx(db, key, TxBNPLAnchor, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"tx_id":   tx.TxID,
		"tx_hash": tx.TxHash,
		"block":   block,
	})
}
