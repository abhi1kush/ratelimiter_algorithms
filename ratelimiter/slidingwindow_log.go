package ratelimiter

import (
	"sort"
	"sync"
	"time"
)

// SlidingWindowLog implements per-user rate limiting using a sorted log of request timestamps.
type SlidingWindowLog struct {
	mu       sync.Mutex
	requests map[string][]int64 // User -> list of request timestamps
	limit    int                // Max requests allowed
	window   time.Duration      // Time window
}

// NewSlidingWindowLog initializes the rate limiter.
func NewSlidingWindowLog(limit int, window time.Duration) *SlidingWindowLog {
	return &SlidingWindowLog{
		requests: make(map[string][]int64),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request is allowed based on the log.
func (sw *SlidingWindowLog) AllowRequest(userID string) bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now().UnixMilli()
	windowStart := now - sw.window.Milliseconds()

	// Remove expired timestamps using binary search (O(log N))
	timestamps := sw.requests[userID]
	idx := sort.Search(len(timestamps), func(i int) bool { return timestamps[i] >= windowStart })
	sw.requests[userID] = timestamps[idx:]

	// Check if the user is within the rate limit
	if len(sw.requests[userID]) >= sw.limit {
		return false
	}

	// Allow request and append timestamp
	sw.requests[userID] = append(sw.requests[userID], now)

	// Cleanup in a short-lived goroutine
	go sw.cleanupOldRequests()

	return true
}

// cleanupOldRequests removes expired entries asynchronously.
func (sw *SlidingWindowLog) cleanupOldRequests() {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now().UnixMilli()
	windowStart := now - sw.window.Milliseconds()

	for userID, timestamps := range sw.requests {
		// Remove old timestamps using binary search
		idx := sort.Search(len(timestamps), func(i int) bool { return timestamps[i] >= windowStart })
		sw.requests[userID] = timestamps[idx:]

		// If no requests left, delete the entry to free memory
		if len(sw.requests[userID]) == 0 {
			delete(sw.requests, userID)
		}
	}
}
