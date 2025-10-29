package core

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestRequestMethod_ToPathString(t *testing.T) {
	tests := []struct {
		method   RequestMethod
		expected string
	}{
		{METHOD_GET, "GET"},
		{METHOD_POST, "POST"},
		{METHOD_PUT, "PUT"},
		{METHOD_PATCH, "PATCH"},
		{METHOD_DELETE, "DELETE"},
		{METHOD_HEAD, "HEAD"},
		{METHOD_OPTIONS, "OPTIONS"},
		{METHOD_TRACE, "TRACE"},
		{METHOD_CONNECT, "CONNECT"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.method.ToPathString() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.method.ToPathString())
			}
		})
	}
}

func TestNewCtx(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	route := Route{}
	ep := Endpoint{}
	ctx := NewCtx(rec, req, route, ep)

	if ctx.Request != req {
		t.Errorf("Expected request to be %v, got %v", req, ctx.Request)
	}
	if ctx.Response == nil {
		t.Error("Expected response to be initialized, got nil")
	}
	if ctx.conClosed == nil {
		t.Error("Expected conClosed channel to be initialized, got nil")
	}
	if ctx.conStarted.IsZero() {
		t.Error("Expected conStarted to be set, got zero time")
	}
}

func TestCtx_Close(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	route := Route{}
	ep := Endpoint{}
	ctx := NewCtx(rec, req, route, ep)

	ctx.Close()

	if ctx.conEnded.IsZero() {
		t.Error("Expected conEnded to be set, got zero time")
	}
	select {
	case <-ctx.conClosed:
		// Expected to receive from channel
	default:
		t.Error("Expected conClosed channel to be closed")
	}
}

func TestCtx_IsClosed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	route := Route{}
	ep := Endpoint{}
	ctx := NewCtx(rec, req, route, ep)

	go func() {
		time.Sleep(10 * time.Millisecond)
		ctx.Close()
	}()

	if !ctx.IsClosed() {
		t.Error("Expected IsClosed to return true")
	}
}

func TestCtx_ConClosed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	route := Route{}
	ep := Endpoint{}
	ctx := NewCtx(rec, req, route, ep)

	select {
	case <-ctx.ConClosed():
		t.Error("Expected ConClosed channel to not be closed initially")
	default:
		// Expected
	}

	ctx.Close()

	select {
	case <-ctx.ConClosed():
		// Expected to receive
	default:
		t.Error("Expected ConClosed channel to be closed after Close()")
	}
}

func TestCtx_GetConnectionDuration(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	route := Route{}
	ep := Endpoint{}
	ctx := NewCtx(rec, req, route, ep)

	// Test before closing
	duration := ctx.GetConnectionDuration()
	if duration <= 0 {
		t.Error("Expected positive duration")
	}

	// Sleep a bit
	time.Sleep(5 * time.Millisecond)

	// Close and test
	ctx.Close()
	duration = ctx.GetConnectionDuration()
	if duration <= 0 {
		t.Error("Expected positive duration after close")
	}
}

func TestRoute_GetPath(t *testing.T) {
	route := &Route{path: "/test"}
	if route.GetPath() != "/test" {
		t.Errorf("Expected path '/test', got %s", route.GetPath())
	}
}

func TestRoute_GetEndpoints(t *testing.T) {
	endpoints := []Endpoint{{method: METHOD_GET}}
	route := &Route{endpoints: endpoints}
	if len(route.GetEndpoints()) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(route.GetEndpoints()))
	}
	if route.GetEndpoints()[0].method != METHOD_GET {
		t.Errorf("Expected method GET, got %v", route.GetEndpoints()[0].method)
	}
}

func TestEndpoint_GetMethod(t *testing.T) {
	endpoint := Endpoint{method: METHOD_POST}
	if endpoint.GetMethod() != METHOD_POST {
		t.Errorf("Expected method POST, got %v", endpoint.GetMethod())
	}
}

func TestEndpoint_GetRouteOptionConfig(t *testing.T) {
	config := RouteOptionConfig{openApiEnabled: true}
	endpoint := Endpoint{routeOptionConfig: config}
	if !endpoint.GetRouteOptionConfig().openApiEnabled {
		t.Error("Expected openApiEnabled to be true")
	}
}

func TestRouteOptionConfig_GetOpenApiEnabled(t *testing.T) {
	config := RouteOptionConfig{openApiEnabled: true}
	if !config.GetOpenApiEnabled() {
		t.Error("Expected openApiEnabled to be true")
	}
}

func TestRouteOptionConfig_GetOpenApiConfig(t *testing.T) {
	config := RouteOptionConfig{
		openApiEnabled: true,
		openApiConfig: struct {
			summary, description string
			tags                 []string
			deprecated           bool
			jwtAuth              bool
		}{
			summary:     "Test Summary",
			description: "Test Description",
			tags:        []string{"test"},
			deprecated:  true,
			jwtAuth:     true,
		},
	}

	result := config.GetOpenApiConfig()
	if result.Summary != "Test Summary" {
		t.Errorf("Expected Summary 'Test Summary', got %s", result.Summary)
	}
	if result.Description != "Test Description" {
		t.Errorf("Expected Description 'Test Description', got %s", result.Description)
	}
	if len(result.Tags) != 1 || result.Tags[0] != "test" {
		t.Errorf("Expected Tags ['test'], got %v", result.Tags)
	}
	if !result.Deprecated {
		t.Error("Expected Deprecated to be true")
	}
	if !result.JwtAuth {
		t.Error("Expected JwtAuth to be true")
	}
}

func TestRouteOptionConfig_GetQueryParams(t *testing.T) {
	queryParams := []struct {
		Name, Description string
		Required          bool
		Type              reflect.Type
		Example           any
	}{
		{Name: "id", Description: "User ID", Required: true, Type: reflect.TypeOf(1), Example: 123},
	}
	config := RouteOptionConfig{queryParams: queryParams}

	result := config.GetQueryParams()
	if len(result) != 1 {
		t.Errorf("Expected 1 query param, got %d", len(result))
	}
	if result[0].Name != "id" {
		t.Errorf("Expected name 'id', got %s", result[0].Name)
	}
}

func TestNewRouteOptionConfig(t *testing.T) {
	tests := []struct {
		name string
		opts []RouteOption
		want RouteOptionConfig
	}{
		{
			name: "Default Config",
			opts: []RouteOption{},
			want: RouteOptionConfig{
				openApiEnabled: true,
			},
		},
		{
			name: "WithObjIn",
			opts: []RouteOption{WithObjIn(struct{ Name string }{}), WithOpenApiDisabled()},
			want: RouteOptionConfig{
				ObjIn:          struct{ Name string }{},
				openApiEnabled: false,
			},
		},
		{
			name: "WithObjOut",
			opts: []RouteOption{WithObjOut(struct{ ID int }{}), WithOpenApiDisabled()},
			want: RouteOptionConfig{
				ObjOut:         struct{ ID int }{},
				openApiEnabled: false,
			},
		},
		{
			name: "WithOpenApiDisabled",
			opts: []RouteOption{WithOpenApiDisabled()},
			want: RouteOptionConfig{
				openApiEnabled: false,
			},
		},
		{
			name: "WithOpenApiEnabled",
			opts: []RouteOption{WithOpenApiEnabled("Summary", "Description")},
			want: RouteOptionConfig{
				openApiEnabled: true,
				openApiConfig: struct {
					summary, description string
					tags                 []string
					deprecated           bool
					jwtAuth              bool
				}{
					summary:     "Summary",
					description: "Description",
				},
			},
		},
		{
			name: "WithOpenApiInfos",
			opts: []RouteOption{WithOpenApiInfos("New Summary", "New Description")},
			want: RouteOptionConfig{
				openApiEnabled: true,
				openApiConfig: struct {
					summary, description string
					tags                 []string
					deprecated           bool
					jwtAuth              bool
				}{
					summary:     "New Summary",
					description: "New Description",
				},
			},
		},
		{
			name: "WithOpenApiDeprecated",
			opts: []RouteOption{WithOpenApiDeprecated()},
			want: RouteOptionConfig{
				openApiEnabled: true,
				openApiConfig: struct {
					summary, description string
					tags                 []string
					deprecated           bool
					jwtAuth              bool
				}{
					deprecated: true,
				},
			},
		},
		{
			name: "WithOpenApiTags",
			opts: []RouteOption{WithOpenApiTags("tag1", "tag2")},
			want: RouteOptionConfig{
				openApiEnabled: true,
				openApiConfig: struct {
					summary, description string
					tags                 []string
					deprecated           bool
					jwtAuth              bool
				}{
					tags: []string{"tag1", "tag2"},
				},
			},
		},
		{
			name: "Use Middleware",
			opts: []RouteOption{Use(func(next Handler) Handler { return next })},
			want: RouteOptionConfig{
				openApiEnabled: true,
				middlewares:    []Middleware{func(next Handler) Handler { return next }},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewRouteOptionConfig(tt.opts...)

			if config.openApiEnabled != tt.want.openApiEnabled {
				t.Errorf("openApiEnabled: got %v, want %v", config.openApiEnabled, tt.want.openApiEnabled)
			}
			if config.openApiConfig.summary != tt.want.openApiConfig.summary {
				t.Errorf("summary: got %q, want %q", config.openApiConfig.summary, tt.want.openApiConfig.summary)
			}
			if config.openApiConfig.description != tt.want.openApiConfig.description {
				t.Errorf("description: got %q, want %q", config.openApiConfig.description, tt.want.openApiConfig.description)
			}
			if !reflect.DeepEqual(config.openApiConfig.tags, tt.want.openApiConfig.tags) {
				t.Errorf("tags: got %v, want %v", config.openApiConfig.tags, tt.want.openApiConfig.tags)
			}
			if config.openApiConfig.deprecated != tt.want.openApiConfig.deprecated {
				t.Errorf("deprecated: got %v, want %v", config.openApiConfig.deprecated, tt.want.openApiConfig.deprecated)
			}
			if (config.ObjIn == nil && tt.want.ObjIn != nil) || (config.ObjIn != nil && tt.want.ObjIn == nil) {
				t.Errorf("ObjIn: got %v, want %v", config.ObjIn, tt.want.ObjIn)
			}
			if (config.ObjOut == nil && tt.want.ObjOut != nil) || (config.ObjOut != nil && tt.want.ObjOut == nil) {
				t.Errorf("ObjOut: got %v, want %v", config.ObjOut, tt.want.ObjOut)
			}
			if len(config.middlewares) != len(tt.want.middlewares) {
				t.Errorf("middlewares length: got %v, want %v", len(config.middlewares), len(tt.want.middlewares))
			}
		})
	}
}
