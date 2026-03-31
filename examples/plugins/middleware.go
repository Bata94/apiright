package plugins

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bata94/apiright/pkg/core"
)

type CustomMiddlewarePlugin struct {
	paths   []string
	headers map[string]string
	logger  core.Logger
}

func NewCustomMiddlewarePlugin() *CustomMiddlewarePlugin {
	return &CustomMiddlewarePlugin{
		paths:   []string{},
		headers: make(map[string]string),
	}
}

func (p *CustomMiddlewarePlugin) Name() string {
	return "custom-middleware"
}

func (p *CustomMiddlewarePlugin) Version() string {
	return "1.0.0"
}

func (p *CustomMiddlewarePlugin) Generate(ctx *core.GenerationContext) error {
	ctx.Log("info", "Custom middleware plugin executed",
		"paths", len(p.paths),
		"headers", len(p.headers))

	for _, path := range p.paths {
		ctx.Log("debug", "Middleware will be applied to path", "path", path)
	}

	for key, value := range p.headers {
		ctx.Log("debug", "Adding header", "key", key, "value", value)
	}

	return nil
}

func (p *CustomMiddlewarePlugin) Validate(schema *core.Schema) error {
	if len(p.paths) == 0 {
		return fmt.Errorf("no paths specified for middleware")
	}
	for _, path := range p.paths {
		if !strings.HasPrefix(path, "/") {
			return fmt.Errorf("path must start with /: %s", path)
		}
	}
	return nil
}

func (p *CustomMiddlewarePlugin) AddPath(path string) {
	p.paths = append(p.paths, path)
}

func (p *CustomMiddlewarePlugin) AddHeader(key, value string) {
	p.headers[key] = value
}

func (p *CustomMiddlewarePlugin) MiddlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key, value := range p.headers {
			w.Header().Set(key, value)
		}
		next.ServeHTTP(w, r)
	})
}

func (p *CustomMiddlewarePlugin) SetLogger(logger core.Logger) {
	p.logger = logger
}
