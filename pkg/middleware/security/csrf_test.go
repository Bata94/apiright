package security

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bata94/apiright/pkg/core"
)

func TestDefaultCSRFConfig(t *testing.T) {
	config := DefaultCSRFConfig()

	if config.TokenLength != 32 {
		t.Errorf("Expected TokenLength 32, got %d", config.TokenLength)
	}

	if config.CookieName != "csrf_token" {
		t.Errorf("Expected CookieName 'csrf_token', got '%s'", config.CookieName)
	}

	if config.CookiePath != "/" {
		t.Errorf("Expected CookiePath '/', got '%s'", config.CookiePath)
	}

	if config.CookieExpires != 24*time.Hour {
		t.Errorf("Expected CookieExpires 24h, got %v", config.CookieExpires)
	}

	if !config.CookieSecure {
		t.Error("Expected CookieSecure to be true")
	}

	if !config.CookieHTTPOnly {
		t.Error("Expected CookieHTTPOnly to be true")
	}

	if config.HeaderName != "X-CSRF-Token" {
		t.Errorf("Expected HeaderName 'X-CSRF-Token', got '%s'", config.HeaderName)
	}
}

func TestGenerateRandomToken(t *testing.T) {
	token1, err := generateRandomToken(32)
	if err != nil {
		t.Errorf("generateRandomToken returned error: %v", err)
	}

	if len(token1) == 0 {
		t.Error("Expected non-empty token")
	}

	token2, err := generateRandomToken(32)
	if err != nil {
		t.Errorf("generateRandomToken returned error: %v", err)
	}

	// Tokens should be different (very unlikely to be the same)
	if token1 == token2 {
		t.Error("Expected different tokens, got identical ones")
	}

	// Test different lengths
	token3, err := generateRandomToken(16)
	if err != nil {
		t.Errorf("generateRandomToken returned error: %v", err)
	}

	if len(token3) == 0 {
		t.Error("Expected non-empty token for length 16")
	}
}

func TestCSRFMiddleware(t *testing.T) {
	mockHandler := func(c *core.Ctx) error {
		c.Response.SetMessage("OK")
		return nil
	}

	tests := []struct {
		name           string
		config         CSRFConfig
		requestMethod  string
		requestHeaders map[string]string
		cookies        []*http.Cookie
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "GET request sets CSRF cookie",
			config:         DefaultCSRFConfig(),
			requestMethod:  http.MethodGet,
			requestHeaders: map[string]string{},
			cookies:        []*http.Cookie{},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				cookies := w.Result().Cookies()
				if len(cookies) != 1 {
					t.Errorf("Expected 1 cookie, got %d", len(cookies))
					return
				}

				cookie := cookies[0]
				if cookie.Name != "csrf_token" {
					t.Errorf("Expected cookie name 'csrf_token', got '%s'", cookie.Name)
				}

				if cookie.Value == "" {
					t.Error("Expected non-empty cookie value")
				}

				if !cookie.HttpOnly {
					t.Error("Expected HttpOnly cookie")
				}

				if !cookie.Secure {
					t.Error("Expected secure cookie")
				}
			},
		},
		{
			name:          "POST request with valid CSRF token",
			config:        DefaultCSRFConfig(),
			requestMethod: http.MethodPost,
			requestHeaders: map[string]string{
				"X-CSRF-Token": "valid-token",
			},
			cookies: []*http.Cookie{
				{Name: "csrf_token", Value: "valid-token"},
			},
			expectedStatus: http.StatusOK,
			checkResponse:  func(t *testing.T, w *httptest.ResponseRecorder) {},
		},
		{
			name:          "POST request with invalid CSRF token",
			config:        DefaultCSRFConfig(),
			requestMethod: http.MethodPost,
			requestHeaders: map[string]string{
				"X-CSRF-Token": "invalid-token",
			},
			cookies: []*http.Cookie{
				{Name: "csrf_token", Value: "valid-token"},
			},
			expectedStatus: http.StatusForbidden,
			checkResponse:  func(t *testing.T, w *httptest.ResponseRecorder) {},
		},
		{
			name:           "POST request without CSRF header",
			config:         DefaultCSRFConfig(),
			requestMethod:  http.MethodPost,
			requestHeaders: map[string]string{},
			cookies: []*http.Cookie{
				{Name: "csrf_token", Value: "some-token"},
			},
			expectedStatus: http.StatusForbidden,
			checkResponse:  func(t *testing.T, w *httptest.ResponseRecorder) {},
		},
		{
			name:           "POST request without CSRF cookie",
			config:         DefaultCSRFConfig(),
			requestMethod:  http.MethodPost,
			requestHeaders: map[string]string{"X-CSRF-Token": "some-token"},
			cookies:        []*http.Cookie{},
			expectedStatus: http.StatusForbidden,
			checkResponse:  func(t *testing.T, w *httptest.ResponseRecorder) {},
		},
		{
			name: "custom header name",
			config: CSRFConfig{
				TokenLength:    32,
				CookieName:     "custom_csrf",
				CookiePath:     "/",
				CookieExpires:  24 * time.Hour,
				CookieSecure:   false,
				CookieHTTPOnly: false,
				HeaderName:     "X-Custom-CSRF",
			},
			requestMethod: http.MethodPost,
			requestHeaders: map[string]string{
				"X-Custom-CSRF": "custom-token",
			},
			cookies: []*http.Cookie{
				{Name: "custom_csrf", Value: "custom-token"},
			},
			expectedStatus: http.StatusOK,
			checkResponse:  func(t *testing.T, w *httptest.ResponseRecorder) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.requestMethod, "/", nil)
			for key, value := range tt.requestHeaders {
				req.Header.Set(key, value)
			}

			// Add cookies to request
			for _, cookie := range tt.cookies {
				req.AddCookie(cookie)
			}

			w := httptest.NewRecorder()
			ctx := core.NewCtx(w, req, core.Route{}, core.Endpoint{})

			middleware := CSRFMiddleware(tt.config)
			handler := middleware(mockHandler)
			err := handler(ctx)

			if err != nil && !strings.Contains(err.Error(), "CSRF token mismatch") && tt.expectedStatus != http.StatusForbidden {
				t.Errorf("Handler returned unexpected error: %v", err)
			}

			ctx.SendingReturn(w, nil)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}
