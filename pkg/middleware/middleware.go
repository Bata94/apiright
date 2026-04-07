// Package middleware provides HTTP and gRPC middleware.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bata94/apiright/pkg/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HTTPMiddleware interface for HTTP middleware
type HTTPMiddleware interface {
	Name() string
	Priority() int
	Handler() func(http.Handler) http.Handler
}

// LoggingMiddleware provides request/response logging
type LoggingMiddleware struct {
	logger core.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger core.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Name returns middleware name
func (lm *LoggingMiddleware) Name() string {
	return "logging"
}

// Priority returns middleware priority
func (lm *LoggingMiddleware) Priority() int {
	return 100 // Low priority (executed first)
}

// Handler returns HTTP middleware handler
func (lm *LoggingMiddleware) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lm.logger.Info("HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.Header.Get("User-Agent"),
			)

			// Call next handler
			next.ServeHTTP(w, r)

			duration := time.Since(start)
			lm.logger.Info("HTTP response",
				"method", r.Method,
				"path", r.URL.Path,
				"duration", duration,
			)
		})
	}
}

// GRPCInterceptor returns gRPC interceptor
func (lm *LoggingMiddleware) GRPCInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		lm.logger.Info("gRPC request",
			"method", info.FullMethod,
			"request", fmt.Sprintf("%+v", req),
		)

		// Call handler
		resp, err := handler(ctx, req)

		duration := time.Since(start)
		lm.logger.Info("gRPC response",
			"method", info.FullMethod,
			"duration", duration,
			"error", err,
		)

		return resp, err
	}
}

// CORSMiddleware provides CORS handling
type CORSMiddleware struct {
	config CORSConfig
	logger core.Logger
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(config CORSConfig, logger core.Logger) *CORSMiddleware {
	return &CORSMiddleware{
		config: config,
		logger: logger,
	}
}

// Name returns middleware name
func (cm *CORSMiddleware) Name() string {
	return "cors"
}

// Priority returns middleware priority
func (cm *CORSMiddleware) Priority() int {
	return 10 // High priority (executed early)
}

// Handler returns HTTP middleware handler
func (cm *CORSMiddleware) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range cm.config.AllowOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				if cm.config.AllowOrigins[0] == "*" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", strings.Join(cm.config.AllowMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(cm.config.AllowHeaders, ", "))
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cm.config.MaxAge))

			if cm.config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GRPCInterceptor returns gRPC interceptor (CORS not applicable to gRPC)
func (cm *CORSMiddleware) GRPCInterceptor() grpc.UnaryServerInterceptor {
	return nil // CORS doesn't apply to gRPC
}

// RateLimitMiddleware provides request rate limiting
type RateLimitMiddleware struct {
	limit    int
	window   time.Duration
	requests map[string][]time.Time
	logger   core.Logger
	mu       sync.Mutex
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(limit int, window time.Duration, logger core.Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limit:    limit,
		window:   window,
		requests: make(map[string][]time.Time),
		logger:   logger,
	}
}

// Name returns middleware name
func (rm *RateLimitMiddleware) Name() string {
	return "rate_limit"
}

// Priority returns middleware priority
func (rm *RateLimitMiddleware) Priority() int {
	return 5 // Very high priority (executed first)
}

// Handler returns HTTP middleware handler
func (rm *RateLimitMiddleware) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := rm.getClientIP(r)

			rm.mu.Lock()
			// Clean old requests
			rm.cleanOldRequests(clientIP)

			// Check current count
			requestCount := len(rm.requests[clientIP])
			if requestCount >= rm.limit {
				rm.mu.Unlock()
				rm.logger.Warn("Rate limit exceeded",
					"ip", clientIP,
					"count", requestCount,
					"limit", rm.limit,
				)

				// Return appropriate content type based on Accept header
				acceptHeader := r.Header.Get("Accept")
				contentType := "application/json"
				if acceptHeader != "" {
					// Simple content type detection
					if strings.Contains(acceptHeader, "application/json") {
						contentType = "application/json"
					} else if strings.Contains(acceptHeader, "text/plain") {
						contentType = "text/plain"
					}
				}

				w.Header().Set("Content-Type", contentType)
				w.WriteHeader(http.StatusTooManyRequests)

				response := fmt.Sprintf("Rate limit exceeded: %d requests per %v", rm.limit, rm.window)
				if _, err := w.Write([]byte(response)); err != nil {
					rm.logger.Warn("failed to write rate limit response", "error", err)
				}
				return
			}

			// Record current request
			rm.requests[clientIP] = append(rm.requests[clientIP], time.Now())
			rm.mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

// GRPCInterceptor returns gRPC interceptor
func (rm *RateLimitMiddleware) GRPCInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Extract client info from context (if available)
		clientIP := rm.extractGRPCClientIP(ctx)

		if clientIP != "" {
			rm.mu.Lock()
			// Clean old requests
			rm.cleanOldRequests(clientIP)

			// Check current count
			requestCount := len(rm.requests[clientIP])
			if requestCount >= rm.limit {
				rm.mu.Unlock()
				rm.logger.Warn("gRPC rate limit exceeded",
					"ip", clientIP,
					"count", requestCount,
					"limit", rm.limit,
				)
				return nil, status.Error(codes.ResourceExhausted, "Rate limit exceeded")
			}

			// Record current request
			rm.requests[clientIP] = append(rm.requests[clientIP], time.Now())
			rm.mu.Unlock()
		}

		return handler(ctx, req)
	}
}

// cleanOldRequests removes old request timestamps
func (rm *RateLimitMiddleware) cleanOldRequests(clientIP string) {
	if requests, exists := rm.requests[clientIP]; exists {
		cutoff := time.Now().Add(-rm.window)
		var validRequests []time.Time

		for _, reqTime := range requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}

		rm.requests[clientIP] = validRequests
	}
}

// getClientIP extracts client IP from HTTP request
func (rm *RateLimitMiddleware) getClientIP(r *http.Request) string {
	// Try X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Try X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

// extractGRPCClientIP extracts client IP from gRPC context
func (rm *RateLimitMiddleware) extractGRPCClientIP(ctx context.Context) string {
	// In a real implementation, this would extract from metadata
	// For now, return empty string to disable rate limiting for gRPC
	return ""
}

// ValidationMiddleware provides request validation
type ValidationMiddleware struct {
	validator core.Validator
	logger    core.Logger
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware(validator core.Validator, logger core.Logger) *ValidationMiddleware {
	return &ValidationMiddleware{
		validator: validator,
		logger:    logger,
	}
}

// Name returns middleware name
func (vm *ValidationMiddleware) Name() string {
	return "validation"
}

// Priority returns middleware priority
func (vm *ValidationMiddleware) Priority() int {
	return 50 // Medium priority
}

// Handler returns HTTP middleware handler
func (vm *ValidationMiddleware) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For now, just add a validation header
			w.Header().Set("X-Validated", "true")
			next.ServeHTTP(w, r)
		})
	}
}

// GRPCInterceptor returns gRPC interceptor
func (vm *ValidationMiddleware) GRPCInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Basic validation
		if vm.validator != nil {
			if err := vm.validator.Validate(req); err != nil {
				vm.logger.Warn("gRPC request validation failed",
					"method", info.FullMethod,
					"error", err,
				)
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}

		return handler(ctx, req)
	}
}

// MiddlewareRegistry manages middleware registration and execution
type MiddlewareRegistry struct {
	middleware []HTTPMiddleware
	logger     core.Logger
	mu         sync.RWMutex
}

// NewMiddlewareRegistry creates a new middleware registry
func NewMiddlewareRegistry(logger core.Logger) *MiddlewareRegistry {
	return &MiddlewareRegistry{
		middleware: []HTTPMiddleware{},
		logger:     logger,
	}
}

// RegisterMiddleware registers a middleware
func (mr *MiddlewareRegistry) RegisterMiddleware(middleware HTTPMiddleware) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	mr.middleware = append(mr.middleware, middleware)

	// Sort by priority (lower number = higher priority)
	for i := 0; i < len(mr.middleware)-1; i++ {
		for j := 0; j < len(mr.middleware)-i-1; j++ {
			if mr.middleware[j].Priority() > mr.middleware[j+1].Priority() {
				mr.middleware[j], mr.middleware[j+1] = mr.middleware[j+1], mr.middleware[j]
			}
		}
	}

	mr.logger.Info("Middleware registered",
		"name", middleware.Name(),
		"priority", middleware.Priority(),
	)
}

// GetHTTPMiddleware returns all HTTP middleware handlers
func (mr *MiddlewareRegistry) GetHTTPMiddleware() []func(http.Handler) http.Handler {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var handlers []func(http.Handler) http.Handler

	for _, mw := range mr.middleware {
		handlers = append(handlers, mw.Handler())
	}

	return handlers
}

// GetGRPCInterceptors returns all gRPC interceptors
func (mr *MiddlewareRegistry) GetGRPCInterceptors() []grpc.UnaryServerInterceptor {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var interceptors []grpc.UnaryServerInterceptor

	for _, mw := range mr.middleware {
		// Type assertion to check if it has GRPCInterceptor method
		switch m := mw.(type) {
		case *LoggingMiddleware:
			if interceptor := m.GRPCInterceptor(); interceptor != nil {
				interceptors = append(interceptors, interceptor)
			}
		case *ValidationMiddleware:
			if interceptor := m.GRPCInterceptor(); interceptor != nil {
				interceptors = append(interceptors, interceptor)
			}
		case interface {
			GRPCInterceptor() grpc.UnaryServerInterceptor
		}:
			if interceptor := m.GRPCInterceptor(); interceptor != nil {
				interceptors = append(interceptors, interceptor)
			}
		}
	}

	return interceptors
}

// ListMiddleware returns all registered middleware
func (mr *MiddlewareRegistry) ListMiddleware() []HTTPMiddleware {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	result := make([]HTTPMiddleware, len(mr.middleware))
	copy(result, mr.middleware)
	return result
}

// GetMiddlewareByName returns middleware by name
func (mr *MiddlewareRegistry) GetMiddlewareByName(name string) (HTTPMiddleware, bool) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	for _, mw := range mr.middleware {
		if mw.Name() == name {
			return mw, true
		}
	}
	return nil, false
}

// RemoveMiddleware removes a middleware by name
func (mr *MiddlewareRegistry) RemoveMiddleware(name string) bool {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	for i, mw := range mr.middleware {
		if mw.Name() == name {
			mr.middleware = append(mr.middleware[:i], mr.middleware[i+1:]...)
			mr.logger.Info("Middleware removed", "name", name)
			return true
		}
	}
	return false
}
