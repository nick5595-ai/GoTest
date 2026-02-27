package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ErrorResponse provides a consistent error format for API responses.
type ErrorResponse struct {
	Error string `json:"error"`
}

// TaskUpdateRequest represents the optional fields for updating a task.
type TaskUpdateRequest struct {
	Title  *string `json:"title"`
	Status *string `json:"status"`
	UserID *int    `json:"userId"`
}

// loggingMiddleware wraps an http.Handler and logs method, path, status, and duration.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		duration := time.Since(start)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

// statusRecorder wraps http.ResponseWriter to capture the status code.
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

// setCORS sets common CORS headers on a response.
func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// writeError writes a JSON error response with the given status code.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

type Server struct {
	dataStore *DataStore
	startTime time.Time
}

func NewServer(dataStore *DataStore) *Server {
	return &Server{
		dataStore: dataStore,
		startTime: time.Now(),
	}
}

func (s *Server) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/users", s.handleUsers)
	mux.HandleFunc("/api/users/", s.handleUserByID)
	mux.HandleFunc("/api/tasks", s.handleTasks)
	mux.HandleFunc("/api/tasks/", s.handleTaskByID)
	mux.HandleFunc("/api/stats", s.handleStats)
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	stats := s.dataStore.GetStats()
	dsHealth := &DataStoreHealth{
		Status:    "ok",
		UserCount: stats.Users.Total,
		TaskCount: stats.Tasks.Total,
	}

	response := HealthResponse{
		Status:    "ok",
		Message:   "Go backend is running",
		Version:   appVersion,
		Uptime:    time.Since(s.startTime).Round(time.Second).String(),
		DataStore: dsHealth,
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch r.Method {
	case http.MethodGet:
		users := s.dataStore.GetUsers()
		response := UsersResponse{
			Users: users,
			Count: len(users),
		}
		writeJSON(w, http.StatusOK, response)

	case http.MethodPost:
		var body struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Role  string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			log.Printf("Error decoding user creation request: %v", err)
			writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))
			return
		}

		user, err := s.dataStore.CreateUser(body.Name, body.Email, body.Role)
		if err != nil {
			log.Printf("Error creating user: %v", err)
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		log.Printf("Created user: id=%d name=%s", user.ID, user.Name)
		writeJSON(w, http.StatusCreated, user)

	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) handleUserByID(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/users/")
	id, err := strconv.Atoi(path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user := s.dataStore.GetUserByID(id)
	if user == nil {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch r.Method {
	case http.MethodGet:
		status := r.URL.Query().Get("status")
		userID := r.URL.Query().Get("userId")

		tasks := s.dataStore.GetTasks(status, userID)
		response := TasksResponse{
			Tasks: tasks,
			Count: len(tasks),
		}
		writeJSON(w, http.StatusOK, response)

	case http.MethodPost:
		var body struct {
			Title  string `json:"title"`
			Status string `json:"status"`
			UserID int    `json:"userId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			log.Printf("Error decoding task creation request: %v", err)
			writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))
			return
		}

		task, err := s.dataStore.CreateTask(body.Title, body.Status, body.UserID)
		if err != nil {
			log.Printf("Error creating task: %v", err)
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		log.Printf("Created task: id=%d title=%s", task.ID, task.Title)
		writeJSON(w, http.StatusCreated, task)

	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	id, err := strconv.Atoi(path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		task := s.dataStore.GetTaskByID(id)
		if task == nil {
			writeError(w, http.StatusNotFound, "Task not found")
			return
		}
		writeJSON(w, http.StatusOK, task)

	case http.MethodPut, http.MethodPatch:
		var body TaskUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			log.Printf("Error decoding task update request: %v", err)
			writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))
			return
		}

		updated, err := s.dataStore.UpdateTask(id, body.Title, body.Status, body.UserID)
		if err != nil {
			log.Printf("Error updating task %d: %v", id, err)
			if err.Error() == "task not found" {
				writeError(w, http.StatusNotFound, err.Error())
			} else {
				writeError(w, http.StatusBadRequest, err.Error())
			}
			return
		}

		log.Printf("Updated task: id=%d", updated.ID)
		writeJSON(w, http.StatusOK, updated)

	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	stats := s.dataStore.GetStats()
	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) Start(port string) {
	mux := s.setupRoutes()
	handler := loggingMiddleware(mux)

	if port == "" {
		port = defaultPort
	}

	log.Printf("Go backend server starting on http://localhost:%s", port)
	log.Printf("Serving data directly from Go backend")

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
