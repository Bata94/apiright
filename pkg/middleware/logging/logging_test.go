package logging

import (
	"fmt"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/logger"
)

func TestLogMiddleware(t *testing.T) {
	// Create a mock logger that captures log messages
	var loggedMessages []string
	var loggedLevels []string

	mockLogger := &mockLogger{
		logFunc: func(level, message string) {
			loggedLevels = append(loggedLevels, level)
			loggedMessages = append(loggedMessages, message)
		},
	}

	mockHandler := func(c *core.Ctx) error {
		c.Response.SetStatus(200)
		c.Response.SetMessage("OK")
		return nil
	}

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req, core.Route{}, core.Endpoint{})

	middleware := LogMiddleware(mockLogger)
	handler := middleware(mockHandler)

	err := handler(ctx)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}

	// Close the connection to trigger logging
	ctx.Close()

	// Wait a bit for the goroutine to complete
	time.Sleep(10 * time.Millisecond)

	if len(loggedMessages) != 1 {
		t.Errorf("Expected 1 log message, got %d", len(loggedMessages))
	}

	if len(loggedLevels) != 1 {
		t.Errorf("Expected 1 log level, got %d", len(loggedLevels))
	}

	if loggedLevels[0] != "info" {
		t.Errorf("Expected log level 'info', got '%s'", loggedLevels[0])
	}

	// Test error logging
	loggedMessages = nil
	loggedLevels = nil

	req2 := httptest.NewRequest("POST", "/error", nil)
	w2 := httptest.NewRecorder()
	ctx2 := core.NewCtx(w2, req2, core.Route{}, core.Endpoint{})

	mockHandlerError := func(c *core.Ctx) error {
		c.Response.SetStatus(500)
		c.Response.SetMessage("Internal Server Error")
		return nil
	}

	handler2 := middleware(mockHandlerError)
	err = handler2(ctx2)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}

	ctx2.Close()
	time.Sleep(10 * time.Millisecond)

	if len(loggedMessages) != 1 {
		t.Errorf("Expected 1 log message for error, got %d", len(loggedMessages))
	}

	if loggedLevels[0] != "error" {
		t.Errorf("Expected log level 'error', got '%s'", loggedLevels[0])
	}
}

// mockLogger implements the Logger interface for testing
type mockLogger struct {
	logFunc func(level, message string)
}

func (m *mockLogger) Debug(args ...any) {
	if len(args) > 0 {
		if msg, ok := args[0].(string); ok {
			m.logFunc("debug", msg)
		}
	}
}

func (m *mockLogger) Debugf(format string, args ...any) {
	m.logFunc("debug", fmt.Sprintf(format, args...))
}

func (m *mockLogger) Info(args ...any) {
	if len(args) > 0 {
		if msg, ok := args[0].(string); ok {
			m.logFunc("info", msg)
		}
	}
}

func (m *mockLogger) Infof(format string, args ...any) {
	m.logFunc("info", fmt.Sprintf(format, args...))
}

func (m *mockLogger) Warn(args ...any) {
	if len(args) > 0 {
		if msg, ok := args[0].(string); ok {
			m.logFunc("warn", msg)
		}
	}
}

func (m *mockLogger) Warnf(format string, args ...any) {
	m.logFunc("warn", fmt.Sprintf(format, args...))
}

func (m *mockLogger) Error(args ...any) {
	if len(args) > 0 {
		if msg, ok := args[0].(string); ok {
			m.logFunc("error", msg)
		}
	}
}

func (m *mockLogger) Errorf(format string, args ...any) {
	m.logFunc("error", fmt.Sprintf(format, args...))
}

func (m *mockLogger) Fatal(args ...any) {
	if len(args) > 0 {
		if msg, ok := args[0].(string); ok {
			m.logFunc("fatal", msg)
		}
	}
}

func (m *mockLogger) Fatalf(format string, args ...any) {
	m.logFunc("fatal", fmt.Sprintf(format, args...))
}

func (m *mockLogger) Panic(args ...any) {
	if len(args) > 0 {
		if msg, ok := args[0].(string); ok {
			m.logFunc("panic", msg)
		}
	}
}

func (m *mockLogger) Panicf(format string, args ...any) {
	m.logFunc("panic", fmt.Sprintf(format, args...))
}

func (m *mockLogger) SetLevel(level logger.LogLevel) {}
func (m *mockLogger) GetLevel() logger.LogLevel      { return logger.InfoLevel }
func (m *mockLogger) SetOutput(output io.Writer)     {}
func (m *mockLogger) SetColors(enabled bool)         {}
