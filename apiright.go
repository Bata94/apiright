package apiright

import (
	"fmt"
	"net/http"
)

func NewAppConfig() AppConfig {
	return AppConfig{
		host: "127.0.0.1",
		port: "5500",
	}
}

type AppConfig struct {
	host, port string
}

func (c AppConfig) GetListenAddress() string {
	return fmt.Sprintf("%s:%s", c.host, c.port)
}

var defCatchallHandler = func(c *Ctx) error {
	Debug("Default CatchAll Handler")
	c.Writer.Write([]byte("Not found!"))
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

func (r *Router) GET(path string, handler Handler) {
	r.addEndpoint(METHOD_GET, path, handler)
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

func (r *Router) addEndpoint(m RequestMethod, p string, h Handler) {
	routeIndex := r.routeExists(p)

	if routeIndex == -1 {
		r.routes = append(r.routes, &Route{
			basePath:  p,
			path:      fmt.Sprint(r.basePath, p),
			endpoints: []Endpoint{},
		})

		routeIndex = len(r.routes) - 1
	}

	r.routes[routeIndex].endpoints = append(r.routes[routeIndex].endpoints, Endpoint{
		method:     m,
		handleFunc: h,
	})
}

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

type Middleware func(Handler) Handler

type Response struct{}

func NewCtx(w http.ResponseWriter, r *http.Request) *Ctx {
	return &Ctx{
		Writer:  w,
		Request: r,
	}
}

type Ctx struct {
	Writer  http.ResponseWriter
	Request *http.Request
}

func NewApp() App {
	handler := http.NewServeMux()
	config := NewAppConfig()

	return App{
		Config:          &config,
		handler:         handler,
		Logger:          GetLogger(),
		router:          newRouter(""),
		defRouteHandler: defCatchallHandler,
	}
}

type App struct {
	Config  *AppConfig
	handler *http.ServeMux
	Logger  Logger

	defRouteHandler Handler
	router          *Router
}

func (a *App) SetDefaultRoute(handler Handler) {
	a.defRouteHandler = handler
}

func (a *App) SetLogger(logger Logger) {
	a.Logger = logger
}

func (a *App) NewRouter(path string) *Router {
	// Creates and adds a new Router, with a BasePath
	newRouter := newRouter(path)

	a.router.groups = append(a.router.groups, newRouter)

	return newRouter
}

func (a App) getHttpHandler() *http.ServeMux {
	return a.handler
}

func (a *App) GET(path string, handler func(*Ctx) error) {
	a.router.addEndpoint(METHOD_GET, path, handler)
}

// TODO: Prob move into a Middleware
func recoverFromPanic(w http.ResponseWriter, logger Logger) {
	if r := recover(); r != nil {
		logger.Errorf("Recovered from panic: %v", r)
		w.Write([]byte("Internal Server Error"))
	}
}

func (a *App) handleFunc(route Route, endPoint Endpoint, router Router) {
	handlerPath := fmt.Sprintf("%s %s", endPoint.method.toPathString(), route.path)
	a.Logger.Debugf("Registering route: %s", handlerPath)

	a.getHttpHandler().HandleFunc(handlerPath, func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w, a.Logger)
		h := endPoint.handleFunc

		if route.basePath == "/" && r.URL.Path != "/" {
			a.Logger.Debugf("Using default route handler for path: %s", r.URL.Path)
			h = a.defRouteHandler
		}

		c := NewCtx(w, r)
		err := h(c)

		if err != nil {
			err = fmt.Errorf("Error in HanlderFunc: %w", err)
			a.Logger.Errorf("Handler error: %v", err)
			w.Write([]byte(err.Error()))
		}
	})
}

func (a App) addRoutesToHandler() {
	a.Logger.Info("Registering available routes")

	a.Logger.Infof("Global Router with %d routes", len(a.router.routes))
	for _, r := range a.router.routes {
		for _, e := range r.endpoints {
			a.handleFunc(*r, e, *a.router)
		}
	}

	for _, group := range a.router.groups {
		a.Logger.Infof("Router group with %d routes", len(group.routes))
		for _, r := range group.routes {
			for _, e := range r.endpoints {
				a.handleFunc(*r, e, *group)
			}
		}
	}
}

func (a App) Run() error {
	a.addRoutesToHandler()

	return http.ListenAndServe(a.Config.GetListenAddress(), a.getHttpHandler())
}
