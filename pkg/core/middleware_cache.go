package core

import (
	"bytes"
	"net/http"
	"sync"
	"time"
)

// ResponseCacheConfig holds the configuration for the response caching middleware.
type ResponseCacheConfig struct {
	// Expiration is the duration after which a cached response expires.
	Expiration time.Duration
}

// DefaultResponseCacheConfig returns a default response cache configuration.
func DefaultResponseCacheConfig() ResponseCacheConfig {
	return ResponseCacheConfig{
		Expiration: 5 * time.Minute, // Default to 5 minutes
	}
}

// cachedResponse stores the cached response data.
type cachedResponse struct {
	status  int
	headers http.Header
	body    []byte
	expires time.Time
}

// responseCache is a simple in-memory cache.
var responseCache = struct {
	sync.RWMutex
	data map[string]cachedResponse
}{
	data: make(map[string]cachedResponse),
}

// ResponseCacheMiddleware returns a middleware that caches responses.
func ResponseCacheMiddleware(config ResponseCacheConfig) Middleware {
	return func(next Handler) Handler {
		return func(c *Ctx) error {
			cacheKey := c.Request.URL.String()

			// Check if response is in cache
			responseCache.RLock()
			cached, found := responseCache.data[cacheKey]
			responseCache.RUnlock()

			if found && time.Now().Before(cached.expires) {
				// Serve from cache
				c.Response.SetStatus(cached.status)
				for k, v := range cached.headers {
					for _, val := range v {
						c.Response.AddHeader(k, val)
					}
				}
				c.Response.SetData(cached.body)
				return nil
			}

			// Capture the response
			w := &responseWriter{ResponseWriter: c.Response.Writer}
			c.Response.Writer = w

			err := next(c)

			if err == nil && c.Response.StatusCode == http.StatusOK {
				// Cache the response
				responseCache.Lock()
				responseCache.data[cacheKey] = cachedResponse{
					status:  c.Response.StatusCode,
					headers: c.Response.Headers,
					body:    w.body.Bytes(),
					expires: time.Now().Add(config.Expiration),
				}
				responseCache.Unlock()
			}

			return err
		}
	}
}

// responseWriter is a wrapper to capture the response body.
type responseWriter struct {
	http.ResponseWriter
	body bytes.Buffer
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}