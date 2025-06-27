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

type StaticSevFileConfig struct {
	preCache bool
	contentType string
}

type StaticServFileOption func(*StaticSevFileConfig)

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

func WithPreCache() StaticServFileOption {
	// Read the file content once when the handler is created.
	// This is efficient for files that don't change frequently.
  return func(c *StaticSevFileConfig) {
    c.preCache = true
  }
}

func WithContentType(contentType string) StaticServFileOption {
  return func(c *StaticSevFileConfig) {
    c.contentType = contentType
  }
}

func (r *Router) ServeStaticFile(urlPath, filePath string, opt ...StaticServFileOption) error {
	config := NewStaticServeFileConfig(opt...)

	if config.preCache {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			err = errors.New(fmt.Sprintf("static directory '%s' does not exist. Please create it and add your files.", filePath))
			log.Error(err)
			return err
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			err = errors.New(fmt.Sprintf("static directory '%s' exists, but is not readable. Please verify permissions.", filePath))
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
	
	a.getHttpHandler().HandleFunc(getPattern, func(w http.ResponseWriter, r *http.Request) {
		// Strip the URL prefix to match the file system path
		http.StripPrefix(urlPath, fs).ServeHTTP(w, r)
	})
	a.getHttpHandler().HandleFunc(headPattern, func(w http.ResponseWriter, r *http.Request) {
		// Strip the URL prefix to match the file system path
		http.StripPrefix(urlPath, fs).ServeHTTP(w, r)
	})
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
			basePath: p,
			path:     fmt.Sprint(r.basePath, p),
			endpoints: endpoints,
		})

		routeIndex = len(r.routes) - 1
	}

	r.routes[routeIndex].endpoints = append(r.routes[routeIndex].endpoints, Endpoint{
		method:            m,
		handleFunc:        h,
		routeOptionConfig: *routeConfig,
	})
}
