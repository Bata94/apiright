package crud

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/bata94/apiright/pkg/apiright"
	"github.com/gorilla/mux"
)

// Config holds the configuration for CRUD operations
type Config struct {
	TableName    string
	IDField      string
	Route        string
	Transformer  Transformer
	Middleware   []apiright.Middleware
	ReadOnly     bool
	WriteOnly    bool
	CustomFields map[string]string // Maps struct field to database column
}

// Transformer defines the interface for transforming between DB and API models
type Transformer interface {
	ToAPI(dbModel interface{}) (interface{}, error)
	FromAPI(apiModel interface{}) (interface{}, error)
}

// DefaultTransformer is a pass-through transformer
type DefaultTransformer struct{}

func (dt *DefaultTransformer) ToAPI(dbModel interface{}) (interface{}, error) {
	return dbModel, nil
}

func (dt *DefaultTransformer) FromAPI(apiModel interface{}) (interface{}, error) {
	return apiModel, nil
}

// CRUDHandler handles CRUD operations for a specific model
type CRUDHandler[T any] struct {
	app         *apiright.App
	config      Config
	modelType   reflect.Type
	transformer Transformer
}

// Register registers CRUD endpoints for a model type
func Register[T any](app *apiright.App, config Config) *CRUDHandler[T] {
	var zero T
	modelType := reflect.TypeOf(zero)
	
	// Set defaults
	if config.IDField == "" {
		config.IDField = "id"
	}
	
	if config.Route == "" {
		config.Route = strings.ToLower(modelType.Name()) + "s"
	}
	
	if config.Transformer == nil {
		config.Transformer = &DefaultTransformer{}
	}
	
	handler := &CRUDHandler[T]{
		app:         app,
		config:      config,
		modelType:   modelType,
		transformer: config.Transformer,
	}
	
	// Register routes
	handler.registerRoutes()
	
	return handler
}

// registerRoutes registers all CRUD routes
func (h *CRUDHandler[T]) registerRoutes() {
	apiGroup := h.app.APIGroup()
	route := "/" + h.config.Route
	
	if !h.config.WriteOnly {
		// GET /resource - List all
		apiGroup.HandleFunc(route, h.applyMiddleware(h.List)).Methods("GET")
		
		// GET /resource/{id} - Get by ID
		apiGroup.HandleFunc(route+"/{id}", h.applyMiddleware(h.Get)).Methods("GET")
	}
	
	if !h.config.ReadOnly {
		// POST /resource - Create
		apiGroup.HandleFunc(route, h.applyMiddleware(h.Create)).Methods("POST")
		
		// PUT /resource/{id} - Update
		apiGroup.HandleFunc(route+"/{id}", h.applyMiddleware(h.Update)).Methods("PUT")
		
		// DELETE /resource/{id} - Delete
		apiGroup.HandleFunc(route+"/{id}", h.applyMiddleware(h.Delete)).Methods("DELETE")
	}
}

// applyMiddleware applies middleware to a handler
func (h *CRUDHandler[T]) applyMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	var wrappedHandler http.Handler = handler
	
	for i := len(h.config.Middleware) - 1; i >= 0; i-- {
		wrappedHandler = h.config.Middleware[i](wrappedHandler)
	}
	
	return wrappedHandler.ServeHTTP
}

// List handles GET /resource
func (h *CRUDHandler[T]) List(w http.ResponseWriter, r *http.Request) {
	db := h.app.Database()
	if db == nil {
		apiright.ErrorResponse(w, "Database not configured", http.StatusInternalServerError)
		return
	}
	
	query := fmt.Sprintf("SELECT * FROM %s", h.config.TableName)
	rows, err := db.Query(query)
	if err != nil {
		apiright.ErrorResponse(w, "Failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	var results []T
	for rows.Next() {
		var item T
		if err := h.scanRow(rows, &item); err != nil {
			apiright.ErrorResponse(w, "Failed to scan row", http.StatusInternalServerError)
			return
		}
		results = append(results, item)
	}
	
	// Transform to API models
	var apiResults []interface{}
	for _, item := range results {
		apiItem, err := h.transformer.ToAPI(item)
		if err != nil {
			apiright.ErrorResponse(w, "Failed to transform model", http.StatusInternalServerError)
			return
		}
		apiResults = append(apiResults, apiItem)
	}
	
	apiright.SuccessResponse(w, apiResults)
}

// Get handles GET /resource/{id}
func (h *CRUDHandler[T]) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	db := h.app.Database()
	if db == nil {
		apiright.ErrorResponse(w, "Database not configured", http.StatusInternalServerError)
		return
	}
	
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = $1", h.config.TableName, h.config.IDField)
	row := db.QueryRow(query, id)
	
	var item T
	if err := h.scanRow(row, &item); err != nil {
		if err == sql.ErrNoRows {
			apiright.ErrorResponse(w, "Resource not found", http.StatusNotFound)
		} else {
			apiright.ErrorResponse(w, "Failed to query database", http.StatusInternalServerError)
		}
		return
	}
	
	// Transform to API model
	apiItem, err := h.transformer.ToAPI(item)
	if err != nil {
		apiright.ErrorResponse(w, "Failed to transform model", http.StatusInternalServerError)
		return
	}
	
	apiright.SuccessResponse(w, apiItem)
}

// Create handles POST /resource
func (h *CRUDHandler[T]) Create(w http.ResponseWriter, r *http.Request) {
	var apiItem interface{}
	if err := json.NewDecoder(r.Body).Decode(&apiItem); err != nil {
		apiright.ErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Transform from API model
	dbItem, err := h.transformer.FromAPI(apiItem)
	if err != nil {
		apiright.ErrorResponse(w, "Failed to transform model", http.StatusBadRequest)
		return
	}
	
	db := h.app.Database()
	if db == nil {
		apiright.ErrorResponse(w, "Database not configured", http.StatusInternalServerError)
		return
	}
	
	// Build insert query
	fields, values, placeholders := h.buildInsertQuery(dbItem)
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING %s", 
		h.config.TableName, fields, placeholders, h.config.IDField)
	
	var newID interface{}
	err = db.QueryRow(query, values...).Scan(&newID)
	if err != nil {
		apiright.ErrorResponse(w, "Failed to create resource", http.StatusInternalServerError)
		return
	}
	
	// Return the created resource
	h.getByID(w, newID)
}

// Update handles PUT /resource/{id}
func (h *CRUDHandler[T]) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	var apiItem interface{}
	if err := json.NewDecoder(r.Body).Decode(&apiItem); err != nil {
		apiright.ErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Transform from API model
	dbItem, err := h.transformer.FromAPI(apiItem)
	if err != nil {
		apiright.ErrorResponse(w, "Failed to transform model", http.StatusBadRequest)
		return
	}
	
	db := h.app.Database()
	if db == nil {
		apiright.ErrorResponse(w, "Database not configured", http.StatusInternalServerError)
		return
	}
	
	// Build update query
	setClause, values := h.buildUpdateQuery(dbItem)
	values = append(values, id)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = $%d", 
		h.config.TableName, setClause, h.config.IDField, len(values))
	
	result, err := db.Exec(query, values...)
	if err != nil {
		apiright.ErrorResponse(w, "Failed to update resource", http.StatusInternalServerError)
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		apiright.ErrorResponse(w, "Resource not found", http.StatusNotFound)
		return
	}
	
	// Return the updated resource
	h.getByID(w, id)
}

// Delete handles DELETE /resource/{id}
func (h *CRUDHandler[T]) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	db := h.app.Database()
	if db == nil {
		apiright.ErrorResponse(w, "Database not configured", http.StatusInternalServerError)
		return
	}
	
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", h.config.TableName, h.config.IDField)
	result, err := db.Exec(query, id)
	if err != nil {
		apiright.ErrorResponse(w, "Failed to delete resource", http.StatusInternalServerError)
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		apiright.ErrorResponse(w, "Resource not found", http.StatusNotFound)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// Helper methods

func (h *CRUDHandler[T]) getByID(w http.ResponseWriter, id interface{}) {
	db := h.app.Database()
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = $1", h.config.TableName, h.config.IDField)
	row := db.QueryRow(query, id)
	
	var item T
	if err := h.scanRow(row, &item); err != nil {
		apiright.ErrorResponse(w, "Failed to retrieve created resource", http.StatusInternalServerError)
		return
	}
	
	apiItem, err := h.transformer.ToAPI(item)
	if err != nil {
		apiright.ErrorResponse(w, "Failed to transform model", http.StatusInternalServerError)
		return
	}
	
	apiright.SuccessResponse(w, apiItem)
}

func (h *CRUDHandler[T]) scanRow(scanner interface{ Scan(...interface{}) error }, dest *T) error {
	// For this simplified implementation, we'll use reflection to scan into struct fields
	destValue := reflect.ValueOf(dest).Elem()
	destType := destValue.Type()
	
	// Create slice of pointers to struct fields
	fieldPointers := make([]interface{}, destType.NumField())
	for i := 0; i < destType.NumField(); i++ {
		fieldPointers[i] = destValue.Field(i).Addr().Interface()
	}
	
	return scanner.Scan(fieldPointers...)
}

func (h *CRUDHandler[T]) buildInsertQuery(item interface{}) (string, []interface{}, string) {
	itemValue := reflect.ValueOf(item)
	if itemValue.Kind() == reflect.Ptr {
		itemValue = itemValue.Elem()
	}
	itemType := itemValue.Type()
	
	var fields []string
	var values []interface{}
	var placeholders []string
	
	placeholderIndex := 1
	for i := 0; i < itemType.NumField(); i++ {
		field := itemType.Field(i)
		fieldValue := itemValue.Field(i)
		
		// Skip ID field for inserts (assuming auto-increment)
		if strings.ToLower(field.Name) == strings.ToLower(h.config.IDField) {
			continue
		}
		
		// Get database column name
		dbColumn := h.getDBColumnName(field)
		
		fields = append(fields, dbColumn)
		values = append(values, fieldValue.Interface())
		placeholders = append(placeholders, fmt.Sprintf("$%d", placeholderIndex))
		placeholderIndex++
	}
	
	return strings.Join(fields, ", "), values, strings.Join(placeholders, ", ")
}

func (h *CRUDHandler[T]) buildUpdateQuery(item interface{}) (string, []interface{}) {
	itemValue := reflect.ValueOf(item)
	if itemValue.Kind() == reflect.Ptr {
		itemValue = itemValue.Elem()
	}
	itemType := itemValue.Type()
	
	var setClauses []string
	var values []interface{}
	
	placeholderIndex := 1
	for i := 0; i < itemType.NumField(); i++ {
		field := itemType.Field(i)
		fieldValue := itemValue.Field(i)
		
		// Skip ID field for updates
		if strings.ToLower(field.Name) == strings.ToLower(h.config.IDField) {
			continue
		}
		
		// Get database column name
		dbColumn := h.getDBColumnName(field)
		
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", dbColumn, placeholderIndex))
		values = append(values, fieldValue.Interface())
		placeholderIndex++
	}
	
	return strings.Join(setClauses, ", "), values
}

func (h *CRUDHandler[T]) getDBColumnName(field reflect.StructField) string {
	// Check for custom field mapping
	if h.config.CustomFields != nil {
		if dbColumn, exists := h.config.CustomFields[field.Name]; exists {
			return dbColumn
		}
	}
	
	// Check for db tag
	if tag := field.Tag.Get("db"); tag != "" && tag != "-" {
		return tag
	}
	
	// Check for json tag as fallback
	if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
		return tag
	}
	
	// Default to lowercase field name
	return strings.ToLower(field.Name)
}