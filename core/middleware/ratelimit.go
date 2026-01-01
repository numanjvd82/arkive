package middleware

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	appcontext "arkive/pkg/context"
)

type RateLimitConfig struct {
	RequestsPerMinute int
	Burst             int
	KeyPrefix         string
	SkipPremium       bool
}

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type limiterStore struct {
	mu      sync.Mutex
	entries map[string]*limiterEntry
	ttl     time.Duration
	max     int
}

var (
	userRateLimiter = newLimiterStore(30*time.Minute, 250000)
	ipRateLimiter   = newLimiterStore(10*time.Minute, 250000)
)

func newLimiterStore(ttl time.Duration, maxEntries int) *limiterStore {
	store := &limiterStore{
		entries: make(map[string]*limiterEntry),
		ttl:     ttl,
		max:     maxEntries,
	}
	go store.cleanupLoop()
	return store
}

func (s *limiterStore) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for key, entry := range s.entries {
			if now.Sub(entry.lastSeen) > s.ttl {
				delete(s.entries, key)
			}
		}
		s.mu.Unlock()
	}
}

func (s *limiterStore) allow(key string, ratePerSec, burst float64) bool {
	now := time.Now()
	s.mu.Lock()
	entry, ok := s.entries[key]
	if !ok {
		s.enforceMax(now)
		if s.max > 0 && len(s.entries) >= s.max {
			s.mu.Unlock()
			return false
		}
	}
	if !ok || int(entry.limiter.Burst()) != int(burst) || entry.limiter.Limit() != rate.Limit(ratePerSec) {
		entry = &limiterEntry{
			limiter:  rate.NewLimiter(rate.Limit(ratePerSec), int(burst)),
			lastSeen: now,
		}
		s.entries[key] = entry
	}

	entry.lastSeen = now

	allowed := entry.limiter.Allow()
	s.mu.Unlock()
	return allowed
}

func (s *limiterStore) enforceMax(now time.Time) {
	if s.max <= 0 || len(s.entries) < s.max {
		return
	}
	for key, entry := range s.entries {
		if now.Sub(entry.lastSeen) > s.ttl {
			delete(s.entries, key)
		}
	}
	if len(s.entries) < s.max {
		return
	}

	var oldestKey string
	var oldestSeen time.Time
	for key, entry := range s.entries {
		if oldestKey == "" || entry.lastSeen.Before(oldestSeen) {
			oldestKey = key
			oldestSeen = entry.lastSeen
		}
	}
	if oldestKey != "" {
		delete(s.entries, oldestKey)
	}
}

func RateLimit(cfg RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.RequestsPerMinute <= 0 {
			c.Next()
			return
		}

		user, ok := appcontext.UserFromContext(c)
		if ok && cfg.SkipPremium && user.IsPremium {
			c.Next()
			return
		}

		key := buildRateLimitKey(c, cfg.KeyPrefix, user.ID, ok)
		ratePerSec := float64(cfg.RequestsPerMinute) / 60
		burst := float64(cfg.Burst)
		if burst <= 0 {
			burst = float64(cfg.RequestsPerMinute)
		}
		store := ipRateLimiter
		if ok && strings.TrimSpace(user.ID) != "" {
			store = userRateLimiter
		}
		if !store.allow(key, ratePerSec, burst) {
			retryAfter := 60
			if cfg.RequestsPerMinute > 0 {
				retryAfter = int(math.Ceil(60.0 / float64(cfg.RequestsPerMinute)))
				if retryAfter < 1 {
					retryAfter = 1
				}
			}
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.Header("X-Rate-Limited", "true")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "try again shortly"})
			return
		}
		c.Next()
	}
}

func buildRateLimitKey(c *gin.Context, prefix, userID string, hasUser bool) string {
	parts := []string{}
	if strings.TrimSpace(prefix) != "" {
		parts = append(parts, strings.TrimSpace(prefix))
	}
	if hasUser && strings.TrimSpace(userID) != "" {
		parts = append(parts, "user:"+userID)
	} else {
		parts = append(parts, "ip:"+c.ClientIP())
	}
	return strings.Join(parts, "|")
}
