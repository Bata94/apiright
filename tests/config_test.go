package apiright_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bata94/apiright/pkg/config"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name       string
		projectDir string
		wantErr    bool
	}{
		{
			name:       "valid config",
			projectDir: "testdata/init",
			wantErr:    false,
		},
		{
			name:       "nonexistent directory returns default config",
			projectDir: "testdata/nonexistent",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadConfig(tt.projectDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && cfg == nil {
				t.Error("LoadConfig() returned nil config")
			}
		})
	}
}

func TestLoadConfigFromAbsolutePath(t *testing.T) {
	absPath, err := filepath.Abs("testdata/init")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	cfg, err := config.LoadConfig(absPath)
	if err != nil {
		t.Errorf("LoadConfig() from absolute path failed: %v", err)
		return
	}

	if cfg.Project.Name == "" {
		t.Error("LoadConfig() returned empty project name")
	}
}

func TestLoadConfigEnvOverride(t *testing.T) {
	projectDir := "testdata/init"

	originalVal := os.Getenv("APIRIGHT_CONFIG_PATH")
	restoreEnv := func() {
		_ = os.Setenv("APIRIGHT_CONFIG_PATH", originalVal)
	}
	defer restoreEnv()

	_ = os.Setenv("APIRIGHT_CONFIG_PATH", "")
	cfg, err := config.LoadConfig(projectDir)
	if err != nil {
		t.Errorf("LoadConfig() with empty env var failed: %v", err)
		return
	}

	if cfg == nil {
		t.Error("LoadConfig() returned nil config")
	}
}
