package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestServer returns a Server backed by a fresh test DataStore (no cache/metrics/auth).
func newTestServer() *Server {
	return NewServer(newTestStore(), ServerConfig{})
}

// newTestServerWithCache returns a Server with caching enabled.
func newTestServerWithCache() *Server {
	return NewServer(newTestStore(), ServerConfig{
		Cache: NewCache(5 * time.Minute),
	})
}

// ---------------------------------------------------------------------------
// Health
// ---------------------------------------------------------------------------

func TestServer_handleHealth_GET(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	s.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status ok, got %s", resp.Status)
	}
}

func TestServer_handleHealth_MethodNotAllowed(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()
	s.handleHealth(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// GET /api/users
// ---------------------------------------------------------------------------

func TestServer_handleUsers_GET(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	s.handleUsers(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp UsersResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Count != 2 {
		t.Errorf("expected 2 users, got %d", resp.Count)
	}
}

// ---------------------------------------------------------------------------
// POST /api/users
// ---------------------------------------------------------------------------

func TestServer_handleUsers_POST_Success(t *testing.T) {
	s := newTestServer()
	body := `{"name":"Charlie","email":"charlie@example.com","role":"tester"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleUsers(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d; body: %s", w.Code, w.Body.String())
	}
	var user User
	if err := json.NewDecoder(w.Body).Decode(&user); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if user.ID != 3 {
		t.Errorf("expected ID 3, got %d", user.ID)
	}
	if user.Name != "Charlie" {
		t.Errorf("expected Charlie, got %s", user.Name)
	}
}

func TestServer_handleUsers_POST_InvalidJSON(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleUsers(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestServer_handleUsers_POST_MissingFields(t *testing.T) {
	s := newTestServer()
	body := `{"name":"","email":"charlie@example.com","role":"tester"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleUsers(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestServer_handleUsers_POST_InvalidEmail(t *testing.T) {
	s := newTestServer()
	body := `{"name":"Charlie","email":"not-valid","role":"tester"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleUsers(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// GET /api/users/:id
// ---------------------------------------------------------------------------

func TestServer_handleUserByID_Found(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/users/1", nil)
	w := httptest.NewRecorder()
	s.handleUserByID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var user User
	if err := json.NewDecoder(w.Body).Decode(&user); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if user.Name != "Alice" {
		t.Errorf("expected Alice, got %s", user.Name)
	}
}

func TestServer_handleUserByID_NotFound(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/users/999", nil)
	w := httptest.NewRecorder()
	s.handleUserByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestServer_handleUserByID_InvalidID(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/users/abc", nil)
	w := httptest.NewRecorder()
	s.handleUserByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// GET /api/tasks
// ---------------------------------------------------------------------------

func TestServer_handleTasks_GET(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	w := httptest.NewRecorder()
	s.handleTasks(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp TasksResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Count != 3 {
		t.Errorf("expected 3 tasks, got %d", resp.Count)
	}
}

func TestServer_handleTasks_GET_FilterStatus(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/tasks?status=pending", nil)
	w := httptest.NewRecorder()
	s.handleTasks(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp TasksResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Count != 1 {
		t.Errorf("expected 1 pending task, got %d", resp.Count)
	}
}

// ---------------------------------------------------------------------------
// POST /api/tasks
// ---------------------------------------------------------------------------

func TestServer_handleTasks_POST_Success(t *testing.T) {
	s := newTestServer()
	body := `{"title":"New Task","status":"pending","userId":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleTasks(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d; body: %s", w.Code, w.Body.String())
	}
	var task Task
	if err := json.NewDecoder(w.Body).Decode(&task); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if task.ID != 4 {
		t.Errorf("expected ID 4, got %d", task.ID)
	}
}

func TestServer_handleTasks_POST_InvalidStatus(t *testing.T) {
	s := newTestServer()
	body := `{"title":"Task","status":"invalid","userId":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleTasks(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestServer_handleTasks_POST_InvalidUserID(t *testing.T) {
	s := newTestServer()
	body := `{"title":"Task","status":"pending","userId":999}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleTasks(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestServer_handleTasks_POST_InvalidJSON(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleTasks(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// GET /api/tasks/:id
// ---------------------------------------------------------------------------

func TestServer_handleTaskByID_GET_Found(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/1", nil)
	w := httptest.NewRecorder()
	s.handleTaskByID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var task Task
	json.NewDecoder(w.Body).Decode(&task)
	if task.Title != "Task One" {
		t.Errorf("expected Task One, got %s", task.Title)
	}
}

func TestServer_handleTaskByID_GET_NotFound(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/999", nil)
	w := httptest.NewRecorder()
	s.handleTaskByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// PUT /api/tasks/:id
// ---------------------------------------------------------------------------

func TestServer_handleTaskByID_PUT_Success(t *testing.T) {
	s := newTestServer()
	body := `{"status":"completed"}`
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleTaskByID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
	var task Task
	json.NewDecoder(w.Body).Decode(&task)
	if task.Status != "completed" {
		t.Errorf("expected completed, got %s", task.Status)
	}
	// Title should remain unchanged
	if task.Title != "Task One" {
		t.Errorf("expected Task One unchanged, got %s", task.Title)
	}
}

func TestServer_handleTaskByID_PUT_AllFields(t *testing.T) {
	s := newTestServer()
	body := `{"title":"Updated","status":"in-progress","userId":2}`
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleTaskByID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var task Task
	json.NewDecoder(w.Body).Decode(&task)
	if task.Title != "Updated" || task.Status != "in-progress" || task.UserID != 2 {
		t.Errorf("unexpected task state: %+v", task)
	}
}

func TestServer_handleTaskByID_PUT_NotFound(t *testing.T) {
	s := newTestServer()
	body := `{"status":"completed"}`
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/999", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleTaskByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestServer_handleTaskByID_PUT_InvalidStatus(t *testing.T) {
	s := newTestServer()
	body := `{"status":"bad"}`
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleTaskByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestServer_handleTaskByID_PUT_InvalidUserID(t *testing.T) {
	s := newTestServer()
	body := `{"userId":999}`
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleTaskByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestServer_handleTaskByID_PUT_InvalidJSON(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/1", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleTaskByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestServer_handleTaskByID_InvalidID(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/abc", nil)
	w := httptest.NewRecorder()
	s.handleTaskByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// GET /api/stats
// ---------------------------------------------------------------------------

func TestServer_handleStats_GET(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	w := httptest.NewRecorder()
	s.handleStats(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var stats StatsResponse
	json.NewDecoder(w.Body).Decode(&stats)
	if stats.Users.Total != 2 {
		t.Errorf("expected 2 users, got %d", stats.Users.Total)
	}
}

func TestServer_handleStats_MethodNotAllowed(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/api/stats", nil)
	w := httptest.NewRecorder()
	s.handleStats(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// CORS preflight
// ---------------------------------------------------------------------------

func TestServer_CORS_Options(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodOptions, "/api/users", nil)
	w := httptest.NewRecorder()
	s.handleUsers(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for OPTIONS, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS origin header")
	}
}

// ---------------------------------------------------------------------------
// Integration: full mux routing
// ---------------------------------------------------------------------------

func TestServer_FullMux_Routing(t *testing.T) {
	s := newTestServer()
	mux := s.setupRoutes()
	handler := loggingMiddleware(mux)

	tests := []struct {
		method string
		path   string
		body   string
		status int
	}{
		{http.MethodGet, "/health", "", http.StatusOK},
		{http.MethodGet, "/api/users", "", http.StatusOK},
		{http.MethodGet, "/api/users/1", "", http.StatusOK},
		{http.MethodGet, "/api/tasks", "", http.StatusOK},
		{http.MethodGet, "/api/stats", "", http.StatusOK},
		{http.MethodPost, "/api/users", `{"name":"X","email":"x@x.com","role":"dev"}`, http.StatusCreated},
		{http.MethodPost, "/api/tasks", `{"title":"T","status":"pending","userId":1}`, http.StatusCreated},
		{http.MethodPut, "/api/tasks/1", `{"status":"completed"}`, http.StatusOK},
	}

	for _, tc := range tests {
		var reqBody *bytes.Buffer
		if tc.body != "" {
			reqBody = bytes.NewBufferString(tc.body)
		} else {
			reqBody = &bytes.Buffer{}
		}
		req := httptest.NewRequest(tc.method, tc.path, reqBody)
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != tc.status {
			t.Errorf("%s %s: expected %d, got %d; body: %s", tc.method, tc.path, tc.status, w.Code, w.Body.String())
		}
	}
}

// ---------------------------------------------------------------------------
// Cache Stats endpoint
// ---------------------------------------------------------------------------

func TestServer_handleCacheStats_Disabled(t *testing.T) {
	s := newTestServer() // no cache
	req := httptest.NewRequest(http.MethodGet, "/api/cache/stats", nil)
	w := httptest.NewRecorder()
	s.handleCacheStats(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "disabled" {
		t.Errorf("expected disabled, got %s", resp["status"])
	}
}

func TestServer_handleCacheStats_Enabled(t *testing.T) {
	s := newTestServerWithCache()
	// Trigger a cache set via users GET
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	s.handleUsers(w, req)

	// Now get cache stats
	req = httptest.NewRequest(http.MethodGet, "/api/cache/stats", nil)
	w = httptest.NewRecorder()
	s.handleCacheStats(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp CacheStatsResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Entries < 1 {
		t.Errorf("expected at least 1 cache entry, got %d", resp.Entries)
	}
}

// ---------------------------------------------------------------------------
// Metrics endpoint
// ---------------------------------------------------------------------------

func TestServer_handleMetrics_Disabled(t *testing.T) {
	s := newTestServer() // no metrics
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	s.handleMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "disabled" {
		t.Errorf("expected disabled, got %s", resp["status"])
	}
}

func TestServer_handleMetrics_Enabled(t *testing.T) {
	m := NewMetrics()
	s := NewServer(newTestStore(), ServerConfig{Metrics: m})

	// Record something via the metrics directly
	m.Record("GET", "/test", 200, time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	s.handleMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp MetricsResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.TotalRequests != 1 {
		t.Errorf("expected 1 total request, got %d", resp.TotalRequests)
	}
}

// ---------------------------------------------------------------------------
// Cache integration: invalidation on mutations
// ---------------------------------------------------------------------------

func TestServer_CacheInvalidation_OnCreateUser(t *testing.T) {
	s := newTestServerWithCache()

	// GET users — populates cache
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	s.handleUsers(w, req)

	// POST user — should invalidate cache
	body := `{"name":"New","email":"new@test.com","role":"dev"}`
	req = httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.handleUsers(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	// GET users again — should get fresh data with 3 users
	req = httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w = httptest.NewRecorder()
	s.handleUsers(w, req)

	var resp UsersResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Count != 3 {
		t.Errorf("expected 3 users after create+invalidation, got %d", resp.Count)
	}
}

func TestServer_CacheInvalidation_OnCreateTask(t *testing.T) {
	s := newTestServerWithCache()

	// GET tasks — populates cache
	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	w := httptest.NewRecorder()
	s.handleTasks(w, req)

	// POST task
	body := `{"title":"New","status":"pending","userId":1}`
	req = httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.handleTasks(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	// GET tasks again
	req = httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	w = httptest.NewRecorder()
	s.handleTasks(w, req)

	var resp TasksResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Count != 4 {
		t.Errorf("expected 4 tasks after create+invalidation, got %d", resp.Count)
	}
}

func TestServer_CacheInvalidation_OnUpdateTask(t *testing.T) {
	s := newTestServerWithCache()

	// GET tasks — populates cache
	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	w := httptest.NewRecorder()
	s.handleTasks(w, req)

	// PUT task
	body := `{"status":"completed"}`
	req = httptest.NewRequest(http.MethodPut, "/api/tasks/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.handleTaskByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Cache should be cleared
	stats := s.cache.Stats()
	if stats.Entries != 0 {
		t.Errorf("expected 0 cache entries after invalidation, got %d", stats.Entries)
	}
}

// ---------------------------------------------------------------------------
// Full middleware chain integration
// ---------------------------------------------------------------------------

func TestServer_FullMiddlewareChain(t *testing.T) {
	m := NewMetrics()
	rl := NewRateLimiter(RateLimiterConfig{Limit: 100, Window: time.Minute})
	s := NewServer(newTestStore(), ServerConfig{
		Cache:       NewCache(5 * time.Minute),
		Metrics:     m,
		RateLimiter: rl,
		AuthCfg:     AuthConfig{APIKey: "test-key", ExemptPaths: []string{"/health", "/metrics"}},
		ValidCfg:    RequestValidationConfig{MaxBodySize: 1 << 20},
	})

	mux := s.setupRoutes()
	var handler http.Handler = mux
	handler = metricsMiddleware(s.metrics)(handler)
	handler = loggingMiddleware(handler)
	handler = requestValidationMiddleware(s.validCfg)(handler)
	handler = s.rateLimiter.Middleware()(handler)
	handler = authMiddleware(s.authCfg)(handler)

	// Health should be exempt from auth
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("health: expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	// API call without key should be 401
	req = httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("no key: expected 401, got %d", w.Code)
	}

	// API call with valid key
	req = httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set("X-API-Key", "test-key")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("valid key: expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	// Check rate limit headers present
	if w.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("expected X-RateLimit-Limit header")
	}

	// Metrics should have recorded requests that passed auth (health exempt + valid key)
	snap := m.Snapshot()
	if snap.TotalRequests < 2 {
		t.Errorf("expected at least 2 recorded requests, got %d", snap.TotalRequests)
	}
}
