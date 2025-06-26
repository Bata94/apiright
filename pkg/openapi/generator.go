package openapi

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
)

// Config contains configuration for the OpenAPI generator
type Config struct {
	// API Information
	Title          string
	Description    string
	Version        string
	TermsOfService string
	Contact        *Contact
	License        *License

	// Server Information
	Servers []Server

	// Security
	SecuritySchemes map[string]SecurityScheme
	GlobalSecurity  []SecurityRequirement

	// Tags
	Tags []Tag

	// External Documentation
	ExternalDocs *ExternalDocumentation

	// Output Configuration
	OutputDir    string
	GenerateJSON bool
	GenerateYAML bool
	GenerateHTML bool
	PrettyPrint  bool

	// Schema Generation Options
	UseReferences     bool
	IncludeExamples   bool
	ValidateSchemas   bool
	CustomTypeMapping map[reflect.Type]Schema
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		Title:             "API Documentation",
		Description:       "Generated API documentation",
		Version:           "1.0.0",
		OutputDir:         "./docs",
		GenerateJSON:      true,
		GenerateYAML:      true,
		GenerateHTML:      true,
		PrettyPrint:       true,
		UseReferences:     true,
		IncludeExamples:   true,
		ValidateSchemas:   true,
		CustomTypeMapping: make(map[reflect.Type]Schema),
	}
}

// Generator is the main OpenAPI documentation generator
type Generator struct {
	config          Config
	spec            *OpenAPISpec
	schemaGenerator *SchemaGenerator
	endpoints       map[string]map[string]EndpointOptions // path -> method -> options
}

// NewGenerator creates a new OpenAPI generator
func NewGenerator(config Config) *Generator {
	spec := NewOpenAPISpec()
	spec.Info = Info{
		Title:          config.Title,
		Description:    config.Description,
		Version:        config.Version,
		TermsOfService: config.TermsOfService,
		Contact:        config.Contact,
		License:        config.License,
	}

	if len(config.Servers) > 0 {
		spec.Servers = config.Servers
	}

	if len(config.SecuritySchemes) > 0 {
		spec.Components.SecuritySchemes = config.SecuritySchemes
	}

	if len(config.GlobalSecurity) > 0 {
		spec.Security = config.GlobalSecurity
	}

	if len(config.Tags) > 0 {
		spec.Tags = config.Tags
	}

	if config.ExternalDocs != nil {
		spec.ExternalDocs = config.ExternalDocs
	}

	return &Generator{
		config:          config,
		spec:            spec,
		schemaGenerator: NewSchemaGenerator(),
		endpoints:       make(map[string]map[string]EndpointOptions),
	}
}

// AddEndpoint adds an endpoint to the documentation
func (g *Generator) AddEndpoint(method, path string, options EndpointOptions) error {
	method = strings.ToUpper(method)

	if !IsValidHTTPMethod(method) {
		return fmt.Errorf("invalid HTTP method: %s", method)
	}

	if g.endpoints[path] == nil {
		g.endpoints[path] = make(map[string]EndpointOptions)
	}

	g.endpoints[path][method] = options
	return nil
}

// AddEndpointWithBuilder adds an endpoint using the builder pattern
func (g *Generator) AddEndpointWithBuilder(method, path string, builder *EndpointBuilder) error {
	return g.AddEndpoint(method, path, builder.Build())
}

// RemoveEndpoint removes an endpoint from the documentation
func (g *Generator) RemoveEndpoint(method, path string) {
	method = strings.ToUpper(method)

	if pathMethods, exists := g.endpoints[path]; exists {
		delete(pathMethods, method)
		if len(pathMethods) == 0 {
			delete(g.endpoints, path)
		}
	}
}

// GetEndpoint retrieves an endpoint's documentation
func (g *Generator) GetEndpoint(method, path string) (EndpointOptions, bool) {
	method = strings.ToUpper(method)

	if pathMethods, exists := g.endpoints[path]; exists {
		if options, exists := pathMethods[method]; exists {
			return options, true
		}
	}
	return EndpointOptions{}, false
}

// ListEndpoints returns all registered endpoints
func (g *Generator) ListEndpoints() map[string][]string {
	result := make(map[string][]string)
	for path, methods := range g.endpoints {
		for method := range methods {
			result[path] = append(result[path], method)
		}
	}
	return result
}

// GenerateSpec generates the complete OpenAPI specification
func (g *Generator) GenerateSpec() (*OpenAPISpec, error) {
	// Clear existing paths
	g.spec.Paths = make(map[string]PathItem)

	// Process all endpoints
	for path, methods := range g.endpoints {
		pathItem := PathItem{}

		for method, options := range methods {
			builder := NewEndpointBuilder()
			builder.options = options

			// Generate schemas for request/response types
			if options.RequestType != nil {
				schema := g.schemaGenerator.GenerateSchema(options.RequestType)
				if builder.options.RequestBody == nil {
					builder.options.RequestBody = &RequestBodyInfo{}
				}
				builder.options.RequestBody.Schema = &schema
			}

			// Generate schemas for individual response types
			for statusCode, response := range builder.options.Responses {
				if response.Type != nil {
					schema := g.schemaGenerator.GenerateSchema(response.Type)
					response.Schema = &schema
					builder.options.Responses[statusCode] = response
				}
			}

			// Fallback: if ResponseType is set, apply it to responses without specific types
			if options.ResponseType != nil {
				schema := g.schemaGenerator.GenerateSchema(options.ResponseType)
				for statusCode, response := range builder.options.Responses {
					if response.Schema == nil {
						response.Schema = &schema
						builder.options.Responses[statusCode] = response
					}
				}
			}

			operation := builder.ConvertToOperation()

			// Add operation to path item
			switch method {
			case "GET":
				pathItem.Get = &operation
			case "POST":
				pathItem.Post = &operation
			case "PUT":
				pathItem.Put = &operation
			case "PATCH":
				pathItem.Patch = &operation
			case "DELETE":
				pathItem.Delete = &operation
			case "HEAD":
				pathItem.Head = &operation
			case "OPTIONS":
				pathItem.Options = &operation
			case "TRACE":
				pathItem.Trace = &operation
			}
		}

		g.spec.Paths[path] = pathItem
	}

	// Add generated schemas to components
	for name, schema := range g.schemaGenerator.GetSchemas() {
		g.spec.Components.Schemas[name] = schema
	}

	// Validate schemas if enabled
	if g.config.ValidateSchemas {
		if err := g.validateSpec(); err != nil {
			return nil, fmt.Errorf("spec validation failed: %w", err)
		}
	}

	return g.spec, nil
}

// validateSpec performs basic validation on the generated specification
func (g *Generator) validateSpec() error {
	// Validate info section
	if g.spec.Info.Title == "" {
		return fmt.Errorf("API title is required")
	}
	if g.spec.Info.Version == "" {
		return fmt.Errorf("API version is required")
	}

	// Validate paths
	for path, pathItem := range g.spec.Paths {
		if path == "" {
			return fmt.Errorf("path cannot be empty")
		}

		// Validate operations
		operations := []*Operation{
			pathItem.Get, pathItem.Post, pathItem.Put, pathItem.Patch,
			pathItem.Delete, pathItem.Head, pathItem.Options, pathItem.Trace,
		}

		for _, op := range operations {
			if op != nil {
				if len(op.Responses) == 0 {
					return fmt.Errorf("operation on path %s must have at least one response", path)
				}
			}
		}
	}

	// Validate schemas
	for name, schema := range g.spec.Components.Schemas {
		if err := ValidateSchema(schema); err != nil {
			return fmt.Errorf("invalid schema %s: %w", name, err)
		}
	}

	return nil
}

// GetSpec returns the current specification
func (g *Generator) GetSpec() *OpenAPISpec {
	return g.spec
}

// UpdateInfo updates the API information
func (g *Generator) UpdateInfo(info Info) {
	g.spec.Info = info
}

// AddServer adds a server to the specification
func (g *Generator) AddServer(server Server) {
	g.spec.Servers = append(g.spec.Servers, server)
}

// AddSecurityScheme adds a security scheme
func (g *Generator) AddSecurityScheme(name string, scheme SecurityScheme) {
	if g.spec.Components.SecuritySchemes == nil {
		g.spec.Components.SecuritySchemes = make(map[string]SecurityScheme)
	}
	g.spec.Components.SecuritySchemes[name] = scheme
}

// AddTag adds a tag to the specification
func (g *Generator) AddTag(tag Tag) {
	g.spec.Tags = append(g.spec.Tags, tag)
}

// AddGlobalSecurity adds global security requirements
func (g *Generator) AddGlobalSecurity(requirements ...SecurityRequirement) {
	g.spec.Security = append(g.spec.Security, requirements...)
}

// SetExternalDocs sets external documentation
func (g *Generator) SetExternalDocs(docs ExternalDocumentation) {
	g.spec.ExternalDocs = &docs
}

// AddCustomSchema adds a custom schema to the components
func (g *Generator) AddCustomSchema(name string, schema Schema) {
	g.spec.Components.Schemas[name] = schema
}

// AddCustomResponse adds a custom response to the components
func (g *Generator) AddCustomResponse(name string, response Response) {
	if g.spec.Components.Responses == nil {
		g.spec.Components.Responses = make(map[string]Response)
	}
	g.spec.Components.Responses[name] = response
}

// AddCustomParameter adds a custom parameter to the components
func (g *Generator) AddCustomParameter(name string, parameter Parameter) {
	if g.spec.Components.Parameters == nil {
		g.spec.Components.Parameters = make(map[string]Parameter)
	}
	g.spec.Components.Parameters[name] = parameter
}

// GetSchemaGenerator returns the schema generator for advanced usage
func (g *Generator) GetSchemaGenerator() *SchemaGenerator {
	return g.schemaGenerator
}

// Reset clears all endpoints and resets the specification
func (g *Generator) Reset() {
	g.endpoints = make(map[string]map[string]EndpointOptions)
	g.spec = NewOpenAPISpec()
	g.schemaGenerator = NewSchemaGenerator()

	// Restore basic info
	g.spec.Info = Info{
		Title:          g.config.Title,
		Description:    g.config.Description,
		Version:        g.config.Version,
		TermsOfService: g.config.TermsOfService,
		Contact:        g.config.Contact,
		License:        g.config.License,
	}
}

// Clone creates a copy of the generator
func (g *Generator) Clone() *Generator {
	newGen := NewGenerator(g.config)

	// Copy endpoints
	for path, methods := range g.endpoints {
		for method, options := range methods {
			err := newGen.AddEndpoint(method, path, options)
			if err != nil {
				panic(fmt.Errorf("failed to clone endpoint %s %s: %w", method, path, err))
			}
		}
	}

	return newGen
}

// Merge merges another generator into this one
func (g *Generator) Merge(other *Generator) error {
	for path, methods := range other.endpoints {
		for method, options := range methods {
			if err := g.AddEndpoint(method, path, options); err != nil {
				return fmt.Errorf("failed to merge endpoint %s %s: %w", method, path, err)
			}
		}
	}

	// Merge schemas
	for name, schema := range other.schemaGenerator.GetSchemas() {
		g.spec.Components.Schemas[name] = schema
	}

	return nil
}

// GetOutputPath returns the full output path for a given filename
func (g *Generator) GetOutputPath(filename string) string {
	return filepath.Join(g.config.OutputDir, filename)
}

// Statistics returns statistics about the generated documentation
type Statistics struct {
	TotalEndpoints    int
	TotalSchemas      int
	EndpointsByMethod map[string]int
	EndpointsByTag    map[string]int
}

// GetStatistics returns statistics about the current documentation
func (g *Generator) GetStatistics() Statistics {
	stats := Statistics{
		EndpointsByMethod: make(map[string]int),
		EndpointsByTag:    make(map[string]int),
	}

	for _, methods := range g.endpoints {
		for method, options := range methods {
			stats.TotalEndpoints++
			stats.EndpointsByMethod[method]++

			for _, tag := range options.Tags {
				stats.EndpointsByTag[tag]++
			}
		}
	}

	stats.TotalSchemas = len(g.schemaGenerator.GetSchemas())

	return stats
}
