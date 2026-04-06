package generator

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/bata94/apiright/pkg/core"
)

// Dialect represents a SQL dialect
type Dialect string

const (
	DialectSQLite   Dialect = "sqlite"
	DialectPostgres Dialect = "postgres"
	DialectMySQL    Dialect = "mysql"
)

// SQLGenerator generates CRUD SQL queries using Go templates
type SQLGenerator struct {
	templates *template.Template
	genSuffix string
	logger    core.Logger
	dialect   Dialect
}

// TableData represents data for SQL template generation
type TableData struct {
	Name            string
	Title           string // Title case for naming
	Columns         []ColumnData
	PrimaryKey      ColumnData
	PrimaryKeyNames []string // Add for composite keys
	ColumnsList     string
	InsertColumns   string
	InsertValues    string
	UpdateSet       string
	PrimaryKeyWhere string
	OrderByClause   string // Add for LIST query ORDER BY
	Dialect         Dialect
	HasReturning    bool // True if dialect supports RETURNING clause
}

// ColumnData represents column data for template generation
type ColumnData struct {
	Name     string
	Type     string
	Nullable bool
	Default  string
	IsPK     bool
	GoType   string
}

// NewSQLGenerator creates a new SQL generator
func NewSQLGenerator(genSuffix string, dialect Dialect, logger core.Logger) *SQLGenerator {
	if dialect == "" {
		dialect = DialectSQLite // Default to SQLite
	}
	return &SQLGenerator{
		genSuffix: genSuffix,
		logger:    logger,
		dialect:   dialect,
	}
}

// GenerateQueries generates CRUD queries for all tables in the schema
func (sg *SQLGenerator) GenerateQueries(schema *core.Schema, ctx *core.GenerationContext) error {
	sg.logger.Info("Starting SQL generation", "tables", len(schema.Tables))

	if err := sg.parseTemplates(); err != nil {
		return fmt.Errorf("failed to parse SQL templates: %w", err)
	}

	for _, table := range schema.Tables {
		if err := sg.generateTableQueries(table, ctx); err != nil {
			return fmt.Errorf("failed to generate queries for table %s: %w", table.Name, err)
		}
	}

	sg.logger.Info("Generated SQL queries", "tables", len(schema.Tables))
	return nil
}

// parseTemplates initializes SQL generation templates
func (sg *SQLGenerator) parseTemplates() error {
	templates := map[string]string{
		"get":    getQueryTemplate,
		"list":   listQueryTemplate,
		"create": sg.getCreateQueryTemplate(),
		"update": sg.getUpdateQueryTemplate(),
		"delete": deleteQueryTemplate,
	}

	sg.templates = template.New("sql").Option("missingkey=error")
	for name, content := range templates {
		_, err := sg.templates.New(name).Parse(content)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}
	}

	return nil
}

// generateTableQueries generates CRUD queries for a single table
func (sg *SQLGenerator) generateTableQueries(table core.Table, ctx *core.GenerationContext) error {
	// Convert table data for template execution
	tableData := sg.prepareTableData(table)

	// Generate each query type
	queries := map[string]string{
		"get":    sg.executeTemplate("get", tableData),
		"list":   sg.executeTemplate("list", tableData),
		"create": sg.executeTemplate("create", tableData),
		"update": sg.executeTemplate("update", tableData),
		"delete": sg.executeTemplate("delete", tableData),
	}

	// Combine all queries into single file
	fileContent := sg.combineQueries(queries, tableData)

	// Write to gen/sql/{table}_ar_gen.sql
	outputPath := ctx.Join(ctx.ProjectDir, "gen", "sql", table.Name+sg.genSuffix+".sql")
	if err := ctx.WriteFile(outputPath, []byte(fileContent), 0644); err != nil {
		return fmt.Errorf("failed to write query file: %w", err)
	}

	sg.logger.Debug("Generated queries for table", "table", table.Name, "path", outputPath)
	return nil
}

// prepareTableData converts core.Table to TableData for template execution
func (sg *SQLGenerator) prepareTableData(table core.Table) TableData {
	var columns []ColumnData
	var primaryKey ColumnData
	var columnNames []string

	var primaryKeys []string
	for _, col := range table.Columns {
		isPK := sg.isPrimaryKey(table, col.Name)

		colData := ColumnData{
			Name:     col.Name,
			Type:     col.Type,
			Nullable: col.Nullable,
			Default:  col.Default,
			IsPK:     isPK,
			GoType:   core.SQLToGoType(col.Type),
		}

		columns = append(columns, colData)
		columnNames = append(columnNames, col.Name)

		if colData.IsPK {
			primaryKey = colData
			primaryKeys = append(primaryKeys, col.Name)
		}
	}

	// Check if dialect supports RETURNING clause
	hasReturning := sg.dialect == DialectPostgres

	tableData := TableData{
		Name:            table.Name,
		Title:           sg.toTitleCase(table.Name),
		Columns:         columns,
		PrimaryKey:      primaryKey,
		PrimaryKeyNames: primaryKeys,
		ColumnsList:     strings.Join(columnNames, ", "),
		Dialect:         sg.dialect,
		HasReturning:    hasReturning,
	}

	// Prepare additional template data
	sg.prepareQuerySpecificData(&tableData)

	return tableData
}

// prepareQuerySpecificData prepares data specific to query types
func (sg *SQLGenerator) prepareQuerySpecificData(data *TableData) {
	var insertColumns []string
	var insertValues []string
	var updateSet []string
	var pkWhere []string

	for _, col := range data.Columns {
		// INSERT columns and values - skip only auto-increment PK for INSERT
		if col.IsPK && col.GoType == "int64" {
			// Skip auto-increment primary key for INSERT only
		} else {
			insertColumns = append(insertColumns, col.Name)
			insertValues = append(insertValues, "?")
		}

		// UPDATE SET clause
		if !col.IsPK {
			updateSet = append(updateSet, fmt.Sprintf("%s = ?", col.Name))
		}

		// WHERE clause for PK - include ALL PK columns
		if col.IsPK {
			pkWhere = append(pkWhere, fmt.Sprintf("%s = ?", col.Name))
		}
	}

	data.InsertColumns = strings.Join(insertColumns, ", ")
	data.InsertValues = strings.Join(insertValues, ", ")
	data.UpdateSet = strings.Join(updateSet, ", ")
	data.PrimaryKeyWhere = strings.Join(pkWhere, " AND ")

	// Set OrderByClause - use primary keys if available, otherwise first column
	if len(data.PrimaryKeyNames) > 0 {
		data.OrderByClause = strings.Join(data.PrimaryKeyNames, ", ")
	} else if len(data.Columns) > 0 {
		data.OrderByClause = data.Columns[0].Name
	}

}

// isPrimaryKey checks if a column is part of the primary key
func (sg *SQLGenerator) isPrimaryKey(table core.Table, columnName string) bool {
	for _, pk := range table.PrimaryKey {
		if pk == columnName {
			return true
		}
	}
	return false
}

// executeTemplate executes a template with table data
func (sg *SQLGenerator) executeTemplate(templateName string, data TableData) string {
	var buf strings.Builder
	err := sg.templates.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		sg.logger.Error("Template execution failed", "template", templateName, "error", err)
		return ""
	}
	return buf.String()
}

// combineQueries combines all generated queries into a single file
func (sg *SQLGenerator) combineQueries(queries map[string]string, data TableData) string {
	var result strings.Builder

	// Add header comment
	fmt.Fprintf(&result, `-- Generated CRUD queries for table %s
-- This file is auto-generated by APIRight. Do not edit directly.
-- Use the queries/ directory for custom queries.

`, data.Name)

	// Add each query in order
	order := []string{"get", "list", "create", "update", "delete"}
	for _, queryType := range order {
		if query, exists := queries[queryType]; exists && query != "" {
			fmt.Fprintf(&result, "-- %s\n%s\n\n", toTitle(queryType), query)
		}
	}

	return result.String()
}

// toTitleCase converts a string to TitleCase for naming
func (sg *SQLGenerator) toTitleCase(s string) string {
	if s == "" {
		return ""
	}

	// Handle common plural forms
	if strings.HasSuffix(s, "ies") {
		return sg.toTitleCase(s[:len(s)-3] + "y")
	}
	if strings.HasSuffix(s, "es") {
		return sg.toTitleCase(s[:len(s)-2])
	}
	if strings.HasSuffix(s, "s") {
		return sg.toTitleCase(s[:len(s)-1])
	}

	// Simple snake_case to TitleCase conversion
	return toTitle(s)
}

// toTitle capitalizes the first letter of a string
func toTitle(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

// SQL generation templates - dialect-aware
// Note: RETURNING is only supported by PostgreSQL and SQLite 3.35+
const (
	getQueryTemplate = `-- name: Get{{.Title}}_ar_gen :one
SELECT {{.ColumnsList}} FROM {{.Name}} WHERE {{.PrimaryKeyWhere}} LIMIT 1;`

	listQueryTemplate = `-- name: List{{.Title}}_ar_gen :many
SELECT {{.ColumnsList}} FROM {{.Name}} ORDER BY {{.OrderByClause}} LIMIT ? OFFSET ?;`

	createQueryTemplatePostgres = `-- name: Create{{.Title}}_ar_gen :one
INSERT INTO {{.Name}} ({{.InsertColumns}}) VALUES ({{.InsertValues}}) RETURNING {{.ColumnsList}};`

	createQueryTemplateSQLite = `-- name: Create{{.Title}}_ar_gen :one
INSERT INTO {{.Name}} ({{.InsertColumns}}) VALUES ({{.InsertValues}});
SELECT last_insert_rowid() AS id;`

	createQueryTemplateMySQL = `-- name: Create{{.Title}}_ar_gen :one
INSERT INTO {{.Name}} ({{.InsertColumns}}) VALUES ({{.InsertValues}});
SELECT LAST_INSERT_ID() AS id;`

	updateQueryTemplatePostgres = `-- name: Update{{.Title}}_ar_gen :one
UPDATE {{.Name}} SET {{.UpdateSet}} WHERE {{.PrimaryKeyWhere}} RETURNING {{.ColumnsList}};`

	updateQueryTemplateGeneric = `-- name: Update{{.Title}}_ar_gen :one
UPDATE {{.Name}} SET {{.UpdateSet}} WHERE {{.PrimaryKeyWhere}};`

	deleteQueryTemplate = `-- name: Delete{{.Title}}_ar_gen :exec
DELETE FROM {{.Name}} WHERE {{.PrimaryKeyWhere}};`
)

// getCreateQueryTemplate returns the appropriate create query template for the dialect
func (sg *SQLGenerator) getCreateQueryTemplate() string {
	switch sg.dialect {
	case DialectPostgres:
		return createQueryTemplatePostgres
	case DialectMySQL:
		return createQueryTemplateMySQL
	default:
		return createQueryTemplateSQLite
	}
}

// getUpdateQueryTemplate returns the appropriate update query template for the dialect
func (sg *SQLGenerator) getUpdateQueryTemplate() string {
	switch sg.dialect {
	case DialectPostgres:
		return updateQueryTemplatePostgres
	default:
		return updateQueryTemplateGeneric
	}
}
