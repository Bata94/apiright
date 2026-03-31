package apiright

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bata94/apiright/pkg/config"
	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/database"
	"github.com/spf13/cobra"
)

type ProjectContext struct {
	Dir      string
	Logger   core.Logger
	Config   *config.Config
	Database *database.Database
}

func GetProjectDir(cmd *cobra.Command) (string, error) {
	if projectDir, _ := cmd.Flags().GetString("project"); projectDir != "" {
		return filepath.Abs(projectDir)
	}
	return os.Getwd()
}

func SetupLogger() (core.Logger, error) {
	logger, err := core.NewLogger(devMode)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	return logger, nil
}

func LoadProjectConfig(dir string) (*config.Config, error) {
	cfg, err := config.LoadConfig(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}

func ConnectDatabase(cfg *config.DatabaseConfig, logger core.Logger) (*database.Database, error) {
	db, err := database.NewDatabase(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}
	if err := db.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	return db, nil
}

func NewProjectContext(cmd *cobra.Command) (*ProjectContext, error) {
	dir, err := GetProjectDir(cmd)
	if err != nil {
		return nil, err
	}

	logger, err := SetupLogger()
	if err != nil {
		return nil, err
	}

	cfg, err := LoadProjectConfig(dir)
	if err != nil {
		return nil, err
	}

	return &ProjectContext{
		Dir:    dir,
		Logger: logger,
		Config: cfg,
	}, nil
}

func (pc *ProjectContext) Close() {
	if pc.Database != nil {
		core.Close("database", pc.Database, pc.Logger)
	}
	core.SyncLogger(pc.Logger)
}
