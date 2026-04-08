package apiright

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bata94/apiright/pkg/config"
	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/generator"
	"github.com/bata94/apiright/pkg/plugins"
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
		WithModulePath(cfg.Project.Module).
		WithConfig("database_type", cfg.Database.Type).
		WithConfig("gen_suffix", cfg.Generation.GenSuffix).
		WithServerConfig(core.ServerConfig{
			Host:       cfg.Server.Host,
			HTTPPort:   cfg.Server.HTTPPort,
			GRPCPort:   cfg.Server.GRPCPort,
			APIVersion: cfg.Server.APIVersion,
			BasePath:   cfg.Server.BasePath,
		})

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

	// Check if user has their own main.go
	userMainPath := filepath.Join(projectDir, "main.go")
	if _, err := os.Stat(userMainPath); err == nil {
		// User has their own main.go, use it
		logger.Info("Using project's main.go")
		return runUserMain(projectDir, logger)
	}

	// Check if adapters exist
	adaptersDir := filepath.Join(projectDir, "gen", "go", "adapters")
	if _, err := os.Stat(adaptersDir); os.IsNotExist(err) {
		return fmt.Errorf("no generated adapters found in %s. Run 'apiright gen' first", adaptersDir)
	}

	// Generate and run temporary server with real adapters
	return runTempServer(projectDir, cfg, logger)
}

// runUserMain runs the project's own main.go
func runUserMain(projectDir string, logger core.Logger) error {
	logger.Info("Compiling project server...")

	cmd := exec.Command("go", "run", ".")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// runTempServer generates and runs a temporary server with real adapters
func runTempServer(projectDir string, cfg *config.Config, logger core.Logger) error {
	// Get apiright directory (current binary location)
	apirightDir := getAPIRightDir()

	// Create temp server generator
	tsg := generator.NewTempServerGenerator(logger)

	// Check cache first
	cacheKey := tsg.GetCacheKey(projectDir, cfg)
	cachedBinary, found := tsg.CheckCache(projectDir, cacheKey)

	if found {
		logger.Info("✅ Using cached server")
		return runBinary(cachedBinary, projectDir)
	}

	logger.Info("🔨 Compiling server with real database services...")
	logger.Info("⏳ This takes 5-10 seconds on first run")

	// Generate temp server files
	tempDir, err := tsg.GenerateTempServer(projectDir, cfg, apirightDir)
	if err != nil {
		return fmt.Errorf("failed to generate temp server: %w", err)
	}

	// Build the binary
	binaryPath := filepath.Join(tempDir, "server")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = tempDir
	buildCmd.Env = os.Environ()

	output, err := buildCmd.CombinedOutput()
	if err != nil {
		// Show full compiler output
		fmt.Fprintf(os.Stderr, "\n❌ Compilation failed:\n%s\n", string(output))
		return fmt.Errorf("failed to build server: %w", err)
	}

	// Cache the binary
	if err := tsg.SaveCache(projectDir, cacheKey, binaryPath); err != nil {
		logger.Warn("Failed to cache binary", "error", err)
	}

	logger.Info("✅ Compilation complete")
	return runBinary(binaryPath, projectDir)
}

// runBinary runs a compiled binary with proper environment
func runBinary(binaryPath, projectDir string) error {
	cmd := exec.Command(binaryPath)
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// getAPIRightDir returns the directory containing the apiright source code
func getAPIRightDir() string {
	// Try to find apiright directory from GOPATH or module cache
	// For now, assume it's in the current working directory's parent
	// or use the executable path

	exec, err := os.Executable()
	if err != nil {
		// Fallback: try common locations
		return "/home/bata/Projects/personal/apiright"
	}

	// If running from go run, exec will be a temp file
	// Try to find the actual source
	if strings.Contains(exec, "/go-build/") || strings.Contains(exec, "/tmp/go-build") {
		return "/home/bata/Projects/personal/apiright"
	}

	return filepath.Dir(exec)
}
