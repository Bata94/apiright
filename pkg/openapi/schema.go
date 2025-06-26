package openapi

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// SchemaGenerator handles the generation of OpenAPI schemas from Go types
type SchemaGenerator struct {
	schemas map[string]Schema
	visited map[reflect.Type]string // Prevent infinite recursion
}

// NewSchemaGenerator creates a new schema generator
func NewSchemaGenerator() *SchemaGenerator {
	return &SchemaGenerator{
		schemas: make(map[string]Schema),
		visited: make(map[reflect.Type]string),
	}
}

// GenerateSchema generates an OpenAPI schema from a Go type
func (sg *SchemaGenerator) GenerateSchema(t reflect.Type) Schema {
	return sg.generateSchemaInternal(t, "")
}

// GenerateSchemaWithName generates an OpenAPI schema with a specific name
func (sg *SchemaGenerator) GenerateSchemaWithName(t reflect.Type, name string) Schema {
	return sg.generateSchemaInternal(t, name)
}

// GetSchemas returns all generated schemas
func (sg *SchemaGenerator) GetSchemas() map[string]Schema {
	return sg.schemas
}

func (sg *SchemaGenerator) generateSchemaInternal(t reflect.Type, name string) Schema {
	// Handle pointers
	if t.Kind() == reflect.Ptr {
		return sg.generateSchemaInternal(t.Elem(), name)
	}

	// Check if we've already processed this type
	if schemaName, exists := sg.visited[t]; exists {
		return Schema{Ref: "#/components/schemas/" + schemaName}
	}

	switch t.Kind() {
	case reflect.String:
		return sg.generateStringSchema(t)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return sg.generateIntegerSchema(t)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return sg.generateIntegerSchema(t)
	case reflect.Float32, reflect.Float64:
		return sg.generateNumberSchema(t)
	case reflect.Bool:
		return Schema{Type: "boolean"}
	case reflect.Array, reflect.Slice:
		return sg.generateArraySchema(t, name)
	case reflect.Map:
		return sg.generateMapSchema(t, name)
	case reflect.Struct:
		return sg.generateStructSchema(t, name)
	case reflect.Interface:
		return Schema{} // Empty schema for any
	default:
		return Schema{Type: "object"}
	}
}

func (sg *SchemaGenerator) generateStringSchema(t reflect.Type) Schema {
	schema := Schema{Type: "string"}

	// Handle special string types
	switch t {
	case reflect.TypeOf(time.Time{}):
		schema.Format = "date-time"
	case reflect.TypeOf(time.Duration(0)):
		schema.Format = "duration"
	}

	return schema
}

func (sg *SchemaGenerator) generateIntegerSchema(t reflect.Type) Schema {
	schema := Schema{Type: "integer"}

	switch t.Kind() {
	case reflect.Int32, reflect.Uint32:
		schema.Format = "int32"
	case reflect.Int64, reflect.Uint64:
		schema.Format = "int64"
	}

	return schema
}

func (sg *SchemaGenerator) generateNumberSchema(t reflect.Type) Schema {
	schema := Schema{Type: "number"}

	switch t.Kind() {
	case reflect.Float32:
		schema.Format = "float"
	case reflect.Float64:
		schema.Format = "double"
	}

	return schema
}

func (sg *SchemaGenerator) generateArraySchema(t reflect.Type, name string) Schema {
	itemType := t.Elem()
	itemSchema := sg.generateSchemaInternal(itemType, "")

	return Schema{
		Type:  "array",
		Items: &itemSchema,
	}
}

func (sg *SchemaGenerator) generateMapSchema(t reflect.Type, name string) Schema {
	valueType := t.Elem()
	valueSchema := sg.generateSchemaInternal(valueType, "")

	return Schema{
		Type:                 "object",
		AdditionalProperties: valueSchema,
	}
}

func (sg *SchemaGenerator) generateStructSchema(t reflect.Type, name string) Schema {
	// Generate schema name if not provided
	if name == "" {
		name = t.Name()
		if name == "" {
			name = "AnonymousStruct"
		}
	}

	// Check if we've already processed this type
	if schemaName, exists := sg.visited[t]; exists {
		return Schema{Ref: "#/components/schemas/" + schemaName}
	}

	// Mark as visited to prevent infinite recursion
	sg.visited[t] = name

	schema := Schema{
		Type:       "object",
		Properties: make(map[string]Schema),
		Required:   []string{},
	}

	// Process struct fields
	for i := range t.NumField() {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON tag information
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue // Skip fields with json:"-"
		}

		fieldName, fieldOptions := sg.parseJSONTag(jsonTag, field.Name)
		if fieldName == "" {
			continue
		}

		// Generate schema for field type
		fieldSchema := sg.generateSchemaInternal(field.Type, "")

		// Add validation from struct tags
		sg.addValidationFromTags(field, &fieldSchema)

		// Add description from comment or tag
		if desc := field.Tag.Get("description"); desc != "" {
			fieldSchema.Description = desc
		}

		schema.Properties[fieldName] = fieldSchema

		// Check if field is required
		if !fieldOptions.omitempty && !sg.isOptionalType(field.Type) {
			schema.Required = append(schema.Required, fieldName)
		}
	}

	// Store the schema
	sg.schemas[name] = schema

	// Return reference to the schema
	return Schema{Ref: "#/components/schemas/" + name}
}

type jsonTagOptions struct {
	omitempty bool
}

func (sg *SchemaGenerator) parseJSONTag(tag, defaultName string) (string, jsonTagOptions) {
	if tag == "" {
		return defaultName, jsonTagOptions{}
	}

	parts := strings.Split(tag, ",")
	name := parts[0]
	if name == "" {
		name = defaultName
	}

	options := jsonTagOptions{}
	for _, part := range parts[1:] {
		switch part {
		case "omitempty":
			options.omitempty = true
		}
	}

	return name, options
}

func (sg *SchemaGenerator) isOptionalType(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.Map
}

func (sg *SchemaGenerator) addValidationFromTags(field reflect.StructField, schema *Schema) {
	// Handle validation tags
	if validate := field.Tag.Get("validate"); validate != "" {
		sg.parseValidationTag(validate, schema)
	}

	// Handle binding tags (common in web frameworks)
	if binding := field.Tag.Get("binding"); binding != "" {
		sg.parseBindingTag(binding, schema)
	}

	// Handle min/max tags
	if min := field.Tag.Get("min"); min != "" {
		if val, err := strconv.ParseFloat(min, 64); err == nil {
			schema.Minimum = &val
		}
	}

	if max := field.Tag.Get("max"); max != "" {
		if val, err := strconv.ParseFloat(max, 64); err == nil {
			schema.Maximum = &val
		}
	}

	// Handle format tag
	if format := field.Tag.Get("format"); format != "" {
		schema.Format = format
	}

	// Handle pattern tag
	if pattern := field.Tag.Get("pattern"); pattern != "" {
		schema.Pattern = pattern
	}

	// Handle example tag
	if example := field.Tag.Get("example"); example != "" {
		schema.Example = example
	}
}

func (sg *SchemaGenerator) parseValidationTag(validate string, schema *Schema) {
	rules := strings.SplitSeq(validate, ",")
	for rule := range rules {
		rule = strings.TrimSpace(rule)

		if rule == "required" {
			// This is handled at the struct level
			continue
		}

		if strings.HasPrefix(rule, "min=") {
			if val, err := strconv.ParseFloat(rule[4:], 64); err == nil {
				schema.Minimum = &val
			}
		}

		if strings.HasPrefix(rule, "max=") {
			if val, err := strconv.ParseFloat(rule[4:], 64); err == nil {
				schema.Maximum = &val
			}
		}

		if strings.HasPrefix(rule, "len=") {
			if val, err := strconv.Atoi(rule[4:]); err == nil {
				schema.MinLength = &val
				schema.MaxLength = &val
			}
		}

		if strings.HasPrefix(rule, "email") {
			schema.Format = "email"
		}

		if strings.HasPrefix(rule, "url") {
			schema.Format = "uri"
		}
	}
}

func (sg *SchemaGenerator) parseBindingTag(binding string, schema *Schema) {
	rules := strings.SplitSeq(binding, ",")
	for rule := range rules {
		rule = strings.TrimSpace(rule)

		if rule == "required" {
			// This is handled at the struct level
			continue
		}
	}
}

// GenerateSchemaFromValue generates a schema from a value (useful for examples)
func (sg *SchemaGenerator) GenerateSchemaFromValue(v any) Schema {
	if v == nil {
		return Schema{}
	}

	t := reflect.TypeOf(v)
	return sg.GenerateSchema(t)
}

// GenerateExample generates an example value for a schema
func GenerateExample(schema Schema) any {
	switch schema.Type {
	case "string":
		if schema.Example != nil {
			return schema.Example
		}
		switch schema.Format {
		case "email":
			return "user@example.com"
		case "date-time":
			return "2023-01-01T00:00:00Z"
		case "uri":
			return "https://example.com"
		default:
			return "string"
		}
	case "integer":
		if schema.Example != nil {
			return schema.Example
		}
		return 0
	case "number":
		if schema.Example != nil {
			return schema.Example
		}
		return 0.0
	case "boolean":
		if schema.Example != nil {
			return schema.Example
		}
		return true
	case "array":
		return []any{}
	case "object":
		if len(schema.Properties) > 0 {
			example := make(map[string]any)
			for name, prop := range schema.Properties {
				example[name] = GenerateExample(prop)
			}
			return example
		}
		return map[string]any{}
	default:
		return nil
	}
}

// ValidateSchema performs basic validation on a schema
func ValidateSchema(schema Schema) error {
	if schema.Type == "" && schema.Ref == "" && len(schema.AllOf) == 0 && len(schema.OneOf) == 0 && len(schema.AnyOf) == 0 {
		return fmt.Errorf("schema must have a type or reference")
	}

	if schema.Type == "array" && schema.Items == nil {
		return fmt.Errorf("array schema must have items definition")
	}

	return nil
}
