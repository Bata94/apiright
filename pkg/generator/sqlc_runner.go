package generator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bata94/apiright/pkg/core"
)

// SQLCRunner handles integration with sqlc tool
type SQLCRunner struct {
	workDir string
	verbose bool
	logger  core.Logger
}

// NewSQLCRunner creates a new sqlc runner
func NewSQLCRunner(workDir string, verbose bool, logger core.Logger) *SQLCRunner {
	return &SQLCRunner{
		workDir: workDir,
		verbose: verbose,
		logger:  logger,
	}
}

// Generate executes sqlc generate command
func (sr *SQLCRunner) Generate(ctx context.Context) error {
	if err := sr.validateEnvironment(); err != nil {
		return sr.formatError("environment_validation", err, "")
	}

	if err := sr.validateConfig(); err != nil {
		return sr.formatError("config_validation", err, "")
	}

	cmd, err := sr.buildCommand(ctx)
	if err != nil {
		return sr.formatError("command_build", err, "")
	}

	if err := sr.executeCommand(cmd); err != nil {
		return sr.formatError("execution", err, "")
	}

	sr.logger.Info("sqlc generation completed successfully")
	return nil
}

// validateEnvironment checks if sqlc is available
func (sr *SQLCRunner) validateEnvironment() error {
	sqlcPath, err := sr.findSQLCExecutable()
	if err != nil {
		return fmt.Errorf("sqlc not found: %w", err)
	}

	// Check version
	versionCmd := exec.CommandContext(context.Background(), sqlcPath, "version")
	output, err := versionCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sqlc version check failed: %w", err)
	}

	version := sr.parseVersionOutput(string(output))
	sr.logger.Debug("Found sqlc", "version", version, "path", sqlcPath)

	return nil
}

// validateConfig checks if sqlc.yaml exists and is valid
func (sr *SQLCRunner) validateConfig() error {
	configPath := filepath.Join(sr.workDir, "sqlc.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("sqlc.yaml not found in %s", sr.workDir)
	}

	// Basic validation of sqlc.yaml structure
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read sqlc.yaml: %w", err)
	}

	configContent := string(content)

	// Check for required fields
	requiredFields := []string{"version", "sql", "gen"}
	for _, field := range requiredFields {
		if !strings.Contains(configContent, field) {
			return fmt.Errorf("sqlc.yaml missing required field: %s", field)
		}
	}

	return nil
}

// buildCommand builds the sqlc generate command
func (sr *SQLCRunner) buildCommand(ctx context.Context) (*exec.Cmd, error) {
	sqlcPath, err := sr.findSQLCExecutable()
	if err != nil {
		return nil, err
	}

	args := []string{"generate"}

	// sqlc doesn't have a --verbose flag, but we can handle verbose output differently
	if sr.verbose {
		sr.logger.Info("sqlc verbose mode enabled (but sqlc doesn't support --verbose flag)")
	}

	cmd := exec.CommandContext(ctx, sqlcPath, args...)
	cmd.Dir = sr.workDir

	sr.logger.Debug("Executing sqlc command", "cmd", cmd.String(), "dir", sr.workDir)

	return cmd, nil
}

// executeCommand runs the sqlc command and processes output
func (sr *SQLCRunner) executeCommand(cmd *exec.Cmd) error {
	if sr.verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Parse sqlc output for user-friendly errors
		return sr.parseSQLError(string(output))
	}

	if sr.verbose {
		sr.logger.Debug("sqlc output", "output", string(output))
	}

	return nil
}

// findSQLCExecutable finds the sqlc executable
func (sr *SQLCRunner) findSQLCExecutable() (string, error) {
	// Check common locations
	paths := []string{
		"sqlc", // in PATH
		"/usr/local/bin/sqlc",
		"/usr/bin/sqlc",
		"C:\\Program Files\\sqlc\\sqlc.exe",
		"C:\\Program Files (x86)\\sqlc\\sqlc.exe",
	}

	for _, path := range paths {
		if _, err := exec.LookPath(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("sqlc executable not found in common paths. Please install sqlc from https://sqlc.dev")
}

// parseVersionOutput parses version output to get version string
func (sr *SQLCRunner) parseVersionOutput(output string) string {
	// Extract version from output like "sqlc version v1.30.0"
	re := regexp.MustCompile(`v(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		return matches[1]
	}
	return "unknown"
}

// parseSQLError parses sqlc error output for user-friendly messages
func (sr *SQLCRunner) parseSQLError(output string) error {
	output = strings.ToLower(output)

	// Common sqlc errors and their user-friendly versions
	errorPatterns := map[string]string{
		"no such table":          "sqlc couldn't find table. Check if your migration files CREATE the tables referenced in your queries.",
		"column.*does not exist": "sqlc couldn't find column. Check column names in your table definitions and queries.",
		"invalid syntax":         "Invalid SQL syntax in your query. Check for typos or incorrect SQL grammar.",
		"type mismatch":          "Type mismatch between column and usage. Check that your query uses the right data types.",
		"ambiguous column":       "Column name exists in multiple tables. Use table.column format to disambiguate.",
		"missing from clause":    "Your query is missing a FROM clause. Add a FROM clause to specify the table.",
		"missing where clause":   "Your UPDATE/DELETE query is missing a WHERE clause. Add a WHERE clause for safety.",
	}

	for pattern, message := range errorPatterns {
		if strings.Contains(output, pattern) {
			return fmt.Errorf("sqlc error: %s", message)
		}
	}

	// If we can't parse it, return the original error
	return fmt.Errorf("sqlc execution failed: %s", output)
}

// formatError creates a formatted error with context
func (sr *SQLCRunner) formatError(errorType string, err error, details string) error {
	switch errorType {
	case "environment_validation":
		return fmt.Errorf("sqlc environment error: %w", err)
	case "config_validation":
		return fmt.Errorf("sqlc configuration error: %w", err)
	case "command_build":
		return fmt.Errorf("sqlc command error: %w", err)
	case "execution":
		return fmt.Errorf("sqlc execution error: %w", err)
	default:
		return fmt.Errorf("sqlc error: %w", err)
	}
}
