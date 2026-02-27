package main

import (
	"sync"
	"testing"
)

// newTestStore returns a fresh DataStore with seed data for testing.
func newTestStore() *DataStore {
	return &DataStore{
		users: []User{
			{ID: 1, Name: "Alice", Email: "alice@example.com", Role: "developer"},
			{ID: 2, Name: "Bob", Email: "bob@example.com", Role: "designer"},
		},
		tasks: []Task{
			{ID: 1, Title: "Task One", Status: "pending", UserID: 1},
			{ID: 2, Title: "Task Two", Status: "in-progress", UserID: 2},
			{ID: 3, Title: "Task Three", Status: "completed", UserID: 1},
		},
	}
}

// ---------------------------------------------------------------------------
// GetUsers
// ---------------------------------------------------------------------------

func TestDataStore_GetUsers(t *testing.T) {
	ds := newTestStore()
	users := ds.GetUsers()
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Name != "Alice" {
		t.Errorf("expected first user Alice, got %s", users[0].Name)
	}
}

// ---------------------------------------------------------------------------
// GetUserByID
// ---------------------------------------------------------------------------

func TestDataStore_GetUserByID_Found(t *testing.T) {
	ds := newTestStore()
	user := ds.GetUserByID(1)
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.Name != "Alice" {
		t.Errorf("expected Alice, got %s", user.Name)
	}
}

func TestDataStore_GetUserByID_NotFound(t *testing.T) {
	ds := newTestStore()
	user := ds.GetUserByID(999)
	if user != nil {
		t.Fatalf("expected nil, got user %+v", user)
	}
}

// ---------------------------------------------------------------------------
// GetTasks
// ---------------------------------------------------------------------------

func TestDataStore_GetTasks_All(t *testing.T) {
	ds := newTestStore()
	tasks := ds.GetTasks("", "")
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}
}

func TestDataStore_GetTasks_FilterByStatus(t *testing.T) {
	ds := newTestStore()
	tasks := ds.GetTasks("pending", "")
	if len(tasks) != 1 {
		t.Fatalf("expected 1 pending task, got %d", len(tasks))
	}
	if tasks[0].Title != "Task One" {
		t.Errorf("expected Task One, got %s", tasks[0].Title)
	}
}

func TestDataStore_GetTasks_FilterByUserID(t *testing.T) {
	ds := newTestStore()
	tasks := ds.GetTasks("", "1")
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks for user 1, got %d", len(tasks))
	}
}

func TestDataStore_GetTasks_FilterByStatusAndUserID(t *testing.T) {
	ds := newTestStore()
	tasks := ds.GetTasks("completed", "1")
	if len(tasks) != 1 {
		t.Fatalf("expected 1 completed task for user 1, got %d", len(tasks))
	}
}

func TestDataStore_GetTasks_InvalidUserID(t *testing.T) {
	ds := newTestStore()
	tasks := ds.GetTasks("", "abc")
	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks for invalid userId, got %d", len(tasks))
	}
}

// ---------------------------------------------------------------------------
// GetTaskByID
// ---------------------------------------------------------------------------

func TestDataStore_GetTaskByID_Found(t *testing.T) {
	ds := newTestStore()
	task := ds.GetTaskByID(2)
	if task == nil {
		t.Fatal("expected task, got nil")
	}
	if task.Title != "Task Two" {
		t.Errorf("expected Task Two, got %s", task.Title)
	}
}

func TestDataStore_GetTaskByID_NotFound(t *testing.T) {
	ds := newTestStore()
	task := ds.GetTaskByID(999)
	if task != nil {
		t.Fatalf("expected nil, got task %+v", task)
	}
}

// ---------------------------------------------------------------------------
// GetStats
// ---------------------------------------------------------------------------

func TestDataStore_GetStats(t *testing.T) {
	ds := newTestStore()
	stats := ds.GetStats()
	if stats.Users.Total != 2 {
		t.Errorf("expected 2 users total, got %d", stats.Users.Total)
	}
	if stats.Tasks.Total != 3 {
		t.Errorf("expected 3 tasks total, got %d", stats.Tasks.Total)
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

func TestDataStore_CreateUser_Success(t *testing.T) {
	ds := newTestStore()
	user, err := ds.CreateUser("Charlie", "charlie@example.com", "manager")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != 3 {
		t.Errorf("expected ID 3, got %d", user.ID)
	}
	if user.Name != "Charlie" {
		t.Errorf("expected Charlie, got %s", user.Name)
	}
	// Verify it appears in GetUsers
	users := ds.GetUsers()
	if len(users) != 3 {
		t.Errorf("expected 3 users after create, got %d", len(users))
	}
}

func TestDataStore_CreateUser_EmptyName(t *testing.T) {
	ds := newTestStore()
	_, err := ds.CreateUser("", "test@example.com", "dev")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestDataStore_CreateUser_EmptyEmail(t *testing.T) {
	ds := newTestStore()
	_, err := ds.CreateUser("Test", "", "dev")
	if err == nil {
		t.Fatal("expected error for empty email")
	}
}

func TestDataStore_CreateUser_InvalidEmail(t *testing.T) {
	ds := newTestStore()
	_, err := ds.CreateUser("Test", "not-an-email", "dev")
	if err == nil {
		t.Fatal("expected error for invalid email")
	}
}

func TestDataStore_CreateUser_EmptyRole(t *testing.T) {
	ds := newTestStore()
	_, err := ds.CreateUser("Test", "test@example.com", "")
	if err == nil {
		t.Fatal("expected error for empty role")
	}
}

// ---------------------------------------------------------------------------
// CreateTask
// ---------------------------------------------------------------------------

func TestDataStore_CreateTask_Success(t *testing.T) {
	ds := newTestStore()
	task, err := ds.CreateTask("New Task", "pending", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != 4 {
		t.Errorf("expected ID 4, got %d", task.ID)
	}
	if task.Title != "New Task" {
		t.Errorf("expected New Task, got %s", task.Title)
	}
	// Verify it appears in GetTasks
	tasks := ds.GetTasks("", "")
	if len(tasks) != 4 {
		t.Errorf("expected 4 tasks after create, got %d", len(tasks))
	}
}

func TestDataStore_CreateTask_EmptyTitle(t *testing.T) {
	ds := newTestStore()
	_, err := ds.CreateTask("", "pending", 1)
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestDataStore_CreateTask_InvalidStatus(t *testing.T) {
	ds := newTestStore()
	_, err := ds.CreateTask("Task", "invalid-status", 1)
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestDataStore_CreateTask_InvalidUserID(t *testing.T) {
	ds := newTestStore()
	_, err := ds.CreateTask("Task", "pending", 999)
	if err == nil {
		t.Fatal("expected error for non-existent userId")
	}
}

// ---------------------------------------------------------------------------
// UpdateTask
// ---------------------------------------------------------------------------

func TestDataStore_UpdateTask_Success_AllFields(t *testing.T) {
	ds := newTestStore()
	title := "Updated Title"
	status := "completed"
	userID := 2
	task, err := ds.UpdateTask(1, &title, &status, &userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "Updated Title" {
		t.Errorf("expected Updated Title, got %s", task.Title)
	}
	if task.Status != "completed" {
		t.Errorf("expected completed, got %s", task.Status)
	}
	if task.UserID != 2 {
		t.Errorf("expected userId 2, got %d", task.UserID)
	}
}

func TestDataStore_UpdateTask_PartialUpdate_StatusOnly(t *testing.T) {
	ds := newTestStore()
	status := "completed"
	task, err := ds.UpdateTask(1, nil, &status, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Status != "completed" {
		t.Errorf("expected completed, got %s", task.Status)
	}
	// Title and UserID should remain unchanged
	if task.Title != "Task One" {
		t.Errorf("expected Title unchanged (Task One), got %s", task.Title)
	}
	if task.UserID != 1 {
		t.Errorf("expected UserID unchanged (1), got %d", task.UserID)
	}
}

func TestDataStore_UpdateTask_NotFound(t *testing.T) {
	ds := newTestStore()
	status := "completed"
	_, err := ds.UpdateTask(999, nil, &status, nil)
	if err == nil {
		t.Fatal("expected error for non-existent task")
	}
}

func TestDataStore_UpdateTask_InvalidStatus(t *testing.T) {
	ds := newTestStore()
	status := "bad-status"
	_, err := ds.UpdateTask(1, nil, &status, nil)
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestDataStore_UpdateTask_InvalidUserID(t *testing.T) {
	ds := newTestStore()
	uid := 999
	_, err := ds.UpdateTask(1, nil, nil, &uid)
	if err == nil {
		t.Fatal("expected error for non-existent userId")
	}
}

// ---------------------------------------------------------------------------
// Concurrent access
// ---------------------------------------------------------------------------

func TestDataStore_ConcurrentAccess(t *testing.T) {
	ds := newTestStore()
	var wg sync.WaitGroup

	// Spawn concurrent reads and writes
	for i := 0; i < 50; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			ds.GetUsers()
		}()
		go func() {
			defer wg.Done()
			ds.GetTasks("", "")
		}()
		go func(n int) {
			defer wg.Done()
			ds.CreateUser("User"+string(rune('A'+n%26)), "concurrent"+string(rune('a'+n%26))+"@test.com", "dev")
		}(i)
	}
	wg.Wait()

	// Just verify no panics occurred and data is consistent
	users := ds.GetUsers()
	if len(users) < 2 {
		t.Errorf("expected at least 2 users, got %d", len(users))
	}
}
