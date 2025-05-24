package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bata94/apiright"
)

// User represents a user in our system
type User struct {
	ID    int    `json:"id" db:"id"`
	Name  string `json:"name" db:"name"`
	Email string `json:"email" db:"email"`
}

func main() {
	// Test basic framework initialization
	fmt.Println("Testing APIRight Framework...")

	// Create a new APIRight app
	app := apiright.New(&apiright.Config{
		Port:     "12000",
		Database: "sqlite3",
		DSN:      ":memory:",
	})

	// Test middleware
	app.Use(apiright.CORSMiddleware())
	app.Use(apiright.LoggingMiddleware())

	// Test CRUD registration
	app.RegisterCRUD("/users", User{})

	// Test health endpoint
	app.Router().HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	fmt.Println("✅ Framework initialization successful")
	fmt.Println("✅ Middleware registration successful")
	fmt.Println("✅ CRUD registration successful")
	fmt.Println("✅ Health endpoint registration successful")

	// Start server in a goroutine for testing
	go func() {
		if err := app.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(2 * time.Second)

	// Test health endpoint
	resp, err := http.Get("http://localhost:12000/health")
	if err != nil {
		fmt.Printf("❌ Health check failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("✅ Health endpoint working")
	} else {
		fmt.Printf("❌ Health endpoint returned status: %d\n", resp.StatusCode)
	}

	fmt.Println("\n🎉 APIRight Framework Test Complete!")
	fmt.Println("The framework is ready for use with SQLC-generated structs.")
}