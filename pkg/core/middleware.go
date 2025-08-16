package core

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bata94/apiright/pkg/logger"
)

// Middleware is a function that wraps a Handler to add functionality.
// TODO: Define a consistent error handling strategy for middlewares, considering whether errors should be returned directly by the middleware or propagated through the Ctx object.
type Middleware func(Handler) Handler

func FavIcon(path string) Middleware {
	return func(next Handler) Handler {
		return func(c *Ctx) error {
			log.Debug("Middleware FavIcon")
			log.Debugf("Path: %s", path)
			log.Debugf("c.Request.URL.Path: %s", c.Request.URL.Path)
			if strings.HasSuffix(c.Request.URL.Path, "/favicon.ico") {
				iconBytes, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				c.Response.AddHeader("Content-Type", "image/x-icon")
				c.Response.SetStatus(200)
				c.Response.SetData(iconBytes)
				return nil
			}
			return next(c)
		}
	}
}

// LogMiddleware is a middleware that logs requests.
func LogMiddleware(logger logger.Logger) Middleware {
	return func(next Handler) Handler {
		return func(c *Ctx) error {
			go func(c *Ctx) {
				<-c.conClosed

				duration := c.conEnded.Sub(c.conStarted)
				// TODO: Enhance log formatting to improve readability, potentially by using a dedicated logging library or custom formatter to add tabs, colors, and structured output.
				infoLog := fmt.Sprintf("[%d] <%d ms> | [%s] %s - %s", c.Response.StatusCode, duration.Microseconds(), c.Request.Method, c.Request.RequestURI, c.Request.RemoteAddr)
				if c.Response.StatusCode >= 400 {
					// TODO: Include the actual error message in the log output when a request results in an error (status code >= 400) for better debugging.
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

// TODO: Implement a function that returns a CORSConfig allowing all origins, methods, and headers, similar to ExposeAllCORSConfig but for all aspects of CORS.

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
func handlePreflight(c *Ctx, config CORSConfig) {
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
func CORSMiddleware(config CORSConfig) Middleware {
	normalizeCORSConfig(&config)

	return func(next Handler) Handler {
		return func(c *Ctx) error {
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

// CSRFConfig holds the configuration for the CSRF middleware
type CSRFConfig struct {
	// TokenLength is the length of the CSRF token
	TokenLength uint8
	// CookieName is the name of the CSRF cookie
	CookieName string
	// CookiePath is the path of the CSRF cookie
	CookiePath string
	// CookieExpires is the expiration time of the CSRF cookie
	CookieExpires time.Duration
	// CookieSecure is the secure flag of the CSRF cookie
	CookieSecure bool
	// CookieHTTPOnly is the HTTPOnly flag of the CSRF cookie
	CookieHTTPOnly bool
	// HeaderName is the name of the CSRF header
	HeaderName string
}

// DefaultCSRFConfig returns a default CSRF configuration
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		TokenLength:    32,
		CookieName:     "csrf_token",
		CookiePath:     "/",
		CookieExpires:  24 * time.Hour,
		CookieSecure:   true,
		CookieHTTPOnly: true,
		HeaderName:     "X-CSRF-Token",
	}
}

// CSRFMiddleware returns a middleware that handles CSRF
func CSRFMiddleware(config CSRFConfig) Middleware {
	return func(next Handler) Handler {
		return func(c *Ctx) error {
			if c.Request.Method == http.MethodGet {
				// Generate and set CSRF cookie
				token, err := generateRandomToken(config.TokenLength)
				if err != nil {
					return err
				}

				cookie := &http.Cookie{
					Name:     config.CookieName,
					Value:    token,
					Path:     config.CookiePath,
					Expires:  time.Now().Add(config.CookieExpires),
					Secure:   config.CookieSecure,
					HttpOnly: config.CookieHTTPOnly,
				}
				c.Response.AddHeader("Set-Cookie", cookie.String())
				// TODO: Thoroughly review and double-check the CSRF token generation, storage (cookie vs. header), and verification logic to ensure it adheres to security best practices and correctly handles all edge cases.
				// } else if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodDelete {
			} else {
				// Verify CSRF token
				csrfTokenFromHeader := c.Request.Header.Get(config.HeaderName)
				csrfTokenFromCookie, err := c.Request.Cookie(config.CookieName)
				if err != nil {
					c.Response.SetStatus(http.StatusForbidden)
					return err
				}

				if csrfTokenFromHeader != csrfTokenFromCookie.Value {
					c.Response.SetStatus(http.StatusForbidden)
					return fmt.Errorf("CSRF token mismatch")
				}
			}

			return next(c)
		}
	}
}

func generateRandomToken(length uint8) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
