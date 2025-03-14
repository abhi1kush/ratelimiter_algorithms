package ratelimiter

import (
	"sync"
	"time"
)

// TokenBucketLimiter implements the Token Bucket algorithm
type TokenBucketLimiter struct {
	capacity   int
	refillRate int
	users      map[string]*TokenBucket
	mu         sync.Mutex
}

// TokenBucketRateLimiter implements RateLimiter using a token bucket algorithm
type TokenBucket struct {
	tokens       int
	lastRefilled time.Time
	mu           sync.Mutex
}

// NewTokenBucketLimiter initializes a Token Bucket limiter
func NewTokenBucketLimiter(capacity, refillRate int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		capacity:   capacity,
		refillRate: refillRate,
		users:      make(map[string]*TokenBucket),
	}
}

// GetUserBucket retrieves or creates a user's token bucket
func (tb *TokenBucketLimiter) GetUserBucket(userID string) *TokenBucket {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if _, exists := tb.users[userID]; !exists {
		tb.users[userID] = &TokenBucket{
			tokens:       tb.capacity,
			lastRefilled: time.Now(),
		}
	}
	return tb.users[userID]
}

// AllowRequest checks if a request is allowed using the Token Bucket algorithm
func (tb *TokenBucketLimiter) AllowRequest(userID string) bool {
	bucket := tb.GetUserBucket(userID)

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefilled).Seconds()
	newTokens := int(elapsed) * tb.refillRate

	if newTokens > 0 {
		bucket.tokens = min(tb.capacity, bucket.tokens+newTokens)
		bucket.lastRefilled = now
	}

	// Allow request if tokens are available
	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}
	return false
}
