package main

import (
	"log"
	"os"
)

const (
	defaultPort = "8080"
	appVersion  = "1.0.0"
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
	Status    string              `json:"status"`
	Message   string              `json:"message"`
	Version   string              `json:"version,omitempty"`
	Uptime    string              `json:"uptime,omitempty"`
	DataStore *DataStoreHealth    `json:"dataStore,omitempty"`
}

type DataStoreHealth struct {
	Status     string `json:"status"`
	UserCount  int    `json:"userCount"`
	TaskCount  int    `json:"taskCount"`
}

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Set up file-based persistence
	dataFile := os.Getenv("DATA_FILE")
	if dataFile == "" {
		dataFile = defaultDataFile
	}
	persistence := NewFilePersistence(dataFile)
	if err := persistence.Load(store); err != nil {
		log.Printf("Warning: could not load data file: %v", err)
	}
	store.onChanged = func() {
		if err := persistence.Save(store); err != nil {
			log.Printf("Error saving data: %v", err)
		}
	}

	server := NewServer(store)
	server.Start(port)
}
