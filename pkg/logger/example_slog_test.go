package logger_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/bata94/apiright/pkg/logger"
)

func TestSlogWrapper(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a slog logger with JSON handler
	slogLogger := logger.NewSlogLogger(logger.InfoLevel, &buf)

	// Wrap it with our SlogWrapper
	wrapper := logger.NewSlogWrapper(slogLogger)

	// Test structured logging
	wrapper.Info("user login", "user_id", 123, "ip", "192.168.1.1")
	wrapper.Error("database error", "error", "connection timeout", "table", "users")

	// Check the output contains structured JSON
	output := buf.String()
	if !strings.Contains(output, `"level":"INFO"`) {
		t.Errorf("Expected INFO level in output, got: %s", output)
	}
	if !strings.Contains(output, `"msg":"user login"`) {
		t.Errorf("Expected message in output, got: %s", output)
	}
	if !strings.Contains(output, `"user_id":123`) {
		t.Errorf("Expected user_id field in output, got: %s", output)
	}
	if !strings.Contains(output, `"level":"ERROR"`) {
		t.Errorf("Expected ERROR level in output, got: %s", output)
	}
	if !strings.Contains(output, `"table":"users"`) {
		t.Errorf("Expected table field in output, got: %s", output)
	}
}

func TestNewStructuredLogger(t *testing.T) {
	var buf bytes.Buffer

	// Create a structured logger
	structuredLogger := logger.NewStructuredLogger(logger.DebugLevel, &buf)

	// Test that it implements the Logger interface
	structuredLogger.Debug("debug message", "key", "value")
	structuredLogger.Info("info message", "count", 42)
	structuredLogger.Warn("warning message", "severity", "high")
	structuredLogger.Error("error message", "code", 500)

	output := buf.String()

	// Should contain JSON structured logs
	if !strings.Contains(output, `"level":"DEBUG"`) {
		t.Errorf("Expected DEBUG level, got: %s", output)
	}
	if !strings.Contains(output, `"level":"INFO"`) {
		t.Errorf("Expected INFO level, got: %s", output)
	}
	if !strings.Contains(output, `"level":"WARN"`) {
		t.Errorf("Expected WARN level, got: %s", output)
	}
	if !strings.Contains(output, `"level":"ERROR"`) {
		t.Errorf("Expected ERROR level, got: %s", output)
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Errorf("Expected structured key-value, got: %s", output)
	}
	if !strings.Contains(output, `"count":42`) {
		t.Errorf("Expected structured count field, got: %s", output)
	}
}

// BenchmarkDefaultLogger benchmarks the default logger performance
func BenchmarkDefaultLogger(b *testing.B) {
	log := logger.NewDefaultLogger()
	log.SetLevel(logger.InfoLevel)
	log.SetOutput(io.Discard) // Discard output to avoid benchmark interference

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info("benchmark test message", "iteration", i, "type", "default")
	}
}

// BenchmarkStructuredLogger benchmarks the structured logger performance
func BenchmarkStructuredLogger(b *testing.B) {
	log := logger.NewStructuredLogger(logger.InfoLevel, io.Discard)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info("benchmark test message", "iteration", i, "type", "structured")
	}
}

// BenchmarkDefaultLoggerSimple benchmarks simple logging without structured args
func BenchmarkDefaultLoggerSimple(b *testing.B) {
	log := logger.NewDefaultLogger()
	log.SetLevel(logger.InfoLevel)
	log.SetOutput(io.Discard) // Discard output to avoid benchmark interference

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info("benchmark test message")
	}
}

// BenchmarkStructuredLoggerSimple benchmarks simple structured logging
func BenchmarkStructuredLoggerSimple(b *testing.B) {
	log := logger.NewStructuredLogger(logger.InfoLevel, io.Discard)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info("benchmark test message")
	}
}
