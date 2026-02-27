# Go Backend Client

This Go application consumes the Node.js backend API and demonstrates various API interactions.

## Setup

1. Make sure Go is installed (version 1.21 or higher)

2. Install dependencies:
```bash
go mod tidy
```

3. Run the application:
```bash
go run main.go
```

Or build and run:
```bash
go build -o go-backend
./go-backend
```

## Configuration

The Go backend connects to the Node.js backend at `http://localhost:3000` by default.

You can override this by setting the `NODE_BACKEND_URL` environment variable:
```bash
NODE_BACKEND_URL=http://localhost:3000 go run main.go
```

## Features

The Go application demonstrates:
- HTTP client implementation with timeout
- JSON request/response handling
- Error handling
- API endpoint consumption
- Structured data types
- Health check verification

## What It Does

The application:
1. Checks the health of the Node.js backend
2. Fetches all users and displays them
3. Fetches a specific user by ID
4. Fetches all tasks
5. Fetches tasks filtered by status
6. Retrieves and displays statistics

## Testing

Make sure the Node.js backend is running before executing the Go application.

```bash
# Terminal 1: Start Node.js backend
cd node-backend
npm start

# Terminal 2: Run Go application
cd go-backend
go run main.go
```
