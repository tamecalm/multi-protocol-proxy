package auth

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter per user
type RateLimiter struct {
	mu      sync.RWMutex
	buckets map[string]*tokenBucket
	limits  map[string]int 
}

type tokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 
	lastRefill time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*tokenBucket),
		limits:  make(map[string]int),
	}
}

func (r *RateLimiter) SetLimit(username string, rpm int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.limits[username] = rpm

	maxTokens := float64(rpm) / 6 
	if maxTokens < 10 {
		maxTokens = 10
	}

	r.buckets[username] = &tokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: float64(rpm) / 60.0, 
		lastRefill: time.Now(),
	}
}

func (r *RateLimiter) Allow(username string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	bucket, exists := r.buckets[username]
	if !exists {
		return true
	}

	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	bucket.tokens += elapsed * bucket.refillRate
	if bucket.tokens > bucket.maxTokens {
		bucket.tokens = bucket.maxTokens
	}
	bucket.lastRefill = now

	if bucket.tokens >= 1 {
		bucket.tokens--
		return true
	}

	return false
}

func (r *RateLimiter) GetRemainingTokens(username string) float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bucket, exists := r.buckets[username]
	if !exists {
		return -1 
	}

	return bucket.tokens
}
