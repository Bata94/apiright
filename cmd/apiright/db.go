package apiright

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewDBCCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database management commands",
		Long:  `Manage database operations like create, drop, reset, seed, and connectivity checks.`,
		Example: `  apiright db ping
  apiright db create
  apiright db reset
  apiright db seed`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:     "create",
		Short:   "Create the database",
		Long:    `Creates a new database based on the configuration.`,
		Example: `  apiright db create`,
		RunE:    runDBCreate,
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "drop",
		Short:   "Drop the database",
		Long:    `Drops the existing database. Use with caution.`,
		Example: `  apiright db drop`,
		RunE:    runDBDrop,
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "reset",
		Short:   "Drop and recreate the database",
		Long:    `Drops the database and recreates it fresh.`,
		Example: `  apiright db reset`,
		RunE:    runDBReset,
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "seed",
		Short:   "Run seed data",
		Long:    `Loads seed data from the seeds directory into the database.`,
		Example: `  apiright db seed`,
		RunE:    runDBSeed,
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "ping",
		Short:   "Check database connectivity",
		Long:    `Tests the database connection and reports status.`,
		Example: `  apiright db ping`,
		RunE:    runDBPing,
	})

	return cmd
}

func runDBCreate(cmd *cobra.Command, args []string) error {
	pc, err := NewProjectContext(cmd)
	if err != nil {
		return err
	}
	defer pc.Close()

	_, err = ConnectDatabase(&pc.Config.Database, pc.Logger)
	if err != nil {
		return err
	}
	defer pc.Close()

	pc.Logger.Info("Database created successfully", "type", pc.Config.Database.Type)
	return nil
}

func runDBDrop(cmd *cobra.Command, args []string) error {
	pc, err := NewProjectContext(cmd)
	if err != nil {
		return err
	}
	defer pc.Close()

	if pc.Config.Database.Type == "sqlite" {
		dbPath := pc.Config.Database.Name
		if dbPath == "" {
			dbPath = "database.db"
		}

		if _, err := os.Stat(dbPath); err == nil {
			if err := os.Remove(dbPath); err != nil {
				return fmt.Errorf("failed to drop database: %w", err)
			}
			pc.Logger.Info("Database dropped", "path", dbPath)
		} else {
			pc.Logger.Info("Database file does not exist", "path", dbPath)
		}
		return nil
	}

	pc.Logger.Warn("Drop command only supported for SQLite in development mode")
	return nil
}

func runDBReset(cmd *cobra.Command, args []string) error {
	if err := runDBDrop(cmd, args); err != nil {
		return err
	}
	return runDBCreate(cmd, args)
}

func runDBSeed(cmd *cobra.Command, args []string) error {
	pc, err := NewProjectContext(cmd)
	if err != nil {
		return err
	}
	defer pc.Close()

	_, err = ConnectDatabase(&pc.Config.Database, pc.Logger)
	if err != nil {
		return err
	}
	defer pc.Close()

	seedDir := pc.Dir + "/seeds"
	if _, err := os.Stat(seedDir); os.IsNotExist(err) {
		pc.Logger.Info("No seeds directory found, skipping seed")
		return nil
	}

	pc.Logger.Info("Seed data would be loaded from", "directory", seedDir)
	pc.Logger.Warn("Seed functionality not yet implemented")
	return nil
}

func runDBPing(cmd *cobra.Command, args []string) error {
	pc, err := NewProjectContext(cmd)
	if err != nil {
		return err
	}
	defer pc.Close()

	db, err := ConnectDatabase(&pc.Config.Database, pc.Logger)
	if err != nil {
		return err
	}
	defer pc.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	pc.Logger.Info("Database ping successful", "type", pc.Config.Database.Type)
	return nil
}
