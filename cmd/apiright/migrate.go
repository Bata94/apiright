package apiright

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Long:  `Run pending database migrations using Goose. Manage database schema versions with up/down commands.`,
		Example: `  apiright migrate up
  apiright migrate down
  apiright migrate status
  apiright migrate create add_users_table`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:     "up",
		Short:   "Run pending migrations",
		Long:    `Runs all pending migrations, bringing the database schema to the latest version.`,
		Example: `  apiright migrate up`,
		RunE:    runMigrateUp,
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "down",
		Short:   "Rollback last migration",
		Long:    `Rolls back the most recent migration, reverting the database schema by one version.`,
		Example: `  apiright migrate down`,
		RunE:    runMigrateDown,
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "status",
		Short:   "Show migration status",
		Long:    `Displays the current migration status, showing which migrations have been applied.`,
		Example: `  apiright migrate status`,
		RunE:    runMigrateStatus,
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "create [name]",
		Short:   "Create a new migration",
		Long:    `Creates a new migration file pair with up/down templates in the migrations directory.`,
		Example: `  apiright migrate create add_users_table`,
		Args:    cobra.ExactArgs(1),
		RunE:    runMigrateCreate,
	})

	return cmd
}

func runMigrateUp(cmd *cobra.Command, args []string) error {
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

	if err := db.Migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	pc.Logger.Info("Migrations completed successfully")
	return nil
}

func runMigrateDown(cmd *cobra.Command, args []string) error {
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

	if err := db.Rollback(); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	pc.Logger.Info("Migration rolled back successfully")
	return nil
}

func runMigrateStatus(cmd *cobra.Command, args []string) error {
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

	status, err := db.Status()
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	fmt.Println("Migration Status:")
	fmt.Println(status)
	return nil
}

func runMigrateCreate(cmd *cobra.Command, args []string) error {
	projectDir, err := GetProjectDir(cmd)
	if err != nil {
		return err
	}

	migrationDir := projectDir + "/migrations"
	if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
		if err := os.MkdirAll(migrationDir, 0755); err != nil {
			return fmt.Errorf("failed to create migrations directory: %w", err)
		}
	}

	fmt.Printf("Migration template created in: %s\n", migrationDir)
	fmt.Println("Please create your migration file with the naming convention:")
	fmt.Println("  XXXX_description_up.sql")
	fmt.Println("  XXXX_description_down.sql")

	return nil
}
