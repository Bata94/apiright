package apiright_test

import (
	"testing"

	"github.com/bata94/apiright/pkg/core"
)

func TestJSONSerialization(t *testing.T) {
	cn := core.NewContentNegotiator()

	testData := map[string]any{
		"name":   "test",
		"age":    42,
		"active": true,
	}

	data, err := cn.SerializeResponse(testData, "application/json")
	if err != nil {
		t.Fatalf("Failed to serialize JSON: %v", err)
	}

	expected := `{"active":true,"age":42,"name":"test"}`
	if string(data) != expected {
		t.Errorf("JSON serialization mismatch:\nexpected: %s\ngot: %s", expected, string(data))
	}

	var result map[string]any
	if err := cn.DeserializeRequest(data, "application/json", &result); err != nil {
		t.Fatalf("Failed to deserialize JSON: %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("Deserialized name mismatch: expected 'test', got '%v'", result["name"])
	}
}

func TestXMLSerialization(t *testing.T) {
	cn := core.NewContentNegotiator()

	testData := map[string]any{
		"name": "test",
		"age":  42,
	}

	data, err := cn.SerializeResponse(testData, "application/xml")
	if err != nil {
		t.Fatalf("Failed to serialize XML: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("XML serialization returned empty data")
	}

	if string(data) == "" {
		t.Error("XML serialization returned empty string")
	}
}

func TestYAMLSerialization(t *testing.T) {
	cn := core.NewContentNegotiator()

	testData := map[string]any{
		"name": "test",
		"age":  42,
	}

	data, err := cn.SerializeResponse(testData, "application/yaml")
	if err != nil {
		t.Fatalf("Failed to serialize YAML: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("YAML serialization returned empty data")
	}
}

func TestTextSerialization(t *testing.T) {
	cn := core.NewContentNegotiator()

	testData := map[string]any{
		"name": "test",
		"age":  42,
	}

	data, err := cn.SerializeResponse(testData, "text/plain")
	if err != nil {
		t.Fatalf("Failed to serialize text: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Text serialization returned empty data")
	}
}

func TestAcceptHeaderParsing(t *testing.T) {
	cn := core.NewContentNegotiator()

	tests := []struct {
		accept   string
		expected string
	}{
		{"application/json", "application/json"},
		{"application/xml", "application/xml"},
		{"application/yaml", "application/yaml"},
		{"text/plain", "text/plain"},
		{"application/protobuf", "application/protobuf"},
		{"", "application/json"},
		{"invalid/type", "application/json"},
	}

	for _, tc := range tests {
		t.Run(tc.accept, func(t *testing.T) {
			result := cn.DetectContentType(tc.accept)
			if result != tc.expected {
				t.Errorf("DetectContentType(%q) = %q, want %q", tc.accept, result, tc.expected)
			}
		})
	}
}

func TestQValuePriority(t *testing.T) {
	cn := core.NewContentNegotiator()

	tests := []struct {
		accept   string
		expected string
	}{
		{"application/json;q=0.9,application/xml;q=0.8", "application/json"},
		{"application/xml;q=0.8,application/json;q=0.9", "application/json"},
		{"text/plain;q=0.5,application/json;q=1.0", "application/json"},
		{"application/yaml;q=0.5,text/plain;q=0.7", "text/plain"},
	}

	for _, tc := range tests {
		t.Run(tc.accept, func(t *testing.T) {
			result := cn.DetectContentType(tc.accept)
			if result != tc.expected {
				t.Errorf("DetectContentType(%q) = %q, want %q", tc.accept, result, tc.expected)
			}
		})
	}
}

func TestWildcardMatching(t *testing.T) {
	cn := core.NewContentNegotiator()

	tests := []struct {
		accept   string
		expected string
	}{
		{"*/*", "application/json"},
		{"application/*", "application/json"},
		{"text/*", "text/plain"},
	}

	for _, tc := range tests {
		t.Run(tc.accept, func(t *testing.T) {
			result := cn.DetectContentType(tc.accept)
			if result != tc.expected {
				t.Errorf("DetectContentType(%q) = %q, want %q", tc.accept, result, tc.expected)
			}
		})
	}
}

func TestUnsupportedType(t *testing.T) {
	cn := core.NewContentNegotiator()

	testData := map[string]any{
		"name": "test",
	}

	_, err := cn.SerializeResponse(testData, "application/unsupported")
	if err == nil {
		t.Error("Expected error for unsupported type, got nil")
	}

	var target map[string]any
	err = cn.DeserializeRequest([]byte("test"), "application/unsupported", &target)
	if err == nil {
		t.Error("Expected error for unsupported type deserialization, got nil")
	}
}

func TestRequestDeserialization(t *testing.T) {
	cn := core.NewContentNegotiator()

	t.Run("JSON", func(t *testing.T) {
		data := []byte(`{"name":"test","age":42}`)
		var result map[string]any
		if err := cn.DeserializeRequest(data, "application/json", &result); err != nil {
			t.Fatalf("Failed to deserialize JSON: %v", err)
		}
		if result["name"] != "test" {
			t.Errorf("Deserialized name mismatch: expected 'test', got '%v'", result["name"])
		}
	})

	t.Run("Text to String", func(t *testing.T) {
		data := []byte("hello world")
		var result string
		if err := cn.DeserializeRequest(data, "text/plain", &result); err != nil {
			t.Fatalf("Failed to deserialize text: %v", err)
		}
		if result != "hello world" {
			t.Errorf("Deserialized text mismatch: expected 'hello world', got '%s'", result)
		}
	})
}

func TestContentNegotiatorSupportedTypes(t *testing.T) {
	cn := core.NewContentNegotiator()

	supported := cn.SupportedTypes()

	expected := []string{
		"application/json",
		"application/xml",
		"application/yaml",
		"application/protobuf",
		"text/plain",
	}

	if len(supported) != len(expected) {
		t.Errorf("SupportedTypes() returned %d types, want %d", len(supported), len(expected))
	}

	for _, exp := range expected {
		found := false
		for _, sup := range supported {
			if sup == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected type %q not found in supported types", exp)
		}
	}
}

func TestIsContentTypeSupported(t *testing.T) {
	cn := core.NewContentNegotiator()

	tests := []struct {
		contentType string
		supported   bool
	}{
		{"application/json", true},
		{"application/xml", true},
		{"application/yaml", true},
		{"application/protobuf", true},
		{"text/plain", true},
		{"text/html", false},
		{"application/octet-stream", false},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(tc.contentType, func(t *testing.T) {
			result := cn.IsContentTypeSupported(tc.contentType)
			if result != tc.supported {
				t.Errorf("IsContentTypeSupported(%q) = %v, want %v", tc.contentType, result, tc.supported)
			}
		})
	}
}

func TestSetSupportedTypes(t *testing.T) {
	cn := core.NewContentNegotiator()

	cn.SetSupportedTypes([]string{"application/json", "text/plain"})

	supported := cn.SupportedTypes()
	if len(supported) != 2 {
		t.Errorf("SetSupportedTypes did not update supported types correctly")
	}

	if !cn.IsContentTypeSupported("application/json") {
		t.Error("SetSupportedTypes did not set application/json as supported")
	}

	if !cn.IsContentTypeSupported("text/plain") {
		t.Error("SetSupportedTypes did not set text/plain as supported")
	}

	if cn.IsContentTypeSupported("application/xml") {
		t.Error("SetSupportedTypes did not remove application/xml from supported types")
	}
}

func TestSetDefaultType(t *testing.T) {
	cn := core.NewContentNegotiator()

	cn.SetDefaultType("application/xml")

	result := cn.DetectContentType("")
	if result != "application/xml" {
		t.Errorf("SetDefaultType did not update default type: got %q, want %q", result, "application/xml")
	}

	result = cn.DetectContentType("*/*")
	if result != "application/xml" {
		t.Errorf("Default type not applied to wildcard: got %q, want %q", result, "application/xml")
	}
}

func TestParseContentHeader(t *testing.T) {
	tests := []struct {
		header   string
		expected string
	}{
		{"application/json", "application/json"},
		{"application/json; charset=utf-8", "application/json"},
		{"text/plain; charset=utf-8", "text/plain"},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.header, func(t *testing.T) {
			info := core.ParseContentHeader(tc.header)
			if info.ContentType != tc.expected {
				t.Errorf("ParseContentHeader(%q) ContentType = %q, want %q", tc.header, info.ContentType, tc.expected)
			}
		})
	}
}
