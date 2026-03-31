package server

import "net/http"

type HTTPMiddleware struct {
	name     string
	priority int
	handler  func(http.Handler) http.Handler
}

func NewHTTPMiddleware(name string, priority int, handler func(http.Handler) http.Handler) *HTTPMiddleware {
	return &HTTPMiddleware{
		name:     name,
		priority: priority,
		handler:  handler,
	}
}

func (m *HTTPMiddleware) Name() string {
	return m.name
}

func (m *HTTPMiddleware) Priority() int {
	return m.priority
}

func (m *HTTPMiddleware) Handler() func(http.Handler) http.Handler {
	return m.handler
}
