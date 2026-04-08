package generator

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/bata94/apiright/pkg/core"
)

// ServiceGenerator generates Go service implementations using sqlc Querier
type ServiceGenerator struct {
	templates *template.Template
	genSuffix string
	logger    core.Logger
}

// ServiceData represents data for service template generation
type ServiceData struct {
	Name         string
	Title        string // Title case for naming
	PackageName  string
	TableName    string
	ModelName    string
	QuerierName  string
	Columns      []ColumnData
	CreateParams []ColumnData // Columns for Create (excludes auto-increment PK)
	UpdateParams []ColumnData // Columns for Update (includes PK)
	PrimaryKey   ColumnData
	Imports      []string
	DBImport     string
}

// NewServiceGenerator creates a new service generator
func NewServiceGenerator(genSuffix string, logger core.Logger) *ServiceGenerator {
	return &ServiceGenerator{
		genSuffix: genSuffix,
		logger:    logger,
	}
}

// GenerateServices generates service implementations for all tables in schema
func (sg *ServiceGenerator) GenerateServices(schema *core.Schema, ctx *core.GenerationContext) error {
	sg.logger.Info("Starting service generation", "tables", len(schema.Tables))

	if err := sg.parseTemplates(); err != nil {
		return fmt.Errorf("failed to parse service templates: %w", err)
	}

	for _, table := range schema.Tables {
		if err := sg.generateTableService(table, ctx); err != nil {
			return fmt.Errorf("failed to generate service for table %s: %w", table.Name, err)
		}
	}

	sg.logger.Info("Generated service implementations", "tables", len(schema.Tables))
	return nil
}

// parseTemplates initializes service generation templates
func (sg *ServiceGenerator) parseTemplates() error {
	templates := map[string]string{
		"service": serviceTemplate,
	}

	sg.templates = template.New("service").Option("missingkey=error")

	// Add template function for PascalCase conversion
	sg.templates.Funcs(template.FuncMap{
		"pascal": sg.toTitleCase,
	})

	for name, content := range templates {
		_, err := sg.templates.New(name).Parse(content)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}
	}

	return nil
}

// generateTableService generates service implementation for a single table
func (sg *ServiceGenerator) generateTableService(table core.Table, ctx *core.GenerationContext) error {
	// Convert table data for template execution
	serviceData := sg.prepareServiceData(table, ctx)

	// Generate service code
	serviceCode := sg.executeTemplate("service", serviceData)

	// Write to gen/go/services/{table}_service.go
	outputPath := ctx.Join(ctx.ProjectDir, "gen", "go", "services", table.Name+"_service.go")
	if err := ctx.WriteFile(outputPath, []byte(serviceCode), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	sg.logger.Debug("Generated service for table", "table", table.Name, "path", outputPath)
	return nil
}

// prepareServiceData converts core.Table to ServiceData for template execution
func (sg *ServiceGenerator) prepareServiceData(table core.Table, ctx *core.GenerationContext) ServiceData {
	var columns []ColumnData
	var primaryKey ColumnData

	// Find primary key
	var pkName string
	if len(table.PrimaryKey) > 0 {
		pkName = table.PrimaryKey[0]
	}

	for _, col := range table.Columns {
		isPK := col.Name == pkName

		colData := ColumnData{
			Name:     col.Name,
			Type:     col.Type,
			Nullable: col.Nullable,
			Default:  col.Default,
			IsPK:     isPK,
			GoType:   core.SQLToGoType(col.Type),
		}

		columns = append(columns, colData)

		if colData.IsPK {
			primaryKey = colData
		}
	}

	// Prepare imports
	imports := []string{
		"context",
		"database/sql",
		"fmt",
		"github.com/bata94/apiright/pkg/core",
	}

	// Check if we need time import
	for _, col := range columns {
		if col.GoType == "time.Time" {
			imports = append(imports, "time")
			break
		}
	}

	tableName := table.Name
	titleName := sg.toTitleCase(singularize(tableName))
	modelName := sg.toTitleCase(tableName)
	querierName := "db.Querier"

	// Build create params (all non-PK columns)
	var createParams []ColumnData
	for _, col := range columns {
		if !col.IsPK {
			createParams = append(createParams, col)
		}
	}

	// Build update params (all columns including PK)
	updateParams := append([]ColumnData{}, columns...)

	// Build dynamic db import path
	// The db package is in gen/go directory with package name "db"
	dbImport := ctx.ModulePath + "/gen/go"
	if dbImport == "/gen/go" {
		// Fallback if module path not set
		dbImport = "github.com/yourmodule/gen/go"
	}

	return ServiceData{
		Name:         tableName,
		Title:        titleName,
		PackageName:  "services",
		TableName:    tableName,
		ModelName:    modelName,
		QuerierName:  querierName,
		Columns:      columns,
		CreateParams: createParams,
		UpdateParams: updateParams,
		PrimaryKey:   primaryKey,
		Imports:      imports,
		DBImport:     dbImport,
	}
}

// executeTemplate executes a template with service data
func (sg *ServiceGenerator) executeTemplate(templateName string, data ServiceData) string {
	var buf strings.Builder
	err := sg.templates.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		sg.logger.Error("Template execution failed", "template", templateName, "error", err)
		return ""
	}
	return buf.String()
}

// toTitleCase converts a string to PascalCase for naming
func (sg *ServiceGenerator) toTitleCase(s string) string {
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

	// Simple snake_case to PascalCase conversion
	words := strings.Split(s, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}
	return strings.Join(words, "")
}

// singularize converts plural to singular form (basic implementation)
func singularize(word string) string {
	if strings.HasSuffix(word, "ies") {
		return word[:len(word)-3] + "y"
	}
	if strings.HasSuffix(word, "es") {
		return word[:len(word)-2]
	}
	if strings.HasSuffix(word, "s") {
		return word[:len(word)-1]
	}
	return word
}

// Service generation template
const serviceTemplate = `// Code generated by APIRight. DO NOT EDIT.
// Generated service for table {{.TableName}}

package services

import (
	{{- range .Imports}}
	"{{.}}"
	{{- end}}
	db "{{.DBImport}}"
)

// Create{{.Title}}Params holds parameters for creating a {{.Title}}
type Create{{.Title}}Params struct {
{{- range .CreateParams}}
	{{pascal .Name}} {{.GoType}}
{{- end}}
}

// Update{{.Title}}Params holds parameters for updating a {{.Title}}
type Update{{.Title}}Params struct {
	Id {{.PrimaryKey.GoType}}
{{- range .UpdateParams}}
	{{pascal .Name}} {{.GoType}}
{{- end}}
}

// {{.Title}}Service provides CRUD operations for {{.TableName}} table
type {{.Title}}Service struct {
	querier {{.QuerierName}}
	logger  core.Logger
}

// New{{.Title}}Service creates a new {{.Title}}Service
func New{{.Title}}Service(querier {{.QuerierName}}, logger core.Logger) *{{.Title}}Service {
	return &{{.Title}}Service{
		querier: querier,
		logger:  logger,
	}
}

// Get{{.Title}} retrieves a single {{.ModelName}} by {{.PrimaryKey.Name}}
func (s *{{.Title}}Service) Get{{.Title}}(ctx context.Context, {{.PrimaryKey.Name}} {{.PrimaryKey.GoType}}) (*db.{{.ModelName}}, error) {
	result, err := s.querier.Get{{.Title}}_ar_gen(ctx, {{.PrimaryKey.Name}})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("{{.TableName}} with %s %v not found", "{{.PrimaryKey.Name}}", {{.PrimaryKey.Name}})
		}
		s.logger.Error("Failed to get {{.TableName}}", "{{.PrimaryKey.Name}}", {{.PrimaryKey.Name}}, "error", err)
		return nil, fmt.Errorf("failed to get {{.TableName}}: %w", err)
	}

	return &result, nil
}

// List{{.Title}} retrieves multiple {{.TableName}} records with pagination
func (s *{{.Title}}Service) List{{.Title}}(ctx context.Context, limit, offset int32) ([]db.{{.ModelName}}, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	results, err := s.querier.List{{.Title}}_ar_gen(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list {{.TableName}}", "limit", limit, "offset", offset, "error", err)
		return nil, fmt.Errorf("failed to list {{.TableName}}: %w", err)
	}

	return results, nil
}

// Create{{.Title}} creates a new {{.TableName}} record
func (s *{{.Title}}Service) Create{{.Title}}(ctx context.Context, params Create{{.Title}}Params) (*db.{{.ModelName}}, error) {
	result, err := s.querier.Create{{.Title}}_ar_gen(ctx, params)
	if err != nil {
		s.logger.Error("Failed to create {{.TableName}}", "params", params, "error", err)
		return nil, fmt.Errorf("failed to create {{.TableName}}: %w", err)
	}
	
	// Get the last inserted ID and fetch the record
	id, err := result.LastInsertId()
	if err != nil {
		s.logger.Error("Failed to get last insert id", "error", err)
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}
	
	return s.Get{{.Title}}(ctx, id)
}

// Update{{.Title}} updates an existing {{.TableName}} record
func (s *{{.Title}}Service) Update{{.Title}}(ctx context.Context, params Update{{.Title}}Params) (*db.{{.ModelName}}, error) {
	_, err := s.querier.Update{{.Title}}_ar_gen(ctx, params)
	if err != nil {
		s.logger.Error("Failed to update {{.TableName}}", "params", params, "error", err)
		return nil, fmt.Errorf("failed to update {{.TableName}}: %w", err)
	}
	
	// Fetch the updated record
	return s.Get{{.Title}}(ctx, params.Id)
}

// Delete{{.Title}} deletes a {{.TableName}} record by {{.PrimaryKey.Name}}
func (s *{{.Title}}Service) Delete{{.Title}}(ctx context.Context, {{.PrimaryKey.Name}} {{.PrimaryKey.GoType}}) error {
	err := s.querier.Delete{{.Title}}_ar_gen(ctx, {{.PrimaryKey.Name}})
	if err != nil {
		s.logger.Error("Failed to delete {{.TableName}}", "{{.PrimaryKey.Name}}", {{.PrimaryKey.Name}}, "error", err)
		return fmt.Errorf("failed to delete {{.TableName}}: %w", err)
	}

	return nil
}

// {{.Title}}ServiceInterface defines the interface for {{.Title}}Service
type {{.Title}}ServiceInterface interface {
	Get{{.Title}}(ctx context.Context, {{.PrimaryKey.Name}} {{.PrimaryKey.GoType}}) (*db.{{.ModelName}}, error)
	List{{.Title}}(ctx context.Context, limit, offset int32) ([]db.{{.ModelName}}, error)
	Create{{.Title}}(ctx context.Context, params Create{{.Title}}Params) (*db.{{.ModelName}}, error)
	Update{{.Title}}(ctx context.Context, params Update{{.Title}}Params) (*db.{{.ModelName}}, error)
	Delete{{.Title}}(ctx context.Context, {{.PrimaryKey.Name}} {{.PrimaryKey.GoType}}) error
}

// Ensure {{.Title}}Service implements the interface
var _ {{.Title}}ServiceInterface = (*{{.Title}}Service)(nil)
`
