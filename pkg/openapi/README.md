# OpenAPI Documentation Generator

A comprehensive OpenAPI 3.0 documentation generator for Go applications. This package provides automatic schema generation from Go types, endpoint documentation, and multiple output formats including JSON, YAML, and HTML with Swagger UI.

## Features

- 🚀 **OpenAPI 3.0 Specification Generation**: Full support for OpenAPI 3.0 spec
- 🔄 **Automatic Schema Generation**: Generate schemas from Go structs using reflection
- 📝 **Endpoint Registration**: Functions to register endpoints with full documentation
- 📊 **Multiple Output Formats**: JSON, YAML, and HTML documentation
- 🎨 **Swagger UI Integration**: Generate beautiful Swagger UI documentation
- ✅ **Validation**: Validate the generated OpenAPI spec
- 🔥 **Hot Reload**: Support for regenerating docs during development
- 🏗️ **Builder Pattern**: Fluent API for building endpoint documentation
- 🔐 **Security Schemes**: Support for various authentication methods
- 📋 **CRUD Generation**: Automatic CRUD endpoint generation
- 🏷️ **Tag Management**: Organize endpoints with tags
- 📈 **Statistics**: Get insights about your API documentation

## Installation

```bash
go get github.com/bata94/apiright/pkg/openapi
```

## Quick Start

```go
package main

import (
    "log"
    "reflect"
    "time"
    
    "github.com/bata94/apiright/pkg/openapi"
)

type User struct {
    ID       string    `json:"id" description:"User ID"`
    Username string    `json:"username" description:"Username"`
    Email    string    `json:"email" description:"Email address"`
    Created  time.Time `json:"created_at" description:"Creation timestamp"`
}

func main() {
    // Create a generator with quick start
    generator := openapi.QuickStart("My API", "A sample API", "1.0.0")

    // Add health check endpoint
    openapi.HealthCheckEndpoint(generator)

    // Add authentication endpoints
    openapi.AuthEndpoints(generator)

    // Add CRUD endpoints for users
    openapi.CRUDEndpoints(generator, "/users", reflect.TypeOf(User{}), "users")

    // Add a custom endpoint
    customBuilder := openapi.NewEndpointBuilder().
        Summary("Get user statistics").
        Description("Returns statistics about user activity").
        Tags("users", "analytics").
        PathParam("id", "User ID", reflect.TypeOf("")).
        QueryParam("period", "Time period for statistics", false, reflect.TypeOf("")).
        Security(openapi.SecurityRequirement{"BearerAuth": []string{}}).
        Response(200, "User statistics", "application/json", nil)

    generator.AddEndpointWithBuilder("GET", "/users/{id}/stats", customBuilder)

    // Generate and write documentation
    writer := openapi.NewWriter(generator)
    if err := writer.WriteFiles(); err != nil {
        log.Fatal(err)
    }

    // Or serve the documentation
    // openapi.GenerateAndServe(generator, 8080)
}
```

## Core Concepts

### Generator

The `Generator` is the main component that manages the OpenAPI specification:

```go
// Create with default config
config := openapi.DefaultConfig()
config.Title = "My API"
config.Description = "API description"
config.Version = "1.0.0"

generator := openapi.NewGenerator(config)

// Or use quick start
generator := openapi.QuickStart("My API", "API description", "1.0.0")
```

### Endpoint Builder

Use the builder pattern to create detailed endpoint documentation:

```go
builder := openapi.NewEndpointBuilder().
    Summary("Create user").
    Description("Create a new user account").
    Tags("users").
    RequestType(reflect.TypeOf(CreateUserRequest{})).
    Response(201, "User created", "application/json", reflect.TypeOf(User{})).
    Response(400, "Invalid input", "application/json", reflect.TypeOf(ErrorResponse{}))

generator.AddEndpointWithBuilder("POST", "/users", builder)
```

### Schema Generation

Automatic schema generation from Go types:

```go
type User struct {
    ID       string    `json:"id" description:"User ID"`
    Username string    `json:"username" validate:"required" description:"Username"`
    Email    string    `json:"email" validate:"required,email" description:"Email"`
    IsActive bool      `json:"is_active" description:"Active status"`
    Created  time.Time `json:"created_at" description:"Creation time"`
}

// Schema is automatically generated when you use the type
builder.RequestType(reflect.TypeOf(User{}))
```

## Advanced Usage

### Custom Configuration

```go
config := openapi.Config{
    Title:           "Advanced API",
    Description:     "Advanced API with custom configuration",
    Version:         "2.0.0",
    OutputDir:       "./docs",
    GenerateJSON:    true,
    GenerateYAML:    true,
    GenerateHTML:    true,
    PrettyPrint:     true,
    UseReferences:   true,
    IncludeExamples: true,
    ValidateSchemas: true,
}

generator := openapi.NewGenerator(config)
```

### Security Schemes

```go
// Add JWT Bearer authentication
generator.AddSecurityScheme("BearerAuth", openapi.SecurityScheme{
    Type:         "http",
    Scheme:       "bearer",
    BearerFormat: "JWT",
    Description:  "JWT Bearer token authentication",
})

// Add API Key authentication
generator.AddSecurityScheme("ApiKeyAuth", openapi.SecurityScheme{
    Type:        "apiKey",
    In:          "header",
    Name:        "X-API-Key",
    Description: "API key authentication",
})

// Use in endpoints
builder.Security(openapi.SecurityRequirement{"BearerAuth": []string{}})
```

### Servers

```go
generator.AddServer(openapi.Server{
    URL:         "https://api.example.com/v1",
    Description: "Production server",
})

generator.AddServer(openapi.Server{
    URL:         "https://staging-api.example.com/v1",
    Description: "Staging server",
})
```

### Tags

```go
generator.AddTag(openapi.Tag{
    Name:        "users",
    Description: "User management operations",
})

generator.AddTag(openapi.Tag{
    Name:        "auth",
    Description: "Authentication operations",
})
```

### Complex Endpoints

```go
builder := openapi.NewEndpointBuilder().
    Summary("Advanced user search").
    Description("Search users with advanced filtering and pagination").
    Tags("users", "search").
    QueryParam("q", "Search query", false, reflect.TypeOf("")).
    QueryParam("page", "Page number", false, reflect.TypeOf(1)).
    QueryParam("limit", "Items per page", false, reflect.TypeOf(20)).
    QueryParam("sort", "Sort field and direction", false, reflect.TypeOf("")).
    QueryParam("filter", "Filter criteria", false, reflect.TypeOf("")).
    HeaderParam("X-Request-ID", "Request ID", false, reflect.TypeOf("")).
    Security(openapi.SecurityRequirement{"BearerAuth": []string{}}).
    Response(200, "Search results", "application/json", reflect.TypeOf(UserListResponse{})).
    Response(400, "Invalid query", "application/json", reflect.TypeOf(ErrorResponse{})).
    Response(401, "Unauthorized", "application/json", reflect.TypeOf(ErrorResponse{}))

generator.AddEndpointWithBuilder("GET", "/users/search", builder)
```

### Request/Response Examples

```go
builder := openapi.NewEndpointBuilder().
    Summary("Create user").
    RequestType(reflect.TypeOf(CreateUserRequest{})).
    RequestExample(CreateUserRequest{
        Username:  "johndoe",
        Email:     "john@example.com",
        FirstName: "John",
        LastName:  "Doe",
    }).
    Response(201, "User created", "application/json", reflect.TypeOf(User{})).
    ResponseExample(User{
        ID:       "user_123",
        Username: "johndoe",
        Email:    "john@example.com",
        Created:  time.Now(),
    })
```

## Validation Tags

The package supports various validation tags for automatic schema validation:

```go
type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50" description:"Username"`
    Email    string `json:"email" validate:"required,email" description:"Email address"`
    Age      int    `json:"age" validate:"min=18,max=120" description:"User age"`
    Website  string `json:"website" validate:"url" description:"Personal website"`
}
```

Supported validation tags:
- `required` - Field is required
- `min=N` - Minimum value/length
- `max=N` - Maximum value/length
- `email` - Email format
- `url` - URL format
- `len=N` - Exact length

## Output Formats

### JSON
```go
writer := openapi.NewWriter(generator)
writer.WriteToFile("./api.json", "json")
```

### YAML
```go
writer.WriteToFile("./api.yaml", "yaml")
```

### HTML with Swagger UI
```go
writer.WriteToFile("./api.html", "html")
```

### Markdown Documentation
```go
writer.GenerateMarkdownDocs()
```

## Utility Functions

### CRUD Generation
```go
// Generates complete CRUD endpoints for a resource
openapi.CRUDEndpoints(generator, "/users", reflect.TypeOf(User{}), "users")
```

### Standard Endpoints
```go
// Add health check endpoint
openapi.HealthCheckEndpoint(generator)

// Add authentication endpoints
openapi.AuthEndpoints(generator)
```

### Helper Functions
```go
// Pointer helpers for optional fields
name := openapi.StringPtr("optional name")
count := openapi.IntPtr(42)
price := openapi.Float64Ptr(19.99)
active := openapi.BoolPtr(true)

// Common parameters
params := openapi.PaginationParams()
filters := openapi.FilterParams()
headers := openapi.CommonHeaders()
```

## Middleware

Automatic endpoint documentation middleware:

```go
generator := openapi.QuickStart("Auto API", "Auto-documented API", "1.0.0")
middleware := openapi.Middleware(generator)

// Use with your HTTP router
router.Use(middleware)
```

## Statistics

Get insights about your API documentation:

```go
stats := generator.GetStatistics()
fmt.Printf("Total endpoints: %d\n", stats.TotalEndpoints)
fmt.Printf("Total schemas: %d\n", stats.TotalSchemas)
fmt.Printf("Endpoints by method: %v\n", stats.EndpointsByMethod)
fmt.Printf("Endpoints by tag: %v\n", stats.EndpointsByTag)
```

## Serving Documentation

### Simple Server
```go
// Generate and serve documentation on port 8080
openapi.GenerateAndServe(generator, 8080)
```

### Custom Server
```go
writer := openapi.NewWriter(generator)
writer.WriteFiles()

// Serve with your own server
http.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs"))))
```

## Best Practices

1. **Use Descriptive Names**: Provide clear summaries and descriptions for endpoints
2. **Tag Organization**: Use tags to group related endpoints
3. **Validation Tags**: Use struct tags for automatic validation
4. **Examples**: Provide examples for better documentation
5. **Security**: Document security requirements for protected endpoints
6. **Error Responses**: Document all possible error responses
7. **Versioning**: Include version information in your API

## Examples

See the `example_test.go` file for comprehensive examples including:
- Basic usage
- Advanced endpoints
- Schema generation
- File generation
- CRUD generation
- Custom validation
- Complete workflows

## Testing

Run the tests:

```bash
go test ./pkg/openapi
```

Run benchmarks:

```bash
go test -bench=. ./pkg/openapi
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This package is part of the APIRight framework and follows the same license terms.