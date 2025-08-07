package core

// ConnectionLimitConfig holds the configuration for the connection limit middleware.
type ConnectionLimitConfig struct {
	MaxConnections int
}

// DefaultConnectionLimitConfig returns a default connection limit configuration.
func DefaultConnectionLimitConfig() ConnectionLimitConfig {
	return ConnectionLimitConfig{
		MaxConnections: 100, // Default to 100 concurrent connections
	}
}

// ConnectionLimitMiddleware returns a middleware that limits the number of concurrent connections.
func ConnectionLimitMiddleware(config ConnectionLimitConfig) Middleware {
	sem := make(chan struct{}, config.MaxConnections)

	return func(next Handler) Handler {
		return func(c *Ctx) error {
			sem <- struct{}{}
			defer func() { <-sem }()
			return next(c)
		}
	}
}
