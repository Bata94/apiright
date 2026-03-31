package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bata94/apiright/pkg/core"
)

// SchemaParser parses SQL migration files to extract schema information
type SchemaParser struct {
	dialect string // sqlite, postgresql, mysql
	logger  core.Logger
}

// NewSchemaParser creates a new schema parser for specified dialect
func NewSchemaParser(dialect string, logger core.Logger) *SchemaParser {
	return &SchemaParser{
		dialect: dialect,
		logger:  logger,
	}
}

// ParseMigrations parses all SQL migration files in specified directory
func (sp *SchemaParser) ParseMigrations(migrationDir string) (*core.Schema, error) {
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration directory: %w", err)
	}

	schema := &core.Schema{
		Tables:  []core.Table{},
		Queries: []core.Query{},
		Enums:   []core.Enum{},
		Types:   []core.Type{},
	}

	// Process migration files in order
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		filePath := filepath.Join(migrationDir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		tables, err := sp.parseSQLFile(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse migration file %s: %w", file.Name(), err)
		}

		schema.Tables = append(schema.Tables, tables...)
	}

	sp.logger.Info("Parsed schema", "tables", len(schema.Tables), "migrations", len(files))

	return schema, nil
}

// parseSQLFile parses SQL content and extracts table definitions
func (sp *SchemaParser) parseSQLFile(sqlContent string) ([]core.Table, error) {
	var tables []core.Table

	// Remove comments and normalize whitespace
	sqlContent = sp.cleanSQL(sqlContent)

	// Split by semicolons to get individual statements
	statements := sp.splitStatements(sqlContent)

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(strings.ToUpper(stmt), "--") {
			continue
		}

		// Parse CREATE TABLE statements
		if strings.HasPrefix(strings.ToUpper(stmt), "CREATE TABLE") {
			table, err := sp.parseCreateTable(stmt)
			if err != nil {
				return nil, fmt.Errorf("failed to parse CREATE TABLE statement: %w", err)
			}
			if table != nil {
				tables = append(tables, *table)
			}
		}

		// Parse CREATE INDEX statements
		if strings.HasPrefix(strings.ToUpper(stmt), "CREATE INDEX") {
			index, err := sp.parseCreateIndex(stmt)
			if err != nil {
				return nil, fmt.Errorf("failed to parse CREATE INDEX statement: %w", err)
			}
			if index != nil {
				// Add index to appropriate table
				sp.addIndexToTable(&tables, index)
			}
		}
	}

	return tables, nil
}

// cleanSQL removes comments and normalizes SQL content
func (sp *SchemaParser) cleanSQL(sql string) string {
	// Remove SQL comments
	lines := strings.Split(sql, "\n")
	var cleaned []string

	for _, line := range lines {
		// Remove inline comments
		if idx := strings.Index(line, "--"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}

		// Remove multi-line comments (basic implementation)
		if strings.Contains(line, "/*") {
			if idx := strings.Index(line, "*/"); idx >= 0 {
				line = line[:strings.Index(line, "/*")] + line[strings.Index(line, "*/")+2:]
			} else {
				continue // Skip comment lines
			}
		}

		if line != "" {
			cleaned = append(cleaned, line)
		}
	}

	return strings.Join(cleaned, " ")
}

// splitStatements splits SQL content into individual statements
func (sp *SchemaParser) splitStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	inQuotes := false
	var quoteChar rune

	for _, ch := range sql {
		if !inQuotes && (ch == '\'' || ch == '"' || ch == '`') {
			inQuotes = true
			quoteChar = ch
		} else if inQuotes && ch == quoteChar {
			inQuotes = false
		}

		if !inQuotes && ch == ';' {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		} else {
			current.WriteRune(ch)
		}
	}

	// Add the last statement if it doesn't end with semicolon
	lastStmt := strings.TrimSpace(current.String())
	if lastStmt != "" {
		statements = append(statements, lastStmt)
	}

	return statements
}

// parseCreateTable parses a CREATE TABLE statement and extracts table information
func (sp *SchemaParser) parseCreateTable(stmt string) (*core.Table, error) {
	// Extract table name and column definitions
	re := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)\s*\((.*)\)`)
	matches := re.FindStringSubmatch(stmt)
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid CREATE TABLE syntax")
	}

	table := core.Table{
		Name:        matches[1],
		Columns:     []core.Column{},
		PrimaryKey:  []string{},
		Indexes:     []core.Index{},
		ForeignKeys: []core.ForeignKey{},
	}

	// Parse column definitions and constraints
	columnDefs := sp.parseColumnDefinitions(matches[2])

	for _, colDef := range columnDefs {
		// Skip constraints that are not columns
		if sp.isConstraint(colDef) {
			// This is handled at table level, skip for now
			continue
		}

		column, err := sp.parseColumnDefinition(colDef)
		if err != nil {
			return nil, fmt.Errorf("failed to parse column definition '%s': %w", colDef, err)
		}

		table.Columns = append(table.Columns, *column)
	}

	// Debug: check for primary key and handle column-level PRIMARY KEY
	if len(table.PrimaryKey) == 0 {
		// Try to find primary key by looking for AUTOINCREMENT flag in column definitions
		for _, col := range table.Columns {
			colTypeUpper := strings.ToUpper(col.Type)
			if col.AutoIncrement && (strings.Contains(colTypeUpper, "INTEGER") || strings.Contains(colTypeUpper, "INT")) {
				table.PrimaryKey = []string{col.Name}
				break
			}
		}
	}
	return &table, nil
}

// parseColumnDefinitions splits column definitions by comma while respecting quotes
func (sp *SchemaParser) parseColumnDefinitions(columnsSection string) []string {
	var defs []string
	var current strings.Builder
	inQuotes := false
	var quoteChar rune
	parenLevel := 0

	for _, ch := range columnsSection {
		if !inQuotes && (ch == '\'' || ch == '"' || ch == '`') {
			inQuotes = true
			quoteChar = ch
		} else if inQuotes && ch == quoteChar {
			inQuotes = false
		}

		if !inQuotes {
			switch ch {
			case '(':
				parenLevel++
			case ')':
				parenLevel--
			}
		}

		if !inQuotes && ch == ',' && parenLevel == 0 {
			def := strings.TrimSpace(current.String())
			if def != "" {
				defs = append(defs, def)
			}
			current.Reset()
		} else {
			current.WriteRune(ch)
		}
	}

	// Add the last definition
	lastDef := strings.TrimSpace(current.String())
	if lastDef != "" {
		defs = append(defs, lastDef)
	}

	return defs
}

// isConstraint checks if a definition is a constraint rather than a column
func (sp *SchemaParser) isConstraint(def string) bool {
	defUpper := strings.ToUpper(strings.TrimSpace(def))
	return strings.HasPrefix(defUpper, "PRIMARY KEY") ||
		strings.HasPrefix(defUpper, "FOREIGN KEY") ||
		strings.HasPrefix(defUpper, "UNIQUE") ||
		strings.HasPrefix(defUpper, "CHECK") ||
		strings.HasPrefix(defUpper, "CONSTRAINT")
}

// parseColumnDefinition parses a single column definition
func (sp *SchemaParser) parseColumnDefinition(def string) (*core.Column, error) {
	// Basic regex for column: name type [constraints]
	re := regexp.MustCompile(`^(\w+)\s+([A-Z]+)(?:\([^)]*\))?\s*(.*)$`)
	matches := re.FindStringSubmatch(def)
	if len(matches) < 3 {
		return nil, fmt.Errorf("invalid column definition syntax: %s", def)
	}

	column := core.Column{
		Name:     matches[1],
		Type:     strings.ToUpper(matches[2]),
		Nullable: true, // Default to nullable
		Default:  "",
	}

	// Parse constraints
	constraints := strings.TrimSpace(matches[3])
	if constraints != "" {
		sp.parseColumnConstraints(constraints, &column)
	}

	return &column, nil
}

// parseColumnConstraints parses column constraints like NOT NULL, DEFAULT, etc.
func (sp *SchemaParser) parseColumnConstraints(constraints string, column *core.Column) {
	parts := strings.Fields(constraints)

	for i := 0; i < len(parts); i++ {
		part := parts[i]
		partUpper := strings.ToUpper(part)

		switch partUpper {
		case "NOT":
			if i+1 < len(parts) && strings.ToUpper(parts[i+1]) == "NULL" {
				column.Nullable = false
				i++ // Skip next part
				continue
			}
		case "DEFAULT":
			if i+1 < len(parts) {
				column.Default = parts[i+1]
				i++ // Skip next part
				continue
			}
		case "PRIMARY":
			if i+1 < len(parts) && strings.ToUpper(parts[i+1]) == "KEY" {
				// Mark column as primary key - table-level constraint handled elsewhere
				i++ // Skip next part
				continue
			}
		case "UNIQUE":
			// Handle UNIQUE constraint
		case "AUTOINCREMENT", "AUTO_INCREMENT":
			// Handle auto-increment
			column.AutoIncrement = true
		}
	}
}

// addIndexToTable adds an index to the appropriate table
func (sp *SchemaParser) addIndexToTable(tables *[]core.Table, index *core.Index) {
	for i, table := range *tables {
		if table.Name == index.Columns[0] { // Simple heuristic - improve this
			(*tables)[i].Indexes = append((*tables)[i].Indexes, *index)
			return
		}
	}
}

// parseCreateIndex parses a CREATE INDEX statement
func (sp *SchemaParser) parseCreateIndex(stmt string) (*core.Index, error) {
	// Basic regex for CREATE INDEX
	re := regexp.MustCompile(`(?i)CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)\s+ON\s+(\w+)\s*\(([^)]+)\)`)
	matches := re.FindStringSubmatch(stmt)
	if len(matches) != 4 {
		return nil, fmt.Errorf("invalid CREATE INDEX syntax")
	}

	columns := strings.Split(strings.ReplaceAll(matches[3], " ", ""), ",")

	return &core.Index{
		Name:    matches[1],
		Columns: columns,
		Unique:  strings.Contains(strings.ToUpper(stmt), "UNIQUE INDEX"),
	}, nil
}
