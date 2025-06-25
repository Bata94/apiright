package core

import (
	"fmt"
	"net/http"
)

var defCatchallHandler = func(c *Ctx) error {
	log.Info("Default CatchAll Handler")
	c.Response.SetStatus(404)
	c.Response.Message = "Not found!"
	return nil
}

func newRouter(path string) *Router {
	return &Router{
		basePath: path,
	}
}

type Router struct {
	groups []*Router
	routes []*Route

	basePath string
}

func (r Router) GetBasePath() string {
	if len(r.basePath) > 1 {
		if string(r.basePath[len(r.basePath)-1]) != "/" {
			return fmt.Sprintf("%s/", r.basePath)
		}
	} else if r.basePath == "" {
		return "/"
	}
	return r.basePath
}

func (r *Router) GET(path string, handler Handler, opt ...RouteOption) {
	r.addEndpoint(METHOD_GET, path, handler, opt...)
}

func (r *Router) POST(path string, handler Handler, opt ...RouteOption) {
	r.addEndpoint(METHOD_POST, path, handler, opt...)
}

func (r *Router) PUT(path string, handler Handler, opt ...RouteOption) {
	r.addEndpoint(METHOD_PUT, path, handler, opt...)
}

func (r *Router) DELETE(path string, handler Handler, opt ...RouteOption) {
	r.addEndpoint(METHOD_DELETE, path, handler, opt...)
}

func (r *Router) OPTIONS(path string, handler Handler, opt ...RouteOption) {
	r.addEndpoint(METHOD_OPTIONS, path, handler, opt...)
}

func (r *Router) routeExists(path string) int {
	// Checks if route exists and returns the index. If false -1 is returned.
	for i, route := range r.routes {
		if route.path == path {
			return i
		}
	}

	return -1
}

func (r *Router) addEndpoint(m RequestMethod, p string, h Handler, opt ...RouteOption) {
	routeConfig := NewRouteOptionConfig(opt...)
	routeIndex := r.routeExists(p)

	if routeIndex == -1 {
		optionEP := Endpoint{
			method:            METHOD_OPTIONS,
			handleFunc:        func(c *Ctx) error {
				c.Response.SetStatus(http.StatusOK)
				return nil
			},
			routeOptionConfig: RouteOptionConfig{},
		}
		r.routes = append(r.routes, &Route{
			basePath:  p,
			path:      fmt.Sprint(r.basePath, p),
			endpoints: []Endpoint{
				optionEP,
			},
		})

		routeIndex = len(r.routes) - 1
	}

	r.routes[routeIndex].endpoints = append(r.routes[routeIndex].endpoints, Endpoint{
		method:            m,
		handleFunc:        h,
		routeOptionConfig: *routeConfig,
	})
}
