package cors

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bata94/apiright/pkg/core"
)

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

// ExposeAllCORSConfig returns a CORSConfig that allows all origins, headers, and methods.
// Use with caution! But might be nice for dev/testing.
func ExposeAllCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

// normalizeCORSConfig processes the CORS configuration to standardize its values.
func normalizeCORSConfig(config *CORSConfig) {
	for i, method := range config.AllowMethods {
		config.AllowMethods[i] = strings.ToUpper(method)
	}
	for i, header := range config.AllowHeaders {
		config.AllowHeaders[i] = http.CanonicalHeaderKey(header)
	}
	for i, header := range config.ExposeHeaders {
		config.ExposeHeaders[i] = http.CanonicalHeaderKey(header)
	}
}

// isOriginAllowed checks if a given origin is present in the list of allowed origins.
// It returns the allowed origin string and a boolean indicating if it's allowed.
func isOriginAllowed(origin string, allowedOrigins []string) (string, bool) {
	if origin == "" {
		return "", false
	}
	for _, o := range allowedOrigins {
		if o == "*" {
			return "*", true
		}
		if strings.EqualFold(o, origin) {
			return origin, true
		}
	}
	return "", false
}

// handlePreflight handles the preflight OPTIONS request by setting the appropriate CORS headers.
func handlePreflight(c *core.Ctx, config CORSConfig) {
	// Set allowed methods
	if len(config.AllowMethods) == 1 && config.AllowMethods[0] == "*" {
		c.Response.AddHeader("Access-Control-Allow-Methods", c.Request.Header.Get("Access-control-request-method"))
	} else {
		c.Response.AddHeader("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
	}

	// Set allowed headers
	if len(config.AllowHeaders) == 1 && config.AllowHeaders[0] == "*" {
		c.Response.AddHeader("Access-Control-Allow-Headers", c.Request.Header.Get("Access-control-request-headers"))
		c.Response.AddHeader("Vary", "Access-Control-Request-Headers")
	} else if len(config.AllowHeaders) > 0 {
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
}

// CORSMiddleware returns a middleware that handles CORS
func CORSMiddleware(config CORSConfig) core.Middleware {
	normalizeCORSConfig(&config)

	return func(next core.Handler) core.Handler {
		return func(c *core.Ctx) error {
			origin := c.Request.Header.Get("Origin")
			allowedOrigin, ok := isOriginAllowed(origin, config.AllowOrigins)
			if !ok {
				return next(c)
			}

			c.Response.AddHeader("Access-Control-Allow-Origin", allowedOrigin)
			if allowedOrigin != "*" {
				c.Response.AddHeader("Vary", "Origin")
			}

			if c.Request.Method == http.MethodOptions {
				handlePreflight(c, config)
				return nil
			}

			if len(config.ExposeHeaders) > 0 {
				c.Response.AddHeader("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
			}

			if config.AllowCredentials {
				c.Response.AddHeader("Access-Control-Allow-Credentials", "true")
			}

			return next(c)
		}
	}
}
