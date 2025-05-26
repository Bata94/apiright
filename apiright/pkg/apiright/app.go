package apiright

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/bata94/apiright/pkg/crud"
	"github.com/bata94/apiright/pkg/transform"
	"github.com/gorilla/mux"
)

// Config holds the application configuration
type Config struct {
	Host         string // Server host
	Port         string // Server port
	Database     string // Database type (postgres, sqlite3)
	DSN          string // Database connection string
	APIVersion   string // API version prefix
	Debug        bool   // Enable debug mode
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

// App represents the main application
type App struct {
	router     *mux.Router
	config     *Config
	db         *sql.DB
	middleware []Middleware
}

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

// RegisterCRUD registers CRUD endpoints for the given model type
func RegisterCRUD[T crud.Model](app *App, path string, tableName string) *crud.CRUDHandler[T] {
	if app.db == nil {
		log.Printf("Warning: No database connection set for CRUD operations on %s", path)
		return nil
	}

	repo := crud.NewRepository[T](app.db, tableName)
	handler := crud.NewCRUDHandler[T](repo)
	
	// Register routes on the API group
	apiGroup := app.APIGroup()
	handler.RegisterRoutes(apiGroup, path)
	
	log.Printf("Registered CRUD endpoints for %s at /%s%s", tableName, app.config.APIVersion, path)
	return handler
}

// RegisterCRUDWithTransform registers CRUD endpoints with model transformation
func RegisterCRUDWithTransform[TDB crud.Model, TAPI any](
	app *App, 
	path string, 
	tableName string,
	transformer *transform.BiDirectionalTransformer[TDB, TAPI],
) *TransformCRUDHandler[TDB, TAPI] {
	if app.db == nil {
		log.Printf("Warning: No database connection set for CRUD operations on %s", path)
		return nil
	}

	repo := crud.NewRepository[TDB](app.db, tableName)
	handler := NewTransformCRUDHandler[TDB, TAPI](repo, transformer)
	
	// Register routes on the API group
	apiGroup := app.APIGroup()
	handler.RegisterRoutes(apiGroup, path)
	
	log.Printf("Registered CRUD endpoints with transformation for %s at /%s%s", tableName, app.config.APIVersion, path)
	return handler
}

// TransformCRUDHandler provides CRUD operations with model transformation
type TransformCRUDHandler[TDB crud.Model, TAPI any] struct {
	repo        *crud.Repository[TDB]
	transformer *transform.BiDirectionalTransformer[TDB, TAPI]
}

// NewTransformCRUDHandler creates a new transform CRUD handler
func NewTransformCRUDHandler[TDB crud.Model, TAPI any](
	repo *crud.Repository[TDB],
	transformer *transform.BiDirectionalTransformer[TDB, TAPI],
) *TransformCRUDHandler[TDB, TAPI] {
	return &TransformCRUDHandler[TDB, TAPI]{
		repo:        repo,
		transformer: transformer,
	}
}

// Create handles POST requests with transformation
func (h *TransformCRUDHandler[TDB, TAPI]) Create(w http.ResponseWriter, r *http.Request) {
	var apiEntity TAPI
	if err := json.NewDecoder(r.Body).Decode(&apiEntity); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Transform API model to DB model
	dbEntity, err := h.transformer.Backward(apiEntity)
	if err != nil {
		http.Error(w, fmt.Sprintf("Transformation error: %v", err), http.StatusBadRequest)
		return
	}

	// Create in database
	createdDB, err := h.repo.Create(dbEntity)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create entity: %v", err), http.StatusInternalServerError)
		return
	}

	// Transform back to API model
	createdAPI, err := h.transformer.Forward(createdDB)
	if err != nil {
		http.Error(w, fmt.Sprintf("Transformation error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdAPI)
}

// GetAll handles GET requests with transformation
func (h *TransformCRUDHandler[TDB, TAPI]) GetAll(w http.ResponseWriter, r *http.Request) {
	dbEntities, err := h.repo.GetAll()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve entities: %v", err), http.StatusInternalServerError)
		return
	}

	// Transform to API models
	apiEntities, err := h.transformer.ForwardSlice(dbEntities)
	if err != nil {
		http.Error(w, fmt.Sprintf("Transformation error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiEntities)
}

// GetByID handles GET requests with transformation
func (h *TransformCRUDHandler[TDB, TAPI]) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}

	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	dbEntity, err := h.repo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Entity not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to retrieve entity: %v", err), http.StatusInternalServerError)
		return
	}

	// Transform to API model
	apiEntity, err := h.transformer.Forward(dbEntity)
	if err != nil {
		http.Error(w, fmt.Sprintf("Transformation error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiEntity)
}

// Update handles PUT requests with transformation
func (h *TransformCRUDHandler[TDB, TAPI]) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}

	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var apiEntity TAPI
	if err := json.NewDecoder(r.Body).Decode(&apiEntity); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Transform API model to DB model
	dbEntity, err := h.transformer.Backward(apiEntity)
	if err != nil {
		http.Error(w, fmt.Sprintf("Transformation error: %v", err), http.StatusBadRequest)
		return
	}

	dbEntity.SetID(id)
	updatedDB, err := h.repo.Update(dbEntity)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update entity: %v", err), http.StatusInternalServerError)
		return
	}

	// Transform back to API model
	updatedAPI, err := h.transformer.Forward(updatedDB)
	if err != nil {
		http.Error(w, fmt.Sprintf("Transformation error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedAPI)
}

// Delete handles DELETE requests
func (h *TransformCRUDHandler[TDB, TAPI]) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}

	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(id); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete entity: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RegisterRoutes registers all CRUD routes for the transform handler
func (h *TransformCRUDHandler[TDB, TAPI]) RegisterRoutes(router *mux.Router, basePath string) {
	router.HandleFunc(basePath, h.Create).Methods("POST")
	router.HandleFunc(basePath, h.GetAll).Methods("GET")
	router.HandleFunc(basePath+"/{id}", h.GetByID).Methods("GET")
	router.HandleFunc(basePath+"/{id}", h.Update).Methods("PUT")
	router.HandleFunc(basePath+"/{id}", h.Delete).Methods("DELETE")
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

// Helper functions

// parseID parses an ID string to int64
func parseID(idStr string) (int64, error) {
	// This could be extended to support different ID types
	return parseInt64(idStr)
}

// parseInt64 parses a string to int64
func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
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