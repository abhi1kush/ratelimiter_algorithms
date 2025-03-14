package ratelimiter

import (
	"net/http"
	"strings"
)

// RateLimiterStrategy defines the interface for different rate-limiting strategies
type RateLimiter interface {
	AllowRequest(userID string) bool
}

// Middleware for rate limiting
func RateLimitMiddleware(next http.Handler, rl RateLimiter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := ExtractUserID(r)

		if !rl.AllowRequest(userID) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ExtractUserID gets user identifier (IP address)
func ExtractUserID(r *http.Request) string {
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}
	return ip
}
