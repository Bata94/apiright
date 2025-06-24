package core

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bata94/apiright/pkg/logger"
)

// Mock logger for testing
type mockLogger struct{}

func (m mockLogger) Debug(args ...interface{})                 {}
func (m mockLogger) Debugf(format string, args ...interface{}) {}
func (m mockLogger) Info(args ...interface{})                  {}
func (m mockLogger) Infof(format string, args ...interface{})  {}
func (m mockLogger) Warn(args ...interface{})                  {}
func (m mockLogger) Warnf(format string, args ...interface{})  {}
func (m mockLogger) Error(args ...interface{})                 {}
func (m mockLogger) Errorf(format string, args ...interface{}) {}
func (m mockLogger) Fatal(args ...interface{})                 {}
func (m mockLogger) Fatalf(format string, args ...interface{}) {}
func (m mockLogger) Panic(args ...interface{})                 {}
func (m mockLogger) Panicf(format string, args ...interface{}) {}
func (m mockLogger) SetLevel(level logger.LogLevel)            {}
func (m mockLogger) GetLevel() logger.LogLevel                 { return logger.InfoLevel }
func (m mockLogger) SetOutput(output io.Writer)                {}

func TestCORSMiddleware(t *testing.T) {
	// Set global logger to mock
	log = &mockLogger{}

	// Create a test handler
	testHandler := func(c *Ctx) error {
		c.Response.SetMessage("Test response")
		return nil
	}

	// Create a CORS middleware with default config
	corsMiddleware := CORSMiddleware(DefaultCORSConfig())

	// Apply the middleware to the test handler
	handler := corsMiddleware(testHandler)

	// Test cases
	tests := []struct {
		name           string
		method         string
		origin         string
		expectedStatus int
		expectedHeader map[string]string
	}{
		{
			name:           "Simple GET request with origin",
			method:         "GET",
			origin:         "http://example.com",
			expectedStatus: http.StatusOK,
			expectedHeader: map[string]string{
				"Access-Control-Allow-Origin": "*",
			},
		},
		{
			name:           "Preflight OPTIONS request",
			method:         "OPTIONS",
			origin:         "http://example.com",
			expectedStatus: http.StatusNoContent,
			expectedHeader: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Max-Age":       "86400",
			},
		},
		{
			name:           "Request without origin",
			method:         "GET",
			origin:         "",
			expectedStatus: http.StatusOK,
			expectedHeader: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request
			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			// Create a response recorder
			w := httptest.NewRecorder()

			// Create a context with a buffered channel to avoid blocking
			ctx := &Ctx{
				Request:    req,
				Response:   NewApiResponse(),
				conClosed:  make(chan bool, 1), // Use buffered channel
					conStarted: time.Now(),
			}

			// Call the handler
			err := handler(ctx)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			// Write response directly to avoid channel blocking
			for k, v := range ctx.Response.Headers {
				w.Header().Set(k, v)
			}
			w.WriteHeader(ctx.Response.StatusCode)
			if ctx.Response.Data == nil {
				w.Write([]byte(ctx.Response.Message))
			} else {
				w.Write(ctx.Response.Data)
			}

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check headers
			for key, value := range tt.expectedHeader {
				if w.Header().Get(key) != value {
					t.Errorf("Expected header %s to be %s, got %s", key, value, w.Header().Get(key))
				}
			}
		})
	}
}