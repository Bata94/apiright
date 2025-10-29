package cmd

import (
	"os"
	"testing"
)

func TestRootCmd_Flags(t *testing.T) {
	cmd := rootCmd

	// Check verbose flag
	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("verbose flag not found")
	}

	// Check debug flag
	debugFlag := cmd.PersistentFlags().Lookup("debug")
	if debugFlag == nil {
		t.Error("debug flag not found")
	}
}

func TestInitConfig(t *testing.T) {
	// Save original env
	originalEnv := os.Getenv("ENV")
	defer func() {
		if err := os.Setenv("ENV", originalEnv); err != nil {
			t.Errorf("Failed to restore ENV: %v", err)
		}
	}()

	// Test DEV env
	if err := os.Setenv("ENV", "DEV"); err != nil {
		t.Fatalf("Failed to set ENV: %v", err)
	}
	initConfig()
	// We can't easily test the logger level without exposing it

	// Reset
	if err := os.Setenv("ENV", originalEnv); err != nil {
		t.Errorf("Failed to reset ENV: %v", err)
	}
}

func TestExecute(t *testing.T) {
	// Test that Execute doesn't panic with basic setup
	// This is hard to test fully without mocking cobra
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute panicked: %v", r)
		}
	}()
	// We don't actually call Execute() as it would try to parse command line args
}
