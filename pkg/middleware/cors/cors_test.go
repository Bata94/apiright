package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bata94/apiright/pkg/core"
)

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig()

	expectedOrigins := []string{"*"}
	if len(config.AllowOrigins) != len(expectedOrigins) || config.AllowOrigins[0] != expectedOrigins[0] {
		t.Errorf("Expected AllowOrigins %v, got %v", expectedOrigins, config.AllowOrigins)
	}

	expectedMethods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions}
	if len(config.AllowMethods) != len(expectedMethods) {
		t.Errorf("Expected %d methods, got %d", len(expectedMethods), len(config.AllowMethods))
	}
	for i, method := range expectedMethods {
		if config.AllowMethods[i] != method {
			t.Errorf("Expected method %s at index %d, got %s", method, i, config.AllowMethods[i])
		}
	}

	if config.AllowCredentials != false {
		t.Errorf("Expected AllowCredentials to be false, got %v", config.AllowCredentials)
	}

	if config.MaxAge != 86400 {
		t.Errorf("Expected MaxAge to be 86400, got %d", config.MaxAge)
	}
}

func TestExposeAllCORSConfig(t *testing.T) {
	config := ExposeAllCORSConfig()

	if config.AllowOrigins[0] != "*" {
		t.Errorf("Expected AllowOrigins to be ['*'], got %v", config.AllowOrigins)
	}

	if config.AllowMethods[0] != "*" {
		t.Errorf("Expected AllowMethods to be ['*'], got %v", config.AllowMethods)
	}

	if config.AllowHeaders[0] != "*" {
		t.Errorf("Expected AllowHeaders to be ['*'], got %v", config.AllowHeaders)
	}
}

func TestNormalizeCORSConfig(t *testing.T) {
	config := CORSConfig{
		AllowMethods:  []string{"get", "post", "PUT"},
		AllowHeaders:  []string{"content-type", "Authorization"},
		ExposeHeaders: []string{"x-custom-header"},
	}

	normalizeCORSConfig(&config)

	expectedMethods := []string{"GET", "POST", "PUT"}
	for i, method := range expectedMethods {
		if config.AllowMethods[i] != method {
			t.Errorf("Expected method %s, got %s", method, config.AllowMethods[i])
		}
	}

	expectedHeaders := []string{"Content-Type", "Authorization"}
	for i, header := range expectedHeaders {
		if config.AllowHeaders[i] != header {
			t.Errorf("Expected header %s, got %s", header, config.AllowHeaders[i])
		}
	}

	if config.ExposeHeaders[0] != "X-Custom-Header" {
		t.Errorf("Expected exposed header 'X-Custom-Header', got %s", config.ExposeHeaders[0])
	}
}

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name            string
		origin          string
		allowedOrigins  []string
		expectedOrigin  string
		expectedAllowed bool
	}{
		{
			name:            "empty origin",
			origin:          "",
			allowedOrigins:  []string{"*"},
			expectedOrigin:  "",
			expectedAllowed: false,
		},
		{
			name:            "wildcard allows all",
			origin:          "http://example.com",
			allowedOrigins:  []string{"*"},
			expectedOrigin:  "*",
			expectedAllowed: true,
		},
		{
			name:            "exact match",
			origin:          "http://example.com",
			allowedOrigins:  []string{"http://example.com"},
			expectedOrigin:  "http://example.com",
			expectedAllowed: true,
		},
		{
			name:            "case insensitive match",
			origin:          "http://EXAMPLE.COM",
			allowedOrigins:  []string{"http://example.com"},
			expectedOrigin:  "http://EXAMPLE.COM",
			expectedAllowed: true,
		},
		{
			name:            "no match",
			origin:          "http://example.com",
			allowedOrigins:  []string{"http://other.com"},
			expectedOrigin:  "",
			expectedAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origin, allowed := isOriginAllowed(tt.origin, tt.allowedOrigins)
			if origin != tt.expectedOrigin {
				t.Errorf("Expected origin %s, got %s", tt.expectedOrigin, origin)
			}
			if allowed != tt.expectedAllowed {
				t.Errorf("Expected allowed %v, got %v", tt.expectedAllowed, allowed)
			}
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	mockHandler := func(c *core.Ctx) error {
		c.Response.SetMessage("OK")
		return nil
	}

	tests := []struct {
		name            string
		config          CORSConfig
		requestMethod   string
		requestHeaders  map[string]string
		expectedStatus  int
		expectedHeaders map[string]string
	}{
		{
			name:            "no origin header",
			config:          DefaultCORSConfig(),
			requestMethod:   http.MethodGet,
			requestHeaders:  map[string]string{},
			expectedStatus:  http.StatusOK,
			expectedHeaders: map[string]string{},
		},
		{
			name:          "allowed origin GET request",
			config:        DefaultCORSConfig(),
			requestMethod: http.MethodGet,
			requestHeaders: map[string]string{
				"Origin": "http://localhost:3000",
			},
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin": "*",
			},
		},
		{
			name:          "disallowed origin",
			config:        CORSConfig{AllowOrigins: []string{"http://allowed.com"}},
			requestMethod: http.MethodGet,
			requestHeaders: map[string]string{
				"Origin": "http://disallowed.com",
			},
			expectedStatus:  http.StatusOK,
			expectedHeaders: map[string]string{},
		},
		{
			name:          "preflight request",
			config:        DefaultCORSConfig(),
			requestMethod: http.MethodOptions,
			requestHeaders: map[string]string{
				"Origin":                        "http://localhost:3000",
				"Access-Control-Request-Method": "POST",
			},
			expectedStatus: http.StatusNoContent,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Max-Age":       "86400",
			},
		},
		{
			name: "preflight with credentials",
			config: CORSConfig{
				AllowOrigins:     []string{"http://localhost:3000"},
				AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
				AllowCredentials: true,
				MaxAge:           3600,
			},
			requestMethod: http.MethodOptions,
			requestHeaders: map[string]string{
				"Origin":                        "http://localhost:3000",
				"Access-Control-Request-Method": "POST",
			},
			expectedStatus: http.StatusNoContent,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://localhost:3000",
				"Access-Control-Allow-Credentials": "true",
				"Vary":                             "Origin",
				"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Max-Age":           "3600",
			},
		},
		{
			name: "preflight with custom headers",
			config: CORSConfig{
				AllowOrigins: []string{"http://localhost:3000"},
				AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
				AllowHeaders: []string{"X-Custom-Header"},
			},
			requestMethod: http.MethodOptions,
			requestHeaders: map[string]string{
				"Origin":                         "http://localhost:3000",
				"Access-Control-Request-Method":  "POST",
				"Access-Control-Request-Headers": "X-Custom-Header",
			},
			expectedStatus: http.StatusNoContent,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "http://localhost:3000",
				"Access-Control-Allow-Headers": "X-Custom-Header",
				"Vary":                         "Origin",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.requestMethod, "/", nil)
			for key, value := range tt.requestHeaders {
				req.Header.Set(key, value)
			}

			w := httptest.NewRecorder()
			ctx := core.NewCtx(w, req, core.Route{}, core.Endpoint{})

			middleware := CORSMiddleware(tt.config)
			handler := middleware(mockHandler)
			err := handler(ctx)

			if err != nil {
				t.Errorf("Handler returned error: %v", err)
			}

			ctx.SendingReturn(w, nil)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			for key, expectedValue := range tt.expectedHeaders {
				actualValue := w.Header().Get(key)
				if actualValue != expectedValue {
					t.Errorf("Expected header %s: %s, got: %s", key, expectedValue, actualValue)
				}
			}
		})
	}
}
