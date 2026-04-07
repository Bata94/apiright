package core

import (
	"context"
)

// Server defines the interface for both HTTP and gRPC servers
type Server interface {
	// Start starts the server with the given context
	Start(ctx context.Context) error

	// Stop gracefully stops the server
	Stop() error

	// RegisterService registers a service with the server
	RegisterService(service any) error

	// Address returns the listening address
	Address() string
}

// Database defines the interface for database operations
type Database interface {
	// Connect establishes a connection to the database
	Connect() error

	// Close closes the database connection
	Close() error

	// Ping checks if the database connection is alive
	Ping() error

	// Migrate runs database migrations
	Migrate() error

	// Connection returns the underlying database connection
	Connection() any
}

// Plugin defines the interface for framework plugins
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Version returns the plugin version
	Version() string

	// Generate executes the plugin's generation logic
	Generate(ctx *GenerationContext) error

	// Validate validates the plugin configuration
	Validate(schema *Schema) error
}

// GenerationContext is defined in context.go

// Middleware defines the interface for HTTP and gRPC middleware
type Middleware interface {
	// Name returns the middleware name
	Name() string

	// Priority returns the middleware priority (lower number = higher priority)
	Priority() int
}

// ContentNegotiator defines the interface for content negotiation
type ContentNegotiator interface {
	// SupportedTypes returns the list of supported content types
	SupportedTypes() []string

	// SerializeResponse serializes data to the specified content type
	SerializeResponse(data any, contentType string) ([]byte, error)

	// DeserializeRequest deserializes data from the specified content type
	DeserializeRequest(data []byte, contentType string, target any) error

	// DetectContentType detects content type from Accept header
	DetectContentType(header string) string
}

// Validator defines the interface for data validation
type Validator interface {
	// Validate validates the given data
	Validate(data any) error

	// ValidateField validates a specific field
	ValidateField(field string, value any) error
}

// Config defines the interface for configuration
type Config interface {
	// Get returns a configuration value
	Get(key string) any

	// GetString returns a string configuration value
	GetString(key string) string

	// GetInt returns an int configuration value
	GetInt(key string) int

	// GetBool returns a bool configuration value
	GetBool(key string) bool

	// Set sets a configuration value
	Set(key string, value any)
}

// Schema represents a database schema
type Schema struct {
	Tables  []Table `json:"tables"`
	Queries []Query `json:"queries"`
	Enums   []Enum  `json:"enums"`
	Types   []Type  `json:"types"`
}

// Table represents a database table
type Table struct {
	Name        string       `json:"name"`
	Columns     []Column     `json:"columns"`
	PrimaryKey  []string     `json:"primary_key"`
	Indexes     []Index      `json:"indexes"`
	ForeignKeys []ForeignKey `json:"foreign_keys"`
}

// Column represents a database column
type Column struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Nullable      bool   `json:"nullable"`
	Default       string `json:"default"`
	AutoIncrement bool   `json:"auto_increment"`
}

// Index represents a database index
type Index struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
}

// ForeignKey represents a foreign key constraint
type ForeignKey struct {
	Name       string   `json:"name"`
	Columns    []string `json:"columns"`
	RefTable   string   `json:"ref_table"`
	RefColumns []string `json:"ref_columns"`
	OnDelete   string   `json:"on_delete"`
	OnUpdate   string   `json:"on_update"`
}

// Query represents a database query
type Query struct {
	Name       string  `json:"name"`
	SQL        string  `json:"sql"`
	ReturnType string  `json:"return_type"`
	Params     []Param `json:"params"`
}

// Param represents a query parameter
type Param struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Enum represents an enum type
type Enum struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// Type represents a custom type
type Type struct {
	Name string `json:"name"`
	SQL  string `json:"sql"`
}

// ServiceMethod represents a service method
type ServiceMethod struct {
	Name        string `json:"name"`
	Input       any    `json:"input"`
	Output      any    `json:"output"`
	Description string `json:"description"`
}

// Service represents a service definition
type Service struct {
	Name    string          `json:"name"`
	Methods []ServiceMethod `json:"methods"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// StatusCode represents HTTP/gRPC status codes
type StatusCode int

const (
	StatusOK StatusCode = iota
	StatusNotFound
	StatusBadRequest
	StatusUnauthorized
	StatusForbidden
	StatusInternalServerError
	StatusConflict
	StatusUnprocessableEntity
)

// ProtoExtension defines the interface for protobuf extensions
type ProtoExtension interface {
	// Name returns the extension name
	Name() string

	// ProtoFiles returns the paths to proto files provided by this extension
	ProtoFiles() []string

	// Imports returns additional imports required by this extension
	Imports() []string
}

// MiddlewareProvider defines the interface for plugins that provide middleware
type MiddlewareProvider interface {
	// Middleware returns a list of middleware instances
	Middleware() []Middleware
}
