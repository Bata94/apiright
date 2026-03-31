// Package server provides HTTP and gRPC server implementations.
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bata94/apiright/pkg/config"
	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/database"
	"github.com/bata94/apiright/pkg/middleware"
	"google.golang.org/grpc"
)

// DualServer implements both HTTP and gRPC servers
type DualServer struct {
	config             *config.ServerConfig
	db                 *database.Database
	contentNeg         *core.ContentNegotiatorImpl
	httpServer         *http.Server
	grpcServer         *grpc.Server
	logger             core.Logger
	mu                 sync.RWMutex
	started            bool
	services           []interface{}
	middlewareRegistry *middleware.MiddlewareRegistry
	serviceRegistry    *ServiceRegistry
}

// NewServer creates a new dual HTTP/gRPC server
func NewServer(cfg *config.ServerConfig, db *database.Database, logger core.Logger) *DualServer {
	return &DualServer{
		config:             cfg,
		db:                 db,
		contentNeg:         core.NewContentNegotiator(),
		logger:             logger,
		services:           []interface{}{},
		middlewareRegistry: middleware.NewMiddlewareRegistry(logger),
		serviceRegistry:    NewServiceRegistry(db, logger),
	}
}

// RegisterGeneratedServices loads and registers all generated services
func (s *DualServer) RegisterGeneratedServices(projectDir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Load generated services from project directory
	if err := s.serviceRegistry.LoadGeneratedServices(projectDir); err != nil {
		return fmt.Errorf("failed to load generated services: %w", err)
	}

	// Discover tables from migrations directory
	tables, err := s.discoverTables(projectDir)
	if err != nil {
		s.logger.Warn("Could not discover tables from migrations, using empty list", "error", err)
		tables = []string{}
	}

	s.logger.Info("Registering services for tables", "count", len(tables))

	for _, tableName := range tables {
		service, err := s.serviceRegistry.CreateServiceFactory(tableName)
		if err != nil {
			s.logger.Warn("Failed to create service", "table", tableName, "error", err)
			continue
		}

		if err := s.RegisterService(service); err != nil {
			s.logger.Error("Failed to register service", "table", tableName, "error", err)
			continue
		}
	}

	if s.httpServer != nil {
		s.logger.Warn("Services registered after HTTP server init - restart server for routes to take effect")
	}

	return nil
}

// discoverTables discovers table names from migration files
func (s *DualServer) discoverTables(projectDir string) ([]string, error) {
	migrationsDir := filepath.Join(projectDir, "migrations")

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	tableSet := make(map[string]bool)
	createTableRe := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?["]?(\w+)["]?`)

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, file.Name()))
		if err != nil {
			continue
		}

		matches := createTableRe.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) > 1 {
				tableSet[match[1]] = true
			}
		}
	}

	tables := make([]string, 0, len(tableSet))
	for table := range tableSet {
		tables = append(tables, table)
	}

	sort.Strings(tables)
	return tables, nil
}

// Start starts HTTP and/or gRPC servers based on config
func (s *DualServer) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("server already started")
	}

	// Initialize HTTP server if enabled
	if s.config.EnableHTTP {
		if err := s.initHTTPServer(); err != nil {
			return fmt.Errorf("failed to initialize HTTP server: %w", err)
		}
	}

	// Initialize gRPC server if enabled
	if s.config.EnableGRPC {
		if err := s.initGRPCServer(); err != nil {
			return fmt.Errorf("failed to initialize gRPC server: %w", err)
		}
	}

	// Start HTTP server in goroutine if enabled
	httpErrChan := make(chan error, 1)
	if s.config.EnableHTTP {
		go func() {
			httpAddr := s.config.GetHTTPAddress()
			s.logger.Info("Starting HTTP server", "address", httpAddr)
			if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				httpErrChan <- fmt.Errorf("HTTP server error: %w", err)
			}
		}()
	}

	// Start gRPC server in goroutine if enabled
	grpcErrChan := make(chan error, 1)
	if s.config.EnableGRPC {
		go func() {
			grpcAddr := s.config.GetGRPCAddress()
			listener, err := net.Listen("tcp", grpcAddr)
			if err != nil {
				grpcErrChan <- fmt.Errorf("failed to listen on gRPC address %s: %w", grpcAddr, err)
				return
			}

			s.logger.Info("Starting gRPC server", "address", grpcAddr)
			if err := s.grpcServer.Serve(listener); err != nil {
				grpcErrChan <- fmt.Errorf("gRPC server error: %w", err)
			}
		}()
	}

	// Create a channel to track if any server is running
	runningChan := make(chan struct{}, 1)

	// Start goroutine to signal when servers are running
	go func() {
		if s.config.EnableHTTP {
			s.logger.Info("HTTP server is running", "base_path", s.config.BasePath, "api_version", s.config.APIVersion)
		}
		if s.config.EnableGRPC {
			s.logger.Info("gRPC server is running")
		}
		close(runningChan)
	}()

	// Wait for context cancellation or server errors
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Shutdown signal received")
			return s.Stop()
		case err := <-httpErrChan:
			return fmt.Errorf("HTTP server failed: %w", err)
		case err := <-grpcErrChan:
			return fmt.Errorf("gRPC server failed: %w", err)
		case <-runningChan:
			// Servers started, now wait for shutdown
		}
	}
}

// Stop gracefully stops enabled servers
func (s *DualServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil
	}

	s.logger.Info("Stopping servers")

	var errors []error

	// Stop HTTP server if enabled
	if s.httpServer != nil && s.config.EnableHTTP {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("HTTP server shutdown error: %w", err))
		}
	}

	// Stop gRPC server if enabled
	if s.grpcServer != nil && s.config.EnableGRPC {
		s.grpcServer.GracefulStop()
	}

	s.started = false

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	s.logger.Info("Servers stopped successfully")
	return nil
}

// RegisterService registers a service with enabled HTTP and/or gRPC servers
func (s *DualServer) RegisterService(service interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.services == nil {
		s.services = make([]interface{}, 0)
	}

	s.services = append(s.services, service)
	serviceType := fmt.Sprintf("%T", service)

	if s.httpServer != nil && s.config.EnableHTTP {
		if err := s.registerHTTPService(service); err != nil {
			return fmt.Errorf("failed to register HTTP service: %w", err)
		}
	}

	if s.grpcServer != nil && s.config.EnableGRPC {
		if err := s.registerGRPCService(service); err != nil {
			return fmt.Errorf("failed to register gRPC service: %w", err)
		}
	}

	s.logger.Debug("Service registered", "service", serviceType, "total", len(s.services))
	return nil
}

// Address returns the primary server address (HTTP)
func (s *DualServer) Address() string {
	return s.config.GetHTTPAddress()
}

// initHTTPServer initializes HTTP server with middleware
func (s *DualServer) initHTTPServer() error {
	// Create HTTP handler
	handler := s.createHTTPHandler()

	// Apply middleware
	for _, mw := range s.getSortedMiddleware() {
		handler = mw(handler)
		s.logger.Debug("Applied HTTP middleware")
	}

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         s.config.GetHTTPAddress(),
		Handler:      handler,
		ReadTimeout:  time.Duration(s.config.Timeout) * time.Second,
		WriteTimeout: time.Duration(s.config.Timeout) * time.Second,
	}

	return nil
}

// createHTTPHandler creates the main HTTP handler with routing
func (s *DualServer) createHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	// Register health check endpoint
	mux.HandleFunc("/health", s.healthCheckHandler)

	// Set up routes for registered services
	s.setupHTTPRoutes(mux)

	// Default route for unmatched paths
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.handleDefaultRoute(w, r)
	})

	return mux
}

// registerGRPCService registers a service with the gRPC server
func (s *DualServer) registerGRPCService(service interface{}) error {
	serviceType := fmt.Sprintf("%T", service)
	s.logger.Info("Registering gRPC service", "service", serviceType)
	s.logger.Debug("Service registered successfully", "service", serviceType)
	return nil
}

// AddMiddleware adds middleware to the server
func (s *DualServer) AddMiddleware(middleware middleware.HTTPMiddleware) {
	s.middlewareRegistry.RegisterMiddleware(middleware)
	s.logger.Info("Middleware added", "name", middleware.Name())
}

// GetHTTPServer returns the underlying HTTP server
func (s *DualServer) GetHTTPServer() *http.Server {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.httpServer
}

// GetGRPCServer returns the underlying gRPC server
func (s *DualServer) GetGRPCServer() *grpc.Server {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.grpcServer
}

// IsStarted returns whether the server is started
func (s *DualServer) IsStarted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.started
}
