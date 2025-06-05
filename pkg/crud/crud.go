package crud

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// Model represents a database model with an ID field
type Model interface {
	GetID() int64
	SetID(int64)
}

// Repository provides CRUD operations for a specific model type
type Repository[T Model] struct {
	db        *sql.DB
	tableName string
	modelType reflect.Type
}

// NewRepository creates a new repository for the given model type
func NewRepository[T Model](db *sql.DB, tableName string) *Repository[T] {
	var zero T
	return &Repository[T]{
		db:        db,
		tableName: tableName,
		modelType: reflect.TypeOf(zero).Elem(),
	}
}

// CRUDHandler provides HTTP handlers for CRUD operations
type CRUDHandler[T Model] struct {
	repo *Repository[T]
}

// NewCRUDHandler creates a new CRUD handler
func NewCRUDHandler[T Model](repo *Repository[T]) *CRUDHandler[T] {
	return &CRUDHandler[T]{repo: repo}
}

// Create handles POST requests to create a new entity
func (h *CRUDHandler[T]) Create(w http.ResponseWriter, r *http.Request) {
	var entity T
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	createdEntity, err := h.repo.Create(entity)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create entity: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdEntity)
}

// GetAll handles GET requests to retrieve all entities
func (h *CRUDHandler[T]) GetAll(w http.ResponseWriter, r *http.Request) {
	entities, err := h.repo.GetAll()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve entities: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}

// GetByID handles GET requests to retrieve a specific entity by ID
func (h *CRUDHandler[T]) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	entity, err := h.repo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Entity not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to retrieve entity: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// Update handles PUT requests to update an entity
func (h *CRUDHandler[T]) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var entity T
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	entity.SetID(id)
	updatedEntity, err := h.repo.Update(entity)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update entity: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedEntity)
}

// Delete handles DELETE requests to remove an entity
func (h *CRUDHandler[T]) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
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

// Repository methods

// Create inserts a new entity into the database
func (r *Repository[T]) Create(entity T) (T, error) {
	var zero T
	
	// Build INSERT query using reflection
	fields, values, placeholders := r.getFieldsAndValues(entity, true) // exclude ID for insert
	
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id",
		r.tableName, strings.Join(fields, ", "), strings.Join(placeholders, ", "))
	
	var id int64
	err := r.db.QueryRow(query, values...).Scan(&id)
	if err != nil {
		return zero, err
	}
	
	entity.SetID(id)
	return entity, nil
}

// GetAll retrieves all entities from the database
func (r *Repository[T]) GetAll() ([]T, error) {
	query := fmt.Sprintf("SELECT * FROM %s", r.tableName)
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []T
	for rows.Next() {
		entity, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}

	return entities, rows.Err()
}

// GetByID retrieves a specific entity by ID
func (r *Repository[T]) GetByID(id int64) (T, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", r.tableName)
	row := r.db.QueryRow(query, id)
	
	return r.scanRow(row)
}

// Update modifies an existing entity in the database
func (r *Repository[T]) Update(entity T) (T, error) {
	fields, values, placeholders := r.getFieldsAndValues(entity, true) // exclude ID
	
	// Build SET clause
	var setPairs []string
	for i, field := range fields {
		setPairs = append(setPairs, fmt.Sprintf("%s = %s", field, placeholders[i]))
	}
	
	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d RETURNING *",
		r.tableName, strings.Join(setPairs, ", "), len(values)+1)
	
	values = append(values, entity.GetID())
	row := r.db.QueryRow(query, values...)
	
	return r.scanRow(row)
}

// Delete removes an entity from the database
func (r *Repository[T]) Delete(id int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", r.tableName)
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	
	return nil
}

// Helper methods

// getFieldsAndValues extracts field names and values from a struct using reflection
func (r *Repository[T]) getFieldsAndValues(entity T, excludeID bool) ([]string, []interface{}, []string) {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	var fields []string
	var values []interface{}
	var placeholders []string
	
	for i := 0; i < v.NumField(); i++ {
		field := r.modelType.Field(i)
		value := v.Field(i)
		
		// Get database column name from tag or use field name
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			dbTag = strings.ToLower(field.Name)
		}
		
		// Skip ID field if requested
		if excludeID && (dbTag == "id" || strings.ToLower(field.Name) == "id") {
			continue
		}
		
		fields = append(fields, dbTag)
		values = append(values, value.Interface())
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(values)))
	}
	
	return fields, values, placeholders
}

// scanRow scans a database row into a struct
func (r *Repository[T]) scanRow(scanner interface {
	Scan(dest ...interface{}) error
}) (T, error) {
	var zero T
	entity := reflect.New(r.modelType).Interface().(T)
	
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	// Create slice of pointers to struct fields for scanning
	var dest []interface{}
	for i := 0; i < v.NumField(); i++ {
		dest = append(dest, v.Field(i).Addr().Interface())
	}
	
	if err := scanner.Scan(dest...); err != nil {
		return zero, err
	}
	
	return entity, nil
}

// RegisterRoutes registers CRUD routes for the handler
func (h *CRUDHandler[T]) RegisterRoutes(router *mux.Router, basePath string) {
	router.HandleFunc(basePath, h.Create).Methods("POST")
	router.HandleFunc(basePath, h.GetAll).Methods("GET")
	router.HandleFunc(basePath+"/{id}", h.GetByID).Methods("GET")
	router.HandleFunc(basePath+"/{id}", h.Update).Methods("PUT")
	router.HandleFunc(basePath+"/{id}", h.Delete).Methods("DELETE")
}