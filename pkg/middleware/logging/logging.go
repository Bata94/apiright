package logging

import (
	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/logger"
)

// LogMiddleware is a middleware that logs requests.
func LogMiddleware(logger logger.Logger) core.Middleware {
	return func(next core.Handler) core.Handler {
		return func(c *core.Ctx) error {
			// Start logging in a goroutine
			go func(c *core.Ctx) {
				// Wait for connection to close
				<-c.ConClosed()
				duration := c.GetConnectionDuration()

				// Use structured logging with key-value pairs
				if c.Response.StatusCode >= 400 {
					logger.Error("request completed", "status", c.Response.StatusCode, "method", c.Request.Method, "path", c.Request.RequestURI, "remote", c.Request.RemoteAddr, "duration_ms", duration.Milliseconds())
				} else {
					logger.Info("request completed", "status", c.Response.StatusCode, "method", c.Request.Method, "path", c.Request.RequestURI, "remote", c.Request.RemoteAddr, "duration_ms", duration.Milliseconds())
				}
			}(c)

			return next(c)
		}
	}
}
