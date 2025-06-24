package openapi

import (
	"reflect"
	"testing"
	"time"
)

// TestBasicFunctionality tests basic functionality of the OpenAPI generator
func TestBasicFunctionality(t *testing.T) {
	// Test generator creation
	config := DefaultConfig()
	config.Title = "Test API"
	config.Description = "Test API Description"
	config.Version = "1.0.0"

	generator := NewGenerator(config)
	if generator == nil {
		t.Fatal("Failed to create generator")
	}

	// Test spec generation
	spec := generator.GetSpec()
	if spec.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got '%s'", spec.Info.Title)
	}

	if spec.OpenAPI != "3.0.3" {
		t.Errorf("Expected OpenAPI version '3.0.3', got '%s'", spec.OpenAPI)
	}
}

func TestEndpointBuilder(t *testing.T) {
	builder := NewEndpointBuilder()
	if builder == nil {
		t.Fatal("Failed to create endpoint builder")
	}

	// Test builder methods
	builder.Summary("Test endpoint").
		Description("Test description").
		Tags("test").
		Response(200, "Success", "application/json", nil)

	options := builder.Build()
	if options.Summary != "Test endpoint" {
		t.Errorf("Expected summary 'Test endpoint', got '%s'", options.Summary)
	}

	if len(options.Tags) != 1 || options.Tags[0] != "test" {
		t.Errorf("Expected tags ['test'], got %v", options.Tags)
	}
}

func TestSchemaGeneration(t *testing.T) {
	type TestStruct struct {
		ID       string    `json:"id"`
		Name     string    `json:"name"`
		Count    int       `json:"count"`
		Active   bool      `json:"active"`
		Created  time.Time `json:"created_at"`
		Optional *string   `json:"optional,omitempty"`
	}

	sg := NewSchemaGenerator()
	schema := sg.GenerateSchemaWithName(reflect.TypeOf(TestStruct{}), "TestStruct")

	if schema.Ref != "#/components/schemas/TestStruct" {
		t.Errorf("Expected reference to TestStruct schema, got '%s'", schema.Ref)
	}

	schemas := sg.GetSchemas()
	if len(schemas) == 0 {
		t.Error("Expected at least 1 schema")
	}

	testSchema, exists := schemas["TestStruct"]
	if !exists {
		t.Fatal("TestStruct schema not found")
	}

	if testSchema.Type != "object" {
		t.Errorf("Expected type 'object', got '%s'", testSchema.Type)
	}

	expectedFields := []string{"id", "name", "count", "active", "created_at", "optional"}
	if len(testSchema.Properties) != len(expectedFields) {
		t.Errorf("Expected %d properties, got %d", len(expectedFields), len(testSchema.Properties))
	}

	for _, field := range expectedFields {
		if _, exists := testSchema.Properties[field]; !exists {
			t.Errorf("Expected field '%s' not found in schema", field)
		}
	}
}

func TestAddEndpoint(t *testing.T) {
	generator := NewGenerator(DefaultConfig())

	builder := NewEndpointBuilder().
		Summary("Test GET endpoint").
		Response(200, "Success", "application/json", nil)

	err := generator.AddEndpointWithBuilder("GET", "/test", builder)
	if err != nil {
		t.Fatalf("Failed to add endpoint: %v", err)
	}

	// Test retrieval
	options, exists := generator.GetEndpoint("GET", "/test")
	if !exists {
		t.Fatal("Endpoint not found after adding")
	}

	if options.Summary != "Test GET endpoint" {
		t.Errorf("Expected summary 'Test GET endpoint', got '%s'", options.Summary)
	}

	// Test listing
	endpoints := generator.ListEndpoints()
	if len(endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(endpoints))
	}

	methods, exists := endpoints["/test"]
	if !exists {
		t.Fatal("Path '/test' not found in endpoints")
	}

	if len(methods) != 1 || methods[0] != "GET" {
		t.Errorf("Expected methods ['GET'], got %v", methods)
	}
}

func TestSpecGeneration(t *testing.T) {
	generator := NewGenerator(DefaultConfig())

	// Add a simple endpoint
	builder := NewEndpointBuilder().
		Summary("Test endpoint").
		Response(200, "Success", "application/json", nil)

	generator.AddEndpointWithBuilder("GET", "/test", builder)

	// Generate spec
	spec, err := generator.GenerateSpec()
	if err != nil {
		t.Fatalf("Failed to generate spec: %v", err)
	}

	if len(spec.Paths) != 1 {
		t.Errorf("Expected 1 path, got %d", len(spec.Paths))
	}

	pathItem, exists := spec.Paths["/test"]
	if !exists {
		t.Fatal("Path '/test' not found in spec")
	}

	if pathItem.Get == nil {
		t.Fatal("GET operation not found")
	}

	if pathItem.Get.Summary != "Test endpoint" {
		t.Errorf("Expected summary 'Test endpoint', got '%s'", pathItem.Get.Summary)
	}
}

func TestQuickStart(t *testing.T) {
	generator := QuickStart("Quick API", "Quick test", "1.0.0")
	if generator == nil {
		t.Fatal("Failed to create generator with QuickStart")
	}

	spec := generator.GetSpec()
	if spec.Info.Title != "Quick API" {
		t.Errorf("Expected title 'Quick API', got '%s'", spec.Info.Title)
	}

	// Should have default servers
	if len(spec.Servers) == 0 {
		t.Error("Expected default servers to be added")
	}

	// Should have default security schemes
	if len(spec.Components.SecuritySchemes) == 0 {
		t.Error("Expected default security schemes to be added")
	}

	// Should have default tags
	if len(spec.Tags) == 0 {
		t.Error("Expected default tags to be added")
	}
}

func TestValidation(t *testing.T) {
	generator := NewGenerator(DefaultConfig())

	// Test invalid HTTP method
	builder := NewEndpointBuilder()
	err := generator.AddEndpointWithBuilder("INVALID", "/test", builder)
	if err == nil {
		t.Error("Expected error for invalid HTTP method")
	}

	// Test valid method
	err = generator.AddEndpointWithBuilder("GET", "/test", builder)
	if err != nil {
		t.Errorf("Unexpected error for valid method: %v", err)
	}
}

func TestStatistics(t *testing.T) {
	generator := NewGenerator(DefaultConfig())

	// Add some endpoints
	builder1 := NewEndpointBuilder().Summary("GET endpoint")
	builder2 := NewEndpointBuilder().Summary("POST endpoint").Tags("api")
	builder3 := NewEndpointBuilder().Summary("Another GET").Tags("api", "users")

	generator.AddEndpointWithBuilder("GET", "/test1", builder1)
	generator.AddEndpointWithBuilder("POST", "/test2", builder2)
	generator.AddEndpointWithBuilder("GET", "/test3", builder3)

	stats := generator.GetStatistics()

	if stats.TotalEndpoints != 3 {
		t.Errorf("Expected 3 total endpoints, got %d", stats.TotalEndpoints)
	}

	if stats.EndpointsByMethod["GET"] != 2 {
		t.Errorf("Expected 2 GET endpoints, got %d", stats.EndpointsByMethod["GET"])
	}

	if stats.EndpointsByMethod["POST"] != 1 {
		t.Errorf("Expected 1 POST endpoint, got %d", stats.EndpointsByMethod["POST"])
	}

	if stats.EndpointsByTag["api"] != 2 {
		t.Errorf("Expected 2 endpoints with 'api' tag, got %d", stats.EndpointsByTag["api"])
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test pointer helpers
	if StringPtr("test") == nil || *StringPtr("test") != "test" {
		t.Error("StringPtr failed")
	}

	if IntPtr(42) == nil || *IntPtr(42) != 42 {
		t.Error("IntPtr failed")
	}

	if Float64Ptr(3.14) == nil || *Float64Ptr(3.14) != 3.14 {
		t.Error("Float64Ptr failed")
	}

	if BoolPtr(true) == nil || *BoolPtr(true) != true {
		t.Error("BoolPtr failed")
	}

	// Test HTTP method validation
	if !IsValidHTTPMethod("GET") {
		t.Error("GET should be valid")
	}

	if IsValidHTTPMethod("INVALID") {
		t.Error("INVALID should not be valid")
	}

	// Test common headers
	headers := CommonHeaders()
	if len(headers) == 0 {
		t.Error("Expected common headers")
	}

	// Test pagination params
	params := PaginationParams()
	if len(params) == 0 {
		t.Error("Expected pagination parameters")
	}
}

// Benchmark tests
func BenchmarkSchemaGeneration(b *testing.B) {
	type BenchStruct struct {
		ID       string    `json:"id"`
		Name     string    `json:"name"`
		Count    int       `json:"count"`
		Active   bool      `json:"active"`
		Created  time.Time `json:"created_at"`
		Data     map[string]interface{} `json:"data"`
		Items    []string  `json:"items"`
	}

	sg := NewSchemaGenerator()
	t := reflect.TypeOf(BenchStruct{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sg.GenerateSchema(t)
	}
}

func BenchmarkSpecGeneration(b *testing.B) {
	generator := NewGenerator(DefaultConfig())

	// Add multiple endpoints
	for i := 0; i < 100; i++ {
		builder := NewEndpointBuilder().
			Summary("Test endpoint").
			Response(200, "Success", "application/json", nil)
		generator.AddEndpointWithBuilder("GET", "/test"+string(rune(i)), builder)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.GenerateSpec()
	}
}