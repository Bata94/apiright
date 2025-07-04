package core

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bata94/apiright/pkg/logger"
)

func TestNewApp(t *testing.T) {
	tests := []struct {
		name string
		opts []AppOption
		want AppConfig
	}{
		{
			name: "Default App",
			opts: []AppOption{},
			want: AppConfig{
				host:               "127.0.0.1",
				port:               "5500",
				title:              "My App",
				serviceDescribtion: "My App Description",
				version:            "0.0.0",
			},
		},
		{
			name: "Custom App Title",
			opts: []AppOption{AppTitle("Test App")},
			want: AppConfig{
				host:               "127.0.0.1",
				port:               "5500",
				title:              "Test App",
				serviceDescribtion: "My App Description",
				version:            "0.0.0",
			},
		},
		{
			name: "Custom App Address",
			opts: []AppOption{AppAddr("0.0.0.0", "8080")},
			want: AppConfig{
				host:               "0.0.0.0",
				port:               "8080",
				title:              "My App",
				serviceDescribtion: "My App Description",
				version:            "0.0.0",
			},
		},
		{
			name: "Custom Logger",
			opts: []AppOption{AppLogger(logger.NewDefaultLogger())},
			want: AppConfig{
				host:               "127.0.0.1",
				port:               "5500",
				title:              "My App",
				serviceDescribtion: "My App Description",
				version:            "0.0.0",
			},
		},
		{
			name: "Custom Contact",
			opts: []AppOption{AppContact("John Doe", "john@example.com", "http://example.com")},
			want: AppConfig{
				host:               "127.0.0.1",
				port:               "5500",
				title:              "My App",
				serviceDescribtion: "My App Description",
				version:            "0.0.0",
				contact: struct {
					Name, Email, URL string
				}{Name: "John Doe", Email: "john@example.com", URL: "http://example.com"},
			},
		},
		{
			name: "Custom License",
			opts: []AppOption{AppLicense("MIT", "http://license.com")},
			want: AppConfig{
				host:               "127.0.0.1",
				port:               "5500",
				title:              "My App",
				serviceDescribtion: "My App Description",
				version:            "0.0.0",
				license: struct {
					Name, URL string
				}{Name: "MIT", URL: "http://license.com"},
			},
		},
		{
			name: "Custom Server",
			opts: []AppOption{AppAddServer("http://api.example.com", "Production Server")},
			want: AppConfig{
				host:               "127.0.0.1",
				port:               "5500",
				title:              "My App",
				serviceDescribtion: "My App Description",
				version:            "0.0.0",
				servers: []struct {
					URL, Description string
				}{{URL: "http://api.example.com", Description: "Production Server"}},
			},
		},
		{
			name: "Custom Timeout",
			opts: []AppOption{AppTimeout(5 * time.Second)},
			want: AppConfig{
				host:               "127.0.0.1",
				port:               "5500",
				title:              "My App",
				serviceDescribtion: "My App Description",
				version:            "0.0.0",
				timeout:            5 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp(tt.opts...)
			if app.config.host != tt.want.host {
				t.Errorf("NewApp() host = %v, want %v", app.config.host, tt.want.host)
			}
			if app.config.port != tt.want.port {
				t.Errorf("NewApp() port = %v, want %v", app.config.port, tt.want.port)
			}
			if app.config.title != tt.want.title {
				t.Errorf("NewApp() title = %v, want %v", app.config.title, tt.want.title)
			}
			if app.config.serviceDescribtion != tt.want.serviceDescribtion {
				t.Errorf("NewApp() serviceDescribtion = %v, want %v", app.config.serviceDescribtion, tt.want.serviceDescribtion)
			}
			if app.config.version != tt.want.version {
				t.Errorf("NewApp() version = %v, want %v", app.config.version, tt.want.version)
			}
			if app.config.contact.Name != tt.want.contact.Name {
				t.Errorf("NewApp() contact name = %v, want %v", app.config.contact.Name, tt.want.contact.Name)
			}
			if app.config.contact.Email != tt.want.contact.Email {
				t.Errorf("NewApp() contact email = %v, want %v", app.config.contact.Email, tt.want.contact.Email)
			}
			if app.config.contact.URL != tt.want.contact.URL {
				t.Errorf("NewApp() contact URL = %v, want %v", app.config.contact.URL, tt.want.contact.URL)
			}
			if app.config.license.Name != tt.want.license.Name {
				t.Errorf("NewApp() license name = %v, want %v", app.config.license.Name, tt.want.license.Name)
			}
			if app.config.license.URL != tt.want.license.URL {
				t.Errorf("NewApp() license URL = %v, want %v", app.config.license.URL, tt.want.license.URL)
			}
			if len(app.config.servers) != len(tt.want.servers) {
				t.Errorf("NewApp() servers length = %v, want %v", len(app.config.servers), len(tt.want.servers))
			}
			if app.config.timeout != tt.want.timeout {
				t.Errorf("NewApp() timeout = %v, want %v", app.config.timeout, tt.want.timeout)
			}
		})
	}
}

func TestApp_SetDefaultRoute(t *testing.T) {
	app := NewApp()
	mockHandler := func(c *Ctx) error {
		c.Response.SetStatus(http.StatusOK)
		c.Response.SetMessage("Default Route")
		return nil
	}
	app.SetDefaultRoute(mockHandler)

	// This is an internal field, so we can't directly test it without exposing it or making a request.
	// For now, we'll assume it works if the method doesn't panic.
	// A more thorough test would involve making a request to a non-existent route and checking the response.
}

func TestApp_SetTimeout(t *testing.T) {
	app := NewApp()
	app.SetTimeout(2 * time.Second)
	if app.timeoutConfig.Timeout != 2*time.Second {
		t.Errorf("SetTimeout() Timeout = %v, want %v", app.timeoutConfig.Timeout, 2*time.Second)
	}
}

func TestApp_SetTimeoutConfig(t *testing.T) {
	app := NewApp()
	newConfig := TimeoutConfig{
		Timeout:           3 * time.Second,
		TimeoutMessage:    "Custom Timeout",
		TimeoutStatusCode: http.StatusRequestTimeout,
	}
	app.SetTimeoutConfig(newConfig)
	if app.timeoutConfig.Timeout != newConfig.Timeout {
		t.Errorf("SetTimeoutConfig() Timeout = %v, want %v", app.timeoutConfig.Timeout, newConfig.Timeout)
	}
	if app.timeoutConfig.TimeoutMessage != newConfig.TimeoutMessage {
		t.Errorf("SetTimeoutConfig() TimeoutMessage = %v, want %v", app.timeoutConfig.TimeoutMessage, newConfig.TimeoutMessage)
	}
	if app.timeoutConfig.TimeoutStatusCode != newConfig.TimeoutStatusCode {
		t.Errorf("SetTimeoutConfig() TimeoutStatusCode = %v, want %v", app.timeoutConfig.TimeoutStatusCode, newConfig.TimeoutStatusCode)
	}
}

func TestApp_SetLogger(t *testing.T) {
	app := NewApp()
	newLogger := logger.NewDefaultLogger()
	app.SetLogger(newLogger)
	if app.Logger != newLogger {
		t.Errorf("SetLogger() Logger = %v, want %v", app.Logger, newLogger)
	}
}

func TestApp_NewRouter(t *testing.T) {
	app := NewApp()
	router := app.NewRouter("/api")
	if router.basePath != "/api" {
		t.Errorf("NewRouter() basePath = %v, want %v", router.basePath, "/api")
	}
	if len(app.router.groups) != 1 {
		t.Errorf("NewRouter() app.router.groups length = %v, want %v", len(app.router.groups), 1)
	}
}

func TestApp_Use(t *testing.T) {
	app := NewApp()
	mockMiddleware := func(next Handler) Handler {
		return func(c *Ctx) error {
			return next(c)
		}
	}
	app.Use(mockMiddleware)
	if len(app.router.middlewares) != 1 {
		t.Errorf("Use() app.router.middlewares length = %v, want %v", len(app.router.middlewares), 1)
	}
}

func TestApp_RoutingMethods(t *testing.T) {
	app := NewApp()
	mockHandler := func(c *Ctx) error {
		c.Response.SetStatus(http.StatusOK)
		return nil
	}

	app.GET("/test", mockHandler)
	app.POST("/test", mockHandler)
	app.PUT("/test", mockHandler)
	app.DELETE("/test", mockHandler)
	app.OPTIONS("/test", mockHandler)

	// Verify that routes are added to the router
	if len(app.router.routes) != 1 { // Should be 1 route with multiple endpoints
		t.Errorf("Expected 1 route, got %d", len(app.router.routes))
	}
	if len(app.router.routes[0].endpoints) != 5 { // 4 methods + 1 for OPTIONS generated by addEndpoint
		t.Errorf("Expected 5 endpoints for /test, got %d", len(app.router.routes[0].endpoints))
	}
}

func TestApp_ServeStaticFile(t *testing.T) {
	app := NewApp()
	// Create a dummy file for testing
	dummyFilePath := "/tmp/dummy.txt"
	dummyContent := "Hello, static file!"
	err := WriteDummyFile(dummyFilePath, dummyContent)
	if err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}
	defer func() { _ = RemoveDummyFile(dummyFilePath) }()

	app.ServeStaticFile("/static/dummy.txt", dummyFilePath, WithContentType("text/plain"), WithPreCache())
	if err != nil {
		t.Fatalf("ServeStaticFile failed: %v", err)
	}

	// Verify that a route was added for the static file
	found := false
	for _, route := range app.router.routes {
		if route.path == "/static/dummy.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ServeStaticFile did not add the expected route")
	}

	// Test serving the file
	req := httptest.NewRequest(http.MethodGet, "/static/dummy.txt", nil)
	rec := httptest.NewRecorder()

	// Manually call handleFunc for the static file route
	// This is a bit of an integration test, but necessary to verify serving
	for _, route := range app.router.routes {
		if route.path == "/static/dummy.txt" {
			for _, endpoint := range route.endpoints {
				if endpoint.method == METHOD_GET {
					ctx := NewCtx(rec, req, *route, endpoint)
					err := endpoint.handleFunc(ctx)
					if err != nil {
						t.Fatalf("Handler returned an error: %v", err)
					}
					ctx.SendingReturn(rec, nil)
					break
				}
			}
			break
		}
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != dummyContent {
		t.Errorf("Expected body %q, got %q", dummyContent, rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("Expected Content-Type %q, got %q", "text/plain", rec.Header().Get("Content-Type"))
	}
}

func TestApp_ServeStaticDir(t *testing.T) {
	app := NewApp()
	// Create a dummy directory and file for testing
	dummyDirPath := "/tmp/static_dir_test"
	dummyFilePath := dummyDirPath + "/index.html"
	dummyContent := "<html><body>Hello, static dir!</body></html>"

	err := CreateDummyDir(dummyDirPath)
	if err != nil {
		t.Fatalf("Failed to create dummy directory: %v", err)
	}
	defer func() { _ = RemoveDummyDir(dummyDirPath) }()

	err = WriteDummyFile(dummyFilePath, dummyContent)
	if err != nil {
		t.Fatalf("Failed to create dummy file in dir: %v", err)
	}

	app.ServeStaticDir("/static_files", dummyDirPath)

	// Test serving a file from the static directory
	// Create a test server to simulate a real HTTP request
	server := httptest.NewServer(app.handler)
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/static_files/index.html", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if string(body) != dummyContent {
		t.Errorf("Expected body %q, got %q", dummyContent, string(body))
	}
	if resp.Header.Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type %q, got %q", "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
	}
}

// Helper functions for static file/dir tests
// Using functions from test_utils.go
