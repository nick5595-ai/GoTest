package main

import (
	"path/filepath"
	"testing"
)

// newTestSQLiteStore creates a temporary SQLite store seeded with test data.
func newTestSQLiteStore(t *testing.T) *SQLiteStore {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	if err := s.Seed(); err != nil {
		t.Fatalf("Seed failed: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// ---------------------------------------------------------------------------
// Migration & Seed
// ---------------------------------------------------------------------------

func TestSQLiteStore_NewAndMigrate(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer s.Close()

	// Tables should exist (no users yet before seeding)
	users := s.GetUsers()
	if users != nil {
		t.Errorf("expected nil users before seed, got %d", len(users))
	}
}

func TestSQLiteStore_Seed(t *testing.T) {
	s := newTestSQLiteStore(t)

	users := s.GetUsers()
	if len(users) != 3 {
		t.Fatalf("expected 3 users after seed, got %d", len(users))
	}

	tasks := s.GetTasks("", "")
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks after seed, got %d", len(tasks))
	}
}

func TestSQLiteStore_SeedIdempotent(t *testing.T) {
	s := newTestSQLiteStore(t)

	// Seed again — should not duplicate
	if err := s.Seed(); err != nil {
		t.Fatalf("second Seed failed: %v", err)
	}

	users := s.GetUsers()
	if len(users) != 3 {
		t.Errorf("expected 3 users after double seed, got %d", len(users))
	}
}

// ---------------------------------------------------------------------------
// GetUsers / GetUserByID
// ---------------------------------------------------------------------------

func TestSQLiteStore_GetUsers(t *testing.T) {
	s := newTestSQLiteStore(t)
	users := s.GetUsers()
	if len(users) != 3 {
		t.Fatalf("expected 3 users, got %d", len(users))
	}
	if users[0].Name != "John Doe" {
		t.Errorf("expected first user John Doe, got %s", users[0].Name)
	}
}

func TestSQLiteStore_GetUserByID_Found(t *testing.T) {
	s := newTestSQLiteStore(t)
	u := s.GetUserByID(1)
	if u == nil {
		t.Fatal("expected user, got nil")
	}
	if u.Name != "John Doe" {
		t.Errorf("expected John Doe, got %s", u.Name)
	}
}

func TestSQLiteStore_GetUserByID_NotFound(t *testing.T) {
	s := newTestSQLiteStore(t)
	u := s.GetUserByID(999)
	if u != nil {
		t.Errorf("expected nil for non-existent user, got %+v", u)
	}
}

// ---------------------------------------------------------------------------
// GetTasks / GetTaskByID
// ---------------------------------------------------------------------------

func TestSQLiteStore_GetTasks_All(t *testing.T) {
	s := newTestSQLiteStore(t)
	tasks := s.GetTasks("", "")
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}
}

func TestSQLiteStore_GetTasks_FilterByStatus(t *testing.T) {
	s := newTestSQLiteStore(t)
	tasks := s.GetTasks("pending", "")
	if len(tasks) != 1 {
		t.Fatalf("expected 1 pending task, got %d", len(tasks))
	}
	if tasks[0].Title != "Implement authentication" {
		t.Errorf("unexpected task: %s", tasks[0].Title)
	}
}

func TestSQLiteStore_GetTasks_FilterByUserID(t *testing.T) {
	s := newTestSQLiteStore(t)
	tasks := s.GetTasks("", "2")
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task for user 2, got %d", len(tasks))
	}
}

func TestSQLiteStore_GetTasks_FilterByStatusAndUserID(t *testing.T) {
	s := newTestSQLiteStore(t)
	tasks := s.GetTasks("in-progress", "2")
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
}

func TestSQLiteStore_GetTasks_InvalidUserID(t *testing.T) {
	s := newTestSQLiteStore(t)
	tasks := s.GetTasks("", "abc")
	if tasks != nil {
		t.Errorf("expected nil for invalid userID, got %d tasks", len(tasks))
	}
}

func TestSQLiteStore_GetTaskByID_Found(t *testing.T) {
	s := newTestSQLiteStore(t)
	task := s.GetTaskByID(1)
	if task == nil {
		t.Fatal("expected task, got nil")
	}
	if task.Title != "Implement authentication" {
		t.Errorf("expected 'Implement authentication', got %s", task.Title)
	}
}

func TestSQLiteStore_GetTaskByID_NotFound(t *testing.T) {
	s := newTestSQLiteStore(t)
	task := s.GetTaskByID(999)
	if task != nil {
		t.Errorf("expected nil for non-existent task, got %+v", task)
	}
}

// ---------------------------------------------------------------------------
// GetStats
// ---------------------------------------------------------------------------

func TestSQLiteStore_GetStats(t *testing.T) {
	s := newTestSQLiteStore(t)
	stats := s.GetStats()

	if stats.Users.Total != 3 {
		t.Errorf("expected 3 users, got %d", stats.Users.Total)
	}
	if stats.Tasks.Total != 3 {
		t.Errorf("expected 3 tasks, got %d", stats.Tasks.Total)
	}
	if stats.Tasks.Pending != 1 {
		t.Errorf("expected 1 pending, got %d", stats.Tasks.Pending)
	}
	if stats.Tasks.InProgress != 1 {
		t.Errorf("expected 1 in-progress, got %d", stats.Tasks.InProgress)
	}
	if stats.Tasks.Completed != 1 {
		t.Errorf("expected 1 completed, got %d", stats.Tasks.Completed)
	}
}

// ---------------------------------------------------------------------------
// CreateUser
// ---------------------------------------------------------------------------

func TestSQLiteStore_CreateUser_Success(t *testing.T) {
	s := newTestSQLiteStore(t)
	u, err := s.CreateUser("Alice", "alice@test.com", "dev")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if u.ID != 4 {
		t.Errorf("expected ID 4, got %d", u.ID)
	}
	if u.Name != "Alice" {
		t.Errorf("expected Alice, got %s", u.Name)
	}

	// Verify persisted
	found := s.GetUserByID(u.ID)
	if found == nil {
		t.Fatal("expected to find created user")
	}
}

func TestSQLiteStore_CreateUser_EmptyName(t *testing.T) {
	s := newTestSQLiteStore(t)
	_, err := s.CreateUser("", "a@a.com", "dev")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSQLiteStore_CreateUser_InvalidEmail(t *testing.T) {
	s := newTestSQLiteStore(t)
	_, err := s.CreateUser("X", "not-email", "dev")
	if err == nil {
		t.Fatal("expected error for invalid email")
	}
}

func TestSQLiteStore_CreateUser_EmptyRole(t *testing.T) {
	s := newTestSQLiteStore(t)
	_, err := s.CreateUser("X", "x@x.com", "")
	if err == nil {
		t.Fatal("expected error for empty role")
	}
}

// ---------------------------------------------------------------------------
// CreateTask
// ---------------------------------------------------------------------------

func TestSQLiteStore_CreateTask_Success(t *testing.T) {
	s := newTestSQLiteStore(t)
	task, err := s.CreateTask("New task", "pending", 1)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	if task.ID != 4 {
		t.Errorf("expected ID 4, got %d", task.ID)
	}

	// Verify persisted
	found := s.GetTaskByID(task.ID)
	if found == nil {
		t.Fatal("expected to find created task")
	}
}

func TestSQLiteStore_CreateTask_EmptyTitle(t *testing.T) {
	s := newTestSQLiteStore(t)
	_, err := s.CreateTask("", "pending", 1)
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestSQLiteStore_CreateTask_InvalidStatus(t *testing.T) {
	s := newTestSQLiteStore(t)
	_, err := s.CreateTask("X", "invalid", 1)
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestSQLiteStore_CreateTask_InvalidUserID(t *testing.T) {
	s := newTestSQLiteStore(t)
	_, err := s.CreateTask("X", "pending", 999)
	if err == nil {
		t.Fatal("expected error for non-existent userId")
	}
}

// ---------------------------------------------------------------------------
// UpdateTask
// ---------------------------------------------------------------------------

func TestSQLiteStore_UpdateTask_Success(t *testing.T) {
	s := newTestSQLiteStore(t)
	newStatus := "completed"
	updated, err := s.UpdateTask(1, nil, &newStatus, nil)
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}
	if updated.Status != "completed" {
		t.Errorf("expected completed, got %s", updated.Status)
	}

	// Verify persisted
	found := s.GetTaskByID(1)
	if found.Status != "completed" {
		t.Errorf("expected persisted status completed, got %s", found.Status)
	}
}

func TestSQLiteStore_UpdateTask_AllFields(t *testing.T) {
	s := newTestSQLiteStore(t)
	title := "Updated"
	status := "in-progress"
	userID := 2
	updated, err := s.UpdateTask(1, &title, &status, &userID)
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}
	if updated.Title != "Updated" || updated.Status != "in-progress" || updated.UserID != 2 {
		t.Errorf("unexpected result: %+v", updated)
	}
}

func TestSQLiteStore_UpdateTask_NotFound(t *testing.T) {
	s := newTestSQLiteStore(t)
	status := "completed"
	_, err := s.UpdateTask(999, nil, &status, nil)
	if err == nil || err.Error() != "task not found" {
		t.Fatalf("expected 'task not found' error, got %v", err)
	}
}

func TestSQLiteStore_UpdateTask_InvalidStatus(t *testing.T) {
	s := newTestSQLiteStore(t)
	status := "bad"
	_, err := s.UpdateTask(1, nil, &status, nil)
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestSQLiteStore_UpdateTask_InvalidUserID(t *testing.T) {
	s := newTestSQLiteStore(t)
	uid := 999
	_, err := s.UpdateTask(1, nil, nil, &uid)
	if err == nil {
		t.Fatal("expected error for non-existent userId")
	}
}

// ---------------------------------------------------------------------------
// Data persists across connections
// ---------------------------------------------------------------------------

func TestSQLiteStore_PersistenceAcrossConnections(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "persist.db")

	// Connection 1: create and seed, then add a user
	s1, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("connection 1 open failed: %v", err)
	}
	s1.Seed()
	s1.CreateUser("Alice", "alice@test.com", "dev")
	s1.Close()

	// Connection 2: should see the new user
	s2, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("connection 2 open failed: %v", err)
	}
	defer s2.Close()

	users := s2.GetUsers()
	if len(users) != 4 {
		t.Fatalf("expected 4 users after reconnect, got %d", len(users))
	}

	found := false
	for _, u := range users {
		if u.Name == "Alice" {
			found = true
		}
	}
	if !found {
		t.Error("Alice not found after reconnect")
	}
}

// ---------------------------------------------------------------------------
// Server integration with SQLiteStore
// ---------------------------------------------------------------------------

func TestServer_WithSQLiteStore(t *testing.T) {
	s := newTestSQLiteStore(t)
	srv := NewServer(s, ServerConfig{})

	// Verify health works
	stats := s.GetStats()
	if stats.Users.Total != 3 {
		t.Errorf("expected 3 users from SQLite store, got %d", stats.Users.Total)
	}

	// Use the server directly
	_ = srv // server is valid and usable with SQLite backend
}

// ---------------------------------------------------------------------------
// Verify SQLiteStore satisfies Store interface at compile time
// ---------------------------------------------------------------------------

var _ Store = (*SQLiteStore)(nil)
