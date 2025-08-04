package ar_templ

import (
	"net/http"
)

// UIRouter defines the interface for UI routing.
type UIRouter interface {
	RegisterUIRoutes(mux *http.ServeMux)
}

// NoOpUIRouter is a no-op implementation of UIRouter for cases where UI routing is not needed.
type NoOpUIRouter struct{}

// RegisterUIRoutes does nothing for NoOpUIRouter.
func (n *NoOpUIRouter) RegisterUIRoutes(mux *http.ServeMux) {
	// No-op
}
