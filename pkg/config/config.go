package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the complete APIRight configuration
type Config struct {
	Project    ProjectConfig    `yaml:"project"`
	Database   DatabaseConfig   `yaml:"database"`
	Server     ServerConfig     `yaml:"server"`
	Generation GenerationConfig `yaml:"generation"`
	Plugins    []PluginConfig   `yaml:"plugins"`
}

// ProjectConfig holds project-specific configuration
type ProjectConfig struct {
	Name    string `yaml:"name"`
	Module  string `yaml:"module"`
	Version string `yaml:"version"`
	Author  string `yaml:"author"`
	License string `yaml:"license"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`
	URL      string `yaml:"url"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	EnableHTTP bool      `yaml:"enable_http"`
	EnableGRPC bool      `yaml:"enable_grpc"`
	APIVersion string    `yaml:"api_version"`
	BasePath   string    `yaml:"base_path"`
	HTTPPort   int       `yaml:"http_port"`
	GRPCPort   int       `yaml:"grpc_port"`
	Host       string    `yaml:"host"`
	Timeout    int       `yaml:"timeout"`
	TLS        TLSConfig `yaml:"tls"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// GenerationConfig holds generation configuration
type GenerationConfig struct {
	OutputDir     string   `yaml:"output_dir"`
	GenSuffix     string   `yaml:"gen_suffix"`
	ContentTypes  []string `yaml:"content_types"`
	GenerateTests bool     `yaml:"generate_tests"`
	GenerateDocs  bool     `yaml:"generate_docs"`
	Validation    bool     `yaml:"validation"`
	Middleware    []string `yaml:"middleware"`
}

// PluginConfig holds plugin configuration
type PluginConfig struct {
	Name    string         `yaml:"name"`
	Version string         `yaml:"version"`
	Enabled bool           `yaml:"enabled"`
	Path    string         `yaml:"path"`
	Config  map[string]any `yaml:"config"`
}

// LoadConfig loads configuration from apiright.yaml file
func LoadConfig(projectDir string) (*Config, error) {
	configPath := filepath.Join(projectDir, "apiright.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to check config file: %w", err)
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults for missing values
	applyDefaults(&config)

	// Expand environment variables
	expandEnv(&config)

	return &config, nil
}

// SaveConfig saves configuration to apiright.yaml file
func SaveConfig(projectDir string, config *Config) error {
	configPath := filepath.Join(projectDir, "apiright.yaml")

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Project: ProjectConfig{
			Name:    "APIRight Project",
			Version: "0.1.0",
			License: "MIT",
		},
		Database: DatabaseConfig{
			Type:    "sqlite",
			Host:    "localhost",
			Port:    5432,
			Name:    "app.db",
			SSLMode: "disable",
		},
		Server: ServerConfig{
			EnableHTTP: true,
			EnableGRPC: true,
			APIVersion: "v0",
			BasePath:   "/api",
			HTTPPort:   8080,
			GRPCPort:   9090,
			Host:       "localhost",
			Timeout:    30,
			TLS: TLSConfig{
				Enabled: false,
			},
		},
		Generation: GenerationConfig{
			OutputDir: "gen",
			GenSuffix: "_ar_gen",
			ContentTypes: []string{
				"application/json",
				"application/xml",
				"application/yaml",
				"application/protobuf",
				"text/plain",
			},
			GenerateTests: true,
			GenerateDocs:  true,
			Validation:    true,
			Middleware:    []string{},
		},
		Plugins: []PluginConfig{},
	}
}

// applyDefaults fills in missing configuration values with defaults
func applyDefaults(config *Config) {
	// Project defaults
	if config.Project.Name == "" {
		config.Project.Name = "APIRight Project"
	}
	if config.Project.Version == "" {
		config.Project.Version = "0.1.0"
	}
	if config.Project.License == "" {
		config.Project.License = "MIT"
	}

	// Database defaults
	if config.Database.Type == "" {
		config.Database.Type = "sqlite"
	}
	if config.Database.Host == "" {
		config.Database.Host = "localhost"
	}
	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}
	if config.Database.SSLMode == "" {
		config.Database.SSLMode = "disable"
	}

	// Server defaults
	if !config.Server.EnableHTTP && !config.Server.EnableGRPC {
		config.Server.EnableHTTP = true
	}
	if config.Server.APIVersion == "" {
		config.Server.APIVersion = "v0"
	}
	if config.Server.BasePath == "" {
		config.Server.BasePath = "/api"
	}
	if config.Server.HTTPPort == 0 {
		config.Server.HTTPPort = 8080
	}
	if config.Server.GRPCPort == 0 {
		config.Server.GRPCPort = 9090
	}
	if config.Server.Host == "" {
		config.Server.Host = "localhost"
	}
	if config.Server.Timeout == 0 {
		config.Server.Timeout = 30
	}

	// Generation defaults
	if config.Generation.OutputDir == "" {
		config.Generation.OutputDir = "gen"
	}
	if config.Generation.GenSuffix == "" {
		config.Generation.GenSuffix = "_ar_gen"
	}
	if len(config.Generation.ContentTypes) == 0 {
		config.Generation.ContentTypes = []string{
			"application/json",
			"application/xml",
			"application/yaml",
			"application/protobuf",
			"text/plain",
		}
	}
}

// ValidateConfig validates the configuration
func ValidateConfig(config *Config) error {
	// Validate project config
	if config.Project.Name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// Validate database config
	if config.Database.Type == "" {
		return fmt.Errorf("database type cannot be empty")
	}
	validDBTypes := []string{"sqlite", "postgres", "mysql"}
	isValidDBType := false
	for _, validType := range validDBTypes {
		if config.Database.Type == validType {
			isValidDBType = true
			break
		}
	}
	if !isValidDBType {
		return fmt.Errorf("invalid database type: %s (must be one of: %v)", config.Database.Type, validDBTypes)
	}

	// Validate server config
	if !config.Server.EnableHTTP && !config.Server.EnableGRPC {
		return fmt.Errorf("at least one of enable_http or enable_grpc must be true")
	}
	if config.Server.EnableHTTP && (config.Server.HTTPPort <= 0 || config.Server.HTTPPort > 65535) {
		return fmt.Errorf("invalid HTTP port: %d", config.Server.HTTPPort)
	}
	if config.Server.EnableGRPC && (config.Server.GRPCPort <= 0 || config.Server.GRPCPort > 65535) {
		return fmt.Errorf("invalid gRPC port: %d", config.Server.GRPCPort)
	}
	if config.Server.EnableHTTP && config.Server.EnableGRPC && config.Server.HTTPPort == config.Server.GRPCPort {
		return fmt.Errorf("HTTP and gRPC ports cannot be the same: %d", config.Server.HTTPPort)
	}
	if config.Server.APIVersion == "" {
		return fmt.Errorf("api_version cannot be empty")
	}

	// Validate generation config
	if config.Generation.OutputDir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}
	if config.Generation.GenSuffix == "" {
		return fmt.Errorf("generation suffix cannot be empty")
	}
	if len(config.Generation.ContentTypes) == 0 {
		return fmt.Errorf("content types cannot be empty")
	}

	return nil
}

// GetDatabaseURL returns the database connection URL
func (c *DatabaseConfig) GetDatabaseURL() string {
	if c.URL != "" {
		return c.URL
	}

	switch c.Type {
	case "sqlite":
		if c.Name == "" {
			c.Name = "app.db"
		}
		return c.Name
	case "postgres":
		port := c.Port
		if port == 0 {
			port = 5432
		}
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			c.User, c.Password, c.Host, port, c.Name, c.SSLMode)
	case "mysql":
		port := c.Port
		if port == 0 {
			port = 3306
		}
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			c.User, c.Password, c.Host, port, c.Name)
	default:
		return ""
	}
}

// GetHTTPAddress returns the HTTP server address
func (c *ServerConfig) GetHTTPAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.HTTPPort)
}

// GetGRPCAddress returns the gRPC server address
func (c *ServerConfig) GetGRPCAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.GRPCPort)
}

// expandEnv expands environment variables in configuration values
// Supports $VAR and ${VAR} syntax
func expandEnv(config *Config) {
	config.Database.Host = os.ExpandEnv(config.Database.Host)
	config.Database.Name = os.ExpandEnv(config.Database.Name)
	config.Database.User = os.ExpandEnv(config.Database.User)
	config.Database.Password = os.ExpandEnv(config.Database.Password)
	config.Database.SSLMode = os.ExpandEnv(config.Database.SSLMode)
	config.Database.URL = os.ExpandEnv(config.Database.URL)
	config.Server.Host = os.ExpandEnv(config.Server.Host)
	config.Server.TLS.CertFile = os.ExpandEnv(config.Server.TLS.CertFile)
	config.Server.TLS.KeyFile = os.ExpandEnv(config.Server.TLS.KeyFile)
}

// MergePluginConfigs merges plugin configurations
func (c *PluginConfig) Merge(other *PluginConfig) *PluginConfig {
	result := *c
	if other.Name != "" {
		result.Name = other.Name
	}
	if other.Version != "" {
		result.Version = other.Version
	}
	if other.Config != nil {
		if result.Config == nil {
			result.Config = make(map[string]any)
		}
		for k, v := range other.Config {
			result.Config[k] = v
		}
	}
	return &result
}
