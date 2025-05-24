package core

import (
	"fmt"
	"net/http"
)

type Server struct {
	Addr string
	Port int
}

type App struct {
	Server *Server
	routes []*Route
}

func InitApp() App {
	fmt.Println("App will be initialized")

	s := Server{
		Addr: "127.0.0.1",
		Port: 5500,
	}

	a := App{
		Server: &s,
		routes: []*Route{},
	}

	return a
}

func (a App) fullAddr() string {
	return fmt.Sprintf("%s:%d", a.Server.Addr, a.Server.Port)
}

func (a App) Run() {
	fmt.Println("App is running")

	router := http.NewServeMux()

	for _, r := range a.routes {
		for _, e := range r.endpoints {
			router.HandleFunc(e.PathString(*r), e.handler)
		}
	}

	fmt.Println(http.ListenAndServe(a.fullAddr(), router))
}
