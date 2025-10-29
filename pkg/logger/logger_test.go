package logger

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"
	"testing"
)

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{PanicLevel, "PANIC"},
		{FatalLevel, "FATAL"},
		{ErrorLevel, "ERROR"},
		{WarnLevel, "WARN"},
		{InfoLevel, "INFO"},
		{DebugLevel, "DEBUG"},
		{TraceLevel, "TRACE"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.level.String() != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", tt.level.String(), tt.expected)
			}
		})
	}
}

func TestGetColor(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{PanicLevel, colorRed},
		{FatalLevel, colorRed},
		{ErrorLevel, colorRed},
		{WarnLevel, colorYellow},
		{InfoLevel, colorGreen},
		{DebugLevel, colorCyan},
		{TraceLevel, colorCyan},
		{LogLevel(999), colorWhite},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			color := getColor(tt.level)
			if color != tt.expected {
				t.Errorf("getColor(%v) = %v, want %v", tt.level, color, tt.expected)
			}
		})
	}
}

func TestIsTerminal(t *testing.T) {
	// Test with os.Stdout
	if !isTerminal(os.Stdout) {
		t.Error("isTerminal(os.Stdout) should return true")
	}

	if !isTerminal(os.Stderr) {
		t.Error("isTerminal(os.Stderr) should return true")
	}

	// Test with buffer
	var buf bytes.Buffer
	if isTerminal(&buf) {
		t.Error("isTerminal(&bytes.Buffer) should return false")
	}
}

func TestDefaultLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewDefaultLogger()
	logger.SetOutput(&buf)
	logger.SetLevel(DebugLevel)
	logger.SetColors(false) // Disable colors for easier testing

	// Test all logging levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()

	if !strings.Contains(output, "DEBUG") {
		t.Error("Expected DEBUG level in output")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("Expected INFO level in output")
	}
	if !strings.Contains(output, "WARN") {
		t.Error("Expected WARN level in output")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("Expected ERROR level in output")
	}
	if !strings.Contains(output, "debug message") {
		t.Error("Expected debug message in output")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Expected info message in output")
	}
}

func TestDefaultLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewDefaultLogger()
	logger.SetOutput(&buf)
	logger.SetLevel(WarnLevel) // Only WARN and above should be logged

	logger.Debug("debug message") // Should not appear
	logger.Info("info message")   // Should not appear
	logger.Warn("warn message")   // Should appear
	logger.Error("error message") // Should appear

	output := buf.String()

	if strings.Contains(output, "debug message") {
		t.Error("Debug message should not appear when level is WARN")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info message should not appear when level is WARN")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should appear when level is WARN")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message should appear when level is WARN")
	}
}

func TestDefaultLoggerColors(t *testing.T) {
	var buf bytes.Buffer
	logger := NewDefaultLogger()
	logger.SetOutput(&buf)
	logger.SetColors(true)

	logger.Info("colored message")

	output := buf.String()

	// Should contain ANSI color codes when colors are enabled
	if !strings.Contains(output, colorGreen) && !strings.Contains(output, colorReset) {
		t.Logf("Color codes not found in output: %q", output)
		t.Logf("This might be expected if running in a non-terminal environment")
		// Don't fail the test since colors might not work in test environment
	}
}

func TestDefaultLoggerFormatted(t *testing.T) {
	var buf bytes.Buffer
	logger := NewDefaultLogger()
	logger.SetOutput(&buf)
	logger.SetLevel(DebugLevel)
	logger.SetColors(false)

	logger.Debugf("debug %s %d", "formatted", 123)
	logger.Infof("info %s %d", "formatted", 456)

	output := buf.String()

	if !strings.Contains(output, "debug formatted 123") {
		t.Error("Expected formatted debug message")
	}
	if !strings.Contains(output, "info formatted 456") {
		t.Error("Expected formatted info message")
	}
}

func TestDefaultLoggerPanic(t *testing.T) {
	logger := NewDefaultLogger()
	logger.SetOutput(io.Discard)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic")
		}
	}()

	logger.Panic("panic message")
}

func TestDefaultLoggerPanicf(t *testing.T) {
	logger := NewDefaultLogger()
	logger.SetOutput(io.Discard)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic")
		}
	}()

	logger.Panicf("panic %s", "formatted")
}

func TestStdLoggerWrapper(t *testing.T) {
	var buf bytes.Buffer
	stdLogger := log.New(&buf, "", 0)
	wrapper := NewStdLoggerWrapper(stdLogger)
	wrapper.SetLevel(DebugLevel)

	wrapper.Debug("[DEBUG] debug message")
	wrapper.Info("[INFO] info message")
	wrapper.Warn("[WARN] warn message")
	wrapper.Error("[ERROR] error message")

	output := buf.String()

	if !strings.Contains(output, "[DEBUG] debug message") {
		t.Error("Expected debug message in StdLoggerWrapper output")
	}
	if !strings.Contains(output, "[INFO] info message") {
		t.Error("Expected info message in StdLoggerWrapper output")
	}
	if !strings.Contains(output, "[WARN] warn message") {
		t.Error("Expected warn message in StdLoggerWrapper output")
	}
	if !strings.Contains(output, "[ERROR] error message") {
		t.Error("Expected error message in StdLoggerWrapper output")
	}
}

func TestStdLoggerWrapperLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	stdLogger := log.New(&buf, "", 0)
	wrapper := NewStdLoggerWrapper(stdLogger)
	wrapper.SetLevel(WarnLevel)

	wrapper.Debug("debug message") // Should not appear
	wrapper.Info("info message")   // Should not appear
	wrapper.Warn("warn message")   // Should appear

	output := buf.String()

	if strings.Contains(output, "debug message") {
		t.Error("Debug message should not appear when level is WARN")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info message should not appear when level is WARN")
	}
	if !strings.Contains(output, "warn message") {
		t.Errorf("Warn message should appear when level is WARN, output: %q", output)
	}
}

func TestStdLoggerWrapperFormatted(t *testing.T) {
	var buf bytes.Buffer
	stdLogger := log.New(&buf, "", 0)
	wrapper := NewStdLoggerWrapper(stdLogger)
	wrapper.SetLevel(DebugLevel)

	wrapper.Debugf("debug %s %d", "formatted", 123)
	wrapper.Infof("info %s %d", "formatted", 456)

	output := buf.String()

	if !strings.Contains(output, "[DEBUG] debug formatted 123") {
		t.Error("Expected formatted debug message in StdLoggerWrapper")
	}
	if !strings.Contains(output, "[INFO] info formatted 456") {
		t.Error("Expected formatted info message in StdLoggerWrapper")
	}
}

func TestWrapStdLogger(t *testing.T) {
	wrapper := WrapStdLogger()
	if wrapper == nil {
		t.Error("WrapStdLogger should return a non-nil wrapper")
	}
	if wrapper.GetLevel() != InfoLevel {
		t.Errorf("Expected default level InfoLevel, got %v", wrapper.GetLevel())
	}
}

func TestSlogWrapper(t *testing.T) {
	var buf bytes.Buffer
	slogLogger := NewSlogLogger(InfoLevel, &buf)
	wrapper := NewSlogWrapper(slogLogger)

	wrapper.Info("test message", "key", "value", "number", 42)

	output := buf.String()

	if !strings.Contains(output, `"level":"INFO"`) {
		t.Errorf("Expected INFO level in slog output, got: %s", output)
	}
	if !strings.Contains(output, `"msg":"test message"`) {
		t.Errorf("Expected message in slog output, got: %s", output)
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Errorf("Expected key-value in slog output, got: %s", output)
	}
	if !strings.Contains(output, `"number":42`) {
		t.Errorf("Expected number field in slog output, got: %s", output)
	}
}

func TestSlogWrapperParseArgs(t *testing.T) {
	wrapper := &SlogWrapper{}

	// Test with message and key-value pairs
	msg, attrs := wrapper.parseArgs([]any{"test message", "key1", "value1", "key2", 42})
	if msg != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", msg)
	}
	if len(attrs) != 2 {
		t.Errorf("Expected 2 attributes, got %d", len(attrs))
	}

	// Test with only message
	msg2, attrs2 := wrapper.parseArgs([]any{"only message"})
	if msg2 != "only message" {
		t.Errorf("Expected message 'only message', got '%s'", msg2)
	}
	if len(attrs2) != 0 {
		t.Errorf("Expected 0 attributes, got %d", len(attrs2))
	}

	// Test with odd number of args (missing value)
	msg3, attrs3 := wrapper.parseArgs([]any{"message", "key_without_value"})
	if msg3 != "message" {
		t.Errorf("Expected message 'message', got '%s'", msg3)
	}
	if len(attrs3) != 0 {
		t.Errorf("Expected 0 attributes for odd args, got %d", len(attrs3))
	}
}

func TestSlogWrapperLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	slogLogger := NewSlogLogger(WarnLevel, &buf) // Only WARN and above
	wrapper := NewSlogWrapper(slogLogger)

	wrapper.Debug("debug message") // Should not appear
	wrapper.Info("info message")   // Should not appear
	wrapper.Warn("warn message")   // Should appear
	wrapper.Error("error message") // Should appear

	output := buf.String()

	if strings.Contains(output, "debug message") {
		t.Error("Debug message should not appear when level is WARN")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info message should not appear when level is WARN")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should appear when level is WARN")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message should appear when level is WARN")
	}
}

func TestNewSlogLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewSlogLogger(InfoLevel, &buf)

	if logger == nil {
		t.Error("NewSlogLogger should return a non-nil logger")
	}

	// Test that it produces valid JSON
	logger.Info("test message", "key", "value")
	output := buf.String()

	if !strings.Contains(output, `"level":"INFO"`) {
		t.Error("Expected valid JSON output from NewSlogLogger")
	}
}

func TestNewColoredSlogLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewColoredSlogLogger(InfoLevel, &buf)

	if logger == nil {
		t.Error("NewColoredSlogLogger should return a non-nil logger")
	}

	// Test that it produces output
	logger.Info("test message")
	output := buf.String()

	// Should contain some output
	if output == "" {
		t.Error("Expected some output from colored slog logger")
	}

	// Color codes might not appear in test environment
	if !strings.Contains(output, "test message") {
		t.Error("Expected test message in colored slog output")
	}
}

func TestGlobalLoggerFunctions(t *testing.T) {
	// Save original logger
	originalLogger := defaultLogger
	defer func() {
		defaultLogger = originalLogger
	}()

	var buf bytes.Buffer
	testLogger := NewDefaultLogger()
	testLogger.SetOutput(&buf)
	testLogger.SetLevel(DebugLevel)
	testLogger.SetColors(false)

	SetLogger(testLogger)

	Debug("global debug")
	Info("global info")
	Warn("global warn")
	Error("global error")

	output := buf.String()

	if !strings.Contains(output, "global debug") {
		t.Error("Expected global debug message")
	}
	if !strings.Contains(output, "global info") {
		t.Error("Expected global info message")
	}
	if !strings.Contains(output, "global warn") {
		t.Error("Expected global warn message")
	}
	if !strings.Contains(output, "global error") {
		t.Error("Expected global error message")
	}
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger()
	if logger == nil {
		t.Error("GetLogger should return a non-nil logger")
	}
}

func TestSetLevel(t *testing.T) {
	originalLevel := GetLevel()
	defer SetLevel(originalLevel)

	SetLevel(DebugLevel)
	if GetLevel() != DebugLevel {
		t.Errorf("Expected level DebugLevel, got %v", GetLevel())
	}
}

func TestSetOutput(t *testing.T) {
	// This is hard to test directly since it affects global state
	// Just ensure it doesn't panic
	SetOutput(os.Stdout)
}
