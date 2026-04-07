package generator

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/bata94/apiright/pkg/core"
)

// ProtoGenerator generates protobuf definitions from schema
type ProtoGenerator struct {
	templates *template.Template
	genSuffix string
	logger    core.Logger
}

// ProtoMessage represents a protobuf message for template generation
type ProtoMessage struct {
	Name      string
	Fields    []ProtoField
	GoName    string
	TableName string
}

// ProtoField represents a protobuf field
type ProtoField struct {
	Name     string
	Type     string
	Number   int
	GoName   string
	GoType   string
	Optional bool
	JSONName string
}

// ProtoService represents a protobuf service
type ProtoService struct {
	Name      string
	GoName    string
	TableName string
	Methods   []ProtoMethod
	Columns   []core.Column
}

// ProtoMethod represents a protobuf service method
type ProtoMethod struct {
	Name          string
	Request       string
	Response      string
	GoName        string
	HTTPMethod    string
	HTTPPath      string
	RequestFields []ProtoField
	ResponseType  string
}

// NewProtoGenerator creates a new protobuf generator
func NewProtoGenerator(genSuffix string, logger core.Logger) *ProtoGenerator {
	return &ProtoGenerator{
		genSuffix: genSuffix,
		logger:    logger,
	}
}

// Generate generates protobuf definitions for all tables in the schema
func (pg *ProtoGenerator) Generate(schema *core.Schema, ctx *core.GenerationContext) error {
	if err := pg.parseTemplates(); err != nil {
		return fmt.Errorf("failed to parse protobuf templates: %w", err)
	}

	// Generate messages file
	if err := pg.generateMessages(schema, ctx); err != nil {
		return fmt.Errorf("failed to generate messages: %w", err)
	}

	// Generate services file
	if err := pg.generateServices(schema, ctx); err != nil {
		return fmt.Errorf("failed to generate services: %w", err)
	}

	pg.logger.Info("Generated protobuf definitions", "tables", len(schema.Tables))
	return nil
}

// parseTemplates initializes protobuf generation templates
func (pg *ProtoGenerator) parseTemplates() error {
	templates := map[string]string{
		"messages": messagesTemplate,
		"services": servicesTemplate,
	}

	pg.templates = template.New("proto").Option("missingkey=error")
	for name, content := range templates {
		_, err := pg.templates.New(name).Parse(content)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}
	}

	return nil
}

// generateMessages generates protobuf messages for all tables
func (pg *ProtoGenerator) generateMessages(schema *core.Schema, ctx *core.GenerationContext) error {
	var messages []ProtoMessage

	for _, table := range schema.Tables {
		message := pg.createMessageFromTable(table)
		messages = append(messages, message)
	}

	// Execute template
	var buf strings.Builder
	data := map[string]any{
		"PackageName": "db",
		"Messages":    messages,
		"GoPackage":   "gen/go/db",
	}

	if err := pg.templates.ExecuteTemplate(&buf, "messages", data); err != nil {
		return fmt.Errorf("failed to execute messages template: %w", err)
	}

	// Write to gen/proto/db_ar_gen.proto
	outputPath := ctx.Join(ctx.ProjectDir, "gen", "proto", "db"+pg.genSuffix+".proto")
	if err := ctx.WriteFile(outputPath, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write messages file: %w", err)
	}

	pg.logger.Debug("Generated protobuf messages", "path", outputPath)
	return nil
}

// generateServices generates protobuf services for all tables
func (pg *ProtoGenerator) generateServices(schema *core.Schema, ctx *core.GenerationContext) error {
	var services []ProtoService

	for _, table := range schema.Tables {
		service := pg.createServiceFromTable(table)
		services = append(services, service)
	}

	// Execute template
	var buf strings.Builder
	data := map[string]any{
		"PackageName": "api",
		"Services":    services,
		"GoPackage":   "gen/go/api",
		"ImportPath":  "gen/proto/db_ar_gen.proto",
	}

	if err := pg.templates.ExecuteTemplate(&buf, "services", data); err != nil {
		return fmt.Errorf("failed to execute services template: %w", err)
	}

	// Write to gen/proto/api_ar_gen.proto
	outputPath := ctx.Join(ctx.ProjectDir, "gen", "proto", "api"+pg.genSuffix+".proto")
	if err := ctx.WriteFile(outputPath, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write services file: %w", err)
	}

	pg.logger.Debug("Generated protobuf services", "path", outputPath)
	return nil
}

// createMessageFromTable creates a protobuf message from a table
func (pg *ProtoGenerator) createMessageFromTable(table core.Table) ProtoMessage {
	message := ProtoMessage{
		Name:      pg.toProtoMessageName(table.Name),
		GoName:    pg.toGoStructName(table.Name),
		TableName: table.Name,
	}

	for i, col := range table.Columns {
		field := ProtoField{
			Name:     pg.toProtoFieldName(col.Name),
			Type:     core.SQLToProtoType(col.Type),
			Number:   i + 1,
			GoName:   pg.toGoFieldName(col.Name),
			GoType:   core.SQLToGoType(col.Type),
			Optional: col.Nullable,
		}
		message.Fields = append(message.Fields, field)
	}

	return message
}

// createServiceFromTable creates a protobuf service from a table
func (pg *ProtoGenerator) createServiceFromTable(table core.Table) ProtoService {
	tableName := table.Name
	titleName := pg.toTitleCase(tableName)

	service := ProtoService{
		Name:      titleName + "Service",
		GoName:    titleName + "Service",
		TableName: tableName,
		Columns:   table.Columns,
	}

	// CRUD methods
	methods := []ProtoMethod{
		{
			Name:          "Get" + titleName,
			Request:       "Get" + titleName + "Request",
			Response:      "Get" + titleName + "Response",
			GoName:        "Get" + titleName,
			HTTPMethod:    "GET",
			HTTPPath:      "/v1/" + pg.pluralize(tableName) + "/{id}",
			RequestFields: pg.generatePrimaryKeyField(table.Columns),
			ResponseType:  titleName,
		},
		{
			Name:          "List" + pg.pluralize(titleName),
			Request:       "List" + pg.pluralize(titleName) + "Request",
			Response:      "List" + pg.pluralize(titleName) + "Response",
			GoName:        "List" + pg.pluralize(titleName),
			HTTPMethod:    "GET",
			HTTPPath:      "/v1/" + pg.pluralize(tableName),
			RequestFields: pg.generatePaginationFields(),
			ResponseType:  "repeated " + titleName,
		},
		{
			Name:          "Create" + titleName,
			Request:       "Create" + titleName + "Request",
			Response:      "Create" + titleName + "Response",
			GoName:        "Create" + titleName,
			HTTPMethod:    "POST",
			HTTPPath:      "/v1/" + pg.pluralize(tableName),
			RequestFields: pg.generateProtoFields(table.Columns, false),
			ResponseType:  titleName,
		},
		{
			Name:          "Update" + titleName,
			Request:       "Update" + titleName + "Request",
			Response:      "Update" + titleName + "Response",
			GoName:        "Update" + titleName,
			HTTPMethod:    "PUT",
			HTTPPath:      "/v1/" + pg.pluralize(tableName) + "/{id}",
			RequestFields: pg.generateUpdateFields(table.Columns),
			ResponseType:  titleName,
		},
		{
			Name:          "Delete" + titleName,
			Request:       "Delete" + titleName + "Request",
			Response:      "Delete" + titleName + "Response",
			GoName:        "Delete" + titleName,
			HTTPMethod:    "DELETE",
			HTTPPath:      "/v1/" + pg.pluralize(tableName) + "/{id}",
			RequestFields: pg.generatePrimaryKeyField(table.Columns),
			ResponseType:  "bool", // Success indicator
		},
	}

	service.Methods = methods
	return service
}

// Helper methods for naming
func (pg *ProtoGenerator) toProtoMessageName(tableName string) string {
	return pg.toTitleCase(tableName)
}

func (pg *ProtoGenerator) toProtoFieldName(columnName string) string {
	return strings.ToLower(columnName)
}

func (pg *ProtoGenerator) toGoStructName(tableName string) string {
	return pg.toTitleCase(tableName)
}

func (pg *ProtoGenerator) toGoFieldName(columnName string) string {
	return pg.toTitleCase(columnName)
}

func (pg *ProtoGenerator) toTitleCase(s string) string {
	if s == "" {
		return ""
	}

	// Handle common plural forms
	if strings.HasSuffix(s, "ies") {
		return pg.toTitleCase(s[:len(s)-3] + "y")
	}
	if strings.HasSuffix(s, "es") {
		return pg.toTitleCase(s[:len(s)-2])
	}
	if strings.HasSuffix(s, "s") {
		return pg.toTitleCase(s[:len(s)-1])
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

func (pg *ProtoGenerator) pluralize(s string) string {
	if s == "" {
		return ""
	}

	// Simple pluralization rules
	if strings.HasSuffix(s, "y") {
		return s[:len(s)-1] + "ies"
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "sh") || strings.HasSuffix(s, "ch") {
		return s + "es"
	}
	return s + "s"
}

// generateProtoFields generates protobuf field definitions
func (pg *ProtoGenerator) generateProtoFields(columns []core.Column, includePK bool) []ProtoField {
	var fields []ProtoField
	for _, col := range columns {
		// Skip primary key if not requested
		if !includePK && pg.isPrimaryKeyField(col) {
			continue
		}

		field := ProtoField{
			Name:     pg.toCamelCase(col.Name),
			Type:     core.SQLToProtoType(col.Type),
			Number:   len(fields) + 1,
			JSONName: col.Name,
		}
		fields = append(fields, field)
	}
	return fields
}

// generatePrimaryKeyField returns the primary key field for request messages
func (pg *ProtoGenerator) generatePrimaryKeyField(columns []core.Column) []ProtoField {
	for _, col := range columns {
		if pg.isPrimaryKeyField(col) {
			return []ProtoField{
				{
					Name:     pg.toCamelCase(col.Name),
					Type:     core.SQLToProtoType(col.Type),
					Number:   1,
					JSONName: col.Name,
				},
			}
		}
	}
	return []ProtoField{}
}

// generateUpdateFields returns primary key + non-PK fields for update requests
func (pg *ProtoGenerator) generateUpdateFields(columns []core.Column) []ProtoField {
	var fields []ProtoField
	fieldNum := 1

	for _, col := range columns {
		field := ProtoField{
			Name:     pg.toCamelCase(col.Name),
			Type:     core.SQLToProtoType(col.Type),
			Number:   fieldNum,
			JSONName: col.Name,
		}
		fields = append(fields, field)
		fieldNum++
	}
	return fields
}

// generatePaginationFields returns pagination fields for list requests
func (pg *ProtoGenerator) generatePaginationFields() []ProtoField {
	return []ProtoField{
		{
			Name:     "limit",
			Type:     "int64",
			Number:   1,
			JSONName: "limit",
		},
		{
			Name:     "offset",
			Type:     "int64",
			Number:   2,
			JSONName: "offset",
		},
	}
}

// isPrimaryKeyField checks if column is a primary key
func (pg *ProtoGenerator) isPrimaryKeyField(col core.Column) bool {
	// Simple heuristic: primary key is usually first column and has specific characteristics
	return strings.Contains(strings.ToUpper(col.Type), "INTEGER") &&
		(strings.Contains(strings.ToUpper(col.Name), "ID") ||
			strings.Contains(strings.ToUpper(col.Name), "UUID") ||
			strings.Contains(strings.ToUpper(col.Name), "KEY"))
}

// toCamelCase converts snake_case to CamelCase
func (pg *ProtoGenerator) toCamelCase(s string) string {
	words := strings.Split(s, "_")
	for i, word := range words {
		if i > 0 && len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}
	return strings.Join(words, "")
}

// Protobuf generation templates
const (
	messagesTemplate = `// Generated protobuf messages for database entities
// This file is auto-generated by APIRight. Do not edit directly.
syntax = "proto3";

package db;
option go_package = "{{.GoPackage}}";

import "google/protobuf/timestamp.proto";

{{range .Messages}}
// {{.GoName}} represents the {{.TableName}} table
message {{.Name}} {
{{range .Fields}}  {{.Type}} {{.Name}} = {{.Number}};
{{end}}}
{{end}}`

	servicesTemplate = `// Generated protobuf services for API endpoints
// This file is auto-generated by APIRight. Do not edit directly.
syntax = "proto3";

package api;
option go_package = "{{.GoPackage}}";

import "google/protobuf/timestamp.proto";
import "{{.ImportPath}}";

{{range .Services}}
// {{.GoName}} provides CRUD operations for {{.TableName}}
service {{.Name}} {
{{range .Methods}}  rpc {{.Name}}({{.Request}}) returns ({{.Response}});
{{end}}

{{range .Methods}}
// Request message for {{.Name}}
message {{.Request}} {
{{range .RequestFields}}  {{.Type}} {{.Name}} = {{.Number}};
{{end}}}

// Response message for {{.Name}}
message {{.Response}} {
  {{.ResponseType}} data = 1;
}

{{end}}
{{end}}`
)
