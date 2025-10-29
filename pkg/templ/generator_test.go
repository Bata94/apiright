package ar_templ

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetGoModulePath(t *testing.T) {
	// This test assumes we're in a Go module
	modulePath, err := getGoModulePath()
	if err != nil {
		t.Fatalf("getGoModulePath failed: %v", err)
	}
	if modulePath == "" {
		t.Error("Expected non-empty module path")
	}
	// Should contain "apiright"
	if len(modulePath) > 0 && !contains(modulePath, "apiright") {
		t.Errorf("Expected module path to contain 'apiright', got %s", modulePath)
	}
}

func TestGetProjectRoot(t *testing.T) {
	projectRoot, err := getProjectRoot()
	if err != nil {
		t.Fatalf("getProjectRoot failed: %v", err)
	}
	if projectRoot == "" {
		t.Error("Expected non-empty project root")
	}

	// Check that go.mod exists in the project root
	goModPath := filepath.Join(projectRoot, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Errorf("go.mod not found at expected location: %s", goModPath)
	}
}

func TestFindRoutes(t *testing.T) {
	// Create a temporary directory with test .templ files
	tempDir, err := os.MkdirTemp("", "templ_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to clean up temp dir: %v", err)
		}
	}()

	// Create test .templ files
	files := []string{"index.templ", "about.templ", "contact-us.templ"}
	for _, file := range files {
		filePath := filepath.Join(tempDir, file)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Create a non-.templ file
	err = os.WriteFile(filepath.Join(tempDir, "notatempl.txt"), []byte("not a templ"), 0644)
	if err != nil {
		t.Fatalf("Failed to create non-templ file: %v", err)
	}

	routes, err := findRoutes(tempDir)
	if err != nil {
		t.Fatalf("findRoutes failed: %v", err)
	}

	expectedRoutes := map[string]string{
		"/":           "Index",
		"/about":      "About",
		"/contact-us": "ContactUs",
	}

	if len(routes) != len(expectedRoutes) {
		t.Errorf("Expected %d routes, got %d", len(expectedRoutes), len(routes))
	}

	for _, route := range routes {
		expectedComponent, exists := expectedRoutes[route.Path]
		if !exists {
			t.Errorf("Unexpected route path: %s", route.Path)
			continue
		}
		if route.ComponentName != expectedComponent {
			t.Errorf("For path %s, expected component name %s, got %s", route.Path, expectedComponent, route.ComponentName)
		}
		delete(expectedRoutes, route.Path)
	}

	if len(expectedRoutes) > 0 {
		t.Errorf("Missing routes: %v", expectedRoutes)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || contains(s[1:len(s)-1], substr)))
}
