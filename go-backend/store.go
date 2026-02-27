package main

// Store defines the interface for data storage backends.
// Both DataStore (in-memory) and SQLiteStore implement this interface.
type Store interface {
	GetUsers() []User
	GetUserByID(id int) *User
	GetTasks(status, userID string) []Task
	GetTaskByID(id int) *Task
	GetStats() StatsResponse
	CreateUser(name, email, role string) (User, error)
	CreateTask(title, status string, userID int) (Task, error)
	UpdateTask(id int, title, status *string, userID *int) (Task, error)
}
