package core

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bata94/apiright/pkg/logger"
)

// Middleware is a function that wraps a Handler to add functionality.
// TODO: Think Errorandling either here or in CTX
type Middleware func(Handler) Handler

// LogMiddleware is a middleware that logs requests.
func LogMiddleware(logger logger.Logger) Middleware {
	return func(next Handler) Handler {
		return func(c *Ctx) error {
			go func(c *Ctx) {
				<-c.conClosed

				duration := c.conEnded.Sub(c.conStarted)
				// TODO: use tabs and colors to make logs more appealing
				infoLog := fmt.Sprintf("[%d] <%d ms> | [%s] %s - %s", c.Response.StatusCode, duration.Microseconds(), c.Request.Method, c.Request.RequestURI, c.Request.RemoteAddr)
				if c.Response.StatusCode >= 400 {
					// TODO: add the error Msg here
					logger.Error(infoLog)
				} else {
					logger.Info(infoLog)
				}
			}(c)

			return next(c)
		}
	}
}

// PanicMiddleware is a middleware that recovers from panics.
func PanicMiddleware() Middleware {
	return func(next Handler) Handler {
		return func(c *Ctx) error {
			defer func() {
				if err := recover(); err != nil {
					c.Response.SetStatus(http.StatusInternalServerError)
					c.Response.Message = fmt.Sprintf("Panic: %v", err)
				}
			}()
			return next(c)
		}
	}
}

// TimeoutConfig holds the configuration for timeout middleware
type TimeoutConfig struct {
	// Timeout is the maximum duration for a request to complete
	// Default is 30 seconds
	Timeout time.Duration

	// TimeoutMessage is the message returned when a request times out
	// Default is "Request timeout"
	TimeoutMessage string

	// TimeoutStatusCode is the HTTP status code returned when a request times out
	// Default is 408 (Request Timeout)
	TimeoutStatusCode int
}

// DefaultTimeoutConfig returns a default timeout configuration
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Timeout:           30 * time.Second,
		TimeoutMessage:    "Request timeout",
		TimeoutStatusCode: http.StatusRequestTimeout,
	}
}

// TimeoutConfigFromApp returns a TimeoutConfig from an App instance.
func TimeoutConfigFromApp(a App) TimeoutConfig {
  return a.timeoutConfig
}

// BUG: Exec of HandlerFunc is not prob stopped!
// TimeoutMiddleware returns a middleware that handles request timeouts
func TimeoutMiddleware(config TimeoutConfig) Middleware {
	return func(next Handler) Handler {
		return func(c *Ctx) error {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(c.Request.Context(), config.Timeout)
			defer cancel()

			// Replace the request context with the timeout context
			c.Request = c.Request.WithContext(ctx)

			// Create a channel to receive the result of the handler
			done := make(chan error, 1)

			// Run the handler in a goroutine
			go func() {
				done <- next(c)
			}()

			// Wait for either the handler to complete or the timeout
			select {
			case err := <-done:
				// Handler completed within timeout
				return err
			case <-ctx.Done():
				// Timeout occurred
				if ctx.Err() == context.DeadlineExceeded {
					c.Response.SetStatus(config.TimeoutStatusCode)
					c.Response.Message = config.TimeoutMessage
					return fmt.Errorf("request timeout after %v", config.Timeout)
				}
				// Context was cancelled for another reason
				return ctx.Err()
			}
		}
	}
}

// CORSConfig holds the configuration for CORS middleware
type CORSConfig struct {
	// AllowOrigins is a list of origins a cross-domain request can be executed from.
	// If the special "*" value is present in the list, all origins will be allowed.
	// Default value is ["*"]
	AllowOrigins []string

	// AllowMethods is a list of methods the client is allowed to use with
	// cross-domain requests. Default value is simple methods (GET, POST, PUT, DELETE)
	AllowMethods []string

	// AllowHeaders is a list of non-simple headers the client is allowed to use with
	// cross-domain requests. Default value is []
	AllowHeaders []string

	// ExposeHeaders indicates which headers are safe to expose to the API of a CORS
	// API specification. Default value is []
	ExposeHeaders []string

	// AllowCredentials indicates whether the request can include user credentials like
	// cookies, HTTP authentication or client side SSL certificates. Default is false.
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached. Default is 0 which stands for no max age.
	MaxAge int
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}

// CORSMiddleware returns a middleware that handles CORS
func CORSMiddleware(config CORSConfig) Middleware {
	// Convert methods to uppercase and headers to canonical form for consistency
	for i, method := range config.AllowMethods {
		config.AllowMethods[i] = strings.ToUpper(method)
	}
	for i, header := range config.AllowHeaders {
		config.AllowHeaders[i] = http.CanonicalHeaderKey(header)
	}
	for i, header := range config.ExposeHeaders {
		config.ExposeHeaders[i] = http.CanonicalHeaderKey(header)
	}

	return func(next Handler) Handler {
		return func(c *Ctx) error {
			// Get origin from request
			origin := c.Request.Header.Get("Origin")

			// Skip if no Origin header is present
			if origin == "" {
				return next(c)
			}

			// Check if the origin is allowed
			allowedOrigin := ""
			for _, o := range config.AllowOrigins {
				if o == "*" {
					allowedOrigin = "*"
					break
				}
				if strings.EqualFold(o, origin) {
					allowedOrigin = origin
					break
				}
			}

			// If origin is not allowed, proceed without CORS headers
			if allowedOrigin == "" {
				return next(c)
			}

			// Set CORS headers
			c.Response.AddHeader("Access-Control-Allow-Origin", allowedOrigin)

			// Set Vary header for proper caching
			if allowedOrigin != "*" {
				c.Response.AddHeader("Vary", "Origin")
			}

			// Handle preflight requests
			if c.Request.Method == http.MethodOptions {
				// Set allowed methods
				c.Response.AddHeader("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))

				// Set allowed headers
				if len(config.AllowHeaders) > 0 {
					c.Response.AddHeader("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
				} else {
					// If no specific headers are defined, allow the requested ones
					reqHeaders := c.Request.Header.Get("Access-Control-Request-Headers")
					if reqHeaders != "" {
						c.Response.AddHeader("Access-Control-Allow-Headers", reqHeaders)
						c.Response.AddHeader("Vary", "Access-Control-Request-Headers")
					}
				}

				// Set max age for preflight cache
				if config.MaxAge > 0 {
					c.Response.AddHeader("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
				}

				// Set credentials flag
				if config.AllowCredentials {
					c.Response.AddHeader("Access-Control-Allow-Credentials", "true")
				}

				// Return 204 No Content for preflight requests
				c.Response.SetStatus(http.StatusNoContent)
				return nil
			}

			// For non-preflight requests, set expose headers and credentials if needed
			if len(config.ExposeHeaders) > 0 {
				c.Response.AddHeader("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
			}

			if config.AllowCredentials {
				c.Response.AddHeader("Access-Control-Allow-Credentials", "true")
			}

			// Continue with the next handler
			return next(c)
		}
	}
}
