package apiright

import (
	"log"
	"net/http"
	"time"
)

// CORSMiddleware adds CORS headers to responses
func CORSMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Create a response writer wrapper to capture status code
			wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			next.ServeHTTP(wrapper, r)
			
			duration := time.Since(start)
			log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapper.statusCode, duration)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// JSONMiddleware ensures request content type is JSON for POST/PUT requests
func JSONMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" || r.Method == "PUT" {
				contentType := r.Header.Get("Content-Type")
				if contentType != "application/json" {
					ErrorResponse(w, "Content-Type must be application/json", http.StatusBadRequest)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware provides basic authentication middleware
func AuthMiddleware(authFunc func(r *http.Request) bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !authFunc(r) {
				ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}