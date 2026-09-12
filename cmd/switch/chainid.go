package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ChainConfig struct {
	ChainID        string `json:"chain_id"`
	NetworkName    string `json:"network_name"`
	CreatedAt      string `json:"created_at"`
	AdminPublicKey string `json:"admin_public_key,omitempty"`
	Description    string `json:"description,omitempty"`
	AirGapAllowed  bool   `json:"air_gap_allowed"`
}

var (
	chainConfig   *ChainConfig
	chainConfigMu sync.RWMutex
)

func defaultChainConfig() *ChainConfig {
	return &ChainConfig{
		ChainID:       "horizon-default",
		NetworkName:   "Horizon Public Network",
		CreatedAt:     time.Now().Format("2006-01-02"),
		AirGapAllowed: false,
	}
}

func initChainConfig() {
	path := os.Getenv("CHAIN_CONFIG")
	if path == "" {
		path = "./config/chain.json"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("chain config not found (%s), using default", path)
		chainConfig = defaultChainConfig()
		return
	}

	var cfg ChainConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("chain config parse error: %v, using default", err)
		chainConfig = defaultChainConfig()
		return
	}

	if cfg.ChainID == "" {
		cfg.ChainID = "horizon-default"
	}
	if cfg.CreatedAt == "" {
		cfg.CreatedAt = time.Now().Format("2006-01-02")
	}
	chainConfig = &cfg
	log.Printf("🔗 Chain: %s (%s) | AirGapAllowed: %v",
		cfg.ChainID, cfg.NetworkName, cfg.AirGapAllowed)
}

func GetChainConfig() *ChainConfig {
	chainConfigMu.RLock()
	defer chainConfigMu.RUnlock()
	if chainConfig == nil {
		return defaultChainConfig()
	}
	return chainConfig
}

func chainIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Chain-ID", GetChainConfig().ChainID)
		c.Next()
	}
}

func chainInfoHandler(c *gin.Context) {
	cfg := GetChainConfig()
	c.JSON(http.StatusOK, gin.H{
		"chain_id":        cfg.ChainID,
		"network_name":    cfg.NetworkName,
		"created_at":      cfg.CreatedAt,
		"description":     cfg.Description,
		"admin_pubkey":    cfg.AdminPublicKey,
		"air_gap_mode":    airGapMode.Load(),
		"air_gap_allowed": cfg.AirGapAllowed,
		"air_gap_blocks":  airGapBlocks.Load(),
	})
}
