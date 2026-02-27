# Go Backend

HTTP server providing the data layer for the three-tier architecture (React → Node.js → **Go**).

## Setup

1. Go 1.21+ required
2. Install dependencies and run:
```bash
go mod tidy
go run .
```
Or build and run:
```bash
go build -o go-backend
./go-backend
```

The server starts on `http://localhost:8080` by default.

## Configuration

| Environment Variable | Default      | Description                     |
|---------------------|--------------|---------------------------------|
| `PORT`              | `8080`       | HTTP server port                |
| `DATA_FILE`         | `data.json`  | Path to persistence JSON file   |

## API Endpoints

### Health

| Method | Path      | Description                                      |
|--------|-----------|--------------------------------------------------|
| GET    | `/health` | Health check with version, uptime, and data store status |

### Users

| Method | Path             | Description          |
|--------|------------------|----------------------|
| GET    | `/api/users`     | List all users       |
| GET    | `/api/users/:id` | Get user by ID       |
| POST   | `/api/users`     | Create a new user    |

**POST /api/users** — JSON body:
```json
{"name": "Alice", "email": "alice@example.com", "role": "developer"}
```
- All fields required, email format validated
- Returns `201` with created user, or `400` on validation error

### Tasks

| Method    | Path             | Description                             |
|-----------|------------------|-----------------------------------------|
| GET       | `/api/tasks`     | List tasks (supports `?status=` and `?userId=` query params) |
| GET       | `/api/tasks/:id` | Get task by ID                          |
| POST      | `/api/tasks`     | Create a new task                       |
| PUT/PATCH | `/api/tasks/:id` | Update an existing task (partial update)|

**POST /api/tasks** — JSON body:
```json
{"title": "Fix bug", "status": "pending", "userId": 1}
```
- `status` must be one of: `pending`, `in-progress`, `completed`
- `userId` must reference an existing user
- Returns `201` with created task, or `400` on validation error

**PUT /api/tasks/:id** — JSON body (all fields optional):
```json
{"title": "Updated title", "status": "completed", "userId": 2}
```
- Only provided fields are updated (partial update)
- Returns `200` with updated task, `404` if not found, or `400` on validation error

### Statistics

| Method | Path         | Description                        |
|--------|--------------|------------------------------------|
| GET    | `/api/stats` | Aggregate counts for users & tasks |

## Features

- **Request Logging**: All requests logged with method, path, status code, and duration
- **CORS**: Full CORS support including preflight OPTIONS handling
- **File Persistence**: Data saved to `data.json` on every mutation; loaded on startup. Atomic writes via temp file + rename. Corrupted files handled gracefully.
- **Thread Safety**: All data access protected by `sync.RWMutex`
- **Consistent Error Responses**: All errors returned as `{"error": "message"}` JSON
- **Enhanced Health Check**: Includes version, uptime, and data store status

## Testing

```bash
go test ./...            # Run all tests
go test -v ./...         # Verbose output
go test -cover ./...     # With coverage report
```

Test files:
- `data_test.go` — Unit tests for DataStore operations (CRUD, filtering, concurrency)
- `server_test.go` — Integration tests for all HTTP endpoints
- `persistence_test.go` — File persistence tests (save/load, corruption, concurrency)

Current coverage: **85%+** across 60 tests.

## Design Decisions

- **Standard library only**: No external dependencies — uses `net/http`, `encoding/json`, `sync`, `log`
- **Async persistence**: `onChanged` callback fires saves in a goroutine to avoid blocking request handlers
- **Atomic file writes**: Write to `.tmp` then rename to prevent corruption on crash
- **Pointer fields for partial updates**: `TaskUpdateRequest` uses `*string`/`*int` to distinguish "not provided" from zero values
