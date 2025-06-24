package openapi_test

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/bata94/apiright/pkg/openapi"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id" description:"Unique user identifier"`
	Username  string    `json:"username" validate:"required" description:"Username (3-50 characters)" example:"johndoe"`
	Email     string    `json:"email" validate:"required,email" description:"Email address" example:"john@example.com"`
	FirstName string    `json:"first_name" description:"First name" example:"John"`
	LastName  string    `json:"last_name" description:"Last name" example:"Doe"`
	Avatar    string    `json:"avatar,omitempty" description:"Avatar URL" example:"https://example.com/avatar.jpg"`
	IsActive  bool      `json:"is_active" description:"Whether the user is active" example:"true"`
	CreatedAt time.Time `json:"created_at" description:"Account creation timestamp"`
	UpdatedAt time.Time `json:"updated_at" description:"Last update timestamp"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Username  string `json:"username" validate:"required,min=3,max=50" description:"Username (3-50 characters)" example:"johndoe"`
	Email     string `json:"email" validate:"required,email" description:"Email address" example:"john@example.com"`
	FirstName string `json:"first_name" validate:"required" description:"First name" example:"John"`
	LastName  string `json:"last_name" validate:"required" description:"Last name" example:"Doe"`
	Password  string `json:"password" validate:"required,min=8" description:"Password (minimum 8 characters)" example:"secretpassword"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	FirstName *string `json:"first_name,omitempty" description:"First name" example:"John"`
	LastName  *string `json:"last_name,omitempty" description:"Last name" example:"Doe"`
	Avatar    *string `json:"avatar,omitempty" description:"Avatar URL" example:"https://example.com/avatar.jpg"`
}

// UserListResponse represents a paginated list of users
type UserListResponse struct {
	Users      []User `json:"users" description:"List of users"`
	Page       int    `json:"page" description:"Current page number"`
	PerPage    int    `json:"per_page" description:"Items per page"`
	Total      int    `json:"total" description:"Total number of users"`
	TotalPages int    `json:"total_pages" description:"Total number of pages"`
}

// UserStatsResponse represents user statistics
type UserStatsResponse struct {
	UserID        string `json:"user_id" description:"User ID"`
	LoginCount    int    `json:"login_count" description:"Number of logins"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty" description:"Last login timestamp"`
	PostCount     int    `json:"post_count" description:"Number of posts created"`
	CommentCount  int    `json:"comment_count" description:"Number of comments made"`
	FollowerCount int    `json:"follower_count" description:"Number of followers"`
	FollowingCount int   `json:"following_count" description:"Number of users being followed"`
}

// ExampleBasicUsage demonstrates basic usage of the OpenAPI generator
func ExampleBasicUsage() {
	// Create a new generator with default configuration
	config := openapi.DefaultConfig()
	config.Title = "User Management API"
	config.Description = "A comprehensive API for managing users in the system"
	config.Version = "2.1.0"
	config.OutputDir = "./docs"

	// Add contact information
	config.Contact = &openapi.Contact{
		Name:  "API Support",
		Email: "support@example.com",
		URL:   "https://example.com/support",
	}

	// Add license information
	config.License = &openapi.License{
		Name: "MIT",
		URL:  "https://opensource.org/licenses/MIT",
	}

	generator := openapi.NewGenerator(config)

	// Add servers
	generator.AddServer(openapi.Server{
		URL:         "https://api.example.com/v2",
		Description: "Production server",
	})
	generator.AddServer(openapi.Server{
		URL:         "https://staging-api.example.com/v2",
		Description: "Staging server",
	})
	generator.AddServer(openapi.Server{
		URL:         "http://localhost:8080/v2",
		Description: "Development server",
	})

	// Add security schemes
	generator.AddSecurityScheme("BearerAuth", openapi.SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "JWT Bearer token authentication",
	})

	generator.AddSecurityScheme("ApiKeyAuth", openapi.SecurityScheme{
		Type:        "apiKey",
		In:          "header",
		Name:        "X-API-Key",
		Description: "API key authentication",
	})

	// Add tags
	generator.AddTag(openapi.Tag{
		Name:        "users",
		Description: "User management operations",
	})
	generator.AddTag(openapi.Tag{
		Name:        "auth",
		Description: "Authentication operations",
	})
	generator.AddTag(openapi.Tag{
		Name:        "health",
		Description: "Health check operations",
	})

	// Add global security requirement
	generator.AddGlobalSecurity(openapi.SecurityRequirement{
		"BearerAuth": []string{},
	})

	fmt.Println("Basic generator created successfully")
}

// ExampleAdvancedEndpoints demonstrates advanced endpoint configuration
func ExampleAdvancedEndpoints() {
	generator := openapi.QuickStart("Advanced API", "Advanced API example", "1.0.0")

	// Health check endpoint
	healthBuilder := openapi.NewEndpointBuilder().
		Summary("Health check").
		Description("Returns the health status of the API").
		Tags("health").
		Response(200, "API is healthy", "application/json", reflect.TypeOf(openapi.SuccessResponse{})).
		Response(503, "API is unhealthy", "application/json", reflect.TypeOf(openapi.ErrorResponse{}))

	generator.AddEndpointWithBuilder("GET", "/health", healthBuilder)

	// List users with advanced filtering and pagination
	listUsersBuilder := openapi.NewEndpointBuilder().
		Summary("List users").
		Description("Retrieve a paginated list of users with optional filtering and sorting").
		Tags("users").
		QueryParam("page", "Page number (1-based)", false, reflect.TypeOf(1)).
		QueryParam("limit", "Number of items per page (1-100)", false, reflect.TypeOf(20)).
		QueryParam("sort", "Sort field and direction (e.g., 'username:asc', 'created_at:desc')", false, reflect.TypeOf("")).
		QueryParam("filter", "Filter users by status (active, inactive, all)", false, reflect.TypeOf("")).
		QueryParam("search", "Search users by username or email", false, reflect.TypeOf("")).
		HeaderParam("X-Request-ID", "Unique request identifier", false, reflect.TypeOf("")).
		Security(openapi.SecurityRequirement{"BearerAuth": []string{}}).
		Response(200, "List of users", "application/json", reflect.TypeOf(UserListResponse{})).
		Response(400, "Invalid query parameters", "application/json", reflect.TypeOf(openapi.ErrorResponse{})).
		Response(401, "Unauthorized", "application/json", reflect.TypeOf(openapi.ErrorResponse{})).
		Response(500, "Internal server error", "application/json", reflect.TypeOf(openapi.ErrorResponse{}))

	generator.AddEndpointWithBuilder("GET", "/users", listUsersBuilder)

	// Create user endpoint
	createUserBuilder := openapi.NewEndpointBuilder().
		Summary("Create user").
		Description("Create a new user account").
		Tags("users").
		RequestType(reflect.TypeOf(CreateUserRequest{})).
		Response(201, "User created successfully", "application/json", reflect.TypeOf(User{})).
		Response(400, "Invalid input data", "application/json", reflect.TypeOf(openapi.ErrorResponse{})).
		Response(409, "User already exists", "application/json", reflect.TypeOf(openapi.ErrorResponse{})).
		Response(422, "Validation error", "application/json", reflect.TypeOf(openapi.ErrorResponse{}))

	generator.AddEndpointWithBuilder("POST", "/users", createUserBuilder)

	// Get user by ID
	getUserBuilder := openapi.NewEndpointBuilder().
		Summary("Get user by ID").
		Description("Retrieve a specific user by their unique identifier").
		Tags("users").
		PathParam("id", "User ID", reflect.TypeOf("")).
		Security(openapi.SecurityRequirement{"BearerAuth": []string{}}).
		Response(200, "User details", "application/json", reflect.TypeOf(User{})).
		Response(404, "User not found", "application/json", reflect.TypeOf(openapi.ErrorResponse{})).
		Response(401, "Unauthorized", "application/json", reflect.TypeOf(openapi.ErrorResponse{}))

	generator.AddEndpointWithBuilder("GET", "/users/{id}", getUserBuilder)

	// Update user
	updateUserBuilder := openapi.NewEndpointBuilder().
		Summary("Update user").
		Description("Update an existing user's information").
		Tags("users").
		PathParam("id", "User ID", reflect.TypeOf("")).
		RequestType(reflect.TypeOf(UpdateUserRequest{})).
		Security(openapi.SecurityRequirement{"BearerAuth": []string{}}).
		Response(200, "User updated successfully", "application/json", reflect.TypeOf(User{})).
		Response(400, "Invalid input data", "application/json", reflect.TypeOf(openapi.ErrorResponse{})).
		Response(404, "User not found", "application/json", reflect.TypeOf(openapi.ErrorResponse{})).
		Response(422, "Validation error", "application/json", reflect.TypeOf(openapi.ErrorResponse{}))

	generator.AddEndpointWithBuilder("PUT", "/users/{id}", updateUserBuilder)

	// Delete user
	deleteUserBuilder := openapi.NewEndpointBuilder().
		Summary("Delete user").
		Description("Delete a user account").
		Tags("users").
		PathParam("id", "User ID", reflect.TypeOf("")).
		Security(openapi.SecurityRequirement{"BearerAuth": []string{}}).
		Response(204, "User deleted successfully", "", nil).
		Response(404, "User not found", "application/json", reflect.TypeOf(openapi.ErrorResponse{})).
		Response(409, "Cannot delete user with active dependencies", "application/json", reflect.TypeOf(openapi.ErrorResponse{}))

	generator.AddEndpointWithBuilder("DELETE", "/users/{id}", deleteUserBuilder)

	// Get user statistics
	userStatsBuilder := openapi.NewEndpointBuilder().
		Summary("Get user statistics").
		Description("Retrieve detailed statistics for a specific user").
		Tags("users", "analytics").
		PathParam("id", "User ID", reflect.TypeOf("")).
		QueryParam("period", "Time period for statistics (7d, 30d, 90d, 1y)", false, reflect.TypeOf("")).
		Security(openapi.SecurityRequirement{"BearerAuth": []string{}}).
		Response(200, "User statistics", "application/json", reflect.TypeOf(UserStatsResponse{})).
		Response(404, "User not found", "application/json", reflect.TypeOf(openapi.ErrorResponse{}))

	generator.AddEndpointWithBuilder("GET", "/users/{id}/stats", userStatsBuilder)

	fmt.Println("Advanced endpoints added successfully")
}

// ExampleSchemaGeneration demonstrates schema generation from Go types
func ExampleSchemaGeneration() {
	generator := openapi.QuickStart("Schema Demo", "Schema generation demo", "1.0.0")
	schemaGen := generator.GetSchemaGenerator()

	// Generate schema for User type
	userSchema := schemaGen.GenerateSchemaWithName(reflect.TypeOf(User{}), "User")
	fmt.Printf("Generated User schema: %+v\n", userSchema)

	// Generate schema for CreateUserRequest
	createReqSchema := schemaGen.GenerateSchemaWithName(reflect.TypeOf(CreateUserRequest{}), "CreateUserRequest")
	fmt.Printf("Generated CreateUserRequest schema: %+v\n", createReqSchema)

	// Get all generated schemas
	schemas := schemaGen.GetSchemas()
	fmt.Printf("Total schemas generated: %d\n", len(schemas))

	for name, schema := range schemas {
		fmt.Printf("Schema %s: %d properties\n", name, len(schema.Properties))
	}
}

// ExampleFileGeneration demonstrates generating documentation files
func ExampleFileGeneration() {
	generator := openapi.QuickStart("File Generation Demo", "Demo for file generation", "1.0.0")

	// Add some sample endpoints
	openapi.HealthCheckEndpoint(generator)
	openapi.AuthEndpoints(generator)

	// Create writer
	writer := openapi.NewWriter(generator)

	// Generate all files
	if err := writer.WriteFiles(); err != nil {
		log.Printf("Error generating files: %v", err)
		return
	}

	// Generate specific formats
	_, err := generator.GenerateSpec()
	if err != nil {
		log.Printf("Error generating spec: %v", err)
		return
	}

	// Write JSON to specific file
	if err := writer.WriteToFile("./custom-api.json", "json"); err != nil {
		log.Printf("Error writing JSON: %v", err)
		return
	}

	// Write YAML to specific file
	if err := writer.WriteToFile("./custom-api.yaml", "yaml"); err != nil {
		log.Printf("Error writing YAML: %v", err)
		return
	}

	// Generate markdown documentation
	if err := writer.GenerateMarkdownDocs(); err != nil {
		log.Printf("Error generating markdown: %v", err)
		return
	}

	// Get list of generated files
	files := writer.GetGeneratedFiles()
	fmt.Printf("Generated files: %v\n", files)

	// Get statistics
	stats := generator.GetStatistics()
	fmt.Printf("Documentation statistics: %+v\n", stats)

	fmt.Println("File generation completed successfully")
}

// ExampleCRUDGeneration demonstrates automatic CRUD endpoint generation
func ExampleCRUDGeneration() {
	generator := openapi.QuickStart("CRUD Demo", "CRUD generation demo", "1.0.0")

	// Generate complete CRUD endpoints for User resource
	if err := openapi.CRUDEndpoints(generator, "/users", reflect.TypeOf(User{}), "users"); err != nil {
		log.Printf("Error generating CRUD endpoints: %v", err)
		return
	}

	// Generate CRUD endpoints for another resource
	type Product struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		InStock     bool    `json:"in_stock"`
	}

	if err := openapi.CRUDEndpoints(generator, "/products", reflect.TypeOf(Product{}), "products"); err != nil {
		log.Printf("Error generating product CRUD endpoints: %v", err)
		return
	}

	// List all endpoints
	endpoints := generator.ListEndpoints()
	fmt.Printf("Generated endpoints:\n")
	for path, methods := range endpoints {
		fmt.Printf("  %s: %v\n", path, methods)
	}

	fmt.Println("CRUD generation completed successfully")
}

// ExampleCustomValidation demonstrates custom validation and examples
func ExampleCustomValidation() {
	generator := openapi.QuickStart("Validation Demo", "Custom validation demo", "1.0.0")

	// Create endpoint with custom validation
	builder := openapi.NewEndpointBuilder().
		Summary("Create user with validation").
		Description("Create a new user with comprehensive validation").
		Tags("users").
		RequestType(reflect.TypeOf(CreateUserRequest{})).
		RequestExample(CreateUserRequest{
			Username:  "johndoe",
			Email:     "john@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Password:  "secretpassword123",
		}).
		Response(201, "User created", "application/json", reflect.TypeOf(User{})).
		ResponseExample(User{
			ID:        "user_123",
			Username:  "johndoe",
			Email:     "john@example.com",
			FirstName: "John",
			LastName:  "Doe",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})

	generator.AddEndpointWithBuilder("POST", "/users", builder)

	// Validate a request
	request := CreateUserRequest{
		Username:  "johndoe",
		Email:     "john@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Password:  "secret",
	}

	if err := openapi.ValidateRequest(generator, "POST", "/users", request); err != nil {
		fmt.Printf("Validation error: %v\n", err)
	} else {
		fmt.Println("Request validation passed")
	}

	fmt.Println("Custom validation example completed")
}

// ExampleMiddleware demonstrates the OpenAPI middleware
func ExampleMiddleware() {
	generator := openapi.QuickStart("Middleware Demo", "Middleware demo", "1.0.0")

	// Create middleware
	_ = openapi.Middleware(generator)

	// This would typically be used with an HTTP router
	// router.Use(middleware)

	fmt.Printf("Middleware created for generator: %s\n", generator.GetSpec().Info.Title)
}

// ExampleCompleteWorkflow demonstrates a complete workflow
func ExampleCompleteWorkflow() {
	// 1. Create generator
	generator := openapi.QuickStart("Complete API", "A complete API example", "1.0.0")

	// 2. Add authentication
	openapi.AuthEndpoints(generator)

	// 3. Add health check
	openapi.HealthCheckEndpoint(generator)

	// 4. Add CRUD endpoints
	openapi.CRUDEndpoints(generator, "/users", reflect.TypeOf(User{}), "users")

	// 5. Add custom endpoints
	customBuilder := openapi.NewEndpointBuilder().
		Summary("Bulk update users").
		Description("Update multiple users in a single request").
		Tags("users", "bulk").
		RequestType(reflect.TypeOf([]UpdateUserRequest{})).
		Security(openapi.SecurityRequirement{"BearerAuth": []string{}}).
		Response(200, "Users updated", "application/json", reflect.TypeOf([]User{}))

	generator.AddEndpointWithBuilder("PATCH", "/users/bulk", customBuilder)

	// 6. Generate documentation
	writer := openapi.NewWriter(generator)
	if err := writer.WriteFiles(); err != nil {
		log.Printf("Error generating documentation: %v", err)
		return
	}

	// 7. Get statistics
	stats := generator.GetStatistics()
	fmt.Printf("Complete workflow statistics:\n")
	fmt.Printf("  Total endpoints: %d\n", stats.TotalEndpoints)
	fmt.Printf("  Total schemas: %d\n", stats.TotalSchemas)
	fmt.Printf("  Endpoints by method: %v\n", stats.EndpointsByMethod)
	fmt.Printf("  Endpoints by tag: %v\n", stats.EndpointsByTag)

	fmt.Println("Complete workflow finished successfully")
}