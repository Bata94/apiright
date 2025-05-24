# APIRight Framework Usage Guide

## Overview

APIRight is a Go framework that automatically converts SQLC-generated structs into ready-to-use CRUD APIs. It provides a seamless transition layer between database models and API endpoints.

## Development Environment Setup

This project uses Nix for reproducible development environments. To get started:

```bash
# Enter the development shell
nix develop

# This provides:
# - Go 1.23.9
# - SQLC v1.29.0
# - Air (live reload)
# - golangci-lint
# - PostgreSQL
# - SQLite
```

## Quick Start

### 1. Basic Setup

```go
package main

import (
    "github.com/bata94/apiright"
)

// Your SQLC-generated struct
type User struct {
    ID    int    `json:"id" db:"id"`
    Name  string `json:"name" db:"name"`
    Email string `json:"email" db:"email"`
}

func main() {
    // Create APIRight app
    app := apiright.New(&apiright.Config{
        Port:     "8080",
        Database: "sqlite3",
        DSN:      "./app.db",
    })

    // Add middleware
    app.Use(apiright.CORSMiddleware())
    app.Use(apiright.LoggingMiddleware())

    // Register CRUD endpoints for User
    app.RegisterCRUD("/users", User{})

    // Start server
    app.Start()
}
```

### 2. Available Endpoints

Once you register a CRUD resource, the following endpoints are automatically available:

- `GET /users` - List all users
- `GET /users/{id}` - Get user by ID
- `POST /users` - Create new user
- `PUT /users/{id}` - Update user
- `DELETE /users/{id}` - Delete user

### 3. With SQLC Integration

```go
// sqlc generated code
type User struct {
    ID        int32     `json:"id" db:"id"`
    Username  string    `json:"username" db:"username"`
    Email     string    `json:"email" db:"email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Your API model (optional transformation)
type UserAPI struct {
    ID       int32  `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
}

func main() {
    app := apiright.New(&apiright.Config{
        Port:     "8080",
        Database: "postgres",
        DSN:      "postgres://user:pass@localhost/db?sslmode=disable",
    })

    // Register with transformation
    app.RegisterCRUDWithTransform("/users", User{}, UserAPI{})

    app.Start()
}
```

## Features

### 1. Automatic CRUD Generation
- Generates REST endpoints automatically
- Supports pagination, filtering, and sorting
- Handles JSON serialization/deserialization

### 2. Database Support
- PostgreSQL (via lib/pq)
- SQLite (via mattn/go-sqlite3)
- Extensible for other databases

### 3. Middleware Support
- CORS handling
- Request logging
- Authentication/Authorization
- Custom middleware support

### 4. Model Transformation
- Convert between DB and API models
- Field mapping and validation
- Custom transformation logic

### 5. Type Safety
- Leverages Go's type system
- Works seamlessly with SQLC-generated code
- Compile-time safety

## Building and Running

### Using Nix (Recommended)

```bash
# Enter development environment
nix develop

# Build the project
go build ./examples/basic

# Run with live reload
air

# Run tests
go test ./...

# Lint code
golangci-lint run
```

### Traditional Go

```bash
# Install dependencies
go mod download

# Build
go build ./examples/basic

# Run
./basic
```

## Example Project Structure

```
your-project/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── models/          # SQLC generated models
│   ├── queries/         # SQLC generated queries
│   └── api/            # API models and transformations
├── migrations/
├── sqlc.yaml
└── go.mod
```

## Configuration Options

```go
type Config struct {
    Port     string // Server port (default: "8080")
    Database string // Database type: "postgres", "sqlite3"
    DSN      string // Database connection string
    Debug    bool   // Enable debug logging
}
```

## Middleware

### Built-in Middleware

```go
// CORS support
app.Use(apiright.CORSMiddleware())

// Request logging
app.Use(apiright.LoggingMiddleware())

// JSON validation
app.Use(apiright.JSONValidationMiddleware())

// Basic auth
app.Use(apiright.BasicAuthMiddleware("user", "pass"))
```

### Custom Middleware

```go
func CustomMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Your middleware logic
            next.ServeHTTP(w, r)
        })
    }
}

app.Use(CustomMiddleware())
```

## Advanced Usage

### Custom Endpoints

```go
// Add custom endpoints alongside CRUD
app.Router().HandleFunc("/users/{id}/profile", getUserProfile).Methods("GET")
app.Router().HandleFunc("/users/search", searchUsers).Methods("GET")
```

### Database Transactions

```go
// The framework supports database transactions
// Custom handlers can access the database connection
func customHandler(w http.ResponseWriter, r *http.Request) {
    // Access database through context or dependency injection
}
```

### Validation and Hooks

```go
// Pre/post processing hooks
app.RegisterCRUDWithHooks("/users", User{}, &apiright.CRUDHooks{
    BeforeCreate: validateUser,
    AfterCreate:  sendWelcomeEmail,
    BeforeUpdate: checkPermissions,
})
```

## Testing

The framework includes comprehensive testing utilities:

```go
func TestUserAPI(t *testing.T) {
    app := apiright.NewTestApp()
    app.RegisterCRUD("/users", User{})
    
    // Test endpoints
    resp := app.TestRequest("GET", "/users", nil)
    assert.Equal(t, 200, resp.StatusCode)
}
```

## Performance Considerations

- Uses connection pooling for database connections
- Supports caching middleware
- Optimized JSON serialization
- Minimal memory allocations in hot paths

## Security

- Built-in CORS protection
- SQL injection prevention through prepared statements
- Input validation and sanitization
- Authentication middleware support

## Deployment

### Docker

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
CMD ["./server"]
```

### Nix

```bash
# Build with Nix
nix build

# Run
./result/bin/server
```

## Contributing

1. Use the Nix development environment
2. Follow Go best practices
3. Add tests for new features
4. Run linting before submitting PRs

## License

MIT License - see LICENSE file for details.