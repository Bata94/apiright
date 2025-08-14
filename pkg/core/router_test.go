package core

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNewRouter(t *testing.T) {
	router := newRouter("/api")
	if router.basePath != "/api" {
		t.Errorf("Expected base path /api, got %s", router.basePath)
	}
}

func TestRouter_Use(t *testing.T) {
	router := newRouter("")
	mockMiddleware := func(next Handler) Handler {
		return func(c *Ctx) error {
			return next(c)
		}
	}
	router.Use(mockMiddleware)

	if len(router.middlewares) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(router.middlewares))
	}
}

func TestRouter_GetBasePath(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		expected string
	}{
		{
			name:     "Empty Base Path",
			basePath: "",
			expected: "/",
		},
		{
			name:     "Root Base Path",
			basePath: "/",
			expected: "/",
		},
		{
			name:     "Simple Base Path",
			basePath: "/api",
			expected: "/api/",
		},
		{
			name:     "Base Path with Trailing Slash",
			basePath: "/api/v1/",
			expected: "/api/v1/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newRouter(tt.basePath)
			if router.GetBasePath() != tt.expected {
				t.Errorf("Expected base path %s, got %s", tt.expected, router.GetBasePath())
			}
		})
	}
}

func TestRouter_AddEndpoint(t *testing.T) {
	router := newRouter("")
	mockHandler := func(c *Ctx) error { return nil }

	router.addEndpoint(METHOD_GET, "/test", mockHandler)

	if len(router.routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(router.routes))
	}
	if router.routes[0].path != "/test" {
		t.Errorf("Expected route path /test, got %s", router.routes[0].path)
	}
	if len(router.routes[0].endpoints) != 2 { // GET + OPTIONS
		t.Errorf("Expected 2 endpoints, got %d", len(router.routes[0].endpoints))
	}

	router.addEndpoint(METHOD_POST, "/test", mockHandler)
	// Should still be 1 route, but now 3 endpoints
	if len(router.routes) != 1 {
		t.Errorf("Expected 1 route after adding POST, got %d", len(router.routes))
	}
	if len(router.routes[0].endpoints) != 3 { // GET + OPTIONS + POST
		t.Errorf("Expected 3 endpoints after adding POST, got %d", len(router.routes[0].endpoints))
	}
}

func TestRouter_RoutingMethods(t *testing.T) {
	router := newRouter("")
	mockHandler := func(c *Ctx) error { return nil }

	router.GET("/get", mockHandler)
	router.POST("/post", mockHandler)
	router.PUT("/put", mockHandler)
	router.DELETE("/delete", mockHandler)
	router.OPTIONS("/options", mockHandler)

	tests := []struct {
		method            RequestMethod
		path              string
		expectedEndpoints int
	}{
		{METHOD_GET, "/get", 2},         // GET + OPTIONS
		{METHOD_POST, "/post", 2},       // POST + OPTIONS
		{METHOD_PUT, "/put", 2},         // PUT + OPTIONS
		{METHOD_DELETE, "/delete", 2},   // DELETE + OPTIONS
		{METHOD_OPTIONS, "/options", 1}, // OPTIONS only, as addEndpoint doesn't add another OPTIONS for /
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			found := false
			for _, route := range router.routes {
				if route.path == tt.path {
					found = true
					if len(route.endpoints) != tt.expectedEndpoints {
						t.Errorf("Path %s: Expected %d endpoints, got %d", tt.path, tt.expectedEndpoints, len(route.endpoints))
					}
					break
				}
			}
			if !found {
				t.Errorf("Route %s not found", tt.path)
			}
		})
	}
}

func TestRouter_ServeStaticFile(t *testing.T) {
	router := newRouter("")
	// Create a dummy file for testing
	dummyFilePath := "/tmp/router_dummy.txt"
	dummyContent := "Hello from router static file!"
	err := WriteDummyFile(dummyFilePath, dummyContent)
	if err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}
	defer func() { _ = RemoveDummyFile(dummyFilePath) }()

	router.ServeStaticFile("/static/router_file.txt", dummyFilePath, WithContentType("text/plain"))

	// Verify that a route was added for the static file
	found := false
	var staticRoute *Route
	for _, route := range router.routes {
		if route.path == "/static/router_file.txt" {
			found = true
			staticRoute = route
			break
		}
	}
	if !found {
		t.Errorf("ServeStaticFile did not add the expected route")
	}

	// Test serving the file
	req := httptest.NewRequest(http.MethodGet, "/static/router_file.txt", nil)
	rec := httptest.NewRecorder()

	// Manually call handleFunc for the static file route
	if staticRoute != nil && len(staticRoute.endpoints) > 0 {
		// Find the GET endpoint
		var getEndpoint *Endpoint
		for i := range staticRoute.endpoints {
			if staticRoute.endpoints[i].method == METHOD_GET {
				getEndpoint = &staticRoute.endpoints[i]
				break
			}
		}

		if getEndpoint != nil {
			ctx := NewCtx(rec, req, *staticRoute, *getEndpoint)
			err := getEndpoint.handleFunc(ctx)
			if err != nil {
				t.Fatalf("Handler returned an error: %v", err)
			}
			ctx.SendingReturn(rec, nil)
		} else {
			t.Fatalf("GET endpoint not found for static file route")
		}
	} else {
		t.Fatalf("No endpoints found for static file route")
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

	// Test case for WithoutPreLoad()
	t.Run("WithoutPreLoad", func(t *testing.T) {
		router := newRouter("") // New router for this sub-test

		// Create a temporary directory and a dummy file for testing static serving
		tempDir, err := os.MkdirTemp("", "static_test_without_preload")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(tempDir) }()

		filePath := filepath.Join(tempDir, "test_without_preload.txt")
		fileContent := "This is a test static file without preload."
		err = os.WriteFile(filePath, []byte(fileContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write dummy file: %v", err)
		}

		router.ServeStaticFile("/static/without_preload.txt", filePath, WithContentType("text/plain"), WithoutPreLoad())

		// Verify that a route was added for the static file
		found := false
		var staticRoute *Route
		for _, route := range router.routes {
			if route.path == "/static/without_preload.txt" {
				found = true
				staticRoute = route
				break
			}
		}
		if !found {
			t.Errorf("ServeStaticFile did not add the expected route for WithoutPreLoad")
		}

		// Test serving the file
		req := httptest.NewRequest(http.MethodGet, "/static/without_preload.txt", nil)
		rec := httptest.NewRecorder()

		// Manually call handleFunc for the static file route
		if staticRoute != nil && len(staticRoute.endpoints) > 0 {
			var getEndpoint *Endpoint
			for i := range staticRoute.endpoints {
				if staticRoute.endpoints[i].method == METHOD_GET {
					getEndpoint = &staticRoute.endpoints[i]
					break
				}
			}

			if getEndpoint != nil {
				ctx := NewCtx(rec, req, *staticRoute, *getEndpoint)
				err := getEndpoint.handleFunc(ctx)
				if err != nil {
					t.Fatalf("Handler returned an error: %v", err)
				}
				ctx.SendingReturn(rec, nil)
			} else {
				t.Fatalf("GET endpoint not found for static file route for WithoutPreLoad")
			}
		} else {
			t.Fatalf("No endpoints found for static file route for WithoutPreLoad")
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d for WithoutPreLoad", http.StatusOK, rec.Code)
		}
		if rec.Body.String() != fileContent {
			t.Errorf("Expected body %q, got %q for WithoutPreLoad", fileContent, rec.Body.String())
		}
		if rec.Header().Get("Content-Type") != "text/plain" {
			t.Errorf("Expected Content-Type %q, got %q for WithoutPreLoad", "text/plain", rec.Header().Get("Content-Type"))
		}
	})
}

// func TestRouter_ServeStaticFile_NotImplemented(t *testing.T) {
// 	router := newRouter("")
// 	err := router.ServeStaticFile("/static/not_implemented.txt", "/tmp/non_existent.txt")
// 	expectedError := "static file '/tmp/non_existent.txt' does not exist. Please ensure the file exists"
// 	if err == nil || err.Error() != expectedError {
// 		t.Errorf("Expected error %q, got %v", expectedError, err)
// 	}
// }

// Helper functions (copied from app_test.go, consider moving to a common test utility file)
// Using functions from test_utils.go
