package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/server"
)

func main() {
	// TODO: Initialize and start server
	log.Println("Starting test-init application...")
	
	// Example server initialization
	srv := server.NewServer()
	
	// TODO: Register services
	
	// Start server in goroutine
	go func() {
		if err := srv.Start(context.Background()); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	if err := srv.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}
	
	log.Println("Server stopped")
}
