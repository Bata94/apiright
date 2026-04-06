// Package database handles database operations and migrations.
package database

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bata94/apiright/pkg/config"
	"github.com/bata94/apiright/pkg/core"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	_ "github.com/lib/pq"              // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3"    // SQLite driver
)

// Database implements the core.Database interface
type Database struct {
	config     *config.DatabaseConfig
	db         *sql.DB
	driverName string
	logger     core.Logger
	migrations []Migration
}

// Migration represents a database migration
type Migration struct {
	Version    int
	Name       string
	SQL        string
	ExecutedAt *time.Time
	Checksum   string
}

// MigrationResult represents the result of a migration operation
type MigrationResult struct {
	Applied  []Migration
	Pending  []Migration
	Errors   []error
	Duration time.Duration
}

// NewDatabase creates a new database instance
func NewDatabase(cfg *config.DatabaseConfig, logger core.Logger) (*Database, error) {
	db := &Database{
		config: cfg,
		logger: logger,
	}

	// Determine driver name
	switch cfg.Type {
	case "sqlite":
		db.driverName = "sqlite3"
	case "postgres", "postgresql":
		db.driverName = "postgres"
	case "mysql":
		db.driverName = "mysql"
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	return db, nil
}

// Connect establishes a connection to the database
func (d *Database) Connect() error {
	dsn := d.config.GetDatabaseURL()

	db, err := sql.Open(d.driverName, dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	d.db = db
	d.logger.Info("Connected to database", "type", d.config.Type, "driver", d.driverName)
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.db != nil {
		if err := d.db.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
		d.logger.Info("Database connection closed")
	}
	return nil
}

// Ping checks if the database connection is alive
func (d *Database) Ping() error {
	if d.db == nil {
		return fmt.Errorf("database not connected")
	}

	if err := d.db.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// Migrate runs database migrations
func (d *Database) Migrate() error {
	start := time.Now()

	// Load migrations from filesystem
	migrations, err := d.loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}
	d.migrations = migrations

	// Ensure migrations table exists
	if err := d.ensureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get executed migrations
	executed, err := d.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	// Find pending migrations
	pending := d.findPendingMigrations(migrations, executed)

	if len(pending) == 0 {
		d.logger.Info("No pending migrations")
		return nil
	}

	// Execute pending migrations
	result := &MigrationResult{
		Applied:  []Migration{},
		Pending:  pending,
		Errors:   []error{},
		Duration: time.Since(start),
	}

	for _, migration := range pending {
		if err := d.executeMigration(migration); err != nil {
			result.Errors = append(result.Errors, err)
			d.logger.Error("Migration failed", "version", migration.Version, "error", err)
			return fmt.Errorf("migration %d failed: %w", migration.Version, err)
		}

		result.Applied = append(result.Applied, migration)
		d.logger.Info("Migration applied", "version", migration.Version, "name", migration.Name)
	}

	d.logger.Info("Migrations completed", "applied", len(result.Applied), "duration", result.Duration)
	return nil
}

// Connection returns the underlying database connection
func (d *Database) Connection() interface{} {
	return d.db
}

// GetDB returns the *sql.DB instance for direct database access
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// loadMigrations loads migration files from the migrations directory
func (d *Database) loadMigrations() ([]Migration, error) {
	migrationsDir := "migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return []Migration{}, nil // No migrations directory
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		// Extract version from filename (e.g., 001_create_users.sql)
		version, name, err := d.parseMigrationFilename(file.Name())
		if err != nil {
			d.logger.Warn("Skipping invalid migration filename", "file", file.Name())
			continue
		}

		// Read migration file
		filePath := filepath.Join(migrationsDir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		migration := Migration{
			Version:  version,
			Name:     name,
			SQL:      string(content),
			Checksum: d.calculateChecksum(string(content)),
		}

		migrations = append(migrations, migration)
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// parseMigrationFilename extracts version and name from migration filename
func (d *Database) parseMigrationFilename(filename string) (int, string, error) {
	// Remove .sql extension
	base := strings.TrimSuffix(filename, ".sql")

	// Split by underscore to get version and name
	parts := strings.SplitN(base, "_", 2)
	if len(parts) < 2 {
		return 0, "", fmt.Errorf("invalid migration filename format: %s", filename)
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("invalid version in filename %s: %w", filename, err)
	}

	name := strings.ReplaceAll(parts[1], "_", " ")
	return version, name, nil
}

// calculateChecksum calculates SHA-256 checksum of migration content
func (d *Database) calculateChecksum(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// ensureMigrationsTable creates the migrations tracking table
func (d *Database) ensureMigrationsTable() error {
	var createSQL string

	switch d.config.Type {
	case "sqlite":
		createSQL = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			checksum TEXT NOT NULL,
			executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`
	case "postgres", "postgresql":
		createSQL = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			checksum TEXT NOT NULL,
			executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`
	case "mysql":
		createSQL = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			checksum VARCHAR(64) NOT NULL,
			executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
	default:
		return fmt.Errorf("unsupported database type for migrations table: %s", d.config.Type)
	}

	_, err := d.db.Exec(createSQL)
	return err
}

// getExecutedMigrations retrieves already executed migrations from database
func (d *Database) getExecutedMigrations() (map[int]Migration, error) {
	rows, err := d.db.Query(`
		SELECT version, name, checksum, executed_at 
		FROM schema_migrations 
		ORDER BY version
	`)
	if err != nil {
		return nil, err
	}
	defer core.Close("rows", rows, d.logger)

	executed := make(map[int]Migration)
	for rows.Next() {
		var migration Migration
		err := rows.Scan(&migration.Version, &migration.Name, &migration.Checksum, &migration.ExecutedAt)
		if err != nil {
			return nil, err
		}
		executed[migration.Version] = migration
	}

	return executed, nil
}

// findPendingMigrations returns migrations that haven't been executed yet
func (d *Database) findPendingMigrations(all []Migration, executed map[int]Migration) []Migration {
	var pending []Migration
	for _, migration := range all {
		if _, exists := executed[migration.Version]; !exists {
			pending = append(pending, migration)
		}
	}
	return pending
}

// executeMigration executes a single migration
func (d *Database) executeMigration(migration Migration) error {
	// Start transaction
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer core.Rollback("transaction", tx, d.logger)

	// Execute migration SQL
	_, err = tx.Exec(migration.SQL)
	if err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration
	_, err = tx.Exec(`
		INSERT INTO schema_migrations (version, name, checksum, executed_at) 
		VALUES (?, ?, ?, ?)
	`, migration.Version, migration.Name, migration.Checksum, time.Now())
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// GetMigrationStatus returns the current migration status
func (d *Database) GetMigrationStatus() (*MigrationResult, error) {
	// Load migrations
	migrations, err := d.loadMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get executed migrations
	executed, err := d.getExecutedMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get executed migrations: %w", err)
	}

	// Find pending migrations
	pending := d.findPendingMigrations(migrations, executed)

	result := &MigrationResult{
		Applied: make([]Migration, 0),
		Pending: pending,
		Errors:  []error{},
	}

	// Add executed migrations to result
	for _, migration := range executed {
		result.Applied = append(result.Applied, migration)
	}

	return result, nil
}

// CreateMigration creates a new migration file
func (d *Database) CreateMigration(name string) error {
	migrationsDir := "migrations"
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Get next version number
	version, err := d.getNextVersion(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get next version: %w", err)
	}

	// Create migration file
	filename := fmt.Sprintf("%03d_%s.sql", version, strings.ReplaceAll(name, " ", "_"))
	filePath := filepath.Join(migrationsDir, filename)

	template := `-- Migration: %s
-- Version: %d
-- Created: %s

-- Add your migration SQL here

`
	content := fmt.Sprintf(template, name, version, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}

	d.logger.Info("Migration file created", "file", filename, "version", version)
	return nil
}

// getNextVersion calculates the next migration version number
func (d *Database) getNextVersion(migrationsDir string) (int, error) {
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}

	maxVersion := 0
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		version, _, err := d.parseMigrationFilename(file.Name())
		if err != nil {
			continue
		}

		if version > maxVersion {
			maxVersion = version
		}
	}

	return maxVersion + 1, nil
}

func (d *Database) Rollback() error {
	executed, err := d.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	if len(executed) == 0 {
		d.logger.Info("No migrations to rollback")
		return nil
	}

	var lastMigration Migration
	for _, m := range executed {
		lastMigration = m
	}

	d.logger.Info("Rolling back migration", "version", lastMigration.Version, "name", lastMigration.Name)

	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer core.Rollback("transaction", tx, d.logger)

	_, err = tx.Exec("DELETE FROM schema_migrations WHERE version = ?", lastMigration.Version)
	if err != nil {
		return fmt.Errorf("failed to delete migration record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	d.logger.Info("Migration rolled back successfully", "version", lastMigration.Version)
	return nil
}

func (d *Database) Status() (string, error) {
	result, err := d.GetMigrationStatus()
	if err != nil {
		return "", err
	}

	var status strings.Builder
	fmt.Fprintf(&status, "Applied: %d\n", len(result.Applied))
	for _, m := range result.Applied {
		executedAt := "N/A"
		if m.ExecutedAt != nil {
			executedAt = m.ExecutedAt.Format("2006-01-02 15:04:05")
		}
		fmt.Fprintf(&status, "  ✓ %d: %s (executed: %s)\n", m.Version, m.Name, executedAt)
	}

	fmt.Fprintf(&status, "Pending: %d\n", len(result.Pending))
	for _, m := range result.Pending {
		fmt.Fprintf(&status, "  ○ %d: %s\n", m.Version, m.Name)
	}

	return status.String(), nil
}
