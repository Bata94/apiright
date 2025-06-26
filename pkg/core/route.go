package core

import (
	"net/http"
	"reflect"
	"time"
)

type Route struct {
	basePath, path string
	endpoints      []Endpoint
}

type RequestMethod int

const (
	METHOD_GET RequestMethod = iota
	METHOD_POST
	METHOD_PUT
	METHOD_PATCH
	METHOD_DELETE
	METHOD_HEAD
	METHOD_OPTIONS
	METHOD_TRACE
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

type Endpoint struct {
	method            RequestMethod
	handleFunc        Handler
	routeOptionConfig RouteOptionConfig
}

type Handler func(*Ctx) error

func NewCtx(w http.ResponseWriter, r *http.Request) *Ctx {
	return &Ctx{
		Request:  r,
		Response: NewApiResponse(),

		conClosed:  make(chan bool),
		conStarted: time.Now(),
	}
}

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

func (c *Ctx) Close() {
	c.conEnded = time.Now()
	c.conClosed <- true
}

func (c *Ctx) IsClosed() bool {
	return <-c.conClosed
}

// TODO: Add this to Router as well and set the Router values as default for Route
type RouteOptionConfig struct {
	openApiEnabled bool
	openApiConfig  struct {
		summary, description string
		tags                 []string
		deprecated           bool
	}

	ObjIn  any
	ObjOut any
}

type RouteOption func(*RouteOptionConfig)

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

func WithObjIn(obj any) RouteOption {
	return func(c *RouteOptionConfig) {
		c.ObjIn = obj
	}
}

func WithObjOut(obj any) RouteOption {
	return func(c *RouteOptionConfig) {
		c.ObjOut = obj
	}
}

func WithOpenApiDisabled() RouteOption {
	return func(c *RouteOptionConfig) {
		c.openApiEnabled = false
	}
}

func WithOpenApiEnabled(summary, description string) RouteOption {
	return func(c *RouteOptionConfig) {
		c.openApiEnabled = true
		c.openApiConfig.summary = summary
		c.openApiConfig.description = description
	}
}

func WithOpenApiInfos(summary, description string) RouteOption {
	return func(c *RouteOptionConfig) {
		c.openApiConfig.summary = summary
		c.openApiConfig.description = description
	}
}

func WithOpenApiDeprecated() RouteOption {
	return func(c *RouteOptionConfig) {
		c.openApiConfig.deprecated = true
	}
}

func WithOpenApiTags(tags ...string) RouteOption {
	return func(c *RouteOptionConfig) {
		c.openApiConfig.tags = tags
	}
}
