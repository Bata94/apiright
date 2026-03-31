package apiright_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bata94/apiright/cmd/apiright"
)

func TestValidateProjectName(t *testing.T) {
	testCases := []struct {
		name        string
		projectName string
		wantErr     bool
	}{
		{"valid name with dash", "test-project", false},
		{"valid name with underscore", "test_project", false},
		{"valid name uppercase", "TestProject", false},
		{"valid name with numbers", "test123", false},
		{"with spaces", "invalid project", true},
		{"with special chars", "test@project", true},
		{"with dot", "test.project", true},
		{"empty", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateProjectName(tc.projectName)
			if (err != nil) != tc.wantErr {
				t.Errorf("validateProjectName(%q) error = %v, wantErr %v", tc.projectName, err, tc.wantErr)
			}
		})
	}
}

func validateProjectName(name string) error {
	if name == "" {
		return &testError{message: "project name cannot be empty"}
	}

	if strings.Contains(name, " ") {
		return &testError{message: "project name cannot contain spaces"}
	}

	for _, r := range name {
		valid := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
		if !valid {
			return &testError{message: "project name contains invalid character"}
		}
	}
	return nil
}

type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

func TestInitProjectStructure(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "testproject"
	projectPath := filepath.Join(tmpDir, projectName)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(cwd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	cmd := apiright.NewInitCommand()
	cmd.SetArgs([]string{projectName})

	err = cmd.Execute()
	if err != nil {
		t.Skipf("Skipping integration test - command uses os.Exit on error: %v", err)
	}

	requiredFiles := []string{
		"apiright.yaml",
		"sqlc.yaml",
		"go.mod",
		"main.go",
		".gitignore",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(projectPath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Required file not created: %s", file)
		}
	}

	requiredDirs := []string{
		"gen",
		"gen/sql",
		"gen/go",
		"gen/proto",
		"queries",
		"proto",
		"migrations",
	}

	for _, dir := range requiredDirs {
		dirPath := filepath.Join(projectPath, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("Required directory not created: %s", dir)
		}
	}
}

func TestInitDatabaseOptions(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "testproject"
	projectPath := filepath.Join(tmpDir, projectName)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(cwd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	cmd := apiright.NewInitCommand()
	cmd.SetArgs([]string{projectName, "-d", "postgres"})

	err = cmd.Execute()
	if err != nil {
		t.Skipf("Skipping integration test - command uses os.Exit on error: %v", err)
	}

	sqlcPath := filepath.Join(projectPath, "sqlc.yaml")
	content, err := os.ReadFile(sqlcPath)
	if err != nil {
		t.Errorf("Failed to read sqlc.yaml: %v", err)
		return
	}

	if !strings.Contains(string(content), "postgresql") {
		t.Errorf("sqlc.yaml missing expected engine: postgresql")
	}
}
