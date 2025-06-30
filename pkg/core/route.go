package core

import (
	"net/http"
	"reflect"
	"time"
)

// Route represents a route in the application.
type Route struct {
	basePath, path string
	endpoints      []Endpoint
}

// RequestMethod is the HTTP request method.
type RequestMethod int

const (
	// METHOD_GET is the GET HTTP method.
	METHOD_GET RequestMethod = iota
	// METHOD_POST is the POST HTTP method.
	METHOD_POST
	// METHOD_PUT is the PUT HTTP method.
	METHOD_PUT
	// METHOD_PATCH is the PATCH HTTP method.
	METHOD_PATCH
	// METHOD_DELETE is the DELETE HTTP method.
	METHOD_DELETE
	// METHOD_HEAD is the HEAD HTTP method.
	METHOD_HEAD
	// METHOD_OPTIONS is the OPTIONS HTTP method.
	METHOD_OPTIONS
	// METHOD_TRACE is the TRACE HTTP method.
	METHOD_TRACE
	// METHOD_CONNECT is the CONNECT HTTP method.
	METHOD_CONNECT
)

var (
	requestMethodPathStrings = []string{
		"GET",
		"POST",
		"PUT",
		"PATCH",
		"DELETE",
		"HEAD",
		"OPTIONS",
		"TRACE",
		"CONNECT",
	}
)

func (m RequestMethod) toPathString() string {
	return requestMethodPathStrings[m]
}

// Endpoint represents an endpoint in the application.
type Endpoint struct {
	method            RequestMethod
	handleFunc        Handler
	middlewares       []Middleware
	routeOptionConfig RouteOptionConfig
}

// Handler is a function that handles a request.
type Handler func(*Ctx) error

// NewCtx creates a new Ctx instance.
func NewCtx(w http.ResponseWriter, r *http.Request) *Ctx {
	return &Ctx{
		Request:  r,
		Response: NewApiResponse(),

		conClosed:  make(chan bool, 1),
		conStarted: time.Now(),
	}
}

// Ctx is the context for a request.
type Ctx struct {
	// TODO: Move to an Interface, prob to use HTML Responses as well
	Response ApiResponse
	Request  *http.Request

	conClosed  chan (bool)
	conStarted time.Time
	conEnded   time.Time

	ObjIn     any
	ObjInType reflect.Type

	ObjOut     any
	ObjOutType reflect.Type
}

// Close closes the connection.
func (c *Ctx) Close() {
	c.conEnded = time.Now()
	c.conClosed <- true
}

// IsClosed returns true if the connection is closed.
func (c *Ctx) IsClosed() bool {
	return <-c.conClosed
}

// RouteOptionConfig holds the configuration for a route.
// TODO: Add this to Router as well and set the Router values as default for Route
type RouteOptionConfig struct {
	openApiEnabled bool
	openApiConfig  struct {
		summary, description string
		tags                 []string
		deprecated           bool
	}

	ObjIn       any
	ObjOut      any
	middlewares []Middleware
}

// RouteOption is a function that configures a RouteOptionConfig.
type RouteOption func(*RouteOptionConfig)

// NewRouteOptionConfig creates a new RouteOptionConfig.
func NewRouteOptionConfig(opts ...RouteOption) *RouteOptionConfig {
	// TODO: Make default settable in AppConfig and pass through
	config := &RouteOptionConfig{
		openApiEnabled: true,
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

// Use adds a middleware to the route.
func Use(m Middleware) RouteOption {
	return func(c *RouteOptionConfig) {
		c.middlewares = append(c.middlewares, m)
	}
}

// WithObjIn sets the input object for the route.
func WithObjIn(obj any) RouteOption {
	return func(c *RouteOptionConfig) {
		c.ObjIn = obj
	}
}

// WithObjOut sets the output object for the route.
func WithObjOut(obj any) RouteOption {
	return func(c *RouteOptionConfig) {
		c.ObjOut = obj
	}
}

// WithOpenApiDisabled disables OpenAPI generation for the route.
func WithOpenApiDisabled() RouteOption {
	return func(c *RouteOptionConfig) {
		c.openApiEnabled = false
	}
}

// WithOpenApiEnabled enables OpenAPI generation for the route.
func WithOpenApiEnabled(summary, description string) RouteOption {
	return func(c *RouteOptionConfig) {
		c.openApiEnabled = true
		c.openApiConfig.summary = summary
		c.openApiConfig.description = description
	}
}

// WithOpenApiInfos sets the OpenAPI summary and description for the route.
func WithOpenApiInfos(summary, description string) RouteOption {
	return func(c *RouteOptionConfig) {
		c.openApiConfig.summary = summary
		c.openApiConfig.description = description
	}
}

// WithOpenApiDeprecated marks the route as deprecated in OpenAPI.
func WithOpenApiDeprecated() RouteOption {
	return func(c *RouteOptionConfig) {
		c.openApiConfig.deprecated = true
	}
}

// WithOpenApiTags adds tags to the route in OpenAPI.
func WithOpenApiTags(tags ...string) RouteOption {
	return func(c *RouteOptionConfig) {
		c.openApiConfig.tags = tags
	}
}
