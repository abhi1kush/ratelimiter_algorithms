package ratelimiter

import (
	"sync"
	"testing"
	"time"
)

// Test that requests within the rate limit are allowed
func TestSlidingWindowLog_Allow_WithinLimit(t *testing.T) {
	limiter := NewSlidingWindowLog(3, 10*time.Second) // 3 requests per 10s
	userID := "user1"

	// Allow 3 requests
	for i := 0; i < 3; i++ {
		if !limiter.AllowRequest(userID) {
			t.Errorf("Expected request %d to be allowed", i+1)
		}
	}

	// 4th request should be blocked
	if limiter.AllowRequest(userID) {
		t.Errorf("Expected 4th request to be blocked but was allowed")
	}
}

// Test that old requests are removed after the window expires
func TestSlidingWindowLog_Allow_WindowExpiration(t *testing.T) {
	limiter := NewSlidingWindowLog(2, 2*time.Second) // 2 requests per 2s
	userID := "user2"

	// Allow 2 requests
	if !limiter.AllowRequest(userID) || !limiter.AllowRequest(userID) {
		t.Errorf("Expected first two requests to be allowed")
	}

	// 3rd request should be blocked
	if limiter.AllowRequest(userID) {
		t.Errorf("Expected request to be blocked but it was allowed")
	}

	// Wait for window to expire
	time.Sleep(2 * time.Second)

	// New request should be allowed
	if !limiter.AllowRequest(userID) {
		t.Errorf("Expected request to be allowed after window reset")
	}
}

// Test concurrent requests handling
func TestSlidingWindowLog_Allow_Concurrency(t *testing.T) {
	limiter := NewSlidingWindowLog(10, 5*time.Second) // 10 requests per 5s
	userID := "user3"
	var wg sync.WaitGroup
	var successCount int
	var mu sync.Mutex

	// Simulate 15 concurrent requests
	for i := 0; i < 15; i++ {
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

	// At most 10 requests should be allowed
	if successCount > 10 {
		t.Errorf("Expected at most 10 successful requests, but got %d", successCount)
	}
}

// Test cleanup function removes old requests
func TestSlidingWindowLog_Cleanup(t *testing.T) {
	limiter := NewSlidingWindowLog(3, 2*time.Second) // 3 requests per 2s
	userID := "user4"

	// Allow 3 requests
	for i := 0; i < 3; i++ {
		limiter.AllowRequest(userID)
	}

	// Wait for window to expire
	time.Sleep(3 * time.Second)

	// Cleanup should remove old requests
	limiter.cleanupOldRequests()

	// New requests should be allowed
	if !limiter.AllowRequest(userID) {
		t.Errorf("Expected request to be allowed after cleanup")
	}
}
