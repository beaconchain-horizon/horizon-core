package ratelimit

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

type Limiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     float64
	capacity float64
	ttl      time.Duration
}

func New(ratePerSec, burst int) *Limiter {
	if ratePerSec <= 0 {
		ratePerSec = 1
	}
	if burst <= 0 {
		burst = 1
	}
	l := &Limiter{
		buckets:  make(map[string]*bucket),
		rate:     float64(ratePerSec),
		capacity: float64(burst),
		ttl:      10 * time.Minute,
	}
	go l.gcLoop()
	return l
}

func (l *Limiter) allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.buckets[key]
	if !ok {
		l.buckets[key] = &bucket{tokens: l.capacity - 1, lastRefill: now}
		return true
	}
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * l.rate
	if b.tokens > l.capacity {
		b.tokens = l.capacity
	}
	b.lastRefill = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (l *Limiter) gcLoop() {
	t := time.NewTicker(5 * time.Minute)
	for range t.C {
		cutoff := time.Now().Add(-l.ttl)
		l.mu.Lock()
		for k, b := range l.buckets {
			if b.lastRefill.Before(cutoff) {
				delete(l.buckets, k)
			}
		}
		l.mu.Unlock()
	}
}

func (l *Limiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if ip == "" {
			ip = "unknown"
		}
		if !l.allow(ip) {
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
