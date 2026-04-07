package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bata94/apiright/pkg/config"
)

// GenerationContext holds state during code generation
type GenerationContext struct {
	ProjectDir   string
	GenDir       string
	UserQueries  string
	ModulePath   string
	Schema       *Schema
	Logger       Logger
	Config       map[string]any
	Output       io.Writer
	ContentTypes []string
}

// NewGenerationContext creates a new generation context
func NewGenerationContext(projectDir string) *GenerationContext {
	return &GenerationContext{
		ProjectDir:  projectDir,
		GenDir:      filepath.Join(projectDir, "gen"),
		UserQueries: filepath.Join(projectDir, "queries"),
		Config:      make(map[string]any),
		Output:      os.Stdout,
		ContentTypes: []string{
			"application/json",
			"application/xml",
			"application/yaml",
			"application/protobuf",
			"text/plain",
		},
	}
}

// WithModulePath sets the module path for the generation context
func (gc *GenerationContext) WithModulePath(modulePath string) *GenerationContext {
	gc.ModulePath = modulePath
	return gc
}

// WithLogger sets the logger for the generation context
func (gc *GenerationContext) WithLogger(logger Logger) *GenerationContext {
	gc.Logger = logger
	return gc
}

// WithOutput sets the output writer for the generation context
func (gc *GenerationContext) WithOutput(output io.Writer) *GenerationContext {
	gc.Output = output
	return gc
}

// WithSchema sets the schema for the generation context
func (gc *GenerationContext) WithSchema(schema *Schema) *GenerationContext {
	gc.Schema = schema
	return gc
}

// WithConfig sets a configuration value
func (gc *GenerationContext) WithConfig(key string, value any) *GenerationContext {
	gc.Config[key] = value
	return gc
}

// GetConfig gets a configuration value
func (gc *GenerationContext) GetConfig(key string) (any, bool) {
	value, exists := gc.Config[key]
	return value, exists
}

// GetConfigString gets a string configuration value
func (gc *GenerationContext) GetConfigString(key string) string {
	if value, exists := gc.GetConfig(key); exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// GetConfigBool gets a boolean configuration value
func (gc *GenerationContext) GetConfigBool(key string) bool {
	if value, exists := gc.GetConfig(key); exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

// Log logs a message if logger is available
func (gc *GenerationContext) Log(level, message string, args ...any) {
	if gc.Logger != nil {
		switch level {
		case "debug":
			gc.Logger.Debug(message, args...)
		case "info":
			gc.Logger.Info(message, args...)
		case "warn":
			gc.Logger.Warn(message, args...)
		case "error":
			gc.Logger.Error(message, args...)
		default:
			gc.Logger.Info(message, args...)
		}
	}
}

// Write writes a message to the output
func (gc *GenerationContext) Write(message string) {
	if gc.Output != nil {
		_, _ = fmt.Fprint(gc.Output, message)
	}
}

// Writeln writes a message with newline to the output
func (gc *GenerationContext) Writeln(message string) {
	if gc.Output != nil {
		_, _ = fmt.Fprintln(gc.Output, message)
	}
}

// Writef writes a formatted message to the output
func (gc *GenerationContext) Writef(format string, args ...any) {
	if gc.Output != nil {
		_, _ = fmt.Fprintf(gc.Output, format, args...)
	}
}

// EnsureDir ensures a directory exists
func (gc *GenerationContext) EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// ReadFile reads a file
func (gc *GenerationContext) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile writes a file
func (gc *GenerationContext) WriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := gc.EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return os.WriteFile(path, data, perm)
}

// Exists checks if a file or directory exists
func (gc *GenerationContext) Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsDir checks if a path is a directory
func (gc *GenerationContext) IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// Join joins path elements
func (gc *GenerationContext) Join(elem ...string) string {
	return filepath.Join(elem...)
}

// Relative returns a relative path from base to target
func (gc *GenerationContext) Relative(base, target string) (string, error) {
	return filepath.Rel(base, target)
}

// Abs returns an absolute path
func (gc *GenerationContext) Abs(path string) (string, error) {
	return filepath.Abs(path)
}

// Base returns the last element of a path
func (gc *GenerationContext) Base(path string) string {
	return filepath.Base(path)
}

// Dir returns the directory part of a path
func (gc *GenerationContext) Dir(path string) string {
	return filepath.Dir(path)
}

// Ext returns the file extension of a path
func (gc *GenerationContext) Ext(path string) string {
	return filepath.Ext(path)
}

// LoadConfig loads configuration from apiright.yaml (convenience method)
func LoadConfig(projectDir string) (*config.Config, error) {
	return config.LoadConfig(projectDir)
}
