package main

import (
	"os"
)

const (
	defaultPort = "8080"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type Task struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
	UserID int    `json:"userId"`
}

type UsersResponse struct {
	Users []User `json:"users"`
	Count int    `json:"count"`
}

type TasksResponse struct {
	Tasks []Task `json:"tasks"`
	Count int    `json:"count"`
}

type StatsResponse struct {
	Users struct {
		Total int `json:"total"`
	} `json:"users"`
	Tasks struct {
		Total      int `json:"total"`
		Pending    int `json:"pending"`
		InProgress int `json:"inProgress"`
		Completed  int `json:"completed"`
	} `json:"tasks"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	server := NewServer(store)
	server.Start(port)
}
