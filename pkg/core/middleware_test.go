package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bata94/apiright/pkg/logger"
)

// TestCORSMiddleware tests the CORSMiddleware
func TestCORSMiddleware(t *testing.T) {
	// Mock handler that does nothing
	mockHandler := func(c *Ctx) error {
		return nil
	}

	// Test cases
	testCases := []struct {
		name            string
		config          CORSConfig
		requestHeaders  map[string]string
		expectedStatus  int
		expectedHeaders map[string]string
	}{
		{
			name:            "No Origin Header",
			config:          DefaultCORSConfig(),
			requestHeaders:  map[string]string{},
			expectedStatus:  http.StatusOK,
			expectedHeaders: map[string]string{},
		},
		{
			name:   "Allowed Origin",
			config: DefaultCORSConfig(),
			requestHeaders: map[string]string{
				"Origin": "http://localhost:3000",
			},
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin": "*",
			},
		},
		{
			name: "Disallowed Origin",
			config: CORSConfig{
				AllowOrigins: []string{"http://example.com"},
			},
			requestHeaders: map[string]string{
				"Origin": "http://disallowed.com",
			},
			expectedStatus:  http.StatusOK,
			expectedHeaders: map[string]string{},
		},
		{
			name:   "Preflight Request",
			config: DefaultCORSConfig(),
			requestHeaders: map[string]string{
				"Origin":                        "http://localhost:3000",
				"Access-Control-Request-Method": "GET",
			},
			expectedStatus: http.StatusNoContent,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Max-Age":       "86400",
			},
		},
		{
			name: "Preflight Request with Credentials",
			config: CORSConfig{
				AllowOrigins:     []string{"http://localhost:3000"},
				AllowCredentials: true,
			},
			requestHeaders: map[string]string{
				"Origin":                        "http://localhost:3000",
				"Access-Control-Request-Method": "POST",
			},
			expectedStatus: http.StatusNoContent,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://localhost:3000",
				"Access-Control-Allow-Credentials": "true",
				"Vary":                             "Origin",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("OPTIONS", "/", nil)
			for key, value := range tc.requestHeaders {
				req.Header.Set(key, value)
			}

			// For non-preflight tests, change method
			if tc.expectedStatus == http.StatusOK {
				req.Method = "GET"
			}

			res := httptest.NewRecorder()
			route := Route{}
			ep := Endpoint{}
			c := NewCtx(res, req, route, ep)

			corsMiddleware := CORSMiddleware(tc.config)
			handler := corsMiddleware(mockHandler)
			err := handler(c)

			if err != nil {
				t.Fatalf("handler error: %v", err)
			}

			if c.Response.StatusCode != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, c.Response.StatusCode)
			}

			for key, value := range tc.expectedHeaders {
				if headerValue, ok := c.Response.Headers[key]; !ok || headerValue != value {
					t.Errorf("expected header %s: %s, got: %s", key, value, headerValue)
				}
			}
		})
	}
}
func TestExposeAllCORSConfig(t *testing.T) {
	config := ExposeAllCORSConfig()
	if len(config.AllowOrigins) != 1 || config.AllowOrigins[0] != "*" {
		t.Errorf("Expected AllowOrigins [\"*\"], got %v", config.AllowOrigins)
	}
	if len(config.AllowMethods) != 1 || config.AllowMethods[0] != "*" {
		t.Errorf("Expected AllowMethods [\"*\"], got %v", config.AllowMethods)
	}
	if len(config.AllowHeaders) != 1 || config.AllowHeaders[0] != "*" {
		t.Errorf("Expected AllowHeaders [\"*\"], got %v", config.AllowHeaders)
	}
	if len(config.ExposeHeaders) != 0 {
		t.Errorf("Expected empty ExposeHeaders, got %v", config.ExposeHeaders)
	}
	if config.AllowCredentials {
		t.Error("Expected AllowCredentials to be false")
	}
	if config.MaxAge != 86400 {
		t.Errorf("Expected MaxAge 86400, got %d", config.MaxAge)
	}
}

func TestLogMiddleware(t *testing.T) {
	app := NewApp()
	app.Use(LogMiddleware(logger.NewDefaultLogger()))

	app.GET("/test", func(c *Ctx) error {
		c.Response.SetMessage("OK")
		return nil
	})

	app.addRoutesToHandler()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	app.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestPanicMiddleware(t *testing.T) {
	app := NewApp()
	app.Use(PanicMiddleware())

	app.GET("/panic", func(c *Ctx) error {
		panic("test panic")
	})

	app.GET("/normal", func(c *Ctx) error {
		c.Response.SetMessage("OK")
		return nil
	})

	app.addRoutesToHandler()

	// Test panic recovery
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	app.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d for panic, got %d", http.StatusInternalServerError, rec.Code)
	}

	// Test normal request
	req = httptest.NewRequest(http.MethodGet, "/normal", nil)
	rec = httptest.NewRecorder()
	app.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d for normal request, got %d", http.StatusOK, rec.Code)
	}
}
