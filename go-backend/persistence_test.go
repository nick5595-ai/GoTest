package main

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestFilePersistence_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	fp := NewFilePersistence(filepath.Join(dir, "test_data.json"))

	// Create a store with data and save it
	ds := &DataStore{
		users: []User{
			{ID: 1, Name: "Alice", Email: "alice@test.com", Role: "dev"},
		},
		tasks: []Task{
			{ID: 1, Title: "Task A", Status: "pending", UserID: 1},
		},
	}

	if err := fp.Save(ds); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load into a new empty store
	ds2 := &DataStore{}
	if err := fp.Load(ds2); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(ds2.users) != 1 || ds2.users[0].Name != "Alice" {
		t.Errorf("expected 1 user (Alice), got %+v", ds2.users)
	}
	if len(ds2.tasks) != 1 || ds2.tasks[0].Title != "Task A" {
		t.Errorf("expected 1 task (Task A), got %+v", ds2.tasks)
	}
}

func TestFilePersistence_Load_FileNotExist(t *testing.T) {
	dir := t.TempDir()
	fp := NewFilePersistence(filepath.Join(dir, "nonexistent.json"))

	ds := &DataStore{
		users: []User{{ID: 1, Name: "Default", Email: "d@d.com", Role: "dev"}},
	}
	if err := fp.Load(ds); err != nil {
		t.Fatalf("Load should not fail for missing file: %v", err)
	}
	// Original data should remain
	if len(ds.users) != 1 || ds.users[0].Name != "Default" {
		t.Errorf("expected default data to remain, got %+v", ds.users)
	}
}

func TestFilePersistence_Load_CorruptedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "corrupt.json")
	os.WriteFile(path, []byte("{this is not valid json}"), 0644)

	fp := NewFilePersistence(path)
	ds := &DataStore{
		users: []User{{ID: 1, Name: "Default", Email: "d@d.com", Role: "dev"}},
	}
	if err := fp.Load(ds); err != nil {
		t.Fatalf("Load should not fail for corrupted file: %v", err)
	}
	// Original data should remain
	if len(ds.users) != 1 || ds.users[0].Name != "Default" {
		t.Errorf("expected default data to remain after corrupt file, got %+v", ds.users)
	}
}

func TestFilePersistence_DataPersistsAcrossRestarts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "persist.json")
	fp := NewFilePersistence(path)

	// Simulate first "server run": create a user and save explicitly
	ds1 := &DataStore{
		users: []User{{ID: 1, Name: "Alice", Email: "a@a.com", Role: "dev"}},
		tasks: []Task{},
	}
	if _, err := ds1.CreateUser("Bob", "bob@test.com", "designer"); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if err := fp.Save(ds1); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Simulate second "server run": load from file
	ds2 := &DataStore{}
	if err := fp.Load(ds2); err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(ds2.users) != 2 {
		t.Fatalf("expected 2 users after reload, got %d", len(ds2.users))
	}
	found := false
	for _, u := range ds2.users {
		if u.Name == "Bob" {
			found = true
		}
	}
	if !found {
		t.Error("Bob not found after reload")
	}
}

func TestFilePersistence_ConcurrentSaves(t *testing.T) {
	dir := t.TempDir()
	fp := NewFilePersistence(filepath.Join(dir, "concurrent.json"))

	ds := &DataStore{
		users: []User{{ID: 1, Name: "Alice", Email: "a@a.com", Role: "dev"}},
		tasks: []Task{},
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fp.Save(ds)
		}()
	}
	wg.Wait()

	// Verify file is valid by loading
	ds2 := &DataStore{}
	if err := fp.Load(ds2); err != nil {
		t.Fatalf("Load after concurrent saves failed: %v", err)
	}
	if len(ds2.users) != 1 {
		t.Errorf("expected 1 user, got %d", len(ds2.users))
	}
}
