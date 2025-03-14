package ratelimiter

import (
	"sync"
	"testing"
	"time"
)

// Test that requests within the token limit are allowed
func TestTokenBucket_Allow_WithinLimit(t *testing.T) {
	maxRequest := 7
	limiter := NewTokenBucketLimiter(maxRequest, 3) // 3 tokens per second
	userID := "user1"

	// Allow 7 requests
	for i := 0; i < maxRequest; i++ {
		if !limiter.AllowRequest(userID) {
			t.Errorf("Expected request %d to be allowed", i+1)
		}
	}

	// 8th request should be blocked
	if limiter.AllowRequest(userID) {
		t.Errorf("Expected 4th request to be blocked but was allowed")
	}
}

// Test that tokens refill over time
func TestTokenBucket_Allow_Refill(t *testing.T) {
	limiter := NewTokenBucketLimiter(2, 1) // 2 tokens per 2s
	userID := "user2"

	// Use all tokens
	limiter.AllowRequest(userID)
	limiter.AllowRequest(userID)

	// 3rd request should be blocked
	if limiter.AllowRequest(userID) {
		t.Errorf("Expected request to be blocked")
	}

	// Wait for refill
	time.Sleep(2 * time.Second)

	// Should allow request again
	if !limiter.AllowRequest(userID) {
		t.Errorf("Expected request to be allowed after refill")
	}
}

// Test concurrent requests handling
func TestTokenBucket_Allow_Concurrency(t *testing.T) {
	limiter := NewTokenBucketLimiter(5, 5) // 5 tokens per second
	userID := "user3"
	var wg sync.WaitGroup
	var successCount int
	var mu sync.Mutex

	// Simulate 10 concurrent requests
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.AllowRequest(userID) {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// At most 5 requests should be allowed
	if successCount > 5 {
		t.Errorf("Expected at most 5 successful requests, but got %d", successCount)
	}
}
