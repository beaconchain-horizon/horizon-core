package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"horizon-core/internal/chain"
)

type PricingTier struct {
	Name       string `json:"name"`
	TomanPerTx int64  `json:"toman_per_tx"`
	PerLicense string `json:"per_license"`
}

// Live IRT/USD rate — update when needed.
const HorizonTomanPerUSD = 267000

var horizonPricingTiers = []PricingTier{
	{Name: "Standard", TomanPerTx: 20000, PerLicense: "20,000 Toman per license"},
	{Name: "Business", TomanPerTx: 30000, PerLicense: "30,000 Toman per license"},
	{Name: "Premium", TomanPerTx: 40000, PerLicense: "40,000 Toman per license"},
	{Name: "Enterprise", TomanPerTx: 50000, PerLicense: "50,000 Toman per license"},
}

func registerPricingHandlers(r *gin.Engine) {
	g := r.Group("/api/v1/pricing")

	g.GET("", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"currency":      "IRT",
			"exchange_rate": gin.H{"toman_per_usd": HorizonTomanPerUSD, "updated": "2026-10-13"},
			"tiers":         horizonPricingTiers,
		})
	})

	g.GET("/calculate", func(c *gin.Context) {
		n, err := strconv.ParseInt(c.DefaultQuery("transactions", "0"), 10, 64)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid count"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"transactions":  n,
			"exchange_rate": HorizonTomanPerUSD,
			"results":       calculateTiers(n),
		})
	})

	g.GET("/live", func(c *gin.Context) {
		if pgDB == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
			return
		}
		info, err := chain.GetChainInfo(pgDB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"transactions":  info.TxTotal,
			"chain_id":      info.ChainID,
			"height":        info.Height,
			"exchange_rate": HorizonTomanPerUSD,
			"results":       calculateTiers(info.TxTotal),
		})
	})
}

func calculateTiers(n int64) []gin.H {
	out := make([]gin.H, 0, len(horizonPricingTiers))
	for _, t := range horizonPricingTiers {
		total := n * t.TomanPerTx
		out = append(out, gin.H{
			"tier":         t.Name,
			"toman_per_tx": t.TomanPerTx,
			"total_toman":  total,
			"total_usd":    float64(total) / float64(HorizonTomanPerUSD),
		})
	}
	return out
}
