package smtp

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	defaultMaxEntries     = 10000
	defaultCleanupEvery   = 5 * time.Minute
	defaultMaxAge         = 30 * time.Minute
)

type limiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// RateLimiter provides per-IP rate limiting using the token bucket algorithm.
// Stale entries (unused for maxAge) are evicted periodically to bound memory.
type RateLimiter struct {
	limiters        map[string]*limiterEntry
	mu              sync.Mutex
	rate            rate.Limit
	burst           int
	maxEntries      int
	lastCleanup     time.Time
	cleanupInterval time.Duration
	maxAge          time.Duration
}

// NewRateLimiter creates a RateLimiter that allows rps requests per second
// with a burst capacity of burst per IP address.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters:        make(map[string]*limiterEntry),
		rate:            rate.Limit(rps),
		burst:           burst,
		maxEntries:      defaultMaxEntries,
		lastCleanup:     time.Now(),
		cleanupInterval: defaultCleanupEvery,
		maxAge:          defaultMaxAge,
	}
}

// Allow reports whether a request from the given IP is permitted.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()

	now := time.Now()

	// Periodic cleanup of stale entries.
	if now.Sub(rl.lastCleanup) >= rl.cleanupInterval {
		rl.cleanup(now)
	}

	entry, ok := rl.limiters[ip]
	if !ok {
		// If at capacity, evict the oldest entry before inserting.
		if len(rl.limiters) >= rl.maxEntries {
			rl.evictOldest()
		}
		entry = &limiterEntry{
			limiter:    rate.NewLimiter(rl.rate, rl.burst),
			lastAccess: now,
		}
		rl.limiters[ip] = entry
	} else {
		entry.lastAccess = now
	}

	rl.mu.Unlock()
	return entry.limiter.Allow()
}

// cleanup removes entries that haven't been accessed for longer than maxAge.
// Caller must hold rl.mu.
func (rl *RateLimiter) cleanup(now time.Time) {
	rl.lastCleanup = now
	for ip, entry := range rl.limiters {
		if now.Sub(entry.lastAccess) >= rl.maxAge {
			delete(rl.limiters, ip)
		}
	}
}

// evictOldest removes the entry with the earliest lastAccess time.
// Caller must hold rl.mu.
func (rl *RateLimiter) evictOldest() {
	var oldestIP string
	var oldestTime time.Time
	first := true
	for ip, entry := range rl.limiters {
		if first || entry.lastAccess.Before(oldestTime) {
			oldestIP = ip
			oldestTime = entry.lastAccess
			first = false
		}
	}
	if oldestIP != "" {
		delete(rl.limiters, oldestIP)
	}
}
