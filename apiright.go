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

func NewRouterGroup() {}

type RouterGroup struct{}

type Route struct{}

type Endpoint struct{}

type Middleware func(http.Handler) http.Handler

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
		Config:  &config,
		handler: handler,
	}
}

type App struct {
	Config  *AppConfig
	handler *http.ServeMux

	routerGroups []*RouterGroup
}

func (a App) getHttpHandler() *http.ServeMux {
	return a.handler
}

func (a *App) GET(path string, handler func(*Ctx) error) {
	a.handleFunc(fmt.Sprint("GET ", path), handler)
}

// TODO: Prob move into a Middleware
func recoverFromPanic(w http.ResponseWriter) {
	if r := recover(); r != nil {
		w.Write([]byte("recoverFromPanic"))
	}
}

func (a *App) handleFunc(p string, h func(*Ctx) error) {
	a.getHttpHandler().HandleFunc(p, func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		c := NewCtx(w, r)
		err := h(c)

		if err != nil {
			err = fmt.Errorf("Error in HanlderFunc: %w", err)
			w.Write([]byte(err.Error()))
		}
	})
}

func (a App) Run() error {
	return http.ListenAndServe(a.Config.GetListenAddress(), a.getHttpHandler())
}

