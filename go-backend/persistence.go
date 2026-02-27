package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

const defaultDataFile = "data.json"

// persistedData represents the JSON structure saved to disk.
type persistedData struct {
	Users []User `json:"users"`
	Tasks []Task `json:"tasks"`
}

// FilePersistence handles reading and writing the DataStore to a JSON file.
type FilePersistence struct {
	mu       sync.Mutex
	filePath string
}

// NewFilePersistence creates a new FilePersistence for the given file path.
func NewFilePersistence(filePath string) *FilePersistence {
	return &FilePersistence{filePath: filePath}
}

// Load reads the JSON file and populates the DataStore. If the file does not
// exist, the store is left unchanged (uses default seed data). Corrupted files
// are logged and skipped so the server can still start.
func (fp *FilePersistence) Load(ds *DataStore) error {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	data, err := os.ReadFile(fp.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("No data file found at %s, using default data", fp.filePath)
			return nil
		}
		return fmt.Errorf("reading data file: %w", err)
	}

	var pd persistedData
	if err := json.Unmarshal(data, &pd); err != nil {
		log.Printf("Warning: corrupted data file %s: %v — using default data", fp.filePath, err)
		return nil
	}

	ds.mu.Lock()
	defer ds.mu.Unlock()
	if pd.Users != nil {
		ds.users = pd.Users
	}
	if pd.Tasks != nil {
		ds.tasks = pd.Tasks
	}

	log.Printf("Loaded %d users and %d tasks from %s", len(ds.users), len(ds.tasks), fp.filePath)
	return nil
}

// Save writes the current DataStore contents to the JSON file atomically by
// writing to a temp file first and then renaming.
func (fp *FilePersistence) Save(ds *DataStore) error {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	ds.mu.RLock()
	pd := persistedData{
		Users: ds.users,
		Tasks: ds.tasks,
	}
	ds.mu.RUnlock()

	data, err := json.MarshalIndent(pd, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling data: %w", err)
	}

	tmpFile := fp.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := os.Rename(tmpFile, fp.filePath); err != nil {
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}
