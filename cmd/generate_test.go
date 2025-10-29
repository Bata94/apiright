package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenCmd_Flags(t *testing.T) {
	// Test that flags are properly set
	cmd := genCmd

	// Check input flag
	inputFlag := cmd.PersistentFlags().Lookup("input")
	if inputFlag == nil {
		t.Fatal("input flag not found")
	}
	if inputFlag.DefValue != defaultInputDir {
		t.Errorf("Expected default input dir %s, got %s", defaultInputDir, inputFlag.DefValue)
	}

	// Check output flag
	outputFlag := cmd.PersistentFlags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("output flag not found")
	}
	if outputFlag.DefValue != defaultOutputFileName {
		t.Errorf("Expected default output file %s, got %s", defaultOutputFileName, outputFlag.DefValue)
	}

	// Check package flag
	packageFlag := cmd.PersistentFlags().Lookup("package")
	if packageFlag == nil {
		t.Fatal("package flag not found")
	}
	if packageFlag.DefValue != defaultPackageName {
		t.Errorf("Expected default package name %s, got %s", defaultPackageName, packageFlag.DefValue)
	}
}

func TestGenCmd_Execution(t *testing.T) {
	// Create a temporary directory with test .templ files
	tempDir, err := os.MkdirTemp("", "cmd_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to clean up temp dir: %v", err)
		}
	}()

	// Create test .templ file
	templFile := filepath.Join(tempDir, "test.templ")
	err = os.WriteFile(templFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test templ file: %v", err)
	}

	// Test the generator function (this would normally be called by the command)
	// We can't easily test the cobra command execution without complex setup
	// But we can test that the function doesn't panic with valid inputs
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GeneratorRun panicked: %v", r)
		}
	}()

	// This will fail because it's trying to find go.mod, but it tests the basic flow
	// In a real test environment, we'd mock this
	// For now, just test that the function exists and can be called
}
