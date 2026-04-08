package generator

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/bata94/apiright/pkg/core"
)

// AdapterGenerator generates adapter services that wrap table-specific services
// and expose them through the generic ServiceInterface
type AdapterGenerator struct {
	templates *template.Template
	genSuffix string
	logger    core.Logger
}

// AdapterData represents data for adapter template generation
type AdapterData struct {
	TableName   string
	Title       string // Title case for naming (e.g., "Post", "User")
	PackageName string
	ModelName   string // Singular model name (e.g., "Post", "User")
	ServiceName string // Full service name (e.g., "PostService")
	PrimaryKey  ColumnData
	ModulePath  string // Full module path from apiright.yaml
}

// NewAdapterGenerator creates a new adapter generator
func NewAdapterGenerator(genSuffix string, logger core.Logger) *AdapterGenerator {
	return &AdapterGenerator{
		genSuffix: genSuffix,
		logger:    logger,
	}
}

// GenerateAdapters generates adapter implementations for all tables in schema
func (ag *AdapterGenerator) GenerateAdapters(schema *core.Schema, ctx *core.GenerationContext) error {
	ag.logger.Info("Starting adapter generation", "tables", len(schema.Tables))

	if err := ag.parseTemplates(); err != nil {
		return fmt.Errorf("failed to parse adapter templates: %w", err)
	}

	for _, table := range schema.Tables {
		if err := ag.generateTableAdapter(table, ctx); err != nil {
			return fmt.Errorf("failed to generate adapter for table %s: %w", table.Name, err)
		}
	}

	// Generate the init.go file that registers all adapters
	if err := ag.generateInitFile(schema.Tables, ctx); err != nil {
		return fmt.Errorf("failed to generate init file: %w", err)
	}

	ag.logger.Info("Generated adapter implementations", "tables", len(schema.Tables))
	return nil
}

// parseTemplates initializes adapter generation templates
func (ag *AdapterGenerator) parseTemplates() error {
	templates := map[string]string{
		"adapter": adapterTemplate,
		"init":    initTemplate,
	}

	ag.templates = template.New("adapter").Option("missingkey=error")
	for name, content := range templates {
		_, err := ag.templates.New(name).Parse(content)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}
	}

	return nil
}

// generateTableAdapter generates adapter implementation for a single table
func (ag *AdapterGenerator) generateTableAdapter(table core.Table, ctx *core.GenerationContext) error {
	// Convert table data for template execution
	adapterData := ag.prepareAdapterData(table, ctx)

	// Generate adapter code
	adapterCode := ag.executeTemplate("adapter", adapterData)

	// Write to gen/go/adapters/{table}_adapter_ar_gen.go
	outputPath := ctx.Join(ctx.ProjectDir, "gen", "go", "adapters", table.Name+"_adapter"+ag.genSuffix+".go")
	if err := ctx.WriteFile(outputPath, []byte(adapterCode), 0644); err != nil {
		return fmt.Errorf("failed to write adapter file: %w", err)
	}

	ag.logger.Debug("Generated adapter for table", "table", table.Name, "path", outputPath)
	return nil
}

// generateInitFile generates the init.go file that registers all adapters
func (ag *AdapterGenerator) generateInitFile(tables []core.Table, ctx *core.GenerationContext) error {
	// Build table registration data
	var tableRegs []TableRegistration
	for _, table := range tables {
		title := ag.toTitleCase(singularize(table.Name))
		tableRegs = append(tableRegs, TableRegistration{
			TableName:   table.Name,
			ServiceName: title + "Service",
		})
	}

	initData := InitData{
		PackageName: "adapters",
		ModulePath:  ctx.ModulePath,
		Tables:      tableRegs,
	}

	// Generate init code
	initCode := ag.executeTemplate("init", initData)

	// Write to gen/go/adapters/init_ar_gen.go
	outputPath := ctx.Join(ctx.ProjectDir, "gen", "go", "adapters", "init"+ag.genSuffix+".go")
	if err := ctx.WriteFile(outputPath, []byte(initCode), 0644); err != nil {
		return fmt.Errorf("failed to write init file: %w", err)
	}

	ag.logger.Debug("Generated init file", "path", outputPath)
	return nil
}

// TableRegistration represents a table to be registered in Init()
type TableRegistration struct {
	TableName   string
	ServiceName string // e.g., "PostService", "UserService"
}

// InitData represents data for init template generation
type InitData struct {
	PackageName string
	ModulePath  string
	Tables      []TableRegistration
}

// prepareAdapterData converts core.Table to AdapterData for template execution
func (ag *AdapterGenerator) prepareAdapterData(table core.Table, ctx *core.GenerationContext) AdapterData {
	// Find primary key from table.PrimaryKey (which is []string)
	var primaryKey ColumnData
	if len(table.PrimaryKey) > 0 {
		// Get the first primary key column
		pkName := table.PrimaryKey[0]
		// Find the column type
		for _, col := range table.Columns {
			if col.Name == pkName {
				primaryKey = ColumnData{
					Name:   col.Name,
					GoType: core.SQLToGoType(col.Type),
				}
				break
			}
		}
	}

	// Convert table name to singular title case
	singularTable := singularize(table.Name)
	titleName := ag.toTitleCase(singularTable)

	return AdapterData{
		TableName:   table.Name,
		Title:       titleName,
		PackageName: "adapters",
		ModelName:   titleName,
		ServiceName: titleName + "Service",
		PrimaryKey:  primaryKey,
		ModulePath:  ctx.ModulePath,
	}
}

// executeTemplate executes a template with adapter data
func (ag *AdapterGenerator) executeTemplate(templateName string, data interface{}) string {
	var buf strings.Builder
	err := ag.templates.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		ag.logger.Error("Template execution failed", "template", templateName, "error", err)
		return ""
	}
	return buf.String()
}

// toTitleCase converts a string to TitleCase for naming
func (ag *AdapterGenerator) toTitleCase(s string) string {
	if s == "" {
		return ""
	}

	// Handle common plural forms
	if strings.HasSuffix(s, "ies") {
		return ag.toTitleCase(s[:len(s)-3] + "y")
	}
	if strings.HasSuffix(s, "es") {
		return ag.toTitleCase(s[:len(s)-2])
	}
	if strings.HasSuffix(s, "s") {
		return ag.toTitleCase(s[:len(s)-1])
	}

	// Simple snake_case to TitleCase conversion
	words := strings.Split(strings.ReplaceAll(s, "_", " "), " ")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// Adapter generation template
const adapterTemplate = `// Code generated by APIRight. DO NOT EDIT.
// Generated adapter for table {{.TableName}}

package {{.PackageName}}

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/bata94/apiright/pkg/core"
	db "{{.ModulePath}}/gen/go"
)

// {{.ServiceName}}Adapter provides CRUD operations and implements ServiceInterface
type {{.ServiceName}}Adapter struct {
	querier db.Querier
	logger  core.Logger
}

// New{{.ServiceName}}Adapter creates a new {{.ServiceName}}Adapter
func New{{.ServiceName}}Adapter(querier db.Querier, logger core.Logger) *{{.ServiceName}}Adapter {
	return &{{.ServiceName}}Adapter{
		querier: querier,
		logger:  logger,
	}
}

// Get retrieves a single {{.ModelName}} by id (supports int64 and string IDs)
func (a *{{.ServiceName}}Adapter) Get(ctx context.Context, id any) (any, error) {
	var {{.PrimaryKey.Name}}Val int64

	switch v := id.(type) {
	case int64:
		{{.PrimaryKey.Name}}Val = v
	case string:
		// Try to parse as int64 first
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("string id not supported for {{.TableName}}: %s", v)
		}
		{{.PrimaryKey.Name}}Val = parsed
	case int:
		{{.PrimaryKey.Name}}Val = int64(v)
	case int32:
		{{.PrimaryKey.Name}}Val = int64(v)
	default:
		return nil, fmt.Errorf("unsupported id type for {{.TableName}}: %T", id)
	}

	result, err := a.querier.Get{{.Title}}_ar_gen(ctx, {{.PrimaryKey.Name}}Val)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("{{.TableName}} with id %v not found", id)
		}
		return nil, err
	}

	return result, nil
}

// List retrieves multiple {{.TableName}} records with pagination
func (a *{{.ServiceName}}Adapter) List(ctx context.Context, limit, offset int32) (any, error) {
	params := db.List{{.Title}}_ar_genParams{
		Limit:  int64(limit),
		Offset: int64(offset),
	}
	return a.querier.List{{.Title}}_ar_gen(ctx, params)
}

// Create creates a new {{.TableName}} record
func (a *{{.ServiceName}}Adapter) Create(ctx context.Context, params any) (any, error) {
	// Convert params map to JSON and unmarshal into sqlc params struct
	paramsMap, ok := params.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid params type for {{.TableName}}: expected map[string]any, got %T", params)
	}

	data, err := json.Marshal(paramsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params for {{.TableName}}: %w", err)
	}

	var createParams db.Create{{.Title}}_ar_genParams
	if err := json.Unmarshal(data, &createParams); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params for {{.TableName}}: %w", err)
	}

	// Execute insert
	if err := a.querier.Create{{.Title}}_ar_gen(ctx, createParams); err != nil {
		return nil, fmt.Errorf("failed to create {{.TableName}}: %w", err)
	}

	// Get the created record - need to fetch by querying for the last inserted
	// For simplicity, we'll return the params as a map
	return paramsMap, nil
}

// Update updates an existing {{.TableName}} record
func (a *{{.ServiceName}}Adapter) Update(ctx context.Context, params any) (any, error) {
	// Convert params map to JSON and unmarshal into sqlc params struct
	paramsMap, ok := params.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid params type for {{.TableName}}: expected map[string]any, got %T", params)
	}

	data, err := json.Marshal(paramsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params for {{.TableName}}: %w", err)
	}

	var updateParams db.Update{{.Title}}_ar_genParams
	if err := json.Unmarshal(data, &updateParams); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params for {{.TableName}}: %w", err)
	}

	// Execute update
	if err := a.querier.Update{{.Title}}_ar_gen(ctx, updateParams); err != nil {
		return nil, fmt.Errorf("failed to update {{.TableName}}: %w", err)
	}

	// Fetch the updated record
	if id, ok := paramsMap["id"]; ok {
		return a.Get(ctx, id)
	}
	
	return paramsMap, nil
}

// Delete deletes a {{.TableName}} record by id
func (a *{{.ServiceName}}Adapter) Delete(ctx context.Context, id any) error {
	var {{.PrimaryKey.Name}}Val int64

	switch v := id.(type) {
	case int64:
		{{.PrimaryKey.Name}}Val = v
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf("string id not supported for {{.TableName}}: %s", v)
		}
		{{.PrimaryKey.Name}}Val = parsed
	case int:
		{{.PrimaryKey.Name}}Val = int64(v)
	case int32:
		{{.PrimaryKey.Name}}Val = int64(v)
	default:
		return fmt.Errorf("unsupported id type for {{.TableName}}: %T", id)
	}

	return a.querier.Delete{{.Title}}_ar_gen(ctx, {{.PrimaryKey.Name}}Val)
}

// TableName returns the table name for this adapter
func (a *{{.ServiceName}}Adapter) TableName() string {
	return "{{.TableName}}"
}

// Ensure {{.ServiceName}}Adapter implements ServiceInterface
var _ ServiceInterface = (*{{.ServiceName}}Adapter)(nil)

// Ensure {{.ServiceName}}Adapter implements TableNamer interface
var _ TableNamer = (*{{.ServiceName}}Adapter)(nil)
`

// Init template for registering all adapters
const initTemplate = `// Code generated by APIRight. DO NOT EDIT.
// Generated adapter initialization

package {{.PackageName}}

import (
	"fmt"

	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/database"
	"github.com/bata94/apiright/pkg/server"
	db "{{.ModulePath}}/gen/go"
)

// Init registers all service adapters with the server
func Init(srv *server.DualServer, dbConn *database.Database, logger core.Logger) error {
	// Create querier from database connection
	querier := db.New(dbConn.GetDB())

	// Register all service adapters
{{- range .Tables}}
	if err := srv.RegisterService(New{{.ServiceName}}Adapter(querier, logger)); err != nil {
		return fmt.Errorf("failed to register {{.TableName}} service: %w", err)
	}
{{- end }}

	return nil
}
`
