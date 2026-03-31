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
	"github.com/bata94/apiright/pkg/plugins"
	"github.com/bata94/apiright/pkg/server"
)

var (
	devMode = flag.Bool("dev", true, "Development mode (colored logs)")
	verbose = flag.Bool("v", false, "Verbose logging")
)

func main() {
	flag.Parse()

	logger, err := core.NewLogger(*devMode)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer core.SyncLogger(logger)

	logger.Info("Starting Todo API", core.String("mode", map[bool]string{true: "development", false: "production"}[*devMode]))

	cfg, err := config.LoadConfig(".")
	if err != nil {
		logger.Error("Failed to load configuration", core.Error(err))
		os.Exit(1)
	}

	db, err := database.NewDatabase(&cfg.Database, logger)
	if err != nil {
		logger.Error("Failed to create database", core.Error(err))
		os.Exit(1)
	}

	if err := db.Connect(); err != nil {
		logger.Error("Failed to connect to database", core.Error(err))
		os.Exit(1)
	}
	defer core.Close("database", db, logger)

	if err := db.Migrate(); err != nil {
		logger.Error("Failed to run migrations", core.Error(err))
		os.Exit(1)
	}

	mwRegistry := middleware.NewMiddlewareRegistry(logger)

	mwRegistry.RegisterHTTP(middleware.LoggingMiddleware(logger))
	mwRegistry.RegisterHTTP(middleware.RequestIDMiddleware())
	mwRegistry.RegisterHTTP(middleware.RecoveryMiddleware(logger))
	mwRegistry.RegisterGRPC(middleware.GRPCLoggingInterceptor(logger))
	mwRegistry.RegisterGRPC(middleware.GRPCRecoveryInterceptor())

	if err := loadPlugins(logger, mwRegistry); err != nil {
		logger.Warn("Some plugins failed to load", core.Error(err))
	}

	srv := server.NewServer(&cfg.Server, db, logger)
	srv.SetMiddlewareRegistry(mwRegistry)

	if err := srv.RegisterGeneratedServices("."); err != nil {
		logger.Warn("Failed to register services", core.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srv.Start(ctx); err != nil {
			logger.Error("Server error", core.Error(err))
			cancel()
		}
	}()

	logger.Info("Server started",
		core.String("http", fmt.Sprintf("http://localhost:%d", cfg.Server.HTTPPort)),
		core.String("grpc", fmt.Sprintf("localhost:%d", cfg.Server.GRPCPort)),
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	cancel()

	if err := srv.Stop(); err != nil {
		logger.Error("Error stopping server", core.Error(err))
	}

	logger.Info("Server stopped")
}

func loadPlugins(logger core.Logger, mwRegistry *middleware.MiddlewareRegistry) error {
	pluginPaths := []string{
		"./plugins/logging.so",
		"./plugins/validation.so",
		"./plugins/middleware.so",
	}

	registry := plugins.NewPluginRegistry(logger)

	for _, path := range pluginPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			logger.Debug("Plugin not found", core.String("path", path))
			continue
		}

		if err := registry.LoadPlugin(path); err != nil {
			logger.Warn("Failed to load plugin", core.String("path", path), core.Error(err))
			continue
		}

		logger.Info("Loaded plugin", core.String("path", path))
	}

	for _, p := range registry.GetPlugins() {
		if middlewarePlugin, ok := p.(plugins.MiddlewareProvider); ok {
			for _, mw := range middlewarePlugin.Middleware() {
				mwRegistry.RegisterHTTP(mw)
			}
		}
	}

	return nil
}
