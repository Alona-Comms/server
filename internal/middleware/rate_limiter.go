package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// IPRateLimiter limits requests by IP address
type IPRateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	r        rate.Limit // requests per second
	b        int        // burst size
}

// r - requests per second
// b - burst size (maximum number of simultaneous requests)
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}
}

func (rl *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.r, rl.b)
		rl.limiters[ip] = limiter

		// Очистка старых лимитеров (опционально, для экономии памяти)
		go rl.cleanupLimiter(ip, 10*time.Minute)
	}

	return limiter
}

func (rl *IPRateLimiter) cleanupLimiter(ip string, after time.Duration) {
	<-time.After(after)
	rl.mu.Lock()
	delete(rl.limiters, ip)
	rl.mu.Unlock()
}

func (rl *IPRateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			limiter := rl.getLimiter(ip)

			if !limiter.Allow() {
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error":       "rate limit exceeded",
					"message":     "too many requests from this IP address",
					"retry_after": "60s",
				})
			}

			return next(c)
		}
	}
}
