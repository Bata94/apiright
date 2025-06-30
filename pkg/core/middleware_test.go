package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
			c := NewCtx(res, req)

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
