package core

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestRequestMethod_toPathString(t *testing.T) {
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
			if tt.method.toPathString() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.method.toPathString())
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
