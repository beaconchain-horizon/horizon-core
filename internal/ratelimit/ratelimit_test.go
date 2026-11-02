package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLimiter_AllowsBurstThenBlocks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := New(1, 3)
	r := gin.New()
	r.Use(l.Middleware())
	r.GET("/x", func(c *gin.Context) { c.Status(200) })
	allowed, blocked := 0, 0
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		switch w.Code {
		case 200:
			allowed++
		case http.StatusTooManyRequests:
			blocked++
		}
	}
	if allowed != 3 || blocked != 7 {
		t.Fatalf("allowed=%d blocked=%d", allowed, blocked)
	}
}
