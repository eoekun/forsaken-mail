package smtp

import (
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiter provides per-IP rate limiting using the token bucket algorithm.
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a RateLimiter that allows rps requests per second
// with a burst capacity of burst per IP address.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(rps),
		burst:    burst,
	}
}

// Allow reports whether a request from the given IP is permitted.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	l, ok := rl.limiters[ip]
	if !ok {
		l = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = l
	}
	rl.mu.Unlock()
	return l.Allow()
}
