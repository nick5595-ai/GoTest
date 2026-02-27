package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultMaxBodySize  = 1 << 20 // 1 MB
	defaultRateLimit    = 100     // requests per window
	defaultRateWindow   = 1 * time.Minute
	apiKeyHeader        = "X-API-Key"
)

// --------------------------------------------------------------------------
// Request Validation Middleware (Phase 3.3)
// --------------------------------------------------------------------------

// RequestValidationConfig holds settings for the validation middleware.
type RequestValidationConfig struct {
	MaxBodySize int64
}

// requestValidationMiddleware rejects requests with bodies larger than the
// configured limit and ensures POST/PUT/PATCH requests have a JSON content type.
func requestValidationMiddleware(cfg RequestValidationConfig) func(http.Handler) http.Handler {
	if cfg.MaxBodySize <= 0 {
		cfg.MaxBodySize = defaultMaxBodySize
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Enforce body size limit
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxBodySize)
			}

			// For mutation methods, require JSON content-type
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				ct := r.Header.Get("Content-Type")
				if ct != "" && !strings.HasPrefix(ct, "application/json") {
					writeError(w, http.StatusUnsupportedMediaType,
						"Content-Type must be application/json")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// --------------------------------------------------------------------------
// API Key Authentication Middleware (Phase 5.1)
// --------------------------------------------------------------------------

// AuthConfig holds settings for the authentication middleware.
type AuthConfig struct {
	APIKey       string   // required API key value; empty string disables auth
	ExemptPaths []string // paths that do not require authentication
}

// authMiddleware protects endpoints with a simple API key check.
// If APIKey is empty, all requests are allowed through (auth disabled).
func authMiddleware(cfg AuthConfig) func(http.Handler) http.Handler {
	exempt := make(map[string]bool, len(cfg.ExemptPaths))
	for _, p := range cfg.ExemptPaths {
		exempt[p] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth if disabled
			if cfg.APIKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Skip exempt paths
			if exempt[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// Skip OPTIONS preflight
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get(apiKeyHeader)
			if key == "" {
				setCORS(w)
				writeError(w, http.StatusUnauthorized, "missing API key")
				return
			}
			if key != cfg.APIKey {
				setCORS(w)
				writeError(w, http.StatusForbidden, "invalid API key")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// --------------------------------------------------------------------------
// Rate Limiting Middleware (Phase 5.2)
// --------------------------------------------------------------------------

// rateLimiterEntry tracks request counts per client.
type rateLimiterEntry struct {
	count     int
	windowEnd time.Time
}

// RateLimiter provides per-IP rate limiting.
type RateLimiter struct {
	mu        sync.Mutex
	clients   map[string]*rateLimiterEntry
	limit     int
	window    time.Duration
	rejected  atomic.Int64
}

// RateLimiterConfig holds settings for the rate limiter.
type RateLimiterConfig struct {
	Limit  int           // max requests per window
	Window time.Duration // window duration
}

// NewRateLimiter creates a RateLimiter with the given config.
func NewRateLimiter(cfg RateLimiterConfig) *RateLimiter {
	if cfg.Limit <= 0 {
		cfg.Limit = defaultRateLimit
	}
	if cfg.Window <= 0 {
		cfg.Window = defaultRateWindow
	}
	return &RateLimiter{
		clients: make(map[string]*rateLimiterEntry),
		limit:   cfg.Limit,
		window:  cfg.Window,
	}
}

// Allow checks whether a request from the given IP is allowed.
// Returns the remaining count and whether the request is allowed.
func (rl *RateLimiter) Allow(ip string) (remaining int, allowed bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.clients[ip]
	if !exists || now.After(entry.windowEnd) {
		rl.clients[ip] = &rateLimiterEntry{
			count:     1,
			windowEnd: now.Add(rl.window),
		}
		return rl.limit - 1, true
	}

	if entry.count >= rl.limit {
		rl.rejected.Add(1)
		return 0, false
	}

	entry.count++
	return rl.limit - entry.count, true
}

// Middleware returns an http middleware that enforces rate limits.
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			remaining, allowed := rl.Allow(ip)

			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Window", rl.window.String())

			if !allowed {
				log.Printf("Rate limit exceeded for %s", ip)
				setCORS(w)
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(rl.window.Seconds())))
				writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// clientIP extracts the client IP from the request, respecting X-Forwarded-For.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	if xff := r.Header.Get("X-Real-Ip"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
