package server

import (
	"context"
	"fmt"

	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/database"
)

// ServiceRegistry manages loading and registration of generated services
type ServiceRegistry struct {
	db       *database.Database
	logger   core.Logger
	services map[string]interface{}
	queriers map[string]interface{}
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(db *database.Database, logger core.Logger) *ServiceRegistry {
	return &ServiceRegistry{
		db:       db,
		logger:   logger,
		services: make(map[string]interface{}),
		queriers: make(map[string]interface{}),
	}
}

// LoadGeneratedServices loads all generated services from gen/go/services
func (sr *ServiceRegistry) LoadGeneratedServices(projectDir string) error {
	sr.logger.Info("Loading generated services", "project", projectDir)

	// For now, we'll create services manually since we can't dynamically import Go packages
	// In a full implementation, this would use Go plugins or reflection to load services

	// Create querier from database connection
	querier, err := sr.createQuerier()
	if err != nil {
		return fmt.Errorf("failed to create querier: %w", err)
	}

	// Create services for each table (this would be auto-generated in a full implementation)
	// For now, we'll create a simple service factory pattern

	sr.queriers["db"] = querier
	sr.logger.Info("Created database querier", "type", fmt.Sprintf("%T", querier))

	return nil
}

// createQuerier creates a sqlc querier from the database connection
func (sr *ServiceRegistry) createQuerier() (interface{}, error) {
	// Get the underlying database connection
	db := sr.db.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// Import the generated db package and create querier
	// Since we can't dynamically import, we'll use a factory pattern
	// In a full implementation, this would be auto-generated

	// For now, return a mock querier that will be replaced by actual generated code
	return &mockQuerier{}, nil
}

// GetService returns a service by name
func (sr *ServiceRegistry) GetService(name string) (interface{}, bool) {
	service, exists := sr.services[name]
	return service, exists
}

// GetQuerier returns a querier by name
func (sr *ServiceRegistry) GetQuerier(name string) (interface{}, bool) {
	querier, exists := sr.queriers[name]
	return querier, exists
}

// RegisterService manually registers a service (for testing or custom services)
func (sr *ServiceRegistry) RegisterService(name string, service interface{}) {
	sr.services[name] = service
	sr.logger.Info("Manually registered service", "name", name, "type", fmt.Sprintf("%T", service))
}

// CreateServiceFactory creates a service for a given table
func (sr *ServiceRegistry) CreateServiceFactory(tableName string) (interface{}, error) {
	sr.logger.Info(fmt.Sprintf("CreateServiceFactory ENTRY for table: %s", tableName))

	// Simple return for testing
	sr.logger.Info(fmt.Sprintf("CreateServiceFactory EXIT for table: %s", tableName))
	return &mockService{
		tableName: tableName,
		db:        sr.db,
		logger:    sr.logger,
	}, nil
}

// mockQuerier provides a mock implementation for testing
type mockQuerier struct{}

// mockService provides a mock service implementation for testing
type mockService struct {
	tableName string
	db        *database.Database
	logger    core.Logger
}

// Get implements the Get method for mock service
func (ms *mockService) Get(ctx context.Context, id interface{}) (interface{}, error) {
	ms.logger.Info("Mock Get called", "table", ms.tableName, "id", id)

	// Return mock data
	return map[string]interface{}{
		"id":         id,
		"name":       fmt.Sprintf("Mock %s", ms.tableName),
		"created_at": "2026-01-28T12:00:00Z",
		"mock":       true,
	}, nil
}

// List implements the List method for mock service
func (ms *mockService) List(ctx context.Context, limit, offset int32) (interface{}, error) {
	ms.logger.Info("Mock List called", "table", ms.tableName, "limit", limit, "offset", offset)

	// Return mock data
	return []interface{}{
		map[string]interface{}{
			"id":         1,
			"name":       fmt.Sprintf("Mock %s 1", ms.tableName),
			"created_at": "2026-01-28T12:00:00Z",
			"mock":       true,
		},
		map[string]interface{}{
			"id":         2,
			"name":       fmt.Sprintf("Mock %s 2", ms.tableName),
			"created_at": "2026-01-28T12:01:00Z",
			"mock":       true,
		},
	}, nil
}

// Create implements the Create method for mock service
func (ms *mockService) Create(ctx context.Context, params interface{}) (interface{}, error) {
	ms.logger.Info("Mock Create called", "table", ms.tableName, "params", params)

	// Return mock created data
	return map[string]interface{}{
		"id":         999,
		"name":       fmt.Sprintf("Created %s", ms.tableName),
		"created_at": "2026-01-28T12:00:00Z",
		"mock":       true,
		"created":    true,
	}, nil
}

// Update implements the Update method for mock service
func (ms *mockService) Update(ctx context.Context, params interface{}) (interface{}, error) {
	ms.logger.Info("Mock Update called", "table", ms.tableName, "params", params)

	// Return mock updated data
	return map[string]interface{}{
		"id":         999,
		"name":       fmt.Sprintf("Updated %s", ms.tableName),
		"updated_at": "2026-01-28T12:00:00Z",
		"mock":       true,
		"updated":    true,
	}, nil
}

// Delete implements the Delete method for mock service
func (ms *mockService) Delete(ctx context.Context, id interface{}) error {
	ms.logger.Info("Mock Delete called", "table", ms.tableName, "id", id)
	return nil
}

// ServiceInterface defines the common interface for all services
type ServiceInterface interface {
	Get(ctx context.Context, id interface{}) (interface{}, error)
	List(ctx context.Context, limit, offset int32) (interface{}, error)
	Create(ctx context.Context, params interface{}) (interface{}, error)
	Update(ctx context.Context, params interface{}) (interface{}, error)
	Delete(ctx context.Context, id interface{}) error
}

// Ensure mockService implements ServiceInterface
var _ ServiceInterface = (*mockService)(nil)
