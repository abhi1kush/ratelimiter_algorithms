package main

import (
	"fmt"
	"net/http"

	"ratelimiter.com/ratelimiter"
)

func main() {
	// Use Token Bucket Strategy (5 requests per second, refill 1 per second)

	// Create a rate limiter with the chosen strategy
	rateLimiter := ratelimiter.NewTokenBucketLimiter(5, 1)
	// Sample HTTP handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Request allowed!")
	})

	// Apply rate limiter middleware
	http.Handle("/", ratelimiter.RateLimitMiddleware(handler, rateLimiter))

	fmt.Println("Server running on port 8080...")
	http.ListenAndServe(":8080", nil)
}
