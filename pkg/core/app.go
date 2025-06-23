package core

import (
	"fmt"
	"net/http"

	"github.com/bata94/apiright/pkg/logger"
)

var (
	log logger.Logger
)

func NewAppConfig() AppConfig {
	// TODO: make setable
	log = logger.NewDefaultLogger()
	log.SetLevel(logger.TraceLevel)

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

func NewApp() App {
	handler := http.NewServeMux()
	config := NewAppConfig()

	return App{
		Config:          &config,
		handler:         handler,
		Logger:          logger.GetLogger(),
		router:          newRouter(""),
		defRouteHandler: defCatchallHandler,
	}
}

type App struct {
	Config  *AppConfig
	handler *http.ServeMux
	Logger  logger.Logger

	defRouteHandler Handler
	router          *Router
}

func (a *App) SetDefaultRoute(handler Handler) {
	a.defRouteHandler = handler
}

func (a *App) SetLogger(logger logger.Logger) {
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

func (a *App) GET(path string, handler Handler, opt ...RouteOption) {
	a.router.addEndpoint(METHOD_GET, path, handler, opt...)
}

func (a App) POST(path string, handler Handler, opt ...RouteOption) {
	a.router.addEndpoint(METHOD_POST, path, handler, opt...)
}

func (a App) PUT(path string, handler Handler, opt ...RouteOption) {
	a.router.addEndpoint(METHOD_PUT, path, handler, opt...)
}

func (a App) DELETE(path string, handler Handler, opt ...RouteOption) {
	a.router.addEndpoint(METHOD_DELETE, path, handler, opt...)
}

// TODO: Prob move into a Middleware and use Ctx
func recoverFromPanic(w http.ResponseWriter, logger logger.Logger) {
	if r := recover(); r != nil {
		logger.Errorf("Recovered from panic: %v", r)
		w.WriteHeader(500)
		w.Write([]byte("Internal Server Error"))
	}
}

func (a *App) handleFunc(route Route, endPoint Endpoint, router Router) {
	handlerPath := fmt.Sprintf("%s %s", endPoint.method.toPathString(), route.path)
	a.Logger.Debugf("Registering route: %s", handlerPath)

	a.getHttpHandler().HandleFunc(handlerPath, func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w, a.Logger)
		h := endPoint.handleFunc

		log.Debugf("route BasePath: %s, r.URL.Path: %s, router.BasePath: %s", route.basePath, r.URL.Path, router.GetBasePath())
		if route.basePath == "/" && r.URL.Path != router.GetBasePath() {
			a.Logger.Debugf("Using default route handler for path: %s", r.URL.Path)
			h = a.defRouteHandler
		}

		c := NewCtx(w, r)
		err := h(c)

		c.Response.SendingReturn(w, c, err)
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
