package main

import (
	"log"
	"os"
	"time"
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
	Status    string           `json:"status"`
	Message   string           `json:"message"`
	Version   string           `json:"version,omitempty"`
	Uptime    string           `json:"uptime,omitempty"`
	DataStore *DataStoreHealth `json:"dataStore,omitempty"`
}

type DataStoreHealth struct {
	Status    string `json:"status"`
	UserCount int    `json:"userCount"`
	TaskCount int    `json:"taskCount"`
}

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Set up storage backend
	var dataStore Store
	storageBackend := os.Getenv("STORAGE_BACKEND")

	switch storageBackend {
	case "sqlite":
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = defaultDBPath
		}
		sqliteStore, err := NewSQLiteStore(dbPath)
		if err != nil {
			log.Fatalf("Failed to initialize SQLite store: %v", err)
		}
		defer sqliteStore.Close()
		if err := sqliteStore.Seed(); err != nil {
			log.Printf("Warning: failed to seed database: %v", err)
		}
		dataStore = sqliteStore
		log.Printf("Using SQLite storage backend: %s", dbPath)

	default:
		// In-memory store with file-based persistence (original behavior)
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
		dataStore = store
		log.Printf("Using in-memory storage with file persistence: %s", dataFile)
	}

	// Set up cache (5 minute TTL)
	cache := NewCache(5 * time.Minute)

	// Set up metrics
	metrics := NewMetrics()

	// Set up rate limiter (100 requests per minute per IP)
	rateLimiter := NewRateLimiter(RateLimiterConfig{
		Limit:  100,
		Window: 1 * time.Minute,
	})

	// Set up authentication (disabled by default; set API_KEY env var to enable)
	authCfg := AuthConfig{
		APIKey:      os.Getenv("API_KEY"),
		ExemptPaths: []string{"/health", "/metrics"},
	}

	// Set up request validation (1MB max body size)
	validCfg := RequestValidationConfig{
		MaxBodySize: 1 << 20,
	}

	server := NewServer(dataStore, ServerConfig{
		Cache:       cache,
		Metrics:     metrics,
		RateLimiter: rateLimiter,
		AuthCfg:     authCfg,
		ValidCfg:    validCfg,
	})
	server.Start(port)
}
