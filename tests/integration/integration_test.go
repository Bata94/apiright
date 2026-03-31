package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bata94/apiright/pkg/core"
)

func TestContentNegotiationIntegration(t *testing.T) {
	cn := core.NewContentNegotiator()

	testData := map[string]interface{}{
		"id":   1,
		"name": "test",
	}

	t.Run("JSON", func(t *testing.T) {
		data, err := cn.SerializeResponse(testData, "application/json")
		if err != nil {
			t.Fatalf("JSON serialization failed: %v", err)
		}
		if len(data) == 0 {
			t.Error("JSON serialization returned empty data")
		}
	})

	t.Run("XML", func(t *testing.T) {
		data, err := cn.SerializeResponse(testData, "application/xml")
		if err != nil {
			t.Fatalf("XML serialization failed: %v", err)
		}
		if len(data) == 0 {
			t.Error("XML serialization returned empty data")
		}
	})

	t.Run("YAML", func(t *testing.T) {
		data, err := cn.SerializeResponse(testData, "application/yaml")
		if err != nil {
			t.Fatalf("YAML serialization failed: %v", err)
		}
		if len(data) == 0 {
			t.Error("YAML serialization returned empty data")
		}
	})
}

func TestSimpleAPIExampleStructure(t *testing.T) {
	examplePath := filepath.Join("..", "..", "examples", "simple-api")

	info, err := os.Stat(examplePath)
	if os.IsNotExist(err) {
		t.Skip("examples/simple-api not found")
	}
	if !info.IsDir() {
		t.Skip("examples/simple-api is not a directory")
	}

	requiredFiles := []string{
		"apiright.yaml",
		"sqlc.yaml",
		"go.mod",
		"migrations",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(examplePath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Required file/directory not found: %s", file)
		}
	}

	migrationPath := filepath.Join(examplePath, "migrations")
	if info, err := os.Stat(migrationPath); err == nil && info.IsDir() {
		files, err := os.ReadDir(migrationPath)
		if err != nil {
			t.Errorf("Failed to read migrations directory: %v", err)
		}
		if len(files) == 0 {
			t.Error("No migration files found in examples/simple-api/migrations")
		}
	}
}

func TestBlogExampleStructure(t *testing.T) {
	examplePath := filepath.Join("..", "..", "examples", "blog")

	info, err := os.Stat(examplePath)
	if os.IsNotExist(err) {
		t.Skip("examples/blog not found")
	}
	if !info.IsDir() {
		t.Skip("examples/blog is not a directory")
	}

	requiredFiles := []string{
		"apiright.yaml",
		"sqlc.yaml",
		"go.mod",
		"migrations",
		"queries",
		"gen",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(examplePath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Required file/directory not found: %s", file)
		}
	}

	genSQLPath := filepath.Join(examplePath, "gen", "sql")
	if info, err := os.Stat(genSQLPath); err == nil && info.IsDir() {
		files, err := os.ReadDir(genSQLPath)
		if err != nil {
			t.Errorf("Failed to read gen/sql directory: %v", err)
		}
		if len(files) == 0 {
			t.Error("No generated SQL files found")
		}
	}
}

func TestSchemaParserFromMigrations(t *testing.T) {
	examplePath := filepath.Join("..", "..", "examples", "simple-api")

	_, err := os.Stat(examplePath)
	if os.IsNotExist(err) {
		t.Skip("examples/simple-api not found")
	}

	migrationPath := filepath.Join(examplePath, "migrations")
	files, err := os.ReadDir(migrationPath)
	if err != nil {
		t.Skip("Could not read migrations directory")
	}

	if len(files) == 0 {
		t.Skip("No migration files")
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationPath, file.Name()))
		if err != nil {
			t.Errorf("Failed to read migration %s: %v", file.Name(), err)
			continue
		}

		if len(content) == 0 {
			t.Errorf("Migration %s is empty", file.Name())
		}

		if !contains(string(content), "CREATE TABLE") {
			t.Errorf("Migration %s does not contain CREATE TABLE", file.Name())
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
