package server

import (
	"fmt"
	"net/http"
	"strings"
)

func (s *DualServer) setupHTTPRoutes(mux *http.ServeMux) {
	s.logger.Info("Setting up HTTP routes",
		"total_services", len(s.services),
		"base_path", s.config.BasePath,
		"api_version", s.config.APIVersion,
	)

	for tableName, service := range s.services {
		serviceType := fmt.Sprintf("%T", service)

		s.logger.Info("Registering routes for service", "service_type", serviceType, "table", tableName)

		basePath := s.config.BasePath + "/" + s.config.APIVersion + "/" + tableName

		mux.HandleFunc(basePath, func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				s.handleListRoute(w, r, tableName)
			case http.MethodPost:
				s.handleCreateRoute(w, r, tableName)
			}
		})

		mux.HandleFunc(basePath+"/", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				s.handleGetRoute(w, r, tableName)
			case http.MethodPut:
				s.handleUpdateRoute(w, r, tableName)
			case http.MethodDelete:
				s.handleDeleteRoute(w, r, tableName)
			}
		})

		s.logger.Info("HTTP routes registered", "table", tableName, "base_path", basePath)
	}

	if len(s.services) == 0 {
		s.logger.Warn("No services registered, only default routes will be available")
	}
}

func (s *DualServer) registerHTTPService(tableName string, service any) error {
	serviceType := fmt.Sprintf("%T", service)
	s.logger.Info("Registering HTTP service", "service", serviceType, "table", tableName)

	s.registerCRUDRoutes(tableName, service)

	s.logger.Debug("HTTP service registered successfully", "service", serviceType, "table", tableName)
	return nil
}

func (s *DualServer) registerCRUDRoutes(tableName string, service any) {
	s.logger.Debug("CRUD routes placeholder registered", "table", tableName)
}

func (s *DualServer) toTableName(serviceName string) string {
	serviceName = strings.TrimSuffix(serviceName, "Service")

	var result []rune
	for i, r := range serviceName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		if r >= 'A' && r <= 'Z' {
			result = append(result, r+32)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func (s *DualServer) getSortedMiddleware() []func(http.Handler) http.Handler {
	return s.middlewareRegistry.GetHTTPMiddleware()
}
