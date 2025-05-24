package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/bata94/apiright"
	_ "github.com/mattn/go-sqlite3"
)

// User represents a user in the database (SQLC-generated style)
type User struct {
	ID        int32     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserAPI represents the API model for users
type UserAPI struct {
	ID    int32  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	// Initialize database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Create users table
	createTableSQL := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	
	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatal("Failed to create table:", err)
	}

	// Insert sample data
	insertSQL := `INSERT INTO users (name, email) VALUES (?, ?)`
	_, err = db.Exec(insertSQL, "John Doe", "john@example.com")
	if err != nil {
		log.Fatal("Failed to insert sample data:", err)
	}

	// Create APIRight app with corrected Config
	app := apiright.New(&apiright.Config{
		Host:     "0.0.0.0",
		Port:     "12000",
		Database: "sqlite3",
		DSN:      ":memory:",
	})

	// Register CRUD endpoints
	app.RegisterCRUD("/users", User{})
	app.RegisterCRUDWithTransform("/api/users", User{}, UserAPI{})

	// Add a health check endpoint
	app.Router().HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "1.0.0",
		}
		apiright.JSONResponse(w, response, http.StatusOK)
	}).Methods("GET")

	// Add custom endpoints
	app.Router().HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"framework": "APIRight",
			"version":   "1.0.0",
			"endpoints": []string{
				"GET /health",
				"GET /info",
				"GET /users",
				"POST /users",
				"GET /users/{id}",
				"PUT /users/{id}",
				"DELETE /users/{id}",
				"GET /api/users (with transformation)",
			},
		}
		apiright.JSONResponse(w, response, http.StatusOK)
	}).Methods("GET")

	// Start the server
	log.Printf("🚀 Starting APIRight server on port 12000")
	log.Printf("📚 Available endpoints:")
	log.Printf("  GET    /health")
	log.Printf("  GET    /info")
	log.Printf("  GET    /users")
	log.Printf("  POST   /users")
	log.Printf("  GET    /users/{id}")
	log.Printf("  PUT    /users/{id}")
	log.Printf("  DELETE /users/{id}")
	log.Printf("  GET    /api/users (with transformation)")
	log.Printf("🌐 Server will be accessible at: http://localhost:12000")

	if err := app.Start(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}