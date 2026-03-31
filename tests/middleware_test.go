package apiright_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/middleware"
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

type testMiddleware struct {
	name     string
	priority int
}

func (tm *testMiddleware) Name() string  { return tm.name }
func (tm *testMiddleware) Priority() int { return tm.priority }
func (tm *testMiddleware) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test-Middleware", tm.name)
			next.ServeHTTP(w, r)
		})
	}
}

func TestMiddlewareRegistry_Register(t *testing.T) {
	logger := &mockLogger{}
	registry := middleware.NewMiddlewareRegistry(logger)

	mw := &testMiddleware{name: "test", priority: 100}
	registry.RegisterMiddleware(mw)

	list := registry.ListMiddleware()
	if len(list) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(list))
	}
	if list[0].Name() != "test" {
		t.Errorf("Expected middleware name 'test', got '%s'", list[0].Name())
	}
}

func TestMiddlewareRegistry_SortByPriority(t *testing.T) {
	logger := &mockLogger{}
	registry := middleware.NewMiddlewareRegistry(logger)

	registry.RegisterMiddleware(&testMiddleware{name: "low", priority: 100})
	registry.RegisterMiddleware(&testMiddleware{name: "high", priority: 10})
	registry.RegisterMiddleware(&testMiddleware{name: "medium", priority: 50})

	list := registry.ListMiddleware()

	if len(list) != 3 {
		t.Fatalf("Expected 3 middleware, got %d", len(list))
	}

	if list[0].Name() != "high" {
		t.Errorf("Expected first middleware 'high', got '%s'", list[0].Name())
	}
	if list[1].Name() != "medium" {
		t.Errorf("Expected second middleware 'medium', got '%s'", list[1].Name())
	}
	if list[2].Name() != "low" {
		t.Errorf("Expected third middleware 'low', got '%s'", list[2].Name())
	}
}

func TestMiddlewareRegistry_GetByName(t *testing.T) {
	logger := &mockLogger{}
	registry := middleware.NewMiddlewareRegistry(logger)

	registry.RegisterMiddleware(&testMiddleware{name: "test", priority: 100})

	mw, found := registry.GetMiddlewareByName("test")
	if !found {
		t.Error("Expected to find middleware 'test'")
	}
	if mw.Name() != "test" {
		t.Errorf("Expected middleware name 'test', got '%s'", mw.Name())
	}

	_, found = registry.GetMiddlewareByName("nonexistent")
	if found {
		t.Error("Expected not to find nonexistent middleware")
	}
}

func TestMiddlewareRegistry_Remove(t *testing.T) {
	logger := &mockLogger{}
	registry := middleware.NewMiddlewareRegistry(logger)

	registry.RegisterMiddleware(&testMiddleware{name: "test", priority: 100})

	removed := registry.RemoveMiddleware("test")
	if !removed {
		t.Error("Expected to remove middleware")
	}

	list := registry.ListMiddleware()
	if len(list) != 0 {
		t.Errorf("Expected 0 middleware after removal, got %d", len(list))
	}

	removed = registry.RemoveMiddleware("nonexistent")
	if removed {
		t.Error("Expected removal to fail for nonexistent middleware")
	}
}

func TestMiddlewareRegistry_GetHTTPMiddleware(t *testing.T) {
	logger := &mockLogger{}
	registry := middleware.NewMiddlewareRegistry(logger)

	registry.RegisterMiddleware(&testMiddleware{name: "mw1", priority: 10})
	registry.RegisterMiddleware(&testMiddleware{name: "mw2", priority: 20})

	handlers := registry.GetHTTPMiddleware()
	if len(handlers) != 2 {
		t.Errorf("Expected 2 handlers, got %d", len(handlers))
	}

	recorder := httptest.NewRecorder()
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	for _, h := range handlers {
		handler = h(handler)
	}

	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/", nil))

	if recorder.Header().Get("X-Test-Middleware") != "mw1" {
		t.Errorf("Expected first middleware 'mw1' (last applied), got '%s'", recorder.Header().Get("X-Test-Middleware"))
	}
}

func TestCORSMiddleware_AllowedOrigin(t *testing.T) {
	logger := &mockLogger{}
	config := middleware.CORSConfig{
		AllowOrigins: []string{"https://example.com"},
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"Content-Type"},
		MaxAge:       86400,
	}
	cors := middleware.NewCORSMiddleware(config, logger)

	handler := cors.Handler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("allowed origin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Origin", "https://example.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
			t.Error("Expected Access-Control-Allow-Origin header")
		}
	})

	t.Run("disallowed origin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Origin", "https://evil.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Error("Expected no Access-Control-Allow-Origin header for disallowed origin")
		}
	})
}

func TestCORSMiddleware_Preflight(t *testing.T) {
	logger := &mockLogger{}
	config := middleware.DefaultCORSConfig()
	cors := middleware.NewCORSMiddleware(config, logger)

	handler := cors.Handler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", rec.Code)
	}
}

func TestCORSMiddleware_WildcardOrigin(t *testing.T) {
	logger := &mockLogger{}
	config := middleware.DefaultCORSConfig()
	cors := middleware.NewCORSMiddleware(config, logger)

	handler := cors.Handler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://any-site.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected '*' for wildcard origin, got '%s'", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestLoggingMiddleware(t *testing.T) {
	logger := &mockLogger{}
	logging := middleware.NewLoggingMiddleware(logger)

	handler := logging.Handler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestRateLimitMiddleware_AllowsUnderLimit(t *testing.T) {
	logger := &mockLogger{}
	rateLimit := middleware.NewRateLimitMiddleware(10, 60, logger)

	handler := rateLimit.Handler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, rec.Code)
		}
	}
}

func TestRateLimitMiddleware_BlocksOverLimit(t *testing.T) {
	logger := &mockLogger{}
	rateLimit := middleware.NewRateLimitMiddleware(2, time.Second, logger)

	handler := rateLimit.Handler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("Request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429, got %d", rec.Code)
	}
}

func TestRateLimitMiddleware_DifferentIPs(t *testing.T) {
	logger := &mockLogger{}
	rateLimit := middleware.NewRateLimitMiddleware(1, 60, logger)

	handler := rateLimit.Handler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = fmt.Sprintf("192.168.1.%d:12345", i+1)

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Different IPs should have separate limits")
		}
	}
}

func TestRateLimitMiddleware_WindowReset(t *testing.T) {
	logger := &mockLogger{}
	rateLimit := middleware.NewRateLimitMiddleware(1, 100*time.Millisecond, logger)

	handler := rateLimit.Handler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("First request: expected 200, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Second request: expected 429, got %d", rec.Code)
	}

	time.Sleep(150 * time.Millisecond)

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("After reset: expected 200, got %d", rec.Code)
	}
}

func TestRateLimitMiddleware_AcceptHeader(t *testing.T) {
	logger := &mockLogger{}
	rateLimit := middleware.NewRateLimitMiddleware(0, time.Second, logger)

	handler := rateLimit.Handler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	tests := []struct {
		accept   string
		expected string
	}{
		{"application/json", "application/json"},
		{"text/plain", "text/plain"},
		{"", "application/json"},
	}

	for _, tc := range tests {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept", tc.accept)
		handler.ServeHTTP(rec, req)
		if rec.Header().Get("Content-Type") != tc.expected {
			t.Errorf("Accept=%s: expected %s, got %s", tc.accept, tc.expected, rec.Header().Get("Content-Type"))
		}
	}
}

func TestDefaultCORSConfig(t *testing.T) {
	config := middleware.DefaultCORSConfig()

	if len(config.AllowOrigins) != 1 || config.AllowOrigins[0] != "*" {
		t.Error("Expected AllowOrigins to contain '*'")
	}

	expectedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	if len(config.AllowMethods) != len(expectedMethods) {
		t.Errorf("Expected %d methods, got %d", len(expectedMethods), len(config.AllowMethods))
	}

	if config.MaxAge != 86400 {
		t.Errorf("Expected MaxAge 86400, got %d", config.MaxAge)
	}
}
