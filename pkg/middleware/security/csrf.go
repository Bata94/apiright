package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/bata94/apiright/pkg/core"
)

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
func CSRFMiddleware(config CSRFConfig) core.Middleware {
	return func(next core.Handler) core.Handler {
		return func(c *core.Ctx) error {
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
