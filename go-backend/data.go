package main

import (
	"strconv"
	"sync"
)

// DataStore holds all application data
type DataStore struct {
	mu    sync.RWMutex
	users []User
	tasks []Task
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
