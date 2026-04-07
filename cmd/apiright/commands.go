package apiright

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bata94/apiright/pkg/config"
	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/database"
	"github.com/bata94/apiright/pkg/generator"
	"github.com/bata94/apiright/pkg/plugins"
	"github.com/bata94/apiright/pkg/server"
	"github.com/spf13/cobra"
)

var devMode bool

// NewGenCommand creates gen command
func NewGenCommand() *cobra.Command {
	var options generator.GenerateOptions

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate CRUD operations from SQL schema",
		Long:  `Generate parses your SQL schema and generates complete CRUD operations, including database queries, Go types, protobuf definitions, and HTTP/gRPC endpoints.`,
		Example: `  apiright gen
  apiright gen --sql-only
  apiright gen --force
  apiright gen --verbose
  apiright gen --go-only`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runGenerate(cmd, options); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().BoolVar(&options.SQLOnly, "sql-only", false, "generate SQL queries only")
	cmd.Flags().BoolVar(&options.GoOnly, "go-only", false, "generate Go code only")
	cmd.Flags().BoolVar(&options.ProtoOnly, "proto-only", false, "generate protobuf only")
	cmd.Flags().BoolVar(&options.Force, "force", false, "force regeneration bypassing cache")
	cmd.Flags().BoolVar(&options.Verbose, "verbose", false, "detailed output with progress")
	cmd.Flags().BoolVar(&options.DryRun, "dry-run", false, "show what would be generated without writing files")

	return cmd
}

// runGenerate executes the generation process
func runGenerate(cmd *cobra.Command, options generator.GenerateOptions) error {
	// Determine project directory
	projectDir, err := GetProjectDir(cmd)
	if err != nil {
		return fmt.Errorf("failed to determine project directory: %w", err)
	}

	// Create logger
	logger, err := core.NewLogger(devMode)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer core.SyncLogger(logger)

	// Load configuration
	cfg, err := config.LoadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	logger.Info("Loaded configuration", "plugins_count", len(cfg.Plugins))

	// Create generation context
	ctx := core.NewGenerationContext(projectDir).
		WithLogger(logger).
		WithConfig("database_type", cfg.Database.Type).
		WithConfig("gen_suffix", cfg.Generation.GenSuffix)

	// Create plugin registry
	pluginRegistry := plugins.NewPluginRegistry()
	pluginLoader := plugins.NewPluginLoader(pluginRegistry)

	// Load plugins from configuration
	if len(cfg.Plugins) > 0 {
		if err := pluginLoader.LoadFromConfig(cfg.Plugins); err != nil {
			return fmt.Errorf("failed to load plugins from config: %w", err)
		}
	}

	// Create generator
	gen, err := generator.NewGenerator(projectDir, options, logger, pluginRegistry)
	if err != nil {
		return fmt.Errorf("failed to create generator: %w", err)
	}

	// Execute generation
	return gen.Generate(ctx, options)
}

// NewServeCommand creates serve command
func NewServeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start development server",
		Long:  `Serve starts the development server with HTTP and gRPC endpoints, supporting all configured content types (JSON, XML, YAML, Protobuf, Plain Text).`,
		Example: `  apiright serve
  apiright serve -H 3000 -G 5000
  apiright serve --host 0.0.0.0
  apiright serve --dev`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runServe(cmd); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().IntP("http-port", "H", 8080, "HTTP server port")
	cmd.Flags().IntP("grpc-port", "G", 9090, "gRPC server port")
	cmd.Flags().String("host", "localhost", "server host")
	cmd.Flags().Bool("dev", false, "enable development mode with hot reload")

	return cmd
}

// runServe executes the server startup
func runServe(cmd *cobra.Command) error {
	// Determine project directory
	projectDir, err := GetProjectDir(cmd)
	if err != nil {
		return fmt.Errorf("failed to determine project directory: %w", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override config with command-line flags
	if httpPort, _ := cmd.Flags().GetInt("http-port"); httpPort != 8080 {
		cfg.Server.HTTPPort = httpPort
	}
	if grpcPort, _ := cmd.Flags().GetInt("grpc-port"); grpcPort != 9090 {
		cfg.Server.GRPCPort = grpcPort
	}
	if host, _ := cmd.Flags().GetString("host"); host != "localhost" {
		cfg.Server.Host = host
	}

	// Create logger
	logger, err := core.NewLogger(devMode)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer core.SyncLogger(logger)

	// Initialize database
	db, err := database.NewDatabase(&cfg.Database, logger)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	// Connect to database
	if err := db.Connect(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer core.Close("database", db, logger)

	// Run migrations
	if err := db.Migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize server
	srv := server.NewServer(&cfg.Server, db, logger)

	logger.Info("Starting APIRight server",
		"http", cfg.Server.GetHTTPAddress(),
		"grpc", cfg.Server.GetGRPCAddress(),
		"database", cfg.Database.Type,
	)

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srv.Start(ctx); err != nil {
			logger.Error("Server error", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	cancel()
	if err := srv.Stop(); err != nil {
		logger.Error("Error stopping server", "error", err)
	}

	logger.Info("Server stopped")
	return nil
}
