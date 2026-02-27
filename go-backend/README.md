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

| Environment Variable | Default      | Description                                      |
|---------------------|--------------|--------------------------------------------------|
| `PORT`              | `8080`       | HTTP server port                                 |
| `STORAGE_BACKEND`   | *(empty)*    | Set to `sqlite` to use SQLite; otherwise in-memory + file |
| `DATA_FILE`         | `data.json`  | Path to persistence JSON file (in-memory mode)   |
| `DB_PATH`           | `data.db`    | Path to SQLite database file (sqlite mode)       |
| `API_KEY`           | *(empty)*    | API key for authentication (disabled when empty)  |

To run with SQLite:
```bash
STORAGE_BACKEND=sqlite go run .
```

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

### Cache

| Method | Path               | Description                          |
|--------|--------------------|--------------------------------------|
| GET    | `/api/cache/stats` | Cache hit/miss stats, entry count    |

### Metrics / Observability

| Method | Path       | Description                                              |
|--------|------------|----------------------------------------------------------|
| GET    | `/metrics` | Request counts by status/method/path, error rate, uptime |

## Features

- **Request Logging**: All requests logged with method, path, status code, and duration
- **CORS**: Full CORS support including preflight OPTIONS handling
- **File Persistence**: Data saved to `data.json` on every mutation; loaded on startup. Atomic writes via temp file + rename. Corrupted files handled gracefully.
- **Thread Safety**: All data access protected by `sync.RWMutex`
- **Consistent Error Responses**: All errors returned as `{"error": "message"}` JSON
- **Enhanced Health Check**: Includes version, uptime, and data store status
- **Response Caching**: GET responses for users and tasks are cached with a 5-minute TTL. Cache is automatically invalidated on POST/PUT mutations. Stats available at `/api/cache/stats`.
- **Request Validation Middleware**: Enforces 1MB max request body size. Rejects non-JSON Content-Type on mutation methods (POST/PUT/PATCH).
- **API Key Authentication**: Set `API_KEY` env var to enable. Requires `X-API-Key` header on all requests except exempt paths (`/health`, `/metrics`). OPTIONS preflight requests are always allowed.
- **Rate Limiting**: 100 requests per minute per IP. Includes `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Window` response headers. Returns `429 Too Many Requests` with `Retry-After` header when exceeded. Supports `X-Forwarded-For` and `X-Real-Ip` for proxy setups.
- **Metrics / Observability**: Tracks total requests, errors, error rate, average duration, and breakdowns by HTTP status class, method, and path. Available at `/metrics`.
- **SQLite Database Backend**: Set `STORAGE_BACKEND=sqlite` to replace in-memory + file storage with a SQLite database. Uses WAL mode for read concurrency, foreign keys, auto-migrations, connection pooling, and idempotent seeding. Pure Go driver (`modernc.org/sqlite`) — no CGO required.

## Testing

```bash
go test ./...            # Run all tests
go test -v ./...         # Verbose output
go test -cover ./...     # With coverage report
```

Test files:
- `data_test.go` — Unit tests for DataStore operations (CRUD, filtering, concurrency)
- `server_test.go` — Integration tests for all HTTP endpoints, cache integration, metrics endpoint, full middleware chain
- `persistence_test.go` — File persistence tests (save/load, corruption, concurrency)
- `cache_test.go` — Cache set/get, expiration, invalidation, stats
- `middleware_test.go` — Request validation, auth (enabled/disabled/exempt/invalid), rate limiting, IP extraction
- `metrics_test.go` — Recording, status/method/path breakdowns, edge cases
- `sqlite_store_test.go` — SQLite CRUD, migrations, seeding, persistence across connections, interface compliance

Current coverage: **82%+** across 123 tests.

## Design Decisions

- **Minimal dependencies**: Core uses only stdlib (`net/http`, `encoding/json`, `sync`, `log`). SQLite backend adds `modernc.org/sqlite` (pure Go, no CGO)
- **Async persistence**: `onChanged` callback fires saves in a goroutine to avoid blocking request handlers
- **Atomic file writes**: Write to `.tmp` then rename to prevent corruption on crash
- **Pointer fields for partial updates**: `TaskUpdateRequest` uses `*string`/`*int` to distinguish "not provided" from zero values
- **Composable middleware chain**: Auth → Rate Limit → Validation → Logging → Metrics → Handler. Each layer is independently testable and optional (nil-safe).
- **Cache invalidation on writes**: All mutations clear the entire cache to guarantee freshness. Simple and correct over complex partial invalidation.
- **Auth disabled by default**: Setting `API_KEY=""` (the default) disables auth entirely, keeping local development frictionless while production can require keys.
- **Per-IP rate limiting with sliding window**: Each IP gets an independent counter that resets after the window expires. Supports proxied requests via forwarding headers.
- **Store interface**: Both `DataStore` (in-memory) and `SQLiteStore` implement the `Store` interface, making backends interchangeable without changing server code.
- **SQLite WAL mode + single writer**: `MaxOpenConns(1)` serializes writes (SQLite limitation) while WAL allows concurrent reads. Foreign keys enforce referential integrity at the DB level.
