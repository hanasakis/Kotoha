package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiterAllow(t *testing.T) {
	t.Run("allows_within_limit", func(t *testing.T) {
		rl := NewRateLimiter(100, time.Second, 10)
		for i := 0; i < 10; i++ {
			if !rl.allow("user-1") {
				t.Errorf("request %d should be allowed", i+1)
			}
		}
	})

	t.Run("blocks_over_burst", func(t *testing.T) {
		rl := NewRateLimiter(100, time.Second, 3)
		for i := 0; i < 3; i++ {
			if !rl.allow("user-2") {
				t.Errorf("request %d should be allowed", i+1)
			}
		}
		if rl.allow("user-2") {
			t.Error("4th request should be blocked (burst=3)")
		}
	})

	t.Run("separate_keys", func(t *testing.T) {
		rl := NewRateLimiter(100, time.Second, 1)
		if !rl.allow("user-a") {
			t.Error("user-a first request should be allowed")
		}
		if !rl.allow("user-b") {
			t.Error("user-b first request should be allowed")
		}
		if rl.allow("user-a") {
			t.Error("user-a second request should be blocked (burst=1)")
		}
	})

	t.Run("refill_after_time", func(t *testing.T) {
		rl := NewRateLimiter(10, time.Second, 1)
		rl.allow("user-3")        // consume the burst token
		if rl.allow("user-3") {   // should be blocked
			t.Fatal("second immediate request should be blocked")
		}
		// Manually refill by advancing lastFill
		b := rl.buckets["user-3"]
		b.lastFill = time.Now().Add(-200 * time.Millisecond) // ~2 tokens worth of time
		b.tokens = 0
		if !rl.allow("user-3") {
			t.Error("third request after refill should be allowed")
		}
	})
}

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("allows_request", func(t *testing.T) {
		r := gin.New()
		r.Use(RateLimit(100, time.Minute, 10))
		r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("got %d, want 200", w.Code)
		}
	})

	t.Run("blocks_after_burst", func(t *testing.T) {
		r := gin.New()
		r.Use(RateLimit(1000, time.Minute, 2))
		r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest("GET", "/test", nil))
			if w.Code != 200 {
				t.Fatalf("request %d: got %d, want 200", i+1, w.Code)
			}
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/test", nil))
		if w.Code != http.StatusTooManyRequests {
			t.Errorf("3rd request: got %d, want 429", w.Code)
		}
	})

	t.Run("uses_user_id_from_context", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("user_id", "test-user-42")
			c.Next()
		})
		r.Use(RateLimit(1000, time.Minute, 1))
		r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/test", nil))
		if w.Code != 200 {
			t.Fatalf("first request: got %d", w.Code)
		}
		w = httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/test", nil))
		if w.Code != http.StatusTooManyRequests {
			t.Errorf("second request for same user: got %d, want 429", w.Code)
		}
	})
}
