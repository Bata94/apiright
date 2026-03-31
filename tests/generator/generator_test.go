package generator_test

import (
	"testing"

	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/generator"
)

type mockLogger struct{}

func (l *mockLogger) Debug(msg string, fields ...interface{})  {}
func (l *mockLogger) Info(msg string, fields ...interface{})   {}
func (l *mockLogger) Warn(msg string, fields ...interface{})   {}
func (l *mockLogger) Error(msg string, fields ...interface{})  {}
func (l *mockLogger) DPanic(msg string, fields ...interface{}) {}
func (l *mockLogger) Panic(msg string, fields ...interface{})  {}
func (l *mockLogger) Fatal(msg string, fields ...interface{})  {}
func (l *mockLogger) With(fields ...interface{}) core.Logger   { return l }
func (l *mockLogger) Sync() error                              { return nil }

func TestSQLGenerator_NewSQLGenerator(t *testing.T) {
	logger := &mockLogger{}
	gen := generator.NewSQLGenerator("_ar_gen", generator.DialectSQLite, logger)

	if gen == nil {
		t.Fatal("Expected non-nil SQLGenerator")
	}
}

func TestSQLGenerator_Dialects(t *testing.T) {
	tests := []struct {
		dialect generator.Dialect
		name    string
	}{
		{generator.DialectSQLite, "sqlite"},
		{generator.DialectPostgres, "postgres"},
		{generator.DialectMySQL, "mysql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &mockLogger{}
			gen := generator.NewSQLGenerator("_ar_gen", tt.dialect, logger)
			if gen == nil {
				t.Errorf("NewSQLGenerator returned nil for dialect %s", tt.dialect)
			}
		})
	}
}

func TestSQLGenerator_DefaultDialect(t *testing.T) {
	logger := &mockLogger{}
	gen := generator.NewSQLGenerator("_ar_gen", "", logger)
	if gen == nil {
		t.Fatal("Expected non-nil SQLGenerator")
	}
}

func TestProtoGenerator_NewProtoGenerator(t *testing.T) {
	logger := &mockLogger{}
	gen := generator.NewProtoGenerator("_ar_gen", logger)

	if gen == nil {
		t.Fatal("Expected non-nil ProtoGenerator")
	}
}

func TestServiceGenerator_NewServiceGenerator(t *testing.T) {
	logger := &mockLogger{}
	gen := generator.NewServiceGenerator("_ar_gen", logger)

	if gen == nil {
		t.Fatal("Expected non-nil ServiceGenerator")
	}
}

func TestOpenAPIGenerator_NewOpenAPIGenerator(t *testing.T) {
	logger := &mockLogger{}
	gen := generator.NewOpenAPIGenerator("_ar_gen", logger)

	if gen == nil {
		t.Fatal("Expected non-nil OpenAPIGenerator")
	}
}

func TestSchemaParser(t *testing.T) {
	logger := &mockLogger{}
	parser := generator.NewSchemaParser("sqlite", logger)

	if parser == nil {
		t.Fatal("Expected non-nil SchemaParser")
	}
}

func TestSQLToGoType(t *testing.T) {
	tests := []struct {
		sqlType  string
		expected string
	}{
		{"INTEGER", "int64"},
		{"INT", "int64"},
		{"BIGINT", "int64"},
		{"SMALLINT", "int32"},
		{"TINYINT", "int32"},
		{"REAL", "float64"},
		{"FLOAT", "float64"},
		{"DOUBLE", "float64"},
		{"TEXT", "string"},
		{"VARCHAR(255)", "string"},
		{"BLOB", "[]byte"},
		{"BOOLEAN", "bool"},
		{"TIMESTAMP", "time.Time"},
		{"DATETIME", "time.Time"},
		{"DATE", "time.Time"},
		{"UUID", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.sqlType, func(t *testing.T) {
			result := core.SQLToGoType(tt.sqlType)
			if result != tt.expected {
				t.Errorf("SQLToGoType(%s) = %s, want %s", tt.sqlType, result, tt.expected)
			}
		})
	}
}

func TestSQLToProtoType(t *testing.T) {
	tests := []struct {
		sqlType  string
		expected string
	}{
		{"INTEGER", "int64"},
		{"BIGINT", "int64"},
		{"REAL", "double"},
		{"FLOAT", "float"},
		{"TEXT", "string"},
		{"VARCHAR(255)", "string"},
		{"BLOB", "bytes"},
		{"BOOLEAN", "bool"},
		{"TIMESTAMP", "google.protobuf.Timestamp"},
		{"DATETIME", "google.protobuf.Timestamp"},
	}

	for _, tt := range tests {
		t.Run(tt.sqlType, func(t *testing.T) {
			result := core.SQLToProtoType(tt.sqlType)
			if result != tt.expected {
				t.Errorf("SQLToProtoType(%s) = %s, want %s", tt.sqlType, result, tt.expected)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user_id", "UserId"},
		{"user", "User"},
		{"users", "Users"},
		{"user_posts", "UserPosts"},
		{"", ""},
		{"i", "I"},
		{"my_id", "MyId"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := core.ToPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToPascalCase(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetExampleValue(t *testing.T) {
	tests := []struct {
		sqlType string
	}{
		{"INTEGER"},
		{"TEXT"},
		{"BOOLEAN"},
		{"DATE"},
		{"TIMESTAMP"},
		{"FLOAT"},
	}

	for _, tt := range tests {
		t.Run(tt.sqlType, func(t *testing.T) {
			result := core.GetExampleValue(tt.sqlType)
			if result == "" {
				t.Errorf("GetExampleValue(%s) returned empty string", tt.sqlType)
			}
		})
	}
}
