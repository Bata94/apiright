package generator

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bata94/apiright/pkg/config"
	"github.com/bata94/apiright/pkg/core"
)

// TempServerGenerator generates a temporary server main.go and go.mod
// that can be compiled to run the project with real database adapters
type TempServerGenerator struct {
	logger core.Logger
}

// NewTempServerGenerator creates a new temp server generator
func NewTempServerGenerator(logger core.Logger) *TempServerGenerator {
	return &TempServerGenerator{
		logger: logger,
	}
}

// GenerateTempServer generates all necessary files for the temp server
func (tsg *TempServerGenerator) GenerateTempServer(projectDir string, cfg *config.Config, apirightDir string) (string, error) {
	tsg.logger.Info("Generating temporary server files")

	// Create temp directory
	tempDir := filepath.Join(projectDir, ".apiright_cache", "serve-temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate main.go
	if err := tsg.generateMainGo(tempDir, projectDir, cfg); err != nil {
		return "", fmt.Errorf("failed to generate main.go: %w", err)
	}

	// Generate go.mod
	if err := tsg.generateGoMod(tempDir, projectDir, cfg, apirightDir); err != nil {
		return "", fmt.Errorf("failed to generate go.mod: %w", err)
	}

	// Run go mod tidy to resolve dependencies
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tempDir
	tidyCmd.Env = os.Environ()
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		tsg.logger.Warn("go mod tidy failed, continuing anyway", "error", err, "output", string(output))
	}

	tsg.logger.Info("Generated temporary server files", "dir", tempDir)
	return tempDir, nil
}

// generateMainGo creates the temporary main.go file
func (tsg *TempServerGenerator) generateMainGo(tempDir, projectDir string, cfg *config.Config) error {
	data := struct {
		ProjectDir  string
		ModulePath  string
		AdapterPath string
		ProjectName string
	}{
		ProjectDir:  projectDir,
		ModulePath:  cfg.Project.Module,
		AdapterPath: cfg.Project.Module + "/gen/go/adapters",
		ProjectName: cfg.Project.Name,
	}

	tmpl, err := template.New("main").Parse(tempMainTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	mainPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(mainPath, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	return nil
}

// generateGoMod creates the temporary go.mod file
func (tsg *TempServerGenerator) generateGoMod(tempDir, projectDir string, cfg *config.Config, apirightDir string) error {
	data := struct {
		ProjectModule string
		ProjectDir    string
		APIRightDir   string
	}{
		ProjectModule: cfg.Project.Module,
		ProjectDir:    projectDir,
		APIRightDir:   apirightDir,
	}

	tmpl, err := template.New("gomod").Parse(tempGoModTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	modPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(modPath, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write go.mod: %w", err)
	}

	return nil
}

// GetCacheKey generates a cache key based on project state
func (tsg *TempServerGenerator) GetCacheKey(projectDir string, cfg *config.Config) string {
	// Hash the combination of:
	// 1. apiright binary version/path
	// 2. go.mod content
	// 3. gen/ directory content

	hash := sha256.New()

	// Add apiright version
	hash.Write([]byte(core.Version))

	// Add project module
	hash.Write([]byte(cfg.Project.Module))

	// Add go.mod if exists
	goModPath := filepath.Join(projectDir, "go.mod")
	if content, err := os.ReadFile(goModPath); err == nil {
		hash.Write(content)
	}

	// Add gen directory hash
	genDir := filepath.Join(projectDir, "gen")
	_ = tsg.hashDirectory(hash, genDir) // Ignore errors, cache will just be less effective

	return fmt.Sprintf("%x", hash.Sum(nil))[:16]
}

// hashDirectory recursively hashes all files in a directory
func (tsg *TempServerGenerator) hashDirectory(h interface{ Write([]byte) (int, error) }, dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Add relative path and content to hash
		relPath, _ := filepath.Rel(dir, path)
		if _, err := h.Write([]byte(relPath)); err != nil {
			return err
		}
		if _, err := h.Write(content); err != nil {
			return err
		}

		return nil
	})
}

// CheckCache checks if a cached binary exists and is valid
func (tsg *TempServerGenerator) CheckCache(projectDir, cacheKey string) (string, bool) {
	cacheDir := filepath.Join(projectDir, ".apiright_cache", "serve-binaries")
	binaryPath := filepath.Join(cacheDir, "server-"+cacheKey)

	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath, true
	}

	return "", false
}

// SaveCache saves the compiled binary to cache
func (tsg *TempServerGenerator) SaveCache(projectDir, cacheKey, binaryPath string) error {
	cacheDir := filepath.Join(projectDir, ".apiright_cache", "serve-binaries")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	cachedPath := filepath.Join(cacheDir, "server-"+cacheKey)

	// Copy binary to cache
	content, err := os.ReadFile(binaryPath)
	if err != nil {
		return err
	}

	return os.WriteFile(cachedPath, content, 0755)
}

const tempMainTemplate = `// Code generated by APIRight serve command. DO NOT EDIT.
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
	"github.com/bata94/apiright/pkg/server"
	"{{.AdapterPath}}"
)

var devMode = flag.Bool("dev", true, "Development mode")

func main() {
	flag.Parse()

	logger, err := core.NewLogger(*devMode)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer core.SyncLogger(logger)

	logger.Info("Starting {{.ProjectName}}",
		core.String("mode", map[bool]string{true: "development", false: "production"}[*devMode]))

	cfg, err := config.LoadConfig("{{.ProjectDir}}")
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

	srv := server.NewServer(&cfg.Server, "{{.ProjectDir}}", db, logger)

	// Initialize real service adapters
	if err := adapters.Init(srv, db, logger); err != nil {
		logger.Error("Failed to initialize service adapters", core.Error(err))
		os.Exit(1)
	}

	logger.Info("Registered real database services",
		core.Int("services", 3))

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
		core.String("grpc", fmt.Sprintf("localhost:%d", cfg.Server.GRPCPort)))

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
`

const tempGoModTemplate = `module apiright-serve-temp

go 1.21

require (
	github.com/bata94/apiright v0.0.0
	{{.ProjectModule}} v0.0.0
)

replace (
	github.com/bata94/apiright => {{.APIRightDir}}
	{{.ProjectModule}} => {{.ProjectDir}}
)
`
