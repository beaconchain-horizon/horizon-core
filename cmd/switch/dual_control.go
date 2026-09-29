package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Dual Control — for transactions above a threshold,
// two admin approvals are required before signing.

var (
	dualThreshold float64 = 1_000_000 // default: 1M units
	dualMutex     sync.Mutex
	dualPending   = make(map[string]*PendingApproval)
)

type PendingApproval struct {
	ID          string    `json:"id"`
	RequestData string    `json:"request_data"`
	RequestedBy string    `json:"requested_by"`
	RequestedAt time.Time `json:"requested_at"`
	ApprovedBy  string    `json:"approved_by,omitempty"`
	ApprovedAt  time.Time `json:"approved_at,omitempty"`
	Status      string    `json:"status"`
}

func initDualControl() {
	if v := os.Getenv("DUAL_CONTROL_THRESHOLD"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			dualThreshold = f
		}
	}
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			dualMutex.Lock()
			cutoff := time.Now().Add(-30 * time.Minute)
			for k, p := range dualPending {
				if p.RequestedAt.Before(cutoff) {
					delete(dualPending, k)
				}
			}
			dualMutex.Unlock()
		}
	}()
}

func RequiresDualControl(amount float64) bool {
	return amount >= dualThreshold
}

func requestApprovalHandler(c *gin.Context) {
	var req struct {
		Data string `json:"data" binding:"required"`
		By   string `json:"by" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h := sha256.Sum256([]byte(req.Data + req.By + time.Now().String()))
	id := hex.EncodeToString(h[:8])

	dualMutex.Lock()
	dualPending[id] = &PendingApproval{
		ID:          id,
		RequestData: req.Data,
		RequestedBy: req.By,
		RequestedAt: time.Now(),
		Status:      "pending",
	}
	dualMutex.Unlock()

	addAudit("DUAL_REQUEST", id, "by="+req.By, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"status": "pending", "id": id})
}

func approveHandler(c *gin.Context) {
	var req struct {
		ID string `json:"id" binding:"required"`
		By string `json:"by" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dualMutex.Lock()
	p, ok := dualPending[req.ID]
	if !ok {
		dualMutex.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if p.RequestedBy == req.By {
		dualMutex.Unlock()
		c.JSON(http.StatusBadRequest, gin.H{"error": "approver must differ from requester"})
		return
	}
	p.ApprovedBy = req.By
	p.ApprovedAt = time.Now()
	p.Status = "approved"
	dualMutex.Unlock()

	addAudit("DUAL_APPROVE", req.ID, "by="+req.By, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"status": "approved", "id": req.ID})
}

func listPendingHandler(c *gin.Context) {
	dualMutex.Lock()
	defer dualMutex.Unlock()
	out := make([]*PendingApproval, 0, len(dualPending))
	for _, p := range dualPending {
		out = append(out, p)
	}
	c.JSON(http.StatusOK, gin.H{"pending": out, "threshold": dualThreshold})
}

func _dualUnused() { _ = fmt.Sprintf("%v", dualThreshold) }
