package core

import (
	"fmt"
	"net/http"

	"github.com/bata94/apiright/v0/utils"
)

type RouteOpts struct {
	Domain string
}

func getRouteOpts(opts ...RouteOpts) RouteOpts {
	if len(opts) == 0 {
		return RouteOpts{}
	}

	// Convert RouteOpts to interface{} for getOpts
	var interfaceOpts []interface{}
	for _, opt := range opts {
		interfaceOpts = append(interfaceOpts, opt)
	}

	o := utils.GetOpts(interfaceOpts...)

	// Type assertion to check if the options are of type RouteOpts
	if routeOpts, ok := o.(RouteOpts); ok {
		return routeOpts
	} else {
		panic("invalid options for RouteOpts")
	}
}

func (a *App) Get(path string, handler func(w http.ResponseWriter, r *http.Request), opts ...RouteOpts) {
	o := getRouteOpts(opts...)

	a.routes = append(a.routes, &Route{
		domain:    o.Domain,
		path:      path,
		endpoints: []Endpoint{{GET, handler}},
	})
}

func (a *App) POST(path string, handler func(w http.ResponseWriter, r *http.Request), opts ...RouteOpts) {
	o := getRouteOpts(opts...)

	a.routes = append(a.routes, &Route{
		domain:    o.Domain,
		path:      path,
		endpoints: []Endpoint{{POST, handler}},
	})
}

func (a *App) PUT(path string, handler func(w http.ResponseWriter, r *http.Request), opts ...RouteOpts) {
	o := getRouteOpts(opts...)

	a.routes = append(a.routes, &Route{
		domain:    o.Domain,
		path:      path,
		endpoints: []Endpoint{{PUT, handler}},
	})
}

func (a *App) DELETE(path string, handler func(w http.ResponseWriter, r *http.Request), opts ...RouteOpts) {
	o := getRouteOpts(opts...)

	a.routes = append(a.routes, &Route{
		domain:    o.Domain,
		path:      path,
		endpoints: []Endpoint{{DELETE, handler}},
	})
}

type Route struct {
	domain    string
	path      string
	endpoints []Endpoint
}

type Endpoint struct {
	method  Method
	handler func(w http.ResponseWriter, r *http.Request)
	// middlewares
}

func (e Endpoint) PathString(r Route) string {
	return fmt.Sprintf("%s %s%s", e.method, r.domain, r.path)
}

type Method int

const (
	GET Method = iota
	POST
	PUT
	DELETE
)

func (m Method) String() string {
	return [...]string{"GET", "POST", "PUT", "DELETE"}[m]
}

func (m Method) EnumIndex() int {
	return int(m)
}
