package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bata94/apiright/pkg/config"
	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/database"
	"github.com/bata94/apiright/pkg/middleware"
	"github.com/bata94/apiright/pkg/server"
)

var devMode = flag.Bool("dev", true, "Development mode")

func main() {
	flag.Parse()

	logger, err := core.NewLogger(*devMode)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer core.SyncLogger(logger)

	logger.Info("Starting Simple API", core.String("mode", map[bool]string{true: "dev", false: "prod"}[*devMode]))

	cfg, err := config.LoadConfig(".")
	if err != nil {
		logger.Error("Failed to load config", core.Error(err))
		os.Exit(1)
	}

	db, err := database.NewDatabase(&cfg.Database, logger)
	if err != nil {
		logger.Error("Failed to create database", core.Error(err))
		os.Exit(1)
	}

	if err := db.Connect(); err != nil {
		logger.Error("Failed to connect", core.Error(err))
		os.Exit(1)
	}
	defer core.Close("database", db, logger)

	if err := db.Migrate(); err != nil {
		logger.Error("Failed to migrate", core.Error(err))
		os.Exit(1)
	}

	mwRegistry := middleware.NewMiddlewareRegistry(logger)
	mwRegistry.RegisterHTTP(middleware.LoggingMiddleware(logger))
	mwRegistry.RegisterHTTP(middleware.RecoveryMiddleware(logger))

	srv := server.NewServer(&cfg.Server, db, logger)
	srv.SetMiddlewareRegistry(mwRegistry)
	srv.RegisterGeneratedServices(".")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srv.Start(ctx); err != nil {
			logger.Error("Server error", core.Error(err))
			cancel()
		}
	}()

	logger.Info("Server ready", core.String("http", fmt.Sprintf("http://localhost:%d", cfg.Server.HTTPPort)))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()
	srv.Stop()
}
