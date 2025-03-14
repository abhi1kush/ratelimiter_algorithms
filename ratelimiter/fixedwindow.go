package ratelimiter

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// FixedWindowLimiter limits requests in fixed time intervals
type FixedWindowLimiter struct {
	limit          int
	window         time.Duration
	requests       map[string]int64 // Stores request count with timestamps
	mu             sync.Mutex
	cleanupRunning bool // Ensures only one cleanup goroutine runs at a time
}

// NewFixedWindowLimiter initializes a fixed window rate limiter
func NewFixedWindowLimiter(limit int, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		limit:    limit,
		window:   window,
		requests: make(map[string]int64),
	}
}

// AllowRequest checks if a request is allowed
func (fw *FixedWindowLimiter) AllowRequest(userID string) bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	now := time.Now().Unix()
	windowKey := fmt.Sprintf("%s-%d", userID, now/int64(fw.window.Seconds()))

	// ✅ Step 1: Start a cleanup goroutine if not already running
	if !fw.cleanupRunning {
		fw.cleanupRunning = true
		go fw.cleanupExpiredKeys()
	}

	// ✅ Step 2: Check request count for the current window
	if fw.requests[windowKey] >= int64(fw.limit) {
		return false
	}

	// ✅ Step 3: Increment request count
	fw.requests[windowKey]++
	return true
}

// cleanupExpiredKeys removes expired keys and exits
func (fw *FixedWindowLimiter) cleanupExpiredKeys() {
	defer func() { fw.cleanupRunning = false }() // Ensure flag resets when goroutine finishes

	time.Sleep(10 * time.Millisecond) // Small delay to batch cleanup requests

	fw.mu.Lock()
	defer fw.mu.Unlock()

	now := time.Now().Unix()
	for key := range fw.requests {
		parts := strings.Split(key, "-")
		if len(parts) < 2 {
			continue
		}

		var windowTime int64
		fmt.Sscanf(parts[1], "%d", &windowTime)

		// Delete entries older than the window duration
		if now-windowTime > int64(fw.window.Seconds()) {
			delete(fw.requests, key)
		}
	}
}
