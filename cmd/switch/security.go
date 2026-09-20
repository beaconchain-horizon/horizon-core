package main

import (
	"os"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ═══════════════════════════════════════════════════════════════
//  SECURITY HARDENING — Middleware جامع
// ═══════════════════════════════════════════════════════════════

// ─── Regexهای اعتبارسنجی ───
var (
	// tenant_id: فقط a-z, A-Z, 0-9, -, _
	tenantIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,64}$`)

	// sensor_id: فقط a-z, A-Z, 0-9, -, _
	sensorIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,128}$`)

	// site_id
	siteIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,128}$`)

	// type: bank|refinery|powerplant|gas|other|...
	typeRegex = regexp.MustCompile(`^[a-z]{2,32}$`)
)

// ═══════════════════════════════════════════════════════════════
//  ValidateTenantID — چک امنیتی
// ═══════════════════════════════════════════════════════════════
func ValidateTenantID(id string) bool {
	if id == "" || len(id) > 64 || len(id) < 3 {
		return false
	}
	// Null byte check
	if strings.ContainsAny(id, "\x00\r\n\t ") {
		return false
	}
	// Path traversal check
	if strings.Contains(id, "..") || strings.ContainsAny(id, "/\\") {
		return false
	}
	// SQL/HTML special chars
	if strings.ContainsAny(id, "'\";<>()&|$`") {
		return false
	}
	return tenantIDRegex.MatchString(id)
}

func ValidateSensorID(id string) bool {
	if id == "" || len(id) > 128 {
		return false
	}
	if strings.ContainsAny(id, "\x00\r\n\t /\\'\";<>()&|$`") {
		return false
	}
	return sensorIDRegex.MatchString(id)
}

func ValidateSiteID(id string) bool {
	if id == "" || len(id) > 128 {
		return false
	}
	if strings.ContainsAny(id, "\x00\r\n\t /\\'\";<>()&|$`") {
		return false
	}
	return siteIDRegex.MatchString(id)
}

func ValidateType(t string) bool {
	return typeRegex.MatchString(t)
}

// ValidateAgentURL — فقط http/https
func ValidateAgentURL(u string) bool {
	if u == "" {
		return true // خالی مجازه
	}
	if len(u) > 512 {
		return false
	}
	// فقط http:// یا https://
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		return false
	}
	// جلوگیری از SSRF به IPهای داخلی
	blocked := []string{
		"127.", "localhost", "0.0.0.0", "::1",
		"169.254.", "10.", "192.168.", "172.16.", "172.17.", "172.18.", "172.19.",
		"172.20.", "172.21.", "172.22.", "172.23.", "172.24.", "172.25.",
		"172.26.", "172.27.", "172.28.", "172.29.", "172.30.", "172.31.",
		"file:", "gopher:", "dict:", "ftp:",
	}
	lower := strings.ToLower(u)
	for _, b := range blocked {
		if strings.Contains(lower, b) {
			return false
		}
	}
	return true
}

// ═══════════════════════════════════════════════════════════════
//  Rate Limiter — توکن bucket
// ═══════════════════════════════════════════════════════════════
type rateBucket struct {
	tokens     float64
	lastRefill time.Time
}

type rateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*rateBucket
	rate     float64 // tokens per second
	capacity float64
}

func NewRateLimiter(rate, capacity float64) *rateLimiter {
	rl := &rateLimiter{
		buckets:  make(map[string]*rateBucket),
		rate:     rate,
		capacity: capacity,
	}
	// Cleanup قدیمی‌ها هر ۵ دقیقه
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

	// Refill
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

// ═══════════════════════════════════════════════════════════════
//  RateLimitMiddleware — محدودیت نرخ درخواست
// ═══════════════════════════════════════════════════════════════
func RateLimitMiddleware(rl *rateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// کلید: IP + مسیر
		key := c.ClientIP()
		if !rl.Allow(key) {
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}

// ═══════════════════════════════════════════════════════════════
//  SecurityHeadersMiddleware — هدرهای امنیتی
// ═══════════════════════════════════════════════════════════════
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// حذف هدرهای افشاگر
		c.Writer.Header().Del("X-Powered-By")
		c.Writer.Header().Del("Server")
		c.Writer.Header().Del("X-AspNet-Version")

		// هدرهای امنیتی
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Referrer-Policy", "no-referrer")
		c.Writer.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Content-Security-Policy (برای HTML)
		c.Writer.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com; img-src 'self' data:; font-src 'self' https://cdnjs.cloudflare.com")

		c.Next()
	}
}

// ═══════════════════════════════════════════════════════════════
//  CORSConfig — CORS محدود
// ═══════════════════════════════════════════════════════════════
func GetCORSOrigins() []string {
	// از env می‌خونیم، اگه نبود پیش‌فرض
	origins := getEnv("CORS_ORIGINS", "")
	if origins == "" {
		return []string{
			"https://beaconchain-horizon.github.io",
			"https://horizon-backend.liara.run",
			"https://horizon-switch.liara.run",
			"http://localhost:8080",
			"http://localhost:3000",
			"http://127.0.0.1:8080",
		}
	}
	parts := strings.Split(origins, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// ═══════════════════════════════════════════════════════════════
//  SanitizeString — پاکسازی ورودی
// ═══════════════════════════════════════════════════════════════
func SanitizeString(s string, maxLen int) string {
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	// حذف null byte و control chars
	s = strings.Map(func(r rune) rune {
		if r == 0 || (r < 32 && r != '\n' && r != '\t') {
			return -1
		}
		return r
	}, s)
	return strings.TrimSpace(s)
}

// getEnv — خواندن env با پیش‌فرض
func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
