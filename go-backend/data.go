package main

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"
)

var validStatuses = map[string]bool{
	"pending":     true,
	"in-progress": true,
	"completed":   true,
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// DataStore holds all application data
type DataStore struct {
	mu        sync.RWMutex
	users     []User
	tasks     []Task
	onChanged func() // called after data mutations, may be nil
}

var store = &DataStore{
	users: []User{
		{ID: 1, Name: "John Doe", Email: "john@example.com", Role: "developer"},
		{ID: 2, Name: "Jane Smith", Email: "jane@example.com", Role: "designer"},
		{ID: 3, Name: "Bob Johnson", Email: "bob@example.com", Role: "manager"},
	},
	tasks: []Task{
		{ID: 1, Title: "Implement authentication", Status: "pending", UserID: 1},
		{ID: 2, Title: "Design user interface", Status: "in-progress", UserID: 2},
		{ID: 3, Title: "Review code changes", Status: "completed", UserID: 3},
	},
}

func (ds *DataStore) GetUsers() []User {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.users
}

func (ds *DataStore) GetUserByID(id int) *User {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	for i := range ds.users {
		if ds.users[i].ID == id {
			return &ds.users[i]
		}
	}
	return nil
}

func (ds *DataStore) GetTasks(status, userID string) []Task {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	
	var filtered []Task
	for _, task := range ds.tasks {
		matchStatus := status == "" || task.Status == status
		
		matchUserID := true
		if userID != "" {
			if id, err := strconv.Atoi(userID); err == nil {
				matchUserID = task.UserID == id
			} else {
				matchUserID = false
			}
		}
		
		if matchStatus && matchUserID {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// GetTaskByID returns a pointer to a task with the given ID, or nil if not found.
func (ds *DataStore) GetTaskByID(id int) *Task {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	for i := range ds.tasks {
		if ds.tasks[i].ID == id {
			return &ds.tasks[i]
		}
	}
	return nil
}

// CreateUser validates and adds a new user to the store. Returns the created user or an error.
func (ds *DataStore) CreateUser(name, email, role string) (User, error) {
	if name == "" {
		return User{}, fmt.Errorf("name is required")
	}
	if email == "" {
		return User{}, fmt.Errorf("email is required")
	}
	if !emailRegex.MatchString(email) {
		return User{}, fmt.Errorf("invalid email format")
	}
	if role == "" {
		return User{}, fmt.Errorf("role is required")
	}

	ds.mu.Lock()
	defer ds.mu.Unlock()

	maxID := 0
	for _, u := range ds.users {
		if u.ID > maxID {
			maxID = u.ID
		}
	}

	user := User{
		ID:    maxID + 1,
		Name:  name,
		Email: email,
		Role:  role,
	}
	ds.users = append(ds.users, user)
	if ds.onChanged != nil {
		go ds.onChanged()
	}
	return user, nil
}

// CreateTask validates and adds a new task to the store. Returns the created task or an error.
func (ds *DataStore) CreateTask(title, status string, userID int) (Task, error) {
	if title == "" {
		return Task{}, fmt.Errorf("title is required")
	}
	if !validStatuses[status] {
		return Task{}, fmt.Errorf("invalid status: must be one of pending, in-progress, completed")
	}

	// Validate userId exists – need read lock for users, but we'll acquire full lock
	ds.mu.Lock()
	defer ds.mu.Unlock()

	userFound := false
	for _, u := range ds.users {
		if u.ID == userID {
			userFound = true
			break
		}
	}
	if !userFound {
		return Task{}, fmt.Errorf("userId %d does not exist", userID)
	}

	maxID := 0
	for _, t := range ds.tasks {
		if t.ID > maxID {
			maxID = t.ID
		}
	}

	task := Task{
		ID:     maxID + 1,
		Title:  title,
		Status: status,
		UserID: userID,
	}
	ds.tasks = append(ds.tasks, task)
	if ds.onChanged != nil {
		go ds.onChanged()
	}
	return task, nil
}

// UpdateTask applies partial updates to an existing task. Returns the updated task or an error.
func (ds *DataStore) UpdateTask(id int, title, status *string, userID *int) (Task, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	var taskIdx int = -1
	for i := range ds.tasks {
		if ds.tasks[i].ID == id {
			taskIdx = i
			break
		}
	}
	if taskIdx == -1 {
		return Task{}, fmt.Errorf("task not found")
	}

	if status != nil {
		if !validStatuses[*status] {
			return Task{}, fmt.Errorf("invalid status: must be one of pending, in-progress, completed")
		}
	}

	if userID != nil {
		userFound := false
		for _, u := range ds.users {
			if u.ID == *userID {
				userFound = true
				break
			}
		}
		if !userFound {
			return Task{}, fmt.Errorf("userId %d does not exist", *userID)
		}
	}

	if title != nil {
		ds.tasks[taskIdx].Title = *title
	}
	if status != nil {
		ds.tasks[taskIdx].Status = *status
	}
	if userID != nil {
		ds.tasks[taskIdx].UserID = *userID
	}

	updated := ds.tasks[taskIdx]
	if ds.onChanged != nil {
		go ds.onChanged()
	}
	return updated, nil
}

func (ds *DataStore) GetStats() StatsResponse {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	
	var stats StatsResponse
	stats.Users.Total = len(ds.users)
	stats.Tasks.Total = len(ds.tasks)
	
	for _, task := range ds.tasks {
		switch task.Status {
		case "pending":
			stats.Tasks.Pending++
		case "in-progress":
			stats.Tasks.InProgress++
		case "completed":
			stats.Tasks.Completed++
		}
	}
	
	return stats
}
