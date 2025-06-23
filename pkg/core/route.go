package core

import (
	"fmt"
	"net/http"
)

type Route struct {
	basePath, path string
	endpoints      []Endpoint
}

func (r Route) fullPath(m RequestMethod) string {
	// Returns the "fullPath" for the net/http Handler input
	return fmt.Sprintf("%s %s", m.toPathString(), r.path)
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
	method     RequestMethod
	handleFunc Handler
}

type Handler func(*Ctx) error

func NewCtx(w http.ResponseWriter, r *http.Request) *Ctx {
	return &Ctx{
		Request:  r,
		Response: NewApiResponse(),
	}
}

type Ctx struct {
	// TODO: Move to an Interface, prob to use HTML Responses as well
	Response ApiResponse
	Request  *http.Request
}

type RouteOptionConfig struct {
	ObjIn  any
	ObjOut any
}

type RouteOption func(*RouteOptionConfig)

func NewRouteOptionConfig(opts ...RouteOption) *RouteOptionConfig {
	config := &RouteOptionConfig{}

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
