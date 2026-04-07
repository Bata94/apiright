package apiright

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bata94/apiright/pkg/config"
	"github.com/bata94/apiright/pkg/database"
	"github.com/bata94/apiright/pkg/server"
	"github.com/spf13/cobra"
)

// Forward declarations for template imports
var _ = config.Config{}
var _ = database.Database{}
var _ = server.DualServer{}

// InitOptions holds options for project initialization
type InitOptions struct {
	ProjectName string
	Database    string
	Module      string
	Example     bool
	Verbose     bool
	Force       bool
}

// NewInitCommand creates init command with actual implementation
func NewInitCommand() *cobra.Command {
	var opts InitOptions

	cmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new APIRight project",
		Long:  `Initialize creates a new APIRight project with standard directory structure, configuration files, and a basic example schema.`,
		Example: `  apiright init myproject
  apiright init myproject -d postgres
  apiright init myproject --module github.com/user/myproject
  apiright init myproject --no-example`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			opts.ProjectName = args[0]
			if err := runInit(&opts); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&opts.Database, "database", "d", "sqlite", "database type (sqlite, postgres, mysql)")
	cmd.Flags().StringVarP(&opts.Module, "module", "m", "", "Go module path (default: github.com/username/project)")
	cmd.Flags().BoolVarP(&opts.Example, "example", "e", true, "include example schema and migrations")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "overwrite existing files")

	return cmd
}

// runInit executes the project initialization
func runInit(opts *InitOptions) error {
	if opts.Verbose {
		fmt.Printf("Initializing new APIRight project: %s\n", opts.ProjectName)
	}

	// Validate project name
	if err := validateProjectName(opts.ProjectName); err != nil {
		return fmt.Errorf("invalid project name: %w", err)
	}

	// Check if directory already exists
	if _, err := os.Stat(opts.ProjectName); err == nil {
		if !opts.Force {
			return fmt.Errorf("directory '%s' already exists (use --force to overwrite)", opts.ProjectName)
		}
		fmt.Printf("Overwriting existing project in: %s\n", opts.ProjectName)
	}

	// Create project directory
	if err := os.Mkdir(opts.ProjectName, 0755); err != nil {
		if !os.IsExist(err) {
			return fmt.Errorf("failed to create project directory: %w", err)
		}
	}

	// Determine module path
	modulePath := opts.Module
	if modulePath == "" {
		modulePath = fmt.Sprintf("github.com/%s/%s", getDefaultUsername(), opts.ProjectName)
	}

	// Create project template
	template := createProjectTemplate(opts.ProjectName, modulePath, opts.Database, opts.Example)

	// Create directories
	for _, dir := range template.Dirs {
		fullPath := filepath.Join(opts.ProjectName, dir)
		if opts.Verbose {
			fmt.Printf("Creating directory: %s\n", fullPath)
		}
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
		}
	}

	// Create files
	for filePath, content := range template.Files {
		fullPath := filepath.Join(opts.ProjectName, filePath)
		if opts.Verbose {
			fmt.Printf("Creating file: %s\n", fullPath)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", fullPath, err)
		}
	}

	fmt.Printf("✅ Successfully initialized APIRight project: %s\n", opts.ProjectName)
	fmt.Printf("📁 Module path: %s\n", modulePath)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n", opts.ProjectName)
	fmt.Printf("  apiright gen\n")
	fmt.Printf("  apiright serve\n")

	return nil
}

// validateProjectName validates the project name
func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	if strings.Contains(name, " ") {
		return fmt.Errorf("project name cannot contain spaces")
	}

	// Check for invalid characters
	for _, r := range name {
		valid := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
		if !valid {
			return fmt.Errorf("project name contains invalid character: %c", r)
		}
	}

	return nil
}

// getDefaultUsername returns a default username for the module path
func getDefaultUsername() string {
	if username := os.Getenv("USER"); username != "" {
		return username
	}
	if username := os.Getenv("USERNAME"); username != "" {
		return username
	}
	return "username"
}

// ProjectTemplate represents a project template
type ProjectTemplate struct {
	Name     string
	Files    map[string]string
	Dirs     []string
	Database string
}

// createProjectTemplate creates a project template
func createProjectTemplate(projectName, modulePath, database string, withExample bool) *ProjectTemplate {
	template := &ProjectTemplate{
		Name:     projectName,
		Database: database,
		Dirs: []string{
			"gen",
			"gen/sql",
			"gen/go",
			"gen/proto",
			"queries",
			"proto",
			"migrations",
		},
		Files: map[string]string{},
	}

	// Go module file
	template.Files["go.mod"] = fmt.Sprintf("module %s\n\ngo 1.21\n\nrequire (\n\tgithub.com/bata94/apiright v0.1.0\n)\n", modulePath)

	// sqlc configuration
	template.Files["sqlc.yaml"] = fmt.Sprintf(`version: "2"
sql:
  - engine: "%s"
    queries: "queries/"
    schema: "migrations/"
    gen:
      go:
        package: "db"
        out: "gen/go"
        sql_package: "database/sql"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
`, getSQLEngine(database))

	// APIRight configuration
	template.Files["apiright.yaml"] = fmt.Sprintf(`# APIRight Configuration
project:
  name: "%s"
  module: "%s"
  
database:
  type: "%s"
  # Additional database configuration
  
server:
  http_port: 8080
  grpc_port: 9090
  host: "localhost"
  
generation:
  output_dir: "gen"
  gen_suffix: "_ar_gen"
  content_types:
    - "application/json"
    - "application/xml"
    - "application/yaml"
    - "application/protobuf"
    - "text/plain"
  generate_tests: true
  generate_docs: true
  validation: true
  
plugins: []
`, projectName, modulePath, database)

	// .gitignore
	template.Files[".gitignore"] = `# Generated files
gen/

# Go
*.sum
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with go test -c
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories
vendor/

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Environment
.env
.env.local
.env.*.local
`

	// Main application file
	template.Files["main.go"] = fmt.Sprintf(`package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bata94/apiright/pkg/config"
	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/database"
	"github.com/bata94/apiright/pkg/server"
)

func main() {
	log.Printf("Starting %s application...", projectName)

	// Load configuration
	cfg, err := config.Load("apiright.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %%v", err)
	}

	// Create logger
	logger := &simpleLogger{verbose: false}

	// Initialize database
	db, err := database.NewDatabase(&cfg.Database, logger)
	if err != nil {
		log.Fatalf("Failed to create database: %%v", err)
	}

	// Connect to database
	if err := db.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %%v", err)
	}

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %%v", err)
	}

	// Initialize server
	srv := server.NewServer(&cfg.Server, db, logger)

	// Register generated services
	if err := srv.RegisterGeneratedServices("."); err != nil {
		log.Fatalf("Failed to register services: %%v", err)
	}

	// Start server in goroutine
	go func() {
		if err := srv.Start(context.Background()); err != nil {
			log.Fatalf("Failed to start server: %%v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := srv.Stop(); err != nil {
		log.Printf("Error stopping server: %%v", err)
	}

	log.Println("Server stopped")
}

type simpleLogger struct {
	verbose bool
}

func (l *simpleLogger) Debug(msg string, fields ...any) {
	if l.verbose {
		fmt.Println("[DEBUG]", msg, fields)
	}
}

func (l *simpleLogger) Info(msg string, fields ...any) {
	fmt.Println("[INFO]", msg)
}

func (l *simpleLogger) Warn(msg string, fields ...any) {
	fmt.Println("[WARN]", msg, fields)
}

func (l *simpleLogger) Error(msg string, fields ...any) {
	fmt.Println("[ERROR]", msg, fields)
}
`, projectName)

	// Add example files if requested
	if withExample {
		addExampleFiles(template)
	}

	return template
}

// getSQLEngine returns the sqlc engine name for the database
func getSQLEngine(database string) string {
	switch database {
	case "postgres":
		return "postgresql"
	case "mysql":
		return "mysql"
	case "sqlite":
		return "sqlite"
	default:
		return "sqlite"
	}
}

// addExampleFiles adds example schema and migration files
func addExampleFiles(template *ProjectTemplate) {
	// Example schema (users table)
	template.Files["migrations/001_create_todos_table.sql"] = `-- Create todos table
CREATE TABLE todos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    completed BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_todos_completed ON todos(completed);
CREATE INDEX idx_todos_created_at ON todos(created_at);
`

	// Example custom query
	template.Files["queries/todos.sql"] = `-- name: GetTodosByStatus :many
SELECT id, title, completed, created_at, updated_at
FROM todos
WHERE completed = ? 
ORDER BY created_at DESC;

-- name: GetTodosCreatedAfter :many
SELECT id, title, completed, created_at, updated_at
FROM todos
WHERE created_at > ? 
ORDER BY created_at DESC;
`
}
