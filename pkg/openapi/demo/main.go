package main

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

func main() {
	fmt.Println("🚀 OpenAPI Documentation Generator Demo")
	fmt.Println("=====================================")

	// Create a generator with quick start
	generator := openapi.QuickStart(
		"User Management API",
		"A comprehensive API for managing users in the system",
		"2.1.0",
	)

	// Add contact information
	generator.GetSpec().Info.Contact = &openapi.Contact{
		Name:  "API Support Team",
		Email: "support@example.com",
		URL:   "https://example.com/support",
	}

	// Add license information
	generator.GetSpec().Info.License = &openapi.License{
		Name: "MIT",
		URL:  "https://opensource.org/licenses/MIT",
	}

	// Add additional servers
	generator.AddServer(openapi.Server{
		URL:         "https://api.example.com/v2",
		Description: "Production server",
	})
	generator.AddServer(openapi.Server{
		URL:         "https://staging-api.example.com/v2",
		Description: "Staging server",
	})

	// Add additional tags
	generator.AddTag(openapi.Tag{
		Name:        "analytics",
		Description: "Analytics and reporting endpoints",
	})

	fmt.Println("✅ Basic configuration completed")

	// Add health check endpoint
	if err := openapi.HealthCheckEndpoint(generator); err != nil {
		log.Fatal("Failed to add health check endpoint:", err)
	}
	fmt.Println("✅ Health check endpoint added")

	// Add authentication endpoints
	if err := openapi.AuthEndpoints(generator); err != nil {
		log.Fatal("Failed to add auth endpoints:", err)
	}
	fmt.Println("✅ Authentication endpoints added")

	// Add CRUD endpoints for users
	if err := openapi.CRUDEndpoints(generator, "/users", reflect.TypeOf(User{}), "users"); err != nil {
		log.Fatal("Failed to add CRUD endpoints:", err)
	}
	fmt.Println("✅ User CRUD endpoints added")

	// Add a custom advanced endpoint
	advancedSearchBuilder := openapi.NewEndpointBuilder().
		Summary("Advanced user search").
		Description("Search users with advanced filtering, sorting, and pagination").
		Tags("users", "search").
		QueryParam("q", "Search query (searches username, email, first name, last name)", false, reflect.TypeOf("")).
		QueryParam("page", "Page number (1-based)", false, reflect.TypeOf(1)).
		QueryParam("limit", "Number of items per page (1-100)", false, reflect.TypeOf(20)).
		QueryParam("sort", "Sort field and direction (e.g., 'username:asc', 'created_at:desc')", false, reflect.TypeOf("")).
		QueryParam("filter", "Filter users by status (active, inactive, all)", false, reflect.TypeOf("")).
		QueryParam("created_after", "Filter users created after this date (ISO 8601)", false, reflect.TypeOf("")).
		QueryParam("created_before", "Filter users created before this date (ISO 8601)", false, reflect.TypeOf("")).
		HeaderParam("X-Request-ID", "Unique request identifier for tracking", false, reflect.TypeOf("")).
		Security(openapi.SecurityRequirement{"BearerAuth": []string{}}).
		Response(200, "Search results with pagination", "application/json", reflect.TypeOf(UserListResponse{})).
		Response(400, "Invalid query parameters", "application/json", reflect.TypeOf(openapi.ErrorResponse{})).
		Response(401, "Unauthorized - invalid or missing token", "application/json", reflect.TypeOf(openapi.ErrorResponse{})).
		Response(422, "Validation error in query parameters", "application/json", reflect.TypeOf(openapi.ErrorResponse{})).
		Response(500, "Internal server error", "application/json", reflect.TypeOf(openapi.ErrorResponse{}))

	if err := generator.AddEndpointWithBuilder("GET", "/users/search", advancedSearchBuilder); err != nil {
		log.Fatal("Failed to add advanced search endpoint:", err)
	}
	fmt.Println("✅ Advanced search endpoint added")

	// Add user statistics endpoint
	userStatsBuilder := openapi.NewEndpointBuilder().
		Summary("Get user statistics").
		Description("Retrieve detailed statistics and analytics for a specific user").
		Tags("users", "analytics").
		PathParam("id", "User ID", reflect.TypeOf("")).
		QueryParam("period", "Time period for statistics (7d, 30d, 90d, 1y, all)", false, reflect.TypeOf("")).
		QueryParam("include_details", "Include detailed breakdown of statistics", false, reflect.TypeOf(false)).
		Security(openapi.SecurityRequirement{"BearerAuth": []string{}}).
		Response(200, "User statistics", "application/json", nil).
		Response(404, "User not found", "application/json", reflect.TypeOf(openapi.ErrorResponse{})).
		Response(403, "Forbidden - insufficient permissions", "application/json", reflect.TypeOf(openapi.ErrorResponse{}))

	if err := generator.AddEndpointWithBuilder("GET", "/users/{id}/stats", userStatsBuilder); err != nil {
		log.Fatal("Failed to add user stats endpoint:", err)
	}
	fmt.Println("✅ User statistics endpoint added")

	// Add bulk operations endpoint
	bulkUpdateBuilder := openapi.NewEndpointBuilder().
		Summary("Bulk update users").
		Description("Update multiple users in a single request for efficient batch operations").
		Tags("users", "bulk").
		RequestType(reflect.TypeOf([]UpdateUserRequest{})).
		RequestExample([]UpdateUserRequest{
			{
				FirstName: openapi.StringPtr("John"),
				LastName:  openapi.StringPtr("Doe"),
			},
			{
				Avatar: openapi.StringPtr("https://example.com/new-avatar.jpg"),
			},
		}).
		Security(openapi.SecurityRequirement{"BearerAuth": []string{}}).
		Response(200, "Users updated successfully", "application/json", reflect.TypeOf([]User{})).
		Response(400, "Invalid request data", "application/json", reflect.TypeOf(openapi.ErrorResponse{})).
		Response(422, "Validation errors in request data", "application/json", reflect.TypeOf(openapi.ErrorResponse{}))

	if err := generator.AddEndpointWithBuilder("PATCH", "/users/bulk", bulkUpdateBuilder); err != nil {
		log.Fatal("Failed to add bulk update endpoint:", err)
	}
	fmt.Println("✅ Bulk update endpoint added")

	// Generate the documentation
	fmt.Println("\n📝 Generating documentation...")

	writer := openapi.NewWriter(generator)
	if err := writer.WriteFiles(); err != nil {
		log.Fatal("Failed to generate documentation:", err)
	}
	fmt.Println("✅ Documentation files generated")

	// Generate markdown documentation
	if err := writer.GenerateMarkdownDocs(); err != nil {
		log.Fatal("Failed to generate markdown documentation:", err)
	}
	fmt.Println("✅ Markdown documentation generated")

	// Get and display statistics
	stats := generator.GetStatistics()
	fmt.Println("\n📊 Documentation Statistics:")
	fmt.Printf("   Total endpoints: %d\n", stats.TotalEndpoints)
	fmt.Printf("   Total schemas: %d\n", stats.TotalSchemas)
	fmt.Printf("   Endpoints by method:\n")
	for method, count := range stats.EndpointsByMethod {
		fmt.Printf("     %s: %d\n", method, count)
	}
	fmt.Printf("   Endpoints by tag:\n")
	for tag, count := range stats.EndpointsByTag {
		fmt.Printf("     %s: %d\n", tag, count)
	}

	// List all generated files
	files := writer.GetGeneratedFiles()
	fmt.Println("\n📁 Generated files:")
	for _, file := range files {
		fmt.Printf("   %s\n", file)
	}

	// List all endpoints
	endpoints := generator.ListEndpoints()
	fmt.Println("\n🔗 Generated endpoints:")
	for path, methods := range endpoints {
		fmt.Printf("   %s: %v\n", path, methods)
	}

	fmt.Println("\n🎉 Demo completed successfully!")
	fmt.Println("\nTo view the documentation:")
	fmt.Println("   1. Open ./docs/index.html in your browser for Swagger UI")
	fmt.Println("   2. Check ./docs/openapi.json for the OpenAPI specification")
	fmt.Println("   3. Check ./docs/openapi.yaml for the YAML format")
	fmt.Println("   4. Check ./docs/README.md for markdown documentation")
	fmt.Println("\nOr run a simple server:")
	fmt.Println("   go run -c 'openapi.GenerateAndServe(generator, 8080)'")
}
