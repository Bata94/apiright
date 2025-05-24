package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/bata94/apiright"
	_ "github.com/mattn/go-sqlite3"
)

// SQLC-generated User struct (simulated)
type User struct {
	ID        int32     `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// API representation (optional transformation layer)
type UserAPI struct {
	ID       int32  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Created  string `json:"created"`
}

// SQLC-generated Post struct (simulated)
type Post struct {
	ID       int32     `json:"id" db:"id"`
	UserID   int32     `json:"user_id" db:"user_id"`
	Title    string    `json:"title" db:"title"`
	Content  string    `json:"content" db:"content"`
	Created  time.Time `json:"created_at" db:"created_at"`
}

// Transform User to UserAPI
func transformUser(user User) UserAPI {
	return UserAPI{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Created:  user.CreatedAt.Format("2006-01-02"),
	}
}

func main() {
	// Initialize database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Create tables (normally done by migrations)
	createTables(db)

	// Create APIRight app
	app := apiright.New(&apiright.Config{
		Port:     "12000",
		Database: "sqlite3",
		DSN:      ":memory:",
	})

	// Add middleware stack
	app.Use(apiright.CORSMiddleware())
	app.Use(apiright.LoggingMiddleware())
	app.Use(apiright.JSONValidationMiddleware())

	// Register CRUD endpoints
	// Direct SQLC struct registration
	app.RegisterCRUD("/users", User{})
	app.RegisterCRUD("/posts", Post{})

	// Register with transformation layer
	app.RegisterCRUDWithTransform("/api/users", User{}, UserAPI{})

	// Add custom endpoints
	app.Router().HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// User-specific endpoints
	app.Router().HandleFunc("/users/{id}/posts", getUserPosts).Methods("GET")
	app.Router().HandleFunc("/users/search", searchUsers).Methods("GET")

	log.Println("🚀 APIRight server starting on port 12000")
	log.Println("📚 Available endpoints:")
	log.Println("  GET    /health")
	log.Println("  GET    /users")
	log.Println("  POST   /users")
	log.Println("  GET    /users/{id}")
	log.Println("  PUT    /users/{id}")
	log.Println("  DELETE /users/{id}")
	log.Println("  GET    /posts")
	log.Println("  POST   /posts")
	log.Println("  GET    /posts/{id}")
	log.Println("  PUT    /posts/{id}")
	log.Println("  DELETE /posts/{id}")
	log.Println("  GET    /api/users (with transformation)")
	log.Println("  GET    /users/{id}/posts")
	log.Println("  GET    /users/search")

	// Start the server
	if err := app.Start(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func createTables(db *sql.DB) {
	userTable := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	postTable := `
	CREATE TABLE posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`

	if _, err := db.Exec(userTable); err != nil {
		log.Fatal("Failed to create users table:", err)
	}

	if _, err := db.Exec(postTable); err != nil {
		log.Fatal("Failed to create posts table:", err)
	}

	// Insert sample data
	insertSampleData(db)
}

func insertSampleData(db *sql.DB) {
	users := []struct {
		username, email string
	}{
		{"john_doe", "john@example.com"},
		{"jane_smith", "jane@example.com"},
		{"bob_wilson", "bob@example.com"},
	}

	for _, user := range users {
		_, err := db.Exec("INSERT INTO users (username, email) VALUES (?, ?)", user.username, user.email)
		if err != nil {
			log.Printf("Failed to insert user %s: %v", user.username, err)
		}
	}

	posts := []struct {
		userID      int
		title, content string
	}{
		{1, "First Post", "This is my first blog post!"},
		{1, "Go Programming", "Learning Go has been amazing..."},
		{2, "API Design", "Best practices for REST APIs..."},
		{3, "Database Tips", "Optimizing SQL queries..."},
	}

	for _, post := range posts {
		_, err := db.Exec("INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)", post.userID, post.title, post.content)
		if err != nil {
			log.Printf("Failed to insert post %s: %v", post.title, err)
		}
	}
}

// Custom endpoint handlers
func getUserPosts(w http.ResponseWriter, r *http.Request) {
	// Implementation would query posts by user ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"User posts endpoint - implementation needed"}`))
}

func searchUsers(w http.ResponseWriter, r *http.Request) {
	// Implementation would search users by query parameters
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"User search endpoint - implementation needed"}`))
}