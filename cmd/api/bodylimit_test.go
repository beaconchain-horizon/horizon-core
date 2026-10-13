package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupBodyLimitRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(bodyLimitMiddleware())
	// /echo reads the body like a real POST handler would,
	// so the middleware's MaxBytesReader path is exercised for
	// chunked requests without a Content-Length.
	r.POST("/echo", func(c *gin.Context) {
		var payload map[string]any
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestBodyLimit_AcceptsSmallBody(t *testing.T) {
	r := setupBodyLimitRouter(t)
	body := strings.NewReader("{\"msg\":\"hello\"}")
	req := httptest.NewRequest(http.MethodPost, "/echo", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("small body should be accepted, got %d", w.Code)
	}
}

func TestBodyLimit_RejectsOversizedBody(t *testing.T) {
	r := setupBodyLimitRouter(t)
	huge := strings.Repeat("A", 9<<20)
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(huge))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized body should return 413, got %d", w.Code)
	}
}

func TestBodyLimit_RejectsChunkedOversized(t *testing.T) {
	r := setupBodyLimitRouter(t)
	huge := strings.Repeat("A", 9<<20)
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(huge))
	req.ContentLength = -1
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusRequestEntityTooLarge && w.Code != http.StatusBadRequest {
		t.Fatalf("chunked oversized should be rejected, got %d", w.Code)
	}
}

func TestBodyLimit_AllowsAtLimit(t *testing.T) {
	r := setupBodyLimitRouter(t)
	// Valid JSON, exactly at the 8 MiB limit.
	// Envelope {"x":""} is 8 bytes, so the fill is 8 MiB - 8.
	payload := `{"x":"` + strings.Repeat("A", (8<<20)-8) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("body at limit should be accepted, got %d", w.Code)
	}
}
