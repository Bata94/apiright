package apiright

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// App represents the main APIRight application
type App struct {
	router     *mux.Router
	db         *sql.DB
	middleware []Middleware
	config     *Config
}

// Config holds the application configuration
type Config struct {
	Host         string // Server host (default: "0.0.0.0")
	Port         string // Server port (default: "8080")
	Database     string // Database type: "postgres", "sqlite3"
	DSN          string // Database connection string
	APIVersion   string // API version prefix (default: "v1")
	Debug        bool   // Enable debug logging
	EnableCORS   bool   // Enable CORS middleware
	EnableLogger bool   // Enable logging middleware
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Host:         "0.0.0.0",
		Port:         "8080",
		Database:     "sqlite3",
		DSN:          ":memory:",
		APIVersion:   "v1",
		Debug:        false,
		EnableCORS:   true,
		EnableLogger: true,
	}
}

// Middleware represents a middleware function
type Middleware func(http.Handler) http.Handler

// New creates a new APIRight application with the given configuration
func New(config *Config) *App {
	if config == nil {
		config = DefaultConfig()
	}
	
	app := &App{
		router:     mux.NewRouter(),
		config:     config,
		middleware: []Middleware{},
	}

	// Setup default middleware
	if app.config.EnableCORS {
		app.Use(CORSMiddleware())
	}
	
	if app.config.EnableLogger {
		app.Use(LoggingMiddleware())
	}

	return app
}

// NewWithOptions creates a new APIRight application with options
func NewWithOptions(opts ...Option) *App {
	config := DefaultConfig()
	
	app := &App{
		router:     mux.NewRouter(),
		config:     config,
		middleware: []Middleware{},
	}

	// Apply options
	for _, opt := range opts {
		opt(app)
	}

	// Setup default middleware
	if app.config.EnableCORS {
		app.Use(CORSMiddleware())
	}
	
	if app.config.EnableLogger {
		app.Use(LoggingMiddleware())
	}

	return app
}

// Option represents a configuration option
type Option func(*App)

// WithDatabase sets the database connection
func WithDatabase(db *sql.DB) Option {
	return func(app *App) {
		app.db = db
	}
}

// WithConfig sets the application configuration
func WithConfig(config *Config) Option {
	return func(app *App) {
		app.config = config
	}
}

// Use adds middleware to the application
func (app *App) Use(middleware Middleware) {
	app.middleware = append(app.middleware, middleware)
}

// Router returns the underlying mux router
func (app *App) Router() *mux.Router {
	return app.router
}

// Database returns the database connection
func (app *App) Database() *sql.DB {
	return app.db
}

// Config returns the application configuration
func (app *App) Config() *Config {
	return app.config
}

// APIGroup creates a new route group with the API version prefix
func (app *App) APIGroup() *mux.Router {
	return app.router.PathPrefix(fmt.Sprintf("/%s", app.config.APIVersion)).Subrouter()
}

// RegisterCRUD registers CRUD endpoints for the given model
func (app *App) RegisterCRUD(path string, model interface{}) {
	// This will be implemented to work with the crud package
	// For now, it's a placeholder that shows the intended API
	log.Printf("Registering CRUD endpoints for %s", path)
}

// RegisterCRUDWithTransform registers CRUD endpoints with model transformation
func (app *App) RegisterCRUDWithTransform(path string, dbModel interface{}, apiModel interface{}) {
	// This will be implemented to work with the crud and transform packages
	// For now, it's a placeholder that shows the intended API
	log.Printf("Registering CRUD endpoints with transformation for %s", path)
}

// Listen starts the HTTP server
func (app *App) Listen(addr string) error {
	// Apply middleware to router
	handler := app.applyMiddleware(app.router)
	
	if addr == "" {
		addr = fmt.Sprintf("%s:%s", app.config.Host, app.config.Port)
	}
	
	log.Printf("🚀 APIRight server starting on %s", addr)
	return http.ListenAndServe(addr, handler)
}

// Start starts the HTTP server using the configured host and port
func (app *App) Start() error {
	return app.Listen("")
}

// applyMiddleware applies all registered middleware to the handler
func (app *App) applyMiddleware(handler http.Handler) http.Handler {
	for i := len(app.middleware) - 1; i >= 0; i-- {
		handler = app.middleware[i](handler)
	}
	return handler
}

// JSONResponse sends a JSON response
func JSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}

// ErrorResponse sends an error response
func ErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]interface{}{
		"error":   true,
		"message": message,
		"status":  statusCode,
	}
	JSONResponse(w, response, statusCode)
}

// SuccessResponse sends a success response
func SuccessResponse(w http.ResponseWriter, data interface{}) {
	response := map[string]interface{}{
		"error": false,
		"data":  data,
	}
	JSONResponse(w, response, http.StatusOK)
}