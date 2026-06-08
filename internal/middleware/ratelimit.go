package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type bucket struct {
	tokens   int
	lastFill time.Time
}

// RateLimiter implements a per-user token bucket rate limiter.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     int           // tokens per period
	period   time.Duration // refill period
	burst    int           // max tokens
	cleanupT *time.Ticker
}

// NewRateLimiter creates a rate limiter with the given rate (tokens per period),
// period (refill interval), and burst (max tokens to accumulate).
// Example: NewRateLimiter(60, time.Minute, 10) allows 60 req/min with burst of 10.
func NewRateLimiter(rate int, period time.Duration, burst int) *RateLimiter {
	if rate <= 0 {
		rate = 60
	}
	if period <= 0 {
		period = time.Minute
	}
	if burst <= 0 {
		burst = rate
	}
	rl := &RateLimiter{
		buckets:  make(map[string]*bucket),
		rate:     rate,
		period:   period,
		burst:    burst,
		cleanupT: time.NewTicker(5 * time.Minute),
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[key]
	now := time.Now()

	if !ok {
		rl.buckets[key] = &bucket{tokens: rl.burst - 1, lastFill: now}
		return true
	}

	elapsed := now.Sub(b.lastFill)
	tokensToAdd := int(elapsed.Seconds() * (float64(rl.rate) / rl.period.Seconds()))
	if tokensToAdd > 0 {
		b.tokens += tokensToAdd
		if b.tokens > rl.burst {
			b.tokens = rl.burst
		}
		b.lastFill = now
	}

	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}

func (rl *RateLimiter) cleanup() {
	for range rl.cleanupT.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		for k, b := range rl.buckets {
			if b.lastFill.Before(cutoff) {
				delete(rl.buckets, k)
			}
		}
		rl.mu.Unlock()
	}
}

// Stop terminates the background cleanup goroutine.
func (rl *RateLimiter) Stop() {
	rl.cleanupT.Stop()
}

// RateLimit returns a Gin middleware that applies per-user rate limiting.
// User identity is derived from JWT user_id or falls back to client IP.
func RateLimit(rate int, period time.Duration, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, period, burst)
	return func(c *gin.Context) {
		key := c.ClientIP()
		if uid, exists := c.Get("user_id"); exists {
			switch v := uid.(type) {
			case string:
				key = v
			case float64:
				key = c.ClientIP() // fallback to IP for numeric type
			}
		}

		if !limiter.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    "common.rate_limit",
				"message": "Too many requests, please try again later",
			})
			return
		}
		c.Next()
	}
}
