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

func TestCSRFMiddleware(t *testing.T) {
	app := NewApp()
	app.Use(CSRFMiddleware(DefaultCSRFConfig()))

	app.GET("/", func(c *Ctx) error {
		c.Response.SetMessage("GET request")
		return nil
	})

	app.POST("/", func(c *Ctx) error {
		c.Response.SetMessage("POST request")
		return nil
	})

	app.addRoutesToHandler()

	// Test GET request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	app.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	// Check if CSRF cookie is set
	cookies := rec.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == DefaultCSRFConfig().CookieName {
			csrfCookie = cookie
			break
		}
	}

	if csrfCookie == nil {
		t.Fatal("CSRF cookie not set")
	}

	// Test POST request with valid CSRF token
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(DefaultCSRFConfig().HeaderName, csrfCookie.Value)
	req.AddCookie(csrfCookie)
	rec = httptest.NewRecorder()
	app.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	// Test POST request with invalid CSRF token
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(DefaultCSRFConfig().HeaderName, "invalid_token")
	req.AddCookie(csrfCookie)
	rec = httptest.NewRecorder()
	app.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, rec.Code)
	}

	// Test POST request without CSRF token
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(csrfCookie)
	rec = httptest.NewRecorder()
	app.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, rec.Code)
	}
}
