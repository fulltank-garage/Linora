package middleware

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimitBucket struct {
	count   int
	resetAt time.Time
}

// RateLimiter uses a fixed time window and is safe for concurrent requests.
// It intentionally keeps the key supplied by the caller so LINE webhook
// limits can be applied per LINE user rather than LINE's shared source IPs.
type RateLimiter struct {
	buckets map[string]rateLimitBucket
	limit   int
	mu      sync.Mutex
	window  time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit < 1 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimiter{buckets: make(map[string]rateLimitBucket), limit: limit, window: window}
}

func (l *RateLimiter) Allow(key string) (bool, time.Duration) {
	key = strings.TrimSpace(key)
	if key == "" {
		key = "unknown"
	}

	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	bucket, found := l.buckets[key]
	if !found || !now.Before(bucket.resetAt) {
		bucket = rateLimitBucket{resetAt: now.Add(l.window)}
	}
	if bucket.count >= l.limit {
		return false, time.Until(bucket.resetAt)
	}
	bucket.count++
	l.buckets[key] = bucket

	if len(l.buckets) > 1024 {
		for bucketKey, candidate := range l.buckets {
			if !now.Before(candidate.resetAt) {
				delete(l.buckets, bucketKey)
			}
		}
	}
	return true, 0
}

func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)
	return func(c *gin.Context) {
		allowed, retryAfter := limiter.Allow(c.ClientIP())
		if !allowed {
			c.Header("Retry-After", strconvItoaCeilSeconds(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Please try again shortly."})
			return
		}
		c.Next()
	}
}

func strconvItoaCeilSeconds(duration time.Duration) string {
	seconds := int(math.Ceil(duration.Seconds()))
	if seconds < 1 {
		seconds = 1
	}
	return strconv.Itoa(seconds)
}
