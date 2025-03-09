package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pack-calculator/config"
	"github.com/pack-calculator/pkg/errors"
)

type RateLimiter struct {
	enabled      bool
	maxRequests  int
	mutex        sync.Mutex
	requests     map[string][]time.Time
	cleanupTimer *time.Timer
}

func NewRateLimiter(cfg config.RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		enabled:     cfg.Enabled,
		maxRequests: cfg.MaxRequests,
		requests:    make(map[string][]time.Time),
	}

	// Start cleanup routine
	rl.cleanupTimer = time.NewTimer(time.Minute)
	go rl.cleanup()

	return rl
}

func (r *RateLimiter) cleanup() {
	for {
		<-r.cleanupTimer.C
		r.mutex.Lock()
		now := time.Now()
		for ip, times := range r.requests {
			var valid []time.Time
			for _, t := range times {
				if now.Sub(t) < time.Second {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(r.requests, ip)
			} else {
				r.requests[ip] = valid
			}
		}
		r.mutex.Unlock()
		r.cleanupTimer.Reset(time.Minute)
	}
}

func (r *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !r.enabled {
			c.Next()
			return
		}

		ip := c.ClientIP()

		r.mutex.Lock()
		now := time.Now()
		times := r.requests[ip]

		// Remove requests older than 1 second
		var valid []time.Time
		for _, t := range times {
			if now.Sub(t) < time.Second {
				valid = append(valid, t)
			}
		}

		if len(valid) >= r.maxRequests {
			r.mutex.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, errors.NewValidationError("rate limit exceeded"))
			return
		}

		// Add current request
		valid = append(valid, now)
		r.requests[ip] = valid
		r.mutex.Unlock()

		c.Next()
	}
}
