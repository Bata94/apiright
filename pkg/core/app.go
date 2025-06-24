package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

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

func (a *App) handleFunc(route Route, endPoint Endpoint, router Router) {
	handlerPath := fmt.Sprintf("%s %s", endPoint.method.toPathString(), route.path)
	a.Logger.Debugf("Registering route: %s", handlerPath)

	a.getHttpHandler().HandleFunc(handlerPath, func(w http.ResponseWriter, r *http.Request) {
		h := endPoint.handleFunc
		var err error

		if route.basePath == "/" && r.URL.Path != router.GetBasePath() {
			a.Logger.Debugf("Using default route handler for path: %s", r.URL.Path)
			h = a.defRouteHandler
		}

		panicMiddleware := PanicMiddleware()
		logMiddleware := LogMiddleware(a.Logger)
		h = panicMiddleware(h)
		h = logMiddleware(h)

		c := NewCtx(w, r)

		if endPoint.routeOptionConfig.ObjIn != nil {
			c.ObjIn = endPoint.routeOptionConfig.ObjIn
			c.ObjInType = reflect.TypeOf(c.ObjIn)

			objInByte, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				c.Response.SetStatus(http.StatusInternalServerError)
				c.Response.SetMessage("Body could not be read, err: " + err.Error())
				goto ClosingFunc
			}

			err = json.Unmarshal(objInByte, &c.ObjIn)
			if err != nil {
				c.Response.SetStatus(http.StatusInternalServerError)
				c.Response.SetMessage("Body could not be parsed to Object, err: " + err.Error())
				goto ClosingFunc
			}

			if reflect.TypeOf(c.ObjIn) != c.ObjInType {
				c.Response.SetStatus(http.StatusInternalServerError)
				c.Response.SetMessage("Parsed Object != wanted Object")
				goto ClosingFunc
			}
		}

		if endPoint.routeOptionConfig.ObjOut != nil {
			c.ObjOut = endPoint.routeOptionConfig.ObjOut
			c.ObjOutType = reflect.TypeOf(c.ObjOut)
		}

		err = h(c)

	ClosingFunc:
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
