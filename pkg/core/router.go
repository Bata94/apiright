package core

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
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

// Router is a router for the application.
type Router struct {
	groups []*Router
	routes []*Route

	basePath    string
	middlewares []Middleware
}

// Use adds a middleware to the router.
func (r *Router) Use(m Middleware) {
	r.middlewares = append(r.middlewares, m)
}

// GetBasePath returns the base path of the router.
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

// GET adds a GET endpoint to the router.
func (r *Router) GET(path string, handler Handler, opt ...RouteOption) {
	r.addEndpoint(METHOD_GET, path, handler, opt...)
}

// POST adds a POST endpoint to the router.
func (r *Router) POST(path string, handler Handler, opt ...RouteOption) {
	r.addEndpoint(METHOD_POST, path, handler, opt...)
}

// PUT adds a PUT endpoint to the router.
func (r *Router) PUT(path string, handler Handler, opt ...RouteOption) {
	r.addEndpoint(METHOD_PUT, path, handler, opt...)
}

// DELETE adds a DELETE endpoint to the router.
func (r *Router) DELETE(path string, handler Handler, opt ...RouteOption) {
	r.addEndpoint(METHOD_DELETE, path, handler, opt...)
}

// OPTIONS adds an OPTIONS endpoint to the router.
func (r *Router) OPTIONS(path string, handler Handler, opt ...RouteOption) {
	r.addEndpoint(METHOD_OPTIONS, path, handler, opt...)
}

// TODO: Add ReadMulti
// TODO: Add BulkUpdate, BulkDelete, BulkCreate
// TODO: Add individual endpoint options
type CrudInterface interface {
	CreateFunc(any) (any, error)
	ReadAllFunc() ([]any, error)
	ReadOneFunc(any) (any, error)
	UpdateFunc(any, any) (any, error)
	DeleteFunc(any) (any, error)
}

// Add full CreateReadUpdateDelete Endpoints, for the basePath and given CrudInterface.
// Only adding Endpoints, for defined functions in CrudInterface.
// RouteOptions will be applied to all, for individual Options will be added later.
// POST   {basePath}/{id} -> Create
// GET    {basePath}/     -> ReadAll
// GET    {basePath}/{id}	-> ReadOne
// PUT    {basePath}/{id}	-> Update
// DELETE {basePath}/{id} -> Delete
// func (r *Router) CRUD(basePath string, ci CrudInterface, opt ...RouteOption) {
// 	pathWithID := fmt.Sprintf("%s/{id}", basePath)
//
// 	if ci.CreateFunc != nil {
// 		r.addEndpoint(METHOD_POST, pathWithID, ci.CreateFunc, opt...)
// 	}
// }

// StaticSevFileConfig holds the configuration for serving a static file.
type StaticSevFileConfig struct {
	preCache    bool
	contentType string
}

// StaticServFileOption is a function that configures a StaticSevFileConfig.
type StaticServFileOption func(*StaticSevFileConfig)

// NewStaticServeFileConfig creates a new StaticSevFileConfig.
func NewStaticServeFileConfig(opts ...StaticServFileOption) *StaticSevFileConfig {
	c := &StaticSevFileConfig{
		preCache:    false,
		contentType: "",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithPreCache caches the file content in memory.
func WithPreCache() StaticServFileOption {
	// Read the file content once when the handler is created.
	// This is efficient for files that don't change frequently.
	return func(c *StaticSevFileConfig) {
		c.preCache = true
	}
}

// WithContentType sets the content type of the file.
func WithContentType(contentType string) StaticServFileOption {
	return func(c *StaticSevFileConfig) {
		c.contentType = contentType
	}
}

// ServeStaticFile serves a static file.
func (r *Router) ServeStaticFile(urlPath, filePath string, opt ...StaticServFileOption) error {
	config := NewStaticServeFileConfig(opt...)

	if config.preCache {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			err = fmt.Errorf("static directory '%s' does not exist. Please create it and add your files", filePath)
			log.Error(err)
			return err
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			err = fmt.Errorf("static directory '%s' exists, but is not readable. Please verify permissions", filePath)
			log.Error(err)
			return err
		}

		h := func(c *Ctx) error {
			c.Response.SetStatus(200)
			c.Response.SetData(content)
			c.Response.AddHeader("Content-Type", config.contentType)
			return nil
		}

		r.addEndpoint(
			METHOD_GET,
			urlPath,
			h,
			WithOpenApiDisabled(),
		)
	} else {
		// TODO: Implement this
		return errors.New("not implemented")
	}

	return nil
}

// TODO: Rethink/Implement do use more of the Framework
func (r *Router) ServeStaticDir(urlPath, dirPath string, a App) {
	fs := http.FileServer(http.Dir(dirPath))
	// Ensure the pattern ends with / to avoid conflicts with method-specific patterns
	pattern := urlPath
	if !strings.HasSuffix(pattern, "/") {
		pattern += "/"
	}

	// Use wildcard patterns to match all paths under the directory
	// In Go 1.22+, we need {path...} for wildcard matching
	getPattern := "GET " + pattern + "{path...}"
	headPattern := "HEAD " + pattern + "{path...}"

	h := func(w http.ResponseWriter, r *http.Request) {
		// Strip the URL prefix to match the file system path
		http.StripPrefix(urlPath, fs).ServeHTTP(w, r)
	}

	a.getHttpHandler().HandleFunc(getPattern, h)
	a.getHttpHandler().HandleFunc(headPattern, h)
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
		var endpoints []Endpoint
		if p != "/" {
			optionEP := Endpoint{
				method: METHOD_OPTIONS,
				handleFunc: func(c *Ctx) error {
					c.Response.SetStatus(http.StatusOK)
					return nil
				},
				routeOptionConfig: RouteOptionConfig{},
			}
			endpoints = []Endpoint{optionEP}
		} else {
			endpoints = []Endpoint{}
		}

		r.routes = append(r.routes, &Route{
			basePath:  p,
			path:      fmt.Sprint(r.basePath, p),
			endpoints: endpoints,
		})

		routeIndex = len(r.routes) - 1
	}

	// Prevent adding duplicate OPTIONS endpoint if it already exists
	if m == METHOD_OPTIONS {
		for _, ep := range r.routes[routeIndex].endpoints {
			if ep.method == METHOD_OPTIONS {
				return
			}
		}
	}

	r.routes[routeIndex].endpoints = append(r.routes[routeIndex].endpoints, Endpoint{
		method:            m,
		handleFunc:        h,
		routeOptionConfig: *routeConfig,
		middlewares:       routeConfig.middlewares,
	})
}
