package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Counter struct {
	Count     int
	StartTime time.Time
}

type RateLimiter struct {
	MaxRequestPerWindow int
	RateLimitWindow     time.Duration

	mu        sync.Mutex
	ipRequest map[string]*Counter
}

func NewRateLimiter(requests int, window time.Duration) *RateLimiter {
	r := &RateLimiter{
		MaxRequestPerWindow: requests,
		RateLimitWindow:     window,
		ipRequest:           make(map[string]*Counter),
	}
	go r.cleanup()
	return r
}

// cleanup periodically removes IP entries whose window has fully elapsed,
// so the map doesn't grow without bound.
func (r *RateLimiter) cleanup() {
	interval := r.RateLimitWindow
	if interval < time.Second {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		r.mu.Lock()
		for ip, c := range r.ipRequest {
			if now.Sub(c.StartTime) >= r.RateLimitWindow {
				delete(r.ipRequest, ip)
			}
		}
		r.mu.Unlock()
	}
}

func (r *RateLimiter) LimitMiddleWare() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()
		now := time.Now()

		r.mu.Lock()
		info, exists := r.ipRequest[ip]
		if !exists || now.Sub(info.StartTime) >= r.RateLimitWindow {
			// First request, or the previous window has expired—reset.
			info = &Counter{Count: 1, StartTime: now}
			r.ipRequest[ip] = info
		} else {
			info.Count++
		}
		// Read the count while we still hold the lock.
		count := info.Count
		r.mu.Unlock()

		if count > r.MaxRequestPerWindow {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"msg": "Too many requests. Try again later."})
			return
		}

		ctx.Next()
	}
}
