package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Signing Session — one-time-use token required for each signing operation.
// This provides an extra layer of protection beyond key unlock.

var (
	sessionMutex   sync.Mutex
	signingSession = make(map[string]time.Time)
	signingSessionTTL     = 5 * time.Minute
)

func generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func requestSigningSessionHandler(c *gin.Context) {
	tok := generateSessionToken()
	sessionMutex.Lock()
	signingSession[tok] = time.Now().Add(signingSessionTTL)
	sessionMutex.Unlock()

	addAudit("SESSION_ISSUED", tok[:16], "ttl=5m", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{
		"session_token": tok,
		"expires_in":    int(signingSessionTTL.Seconds()),
	})
}

func ConsumeSessionToken(tok string) bool {
	if tok == "" {
		return false
	}
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	exp, ok := signingSession[tok]
	if !ok {
		return false
	}
	delete(signingSession, tok)
	return time.Now().Before(exp)
}

func initSigningSessions() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for range ticker.C {
			sessionMutex.Lock()
			now := time.Now()
			for k, exp := range signingSession {
				if now.After(exp) {
					delete(signingSession, k)
				}
			}
			sessionMutex.Unlock()
		}
	}()
}
