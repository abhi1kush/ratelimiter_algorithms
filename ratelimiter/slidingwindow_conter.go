package ratelimiter

import (
	"sync"
	"time"
)

// SlidingWindowCounter limits requests using a counter-based approach.
type SlidingWindowCounter struct {
	mu       sync.Mutex
	counters map[string]map[int64]int // userID -> {timestamp: count}
	limit    int                      // Max requests allowed
	window   time.Duration            // Time window
}

// NewSlidingWindowCounter initializes the rate limiter.
func NewSlidingWindowCounter(limit int, window time.Duration) *SlidingWindowCounter {
	return &SlidingWindowCounter{
		counters: make(map[string]map[int64]int),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request is within the allowed rate.
func (swc *SlidingWindowCounter) AllowRequest(userID string) bool {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	now := time.Now().UnixMilli()
	windowSize := swc.window.Milliseconds()
	currentWindow := now / windowSize // Current time window
	previousWindow := currentWindow - 1

	// Ensure user has a counter map
	if _, exists := swc.counters[userID]; !exists {
		swc.counters[userID] = make(map[int64]int)
	}

	// Cleanup old counters
	go swc.cleanupOldCounters()

	// Get current and previous counts
	currentCount := swc.counters[userID][currentWindow]
	previousCount := swc.counters[userID][previousWindow]

	// Calculate weight (linear interpolation)
	elapsedTime := now % windowSize
	weight := float64(windowSize-elapsedTime) / float64(windowSize)
	estimatedCount := float64(previousCount)*weight + float64(currentCount)

	// Check if request is allowed
	if estimatedCount >= float64(swc.limit) {
		return false
	}

	// Allow request and update counter
	swc.counters[userID][currentWindow]++

	return true
}

// cleanupOldCounters removes outdated counters asynchronously.
func (swc *SlidingWindowCounter) cleanupOldCounters() {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	now := time.Now().UnixMilli()
	windowSize := swc.window.Milliseconds()
	oldestAllowedWindow := now/windowSize - 1 // Keep only current & previous window

	for userID, counterMap := range swc.counters {
		for timestamp := range counterMap {
			if timestamp < oldestAllowedWindow {
				delete(counterMap, timestamp)
			}
		}
		if len(counterMap) == 0 {
			delete(swc.counters, userID)
		}
	}
}
