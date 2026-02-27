package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Request Validation Middleware (Phase 3.3)
// ---------------------------------------------------------------------------

func TestRequestValidation_RejectsLargeBody(t *testing.T) {
	cfg := RequestValidationConfig{MaxBodySize: 64} // 64 bytes max
	mw := requestValidationMiddleware(cfg)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to read the body — MaxBytesReader will trigger an error
		var v map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "body too large")
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(inner)
	largeBody := strings.Repeat("x", 200)
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(largeBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for large body, got %d", w.Code)
	}
}

func TestRequestValidation_RejectsNonJSONContentType(t *testing.T) {
	cfg := RequestValidationConfig{MaxBodySize: 1 << 20}
	mw := requestValidationMiddleware(cfg)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(inner)
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString("data"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415 for non-JSON content type, got %d", w.Code)
	}
}

func TestRequestValidation_AllowsGETWithoutContentType(t *testing.T) {
	cfg := RequestValidationConfig{MaxBodySize: 1 << 20}
	mw := requestValidationMiddleware(cfg)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(inner)
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET, got %d", w.Code)
	}
}

func TestRequestValidation_AllowsJSONContentType(t *testing.T) {
	cfg := RequestValidationConfig{MaxBodySize: 1 << 20}
	mw := requestValidationMiddleware(cfg)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(inner)
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for JSON content type, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Authentication Middleware (Phase 5.1)
// ---------------------------------------------------------------------------

func TestAuth_DisabledWhenNoKey(t *testing.T) {
	cfg := AuthConfig{APIKey: ""}
	mw := authMiddleware(cfg)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(inner)
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 when auth disabled, got %d", w.Code)
	}
}

func TestAuth_MissingKey(t *testing.T) {
	cfg := AuthConfig{APIKey: "secret-key"}
	mw := authMiddleware(cfg)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(inner)
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing key, got %d", w.Code)
	}
}

func TestAuth_InvalidKey(t *testing.T) {
	cfg := AuthConfig{APIKey: "secret-key"}
	mw := authMiddleware(cfg)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(inner)
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for invalid key, got %d", w.Code)
	}
}

func TestAuth_ValidKey(t *testing.T) {
	cfg := AuthConfig{APIKey: "secret-key"}
	mw := authMiddleware(cfg)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(inner)
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set("X-API-Key", "secret-key")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid key, got %d", w.Code)
	}
}

func TestAuth_ExemptPaths(t *testing.T) {
	cfg := AuthConfig{
		APIKey:      "secret-key",
		ExemptPaths: []string{"/health", "/metrics"},
	}
	mw := authMiddleware(cfg)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(inner)

	for _, path := range []string{"/health", "/metrics"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200 for exempt path %s, got %d", path, w.Code)
		}
	}
}

func TestAuth_OptionsPassthrough(t *testing.T) {
	cfg := AuthConfig{APIKey: "secret-key"}
	mw := authMiddleware(cfg)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	handler := mw(inner)
	req := httptest.NewRequest(http.MethodOptions, "/api/users", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for OPTIONS preflight, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Rate Limiting (Phase 5.2)
// ---------------------------------------------------------------------------

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{Limit: 5, Window: time.Minute})

	for i := 0; i < 5; i++ {
		remaining, allowed := rl.Allow("127.0.0.1")
		if !allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
		if remaining != 5-i-1 {
			t.Errorf("request %d: expected remaining %d, got %d", i+1, 5-i-1, remaining)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{Limit: 3, Window: time.Minute})

	for i := 0; i < 3; i++ {
		_, allowed := rl.Allow("127.0.0.1")
		if !allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	remaining, allowed := rl.Allow("127.0.0.1")
	if allowed {
		t.Fatal("4th request should be blocked")
	}
	if remaining != 0 {
		t.Errorf("expected remaining 0, got %d", remaining)
	}
}

func TestRateLimiter_SeparateIPs(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{Limit: 2, Window: time.Minute})

	rl.Allow("1.1.1.1")
	rl.Allow("1.1.1.1")

	// Different IP should still be allowed
	_, allowed := rl.Allow("2.2.2.2")
	if !allowed {
		t.Fatal("different IP should be allowed")
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{Limit: 2, Window: 50 * time.Millisecond})

	rl.Allow("127.0.0.1")
	rl.Allow("127.0.0.1")

	// Should be blocked
	_, allowed := rl.Allow("127.0.0.1")
	if allowed {
		t.Fatal("should be blocked at limit")
	}

	// Wait for window to expire
	time.Sleep(100 * time.Millisecond)

	_, allowed = rl.Allow("127.0.0.1")
	if !allowed {
		t.Fatal("should be allowed after window reset")
	}
}

func TestRateLimiter_Middleware(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{Limit: 2, Window: time.Minute})
	mw := rl.Middleware()

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(inner)

	// First 2 requests should pass
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
		}
		// Check rate limit headers
		if w.Header().Get("X-RateLimit-Limit") != "2" {
			t.Errorf("expected X-RateLimit-Limit=2, got %s", w.Header().Get("X-RateLimit-Limit"))
		}
	}

	// 3rd request should be rejected
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header")
	}
}

func TestClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")
	ip := clientIP(req)
	if ip != "10.0.0.1" {
		t.Errorf("expected 10.0.0.1, got %s", ip)
	}
}

func TestClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-Ip", "10.0.0.5")
	ip := clientIP(req)
	if ip != "10.0.0.5" {
		t.Errorf("expected 10.0.0.5, got %s", ip)
	}
}

func TestClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:54321"
	ip := clientIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("expected 192.168.1.1, got %s", ip)
	}
}
