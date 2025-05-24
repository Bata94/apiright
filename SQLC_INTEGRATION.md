# SQLC Integration Guide

This guide shows how to integrate APIRight with SQLC-generated code for a complete type-safe API solution.

## Overview

SQLC generates type-safe Go code from SQL queries, and APIRight automatically converts these structs into REST API endpoints. This combination provides:

- Type safety from database to API
- Automatic CRUD operations
- Zero boilerplate for basic operations
- Custom query support through SQLC
- Transformation layer for API customization

## Project Structure

```
your-project/
├── cmd/
│   └── server/
│       └── main.go              # Main application
├── internal/
│   ├── models/                  # SQLC generated models
│   │   ├── db.go
│   │   ├── models.go
│   │   └── queries.sql.go
│   ├── api/                     # API models and transformations
│   │   ├── user.go
│   │   ├── post.go
│   │   └── transforms.go
│   └── database/
│       └── connection.go
├── migrations/
│   ├── 001_initial.sql
│   └── 002_add_posts.sql
├── schema/
│   └── schema.sql               # Database schema
├── queries/
│   └── queries.sql              # SQL queries for SQLC
├── sqlc.yaml                    # SQLC configuration
├── flake.nix                    # Nix development environment
└── go.mod
```

## Step 1: Setup SQLC

### 1.1 Create sqlc.yaml

```yaml
version: "2"
sql:
  - engine: "sqlite"
    queries: "queries/"
    schema: "schema/"
    gen:
      go:
        package: "models"
        out: "internal/models"
        sql_package: "database/sql"
        emit_json_tags: true
        emit_db_tags: true
        emit_prepared_queries: true
        emit_interface: false
        emit_exact_table_names: false
        emit_empty_slices: true
```

### 1.2 Define Database Schema

```sql
-- schema/schema.sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    published BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### 1.3 Define Queries

```sql
-- queries/queries.sql

-- name: GetUser :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC;

-- name: CreateUser :one
INSERT INTO users (username, email, password_hash)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateUser :one
UPDATE users 
SET username = ?, email = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: GetPost :one
SELECT * FROM posts WHERE id = ? LIMIT 1;

-- name: ListPosts :many
SELECT * FROM posts ORDER BY created_at DESC;

-- name: CreatePost :one
INSERT INTO posts (user_id, title, content, published)
VALUES (?, ?, ?, ?)
RETURNING *;
```

### 1.4 Generate Code

```bash
# In nix development shell
nix develop

# Generate SQLC code
sqlc generate
```

This generates:
- `internal/models/models.go` - Struct definitions
- `internal/models/queries.sql.go` - Query functions
- `internal/models/db.go` - Database interface

## Step 2: Create API Models (Optional)

Create API-specific models for transformation:

```go
// internal/api/user.go
package api

import "time"

// UserResponse represents the API response for a user
type UserResponse struct {
    ID       int32  `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Created  string `json:"created"`
}

// UserRequest represents the API request for creating/updating a user
type UserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password,omitempty" validate:"min=8"`
}

// PostResponse represents the API response for a post
type PostResponse struct {
    ID        int32  `json:"id"`
    UserID    int32  `json:"user_id"`
    Title     string `json:"title"`
    Content   string `json:"content"`
    Published bool   `json:"published"`
    Created   string `json:"created"`
    Author    string `json:"author,omitempty"`
}

// PostRequest represents the API request for creating/updating a post
type PostRequest struct {
    Title     string `json:"title" validate:"required,min=1,max=200"`
    Content   string `json:"content" validate:"required,min=1"`
    Published bool   `json:"published"`
}
```

## Step 3: Create Transformation Functions

```go
// internal/api/transforms.go
package api

import (
    "your-project/internal/models"
    "time"
)

// TransformUser converts a database User to API UserResponse
func TransformUser(user models.User) UserResponse {
    return UserResponse{
        ID:       user.ID,
        Username: user.Username,
        Email:    user.Email,
        Created:  user.CreatedAt.Format("2006-01-02"),
    }
}

// TransformUserRequest converts API UserRequest to database params
func TransformUserRequest(req UserRequest) models.CreateUserParams {
    return models.CreateUserParams{
        Username:     req.Username,
        Email:        req.Email,
        PasswordHash: hashPassword(req.Password), // implement this
    }
}

// TransformPost converts a database Post to API PostResponse
func TransformPost(post models.Post, author *models.User) PostResponse {
    resp := PostResponse{
        ID:        post.ID,
        UserID:    post.UserID,
        Title:     post.Title,
        Content:   post.Content,
        Published: post.Published,
        Created:   post.CreatedAt.Format("2006-01-02"),
    }
    
    if author != nil {
        resp.Author = author.Username
    }
    
    return resp
}

// TransformPostRequest converts API PostRequest to database params
func TransformPostRequest(req PostRequest, userID int32) models.CreatePostParams {
    return models.CreatePostParams{
        UserID:    userID,
        Title:     req.Title,
        Content:   req.Content,
        Published: req.Published,
    }
}
```

## Step 4: Setup APIRight Application

```go
// cmd/server/main.go
package main

import (
    "database/sql"
    "log"
    "net/http"

    "github.com/bata94/apiright"
    "your-project/internal/api"
    "your-project/internal/models"
    
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    // Initialize database
    db, err := sql.Open("sqlite3", "./app.db")
    if err != nil {
        log.Fatal("Failed to open database:", err)
    }
    defer db.Close()

    // Initialize SQLC queries
    queries := models.New(db)

    // Create APIRight app
    app := apiright.New(&apiright.Config{
        Port:     "8080",
        Database: "sqlite3",
        DSN:      "./app.db",
    })

    // Add middleware
    app.Use(apiright.CORSMiddleware())
    app.Use(apiright.LoggingMiddleware())
    app.Use(apiright.JSONValidationMiddleware())

    // Register CRUD endpoints with SQLC models
    app.RegisterCRUD("/users", models.User{})
    app.RegisterCRUD("/posts", models.Post{})

    // Register with transformation layer
    app.RegisterCRUDWithTransform("/api/users", models.User{}, api.UserResponse{})
    app.RegisterCRUDWithTransform("/api/posts", models.Post{}, api.PostResponse{})

    // Add custom endpoints using SQLC queries
    app.Router().HandleFunc("/api/users/{id}/posts", func(w http.ResponseWriter, r *http.Request) {
        // Use SQLC generated queries
        userID := getUserIDFromPath(r) // implement this
        posts, err := queries.ListPostsByUser(r.Context(), userID)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        // Transform and return
        var response []api.PostResponse
        for _, post := range posts {
            response = append(response, api.TransformPost(post, nil))
        }
        
        writeJSONResponse(w, response) // implement this
    }).Methods("GET")

    // Health check
    app.Router().HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"healthy"}`))
    })

    log.Println("🚀 Server starting on port 8080")
    if err := app.Start(); err != nil {
        log.Fatal("Server failed to start:", err)
    }
}
```

## Step 5: Advanced Integration

### 5.1 Custom CRUD Handlers with SQLC

```go
// Custom user handler that uses SQLC queries
func createUserHandler(queries *models.Queries) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req api.UserRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }

        // Validate request
        if err := validateStruct(req); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Use SQLC generated function
        params := api.TransformUserRequest(req)
        user, err := queries.CreateUser(r.Context(), params)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // Transform and return
        response := api.TransformUser(user)
        writeJSONResponse(w, response)
    }
}
```

### 5.2 Pagination and Filtering

```go
// Add pagination support to SQLC queries
-- name: ListUsersWithPagination :many
SELECT * FROM users 
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

// Use in handler
func listUsersHandler(queries *models.Queries) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        page := getPageFromQuery(r) // implement this
        limit := getLimitFromQuery(r) // implement this
        offset := (page - 1) * limit

        users, err := queries.ListUsersWithPagination(r.Context(), models.ListUsersWithPaginationParams{
            Limit:  int64(limit),
            Offset: int64(offset),
        })
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        total, err := queries.CountUsers(r.Context())
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        response := PaginatedResponse{
            Data:  transformUsers(users),
            Page:  page,
            Limit: limit,
            Total: int(total),
        }

        writeJSONResponse(w, response)
    }
}
```

## Step 6: Testing

```go
// tests/integration_test.go
package tests

import (
    "testing"
    "net/http/httptest"
    "your-project/internal/models"
    "github.com/bata94/apiright"
)

func TestUserCRUD(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()

    queries := models.New(db)

    // Create test app
    app := apiright.NewTestApp()
    app.RegisterCRUD("/users", models.User{})

    // Test create user
    req := httptest.NewRequest("POST", "/users", strings.NewReader(`{
        "username": "testuser",
        "email": "test@example.com",
        "password": "password123"
    }`))
    
    resp := app.TestRequest(req)
    assert.Equal(t, 201, resp.StatusCode)

    // Test get user
    req = httptest.NewRequest("GET", "/users/1", nil)
    resp = app.TestRequest(req)
    assert.Equal(t, 200, resp.StatusCode)
}
```

## Step 7: Development Workflow

### 7.1 Using Nix Development Shell

```bash
# Enter development environment
nix develop

# Generate SQLC code after schema changes
sqlc generate

# Run with live reload
air

# Run tests
go test ./...

# Lint code
golangci-lint run
```

### 7.2 Database Migrations

```bash
# Create migration
migrate create -ext sql -dir migrations -seq add_posts_table

# Run migrations
migrate -path migrations -database "sqlite3://app.db" up

# Rollback
migrate -path migrations -database "sqlite3://app.db" down 1
```

## Benefits of This Integration

1. **Type Safety**: End-to-end type safety from database to API
2. **Zero Boilerplate**: Automatic CRUD operations
3. **Performance**: Compiled queries with SQLC
4. **Flexibility**: Custom queries and transformations
5. **Maintainability**: Clear separation of concerns
6. **Testing**: Easy to test with generated interfaces

## Best Practices

1. **Use Transactions**: Wrap related operations in database transactions
2. **Validate Input**: Always validate API requests before database operations
3. **Transform Data**: Use transformation layer to control API responses
4. **Handle Errors**: Provide meaningful error messages
5. **Add Logging**: Log important operations and errors
6. **Use Middleware**: Leverage middleware for cross-cutting concerns
7. **Test Thoroughly**: Write tests for both SQLC queries and API endpoints

This integration provides a powerful, type-safe foundation for building APIs with minimal boilerplate while maintaining full control over database operations and API design.