package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	_ "modernc.org/sqlite"
)

const defaultDBPath = "data.db"

// SQLiteStore implements the Store interface backed by a SQLite database.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens (or creates) a SQLite database at the given path,
// runs migrations, and returns a ready-to-use store.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(1) // SQLite does not support concurrent writes
	db.SetMaxIdleConns(1)

	// Enable WAL mode for better read concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting WAL mode: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return s, nil
}

// migrate creates the schema if it does not already exist.
func (s *SQLiteStore) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id    INTEGER PRIMARY KEY AUTOINCREMENT,
			name  TEXT NOT NULL,
			email TEXT NOT NULL,
			role  TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id      INTEGER PRIMARY KEY AUTOINCREMENT,
			title   TEXT NOT NULL,
			status  TEXT NOT NULL DEFAULT 'pending',
			user_id INTEGER NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
	}

	for _, m := range migrations {
		if _, err := s.db.Exec(m); err != nil {
			return fmt.Errorf("executing migration: %w", err)
		}
	}

	log.Printf("SQLite migrations applied successfully")
	return nil
}

// Seed inserts initial data only if the tables are empty.
func (s *SQLiteStore) Seed() error {
	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil // already seeded
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	users := []User{
		{Name: "John Doe", Email: "john@example.com", Role: "developer"},
		{Name: "Jane Smith", Email: "jane@example.com", Role: "designer"},
		{Name: "Bob Johnson", Email: "bob@example.com", Role: "manager"},
	}
	for _, u := range users {
		if _, err := tx.Exec("INSERT INTO users (name, email, role) VALUES (?, ?, ?)",
			u.Name, u.Email, u.Role); err != nil {
			return err
		}
	}

	tasks := []struct {
		Title  string
		Status string
		UserID int
	}{
		{"Implement authentication", "pending", 1},
		{"Design user interface", "in-progress", 2},
		{"Review code changes", "completed", 3},
	}
	for _, t := range tasks {
		if _, err := tx.Exec("INSERT INTO tasks (title, status, user_id) VALUES (?, ?, ?)",
			t.Title, t.Status, t.UserID); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Seeded database with %d users and %d tasks", len(users), len(tasks))
	return nil
}

// Close closes the underlying database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// ---------------------------------------------------------------------------
// Store interface implementation
// ---------------------------------------------------------------------------

func (s *SQLiteStore) GetUsers() []User {
	rows, err := s.db.Query("SELECT id, name, email, role FROM users ORDER BY id")
	if err != nil {
		log.Printf("Error querying users: %v", err)
		return nil
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role); err != nil {
			log.Printf("Error scanning user: %v", err)
			continue
		}
		users = append(users, u)
	}
	return users
}

func (s *SQLiteStore) GetUserByID(id int) *User {
	var u User
	err := s.db.QueryRow("SELECT id, name, email, role FROM users WHERE id = ?", id).
		Scan(&u.ID, &u.Name, &u.Email, &u.Role)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error querying user %d: %v", id, err)
		}
		return nil
	}
	return &u
}

func (s *SQLiteStore) GetTasks(status, userID string) []Task {
	query := "SELECT id, title, status, user_id FROM tasks WHERE 1=1"
	var args []interface{}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if userID != "" {
		if id, err := strconv.Atoi(userID); err == nil {
			query += " AND user_id = ?"
			args = append(args, id)
		} else {
			return nil // invalid userID
		}
	}

	query += " ORDER BY id"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		log.Printf("Error querying tasks: %v", err)
		return nil
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Status, &t.UserID); err != nil {
			log.Printf("Error scanning task: %v", err)
			continue
		}
		tasks = append(tasks, t)
	}
	return tasks
}

func (s *SQLiteStore) GetTaskByID(id int) *Task {
	var t Task
	err := s.db.QueryRow("SELECT id, title, status, user_id FROM tasks WHERE id = ?", id).
		Scan(&t.ID, &t.Title, &t.Status, &t.UserID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error querying task %d: %v", id, err)
		}
		return nil
	}
	return &t
}

func (s *SQLiteStore) GetStats() StatsResponse {
	var stats StatsResponse

	s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.Users.Total)
	s.db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&stats.Tasks.Total)
	s.db.QueryRow("SELECT COUNT(*) FROM tasks WHERE status = 'pending'").Scan(&stats.Tasks.Pending)
	s.db.QueryRow("SELECT COUNT(*) FROM tasks WHERE status = 'in-progress'").Scan(&stats.Tasks.InProgress)
	s.db.QueryRow("SELECT COUNT(*) FROM tasks WHERE status = 'completed'").Scan(&stats.Tasks.Completed)

	return stats
}

func (s *SQLiteStore) CreateUser(name, email, role string) (User, error) {
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

	result, err := s.db.Exec("INSERT INTO users (name, email, role) VALUES (?, ?, ?)",
		name, email, role)
	if err != nil {
		return User{}, fmt.Errorf("inserting user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return User{}, fmt.Errorf("getting last insert id: %w", err)
	}

	return User{ID: int(id), Name: name, Email: email, Role: role}, nil
}

func (s *SQLiteStore) CreateTask(title, status string, userID int) (Task, error) {
	if title == "" {
		return Task{}, fmt.Errorf("title is required")
	}
	if !validStatuses[status] {
		return Task{}, fmt.Errorf("invalid status: must be one of pending, in-progress, completed")
	}

	// Validate userID exists
	var exists int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", userID).Scan(&exists); err != nil {
		return Task{}, fmt.Errorf("checking user: %w", err)
	}
	if exists == 0 {
		return Task{}, fmt.Errorf("userId %d does not exist", userID)
	}

	result, err := s.db.Exec("INSERT INTO tasks (title, status, user_id) VALUES (?, ?, ?)",
		title, status, userID)
	if err != nil {
		return Task{}, fmt.Errorf("inserting task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Task{}, fmt.Errorf("getting last insert id: %w", err)
	}

	return Task{ID: int(id), Title: title, Status: status, UserID: userID}, nil
}

func (s *SQLiteStore) UpdateTask(id int, title, status *string, userID *int) (Task, error) {
	// Check task exists
	existing := s.GetTaskByID(id)
	if existing == nil {
		return Task{}, fmt.Errorf("task not found")
	}

	if status != nil {
		if !validStatuses[*status] {
			return Task{}, fmt.Errorf("invalid status: must be one of pending, in-progress, completed")
		}
	}

	if userID != nil {
		var exists int
		if err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", *userID).Scan(&exists); err != nil {
			return Task{}, fmt.Errorf("checking user: %w", err)
		}
		if exists == 0 {
			return Task{}, fmt.Errorf("userId %d does not exist", *userID)
		}
	}

	// Apply updates
	if title != nil {
		existing.Title = *title
	}
	if status != nil {
		existing.Status = *status
	}
	if userID != nil {
		existing.UserID = *userID
	}

	_, err := s.db.Exec("UPDATE tasks SET title = ?, status = ?, user_id = ? WHERE id = ?",
		existing.Title, existing.Status, existing.UserID, id)
	if err != nil {
		return Task{}, fmt.Errorf("updating task: %w", err)
	}

	return *existing, nil
}
