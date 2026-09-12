package main

import (
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

var (
	airGapMode    atomic.Bool
	airGapBlocks  atomic.Int64
	errAirGapMode = errors.New("air-gap mode: outbound connection blocked")
)

func initAirGap() {
	if os.Getenv("AIR_GAP_MODE") == "true" {
		airGapMode.Store(true)
		log.Println("🔒 AIR-GAP MODE ENABLED — only local network allowed")
	} else {
		log.Println("🌐 Air-gap mode: OFF (normal operation)")
	}
}

func isLocalIP(ipStr string) bool {
	host, _, err := net.SplitHostPort(ipStr)
	if err != nil {
		host = ipStr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() {
		return true
	}
	if ip.IsPrivate() {
		return true
	}
	if ip.IsLinkLocalUnicast() {
		return true
	}
	return false
}

func airgapMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if airGapMode.Load() {
			ip := c.ClientIP()
			if !isLocalIP(ip) {
				airGapBlocks.Add(1)
				log.Printf("🚫 AIR-GAP BLOCK: external %s → %s", ip, c.Request.URL.Path)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error":   "air-gap mode: external access denied",
					"your_ip": ip,
					"mode":    "air-gap",
				})
				return
			}
			c.Header("X-Air-Gap-Mode", "enabled")
		}
		c.Next()
	}
}

func outboundGuard(target string) error {
	if airGapMode.Load() {
		airGapBlocks.Add(1)
		log.Printf("🚫 OUTBOUND BLOCKED (air-gap): %s", target)
		return errAirGapMode
	}
	return nil
}
