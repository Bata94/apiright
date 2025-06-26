// Package openapi provides comprehensive OpenAPI 3.0 documentation generation
// for Go web applications. It supports automatic schema generation from Go types,
// endpoint documentation, and multiple output formats including JSON, YAML, and HTML.
package openapi

import (
	"fmt"
	"net/http"
	"reflect"
	"time"
)

// Version of the OpenAPI generator package
const Version = "1.0.0"

// Quick start functions for common use cases

// TODO: Add a Logger to the Package and handle panics more gracefully

func NewBasicGenerator(title, description, version string) *Generator {
	config := DefaultConfig()
	config.Title = title
	config.Description = description
	config.Version = version

	return NewGenerator(config)
}

// QuickStart creates a generator with sensible defaults and basic configuration
func QuickStart(title, description, version string) *Generator {
	config := DefaultConfig()
	config.Title = title
	config.Description = description
	config.Version = version

	// Add common servers
	config.Servers = []Server{
		{
			URL:         "http://localhost:8080",
			Description: "Development server",
		},
		{
			URL:         "https://api.example.com",
			Description: "Production server",
		},
	}

	// Add common security schemes
	config.SecuritySchemes = map[string]SecurityScheme{
		"BearerAuth": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "JWT Bearer token authentication",
		},
		"ApiKeyAuth": {
			Type:        "apiKey",
			In:          "header",
			Name:        "X-API-Key",
			Description: "API key authentication",
		},
	}

	// Add common tags
	config.Tags = []Tag{
		{
			Name:        "auth",
			Description: "Authentication endpoints",
		},
		{
			Name:        "users",
			Description: "User management endpoints",
		},
		{
			Name:        "health",
			Description: "Health check endpoints",
		},
	}

	return NewGenerator(config)
}

// SimpleEndpoint creates a simple endpoint with minimal configuration
func SimpleEndpoint(method, path, summary string) *EndpointBuilder {
	return NewEndpointBuilder().
		Summary(summary).
		Response(200, "Successful response", "application/json", nil)
}

// RESTEndpoint creates a REST endpoint with standard responses
func RESTEndpoint(method, path, summary string, resourceType reflect.Type) *EndpointBuilder {
	builder := NewEndpointBuilder().
		Summary(summary).
		Tags("api")

	// Add standard responses based on method
	switch method {
	case "GET":
		builder.Response(200, "Successful response", "application/json", resourceType).
			Response(404, "Resource not found", "application/json", reflect.TypeOf(ErrorResponse{}))
	case "POST":
		builder.RequestType(resourceType).
			Response(201, "Resource created", "application/json", resourceType).
			Response(400, "Invalid input", "application/json", reflect.TypeOf(ErrorResponse{}))
	case "PUT":
		builder.RequestType(resourceType).
			Response(200, "Resource updated", "application/json", resourceType).
			Response(404, "Resource not found", "application/json", reflect.TypeOf(ErrorResponse{}))
	case "DELETE":
		builder.Response(204, "Resource deleted", "", nil).
			Response(404, "Resource not found", "application/json", reflect.TypeOf(ErrorResponse{}))
	}

	// Add common error responses
	builder.Response(401, "Unauthorized", "application/json", reflect.TypeOf(ErrorResponse{})).
		Response(500, "Internal server error", "application/json", reflect.TypeOf(ErrorResponse{}))

	return builder
}

// CRUDEndpoints creates a complete set of CRUD endpoints for a resource
func CRUDEndpoints(generator *Generator, basePath string, resourceType reflect.Type, resourceName string) error {
	// List resources
	listBuilder := RESTEndpoint("GET", basePath, fmt.Sprintf("List %s", resourceName),
		reflect.TypeOf([]interface{}{})).
		Tags(resourceName).
		QueryParam("page", "Page number", false, reflect.TypeOf(0)).
		QueryParam("limit", "Items per page", false, reflect.TypeOf(0))

	if err := generator.AddEndpointWithBuilder("GET", basePath, listBuilder); err != nil {
		return err
	}

	// Create resource
	createBuilder := RESTEndpoint("POST", basePath, fmt.Sprintf("Create %s", resourceName), resourceType).
		Tags(resourceName)

	if err := generator.AddEndpointWithBuilder("POST", basePath, createBuilder); err != nil {
		return err
	}

	// Get resource by ID
	getPath := basePath + "/{id}"
	getBuilder := RESTEndpoint("GET", getPath, fmt.Sprintf("Get %s by ID", resourceName), resourceType).
		Tags(resourceName).
		PathParam("id", "Resource ID", reflect.TypeOf(""))

	if err := generator.AddEndpointWithBuilder("GET", getPath, getBuilder); err != nil {
		return err
	}

	// Update resource
	updateBuilder := RESTEndpoint("PUT", getPath, fmt.Sprintf("Update %s", resourceName), resourceType).
		Tags(resourceName).
		PathParam("id", "Resource ID", reflect.TypeOf(""))

	if err := generator.AddEndpointWithBuilder("PUT", getPath, updateBuilder); err != nil {
		return err
	}

	// Delete resource
	deleteBuilder := RESTEndpoint("DELETE", getPath, fmt.Sprintf("Delete %s", resourceName), nil).
		Tags(resourceName).
		PathParam("id", "Resource ID", reflect.TypeOf(""))

	if err := generator.AddEndpointWithBuilder("DELETE", getPath, deleteBuilder); err != nil {
		return err
	}

	return nil
}

// HealthCheckEndpoint creates a standard health check endpoint
func HealthCheckEndpoint(generator *Generator) error {
	type HealthResponse struct {
		Status    string    `json:"status" description:"Health status"`
		Timestamp time.Time `json:"timestamp" description:"Check timestamp"`
		Version   string    `json:"version,omitempty" description:"Application version"`
		Uptime    string    `json:"uptime,omitempty" description:"Application uptime"`
	}

	builder := NewEndpointBuilder().
		Summary("Health check").
		Description("Returns the health status of the API").
		Tags("health").
		Response(200, "API is healthy", "application/json", reflect.TypeOf(HealthResponse{})).
		Response(503, "API is unhealthy", "application/json", reflect.TypeOf(ErrorResponse{}))

	return generator.AddEndpointWithBuilder("GET", "/health", builder)
}

// AuthEndpoints creates standard authentication endpoints
func AuthEndpoints(generator *Generator) error {
	type LoginRequest struct {
		Username string `json:"username" validate:"required" description:"Username or email"`
		Password string `json:"password" validate:"required" description:"Password"`
	}

	type UserInfo struct {
		ID       string `json:"id" description:"User ID"`
		Username string `json:"username" description:"Username"`
		Email    string `json:"email" description:"Email address"`
	}

	type LoginResponse struct {
		Token     string    `json:"token" description:"JWT access token"`
		ExpiresAt time.Time `json:"expires_at" description:"Token expiration time"`
		User      UserInfo  `json:"user" description:"User information"`
	}

	// Login endpoint
	loginBuilder := NewEndpointBuilder().
		Summary("User login").
		Description("Authenticate user and return access token").
		Tags("auth").
		RequestType(reflect.TypeOf(LoginRequest{})).
		Response(200, "Login successful", "application/json", reflect.TypeOf(LoginResponse{})).
		Response(401, "Invalid credentials", "application/json", reflect.TypeOf(ErrorResponse{})).
		Response(400, "Invalid request", "application/json", reflect.TypeOf(ErrorResponse{}))

	if err := generator.AddEndpointWithBuilder("POST", "/auth/login", loginBuilder); err != nil {
		return err
	}

	// Logout endpoint
	logoutBuilder := NewEndpointBuilder().
		Summary("User logout").
		Description("Invalidate the current access token").
		Tags("auth").
		Security(SecurityRequirement{"BearerAuth": []string{}}).
		Response(200, "Logout successful", "application/json", nil).
		Response(401, "Unauthorized", "application/json", reflect.TypeOf(ErrorResponse{}))

	if err := generator.AddEndpointWithBuilder("POST", "/auth/logout", logoutBuilder); err != nil {
		return err
	}

	// Get current user endpoint
	meBuilder := NewEndpointBuilder().
		Summary("Get current user").
		Description("Get information about the currently authenticated user").
		Tags("auth", "users").
		Security(SecurityRequirement{"BearerAuth": []string{}}).
		Response(200, "User information", "application/json", reflect.TypeOf(UserInfo{})).
		Response(401, "Unauthorized", "application/json", reflect.TypeOf(ErrorResponse{}))

	if err := generator.AddEndpointWithBuilder("GET", "/auth/me", meBuilder); err != nil {
		return err
	}

	return nil
}

// GenerateAndServe generates documentation and starts a simple HTTP server to serve it
func GenerateAndServe(generator *Generator, port int) error {
	writer := NewWriter(generator)

	// Generate documentation files
	if err := writer.WriteFiles(); err != nil {
		return fmt.Errorf("failed to generate documentation: %w", err)
	}

	// Serve the documentation
	fs := http.FileServer(http.Dir(generator.config.OutputDir))
	http.Handle("/", fs)

	fmt.Printf("OpenAPI documentation server starting on port %d\n", port)
	fmt.Printf("Visit http://localhost:%d to view the documentation\n", port)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

// Middleware creates HTTP middleware that automatically documents endpoints
// func Middleware(generator *Generator) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			// Auto-document the endpoint if not already documented
// 			if _, exists := generator.GetEndpoint(r.Method, r.URL.Path); !exists {
// 				builder := SimpleEndpoint(r.Method, r.URL.Path,
// 					fmt.Sprintf("%s %s", r.Method, r.URL.Path))
// 					err := generator.AddEndpointWithBuilder(r.Method, r.URL.Path, builder)
// 					if err != nil {
// 						panic(fmt.Errorf("middleware failed to document endpoint: %w", err))
// 					}
// 			}
//
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }

// ValidateRequest validates a request against the OpenAPI specification
func ValidateRequest(generator *Generator, method, path string, body interface{}) error {
	options, exists := generator.GetEndpoint(method, path)
	if !exists {
		return fmt.Errorf("endpoint not documented: %s %s", method, path)
	}

	if options.RequestType != nil && body != nil {
		bodyType := reflect.TypeOf(body)
		if bodyType != options.RequestType {
			return fmt.Errorf("request body type mismatch: expected %v, got %v",
				options.RequestType, bodyType)
		}
	}

	return nil
}

// Helper functions for common patterns

// StringPtr returns a pointer to a string (useful for optional schema fields)
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}

// Float64Ptr returns a pointer to a float64
func Float64Ptr(f float64) *float64 {
	return &f
}

// BoolPtr returns a pointer to a bool
func BoolPtr(b bool) *bool {
	return &b
}

// CommonHeaders returns common HTTP headers for API responses
func CommonHeaders() map[string]HeaderInfo {
	return map[string]HeaderInfo{
		"X-Request-ID": {
			Description: "Unique request identifier",
			Type:        reflect.TypeOf(""),
			Example:     "123e4567-e89b-12d3-a456-426614174000",
		},
		"X-Rate-Limit-Remaining": {
			Description: "Number of requests remaining in the current window",
			Type:        reflect.TypeOf(0),
			Example:     100,
		},
		"X-Rate-Limit-Reset": {
			Description: "Unix timestamp when the rate limit resets",
			Type:        reflect.TypeOf(int64(0)),
			Example:     1640995200,
		},
	}
}

// PaginationParams returns common pagination parameters
func PaginationParams() []ParameterInfo {
	return []ParameterInfo{
		{
			Name:        "page",
			Description: "Page number (1-based)",
			Required:    false,
			Type:        reflect.TypeOf(0),
			Example:     1,
		},
		{
			Name:        "limit",
			Description: "Number of items per page",
			Required:    false,
			Type:        reflect.TypeOf(0),
			Example:     20,
		},
		{
			Name:        "sort",
			Description: "Sort field and direction (e.g., 'name:asc', 'created_at:desc')",
			Required:    false,
			Type:        reflect.TypeOf(""),
			Example:     "created_at:desc",
		},
	}
}

// FilterParams returns common filtering parameters
func FilterParams() []ParameterInfo {
	return []ParameterInfo{
		{
			Name:        "q",
			Description: "Search query",
			Required:    false,
			Type:        reflect.TypeOf(""),
			Example:     "search term",
		},
		{
			Name:        "filter",
			Description: "Filter expression",
			Required:    false,
			Type:        reflect.TypeOf(""),
			Example:     "status:active",
		},
	}
}

// Example usage and documentation

// ExampleUsage demonstrates how to use the OpenAPI generator
func ExampleUsage() {
	// Create a generator with quick start
	generator := QuickStart("My API", "A sample API", "1.0.0")

	// Add health check endpoint
	err := HealthCheckEndpoint(generator)
	if err != nil {
		panic(fmt.Errorf("error in ExampleUsage, adding HealthCheckEndpoint: %w", err))
	}

	// Add authentication endpoints
	err = AuthEndpoints(generator)
	if err != nil {
		panic(fmt.Errorf("error in ExampleUsage, adding AuthEndpoints: %w", err))
	}

	// Define a user type
	type User struct {
		ID       string    `json:"id" description:"User ID"`
		Username string    `json:"username" description:"Username"`
		Email    string    `json:"email" description:"Email address"`
		Created  time.Time `json:"created_at" description:"Creation timestamp"`
	}

	// Add CRUD endpoints for users
	err = CRUDEndpoints(generator, "/users", reflect.TypeOf(User{}), "users")
	if err != nil {
		panic(fmt.Errorf("error in ExampleUsage, adding CRUDEndpoints: %w", err))
	}

	// Add a custom endpoint
	customBuilder := NewEndpointBuilder().
		Summary("Get user statistics").
		Description("Returns statistics about user activity").
		Tags("users", "analytics").
		PathParam("id", "User ID", reflect.TypeOf("")).
		QueryParam("period", "Time period for statistics", false, reflect.TypeOf("")).
		Security(SecurityRequirement{"BearerAuth": []string{}}).
		Response(200, "User statistics", "application/json", nil)

	err = generator.AddEndpointWithBuilder("GET", "/users/{id}/stats", customBuilder)
	if err != nil {
		panic(fmt.Errorf("error in ExampleUsage, adding Endpoint: %w", err))
	}

	// Generate and write documentation
	writer := NewWriter(generator)
	err = writer.WriteFiles()
	if err != nil {
		panic(fmt.Errorf("error in ExampleUsage, writing files: %w", err))
	}

	// Or serve the documentation
	// GenerateAndServe(generator, 8080)
}
