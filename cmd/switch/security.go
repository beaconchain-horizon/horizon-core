package main

import (
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	tenantIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,64}$`)
	typeRegex     = regexp.MustCompile(`^[a-z]{2,32}$`)
)

func ValidateTenantID(id string) bool {
	if id == "" || len(id) > 64 || len(id) < 3 {
		return false
	}
	if strings.ContainsAny(id, "\x00\r\n\t ") {
		return false
	}
	if strings.Contains(id, "..") || strings.ContainsAny(id, "/\\") {
		return false
	}
	if strings.ContainsAny(id, "'\";<>()&|$`") {
		return false
	}
	return tenantIDRegex.MatchString(id)
}

func ValidateType(t string) bool {
	return typeRegex.MatchString(t)
}

// ValidateAgentURL — نسخه کامل امن
func ValidateAgentURL(u string) bool {
	if u == "" {
		return true
	}
	if len(u) > 512 {
		return false
	}
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		return false
	}

	parsed, err := url.Parse(u)
	if err != nil || parsed.Host == "" {
		return false
	}

	host := parsed.Hostname()
	if host == "" {
		return false
	}

	// چک ۱: suffix های داخلی
	lower := strings.ToLower(host)
	blockedSuffixes := []string{
		"localhost", ".localhost", ".localdomain", ".internal", ".local", ".home.arpa",
		".nip.io", ".xip.io", ".sslip.io",
	}
	for _, b := range blockedSuffixes {
		if lower == b || strings.HasSuffix(lower, b) {
			return false
		}
	}

	// چک ۲: hosts صریح
	blockedHosts := []string{
		"localhost.localdomain",
		"ip6-localhost",
		"ip6-loopback",
		"broadcasthost",
		"metadata.google.internal",
	}
	for _, b := range blockedHosts {
		if lower == b {
			return false
		}
	}

	// چک ۳: IP استاندارد
	if ip := net.ParseIP(host); ip != nil {
		return !isPrivateIP(ip)
	}

	// چک ۴: IP با فرمت عجیب
	if ip := parseWeirdIP(host); ip != nil {
		return !isPrivateIP(ip)
	}

	return true
}

// parseWeirdIP — پارس octal/hex/decimal/short IP
func parseWeirdIP(host string) net.IP {
	parts := strings.Split(host, ".")

	// تک‌عددی
	if len(parts) == 1 {
		p := parts[0]
		base := 10
		if strings.HasPrefix(p, "0x") || strings.HasPrefix(p, "0X") {
			base = 16
			p = p[2:]
		} else if len(p) > 1 && p[0] == '0' {
			base = 8
			p = p[1:]
		}

		val, err := strconv.ParseInt(p, base, 64)
		if err != nil || val < 0 || val > 0xFFFFFFFF {
			return nil
		}
		return net.IPv4(
			byte(val>>24),
			byte(val>>16),
			byte(val>>8),
			byte(val),
		)
	}

	// 2، 3، 4 قسمتی
	if len(parts) > 4 {
		return nil
	}

	nums := make([]int64, len(parts))
	for i, p := range parts {
		base := 10
		if strings.HasPrefix(p, "0x") || strings.HasPrefix(p, "0X") {
			base = 16
			p = p[2:]
		} else if len(p) > 1 && p[0] == '0' {
			base = 8
			p = p[1:]
		}

		val, err := strconv.ParseInt(p, base, 32)
		if err != nil || val < 0 {
			return nil
		}
		nums[i] = val
	}

	switch len(nums) {
	case 4:
		if nums[0] > 255 || nums[1] > 255 || nums[2] > 255 || nums[3] > 255 {
			return nil
		}
		return net.IPv4(byte(nums[0]), byte(nums[1]), byte(nums[2]), byte(nums[3]))
	case 3:
		if nums[0] > 255 || nums[1] > 255 || nums[2] > 0xFFFF {
			return nil
		}
		return net.IPv4(byte(nums[0]), byte(nums[1]), byte(nums[2]>>8), byte(nums[2]))
	case 2:
		if nums[0] > 255 || nums[1] > 0xFFFFFF {
			return nil
		}
		return net.IPv4(byte(nums[0]), byte(nums[1]>>16), byte(nums[1]>>8), byte(nums[1]))
	}

	return nil
}

func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return true
	}

	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 10 {
			return true
		}
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return true
		}
		if ip4[0] == 192 && ip4[1] == 168 {
			return true
		}
		if ip4[0] == 127 {
			return true
		}
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}
		if ip4[0] == 0 {
			return true
		}
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return true
		}
		if ip4[0] >= 224 && ip4[0] <= 239 {
			return true
		}
		if ip4[0] >= 240 {
			return true
		}
	}

	if ip.To4() == nil {
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return true
		}
		if ip.Equal(net.IPv6loopback) {
			return true
		}
		if len(ip) == 16 && (ip[0]&0xfe) == 0xfc {
			return true
		}
		if len(ip) == 16 && ip[0] == 0xfe && (ip[1]&0xc0) == 0x80 {
			return true
		}
	}

	return false
}

func SanitizeString(s string, maxLen int) string {
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	s = strings.Map(func(r rune) rune {
		if r == 0 || (r < 32 && r != '\n' && r != '\t') {
			return -1
		}
		return r
	}, s)
	return strings.TrimSpace(s)
}

type rateBucket struct {
	tokens     float64
	lastRefill time.Time
}

type rateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*rateBucket
	rate     float64
	capacity float64
}

func NewRateLimiter(rate, capacity float64) *rateLimiter {
	rl := &rateLimiter{
		buckets:  make(map[string]*rateBucket),
		rate:     rate,
		capacity: capacity,
	}
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()
	return rl
}

func (rl *rateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok {
		b = &rateBucket{tokens: rl.capacity, lastRefill: now}
		rl.buckets[key] = b
	}
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * rl.rate
	if b.tokens > rl.capacity {
		b.tokens = rl.capacity
	}
	b.lastRefill = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	cutoff := time.Now().Add(-10 * time.Minute)
	for k, b := range rl.buckets {
		if b.lastRefill.Before(cutoff) {
			delete(rl.buckets, k)
		}
	}
}

func RateLimitMiddleware(rl *rateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.Allow(c.ClientIP()) {
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Del("X-Powered-By")
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Referrer-Policy", "no-referrer")
		c.Next()
	}
}
