package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestServer returns a Server backed by a fresh test DataStore.
func newTestServer() *Server {
	return NewServer(newTestStore())
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
