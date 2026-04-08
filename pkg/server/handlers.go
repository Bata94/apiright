package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (s *DualServer) handleListRoute(w http.ResponseWriter, r *http.Request, tableName string) {
	contentType := s.detectContentType(r)

	serviceName := s.toServiceName(tableName)
	service, exists := s.serviceRegistry.GetService(serviceName)

	var response any
	var err error

	if exists {
		if serviceInterface, ok := service.(ServiceInterface); ok {
			limit := int32(50)
			offset := int32(0)

			if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
				if parsedLimit, parseErr := parseInt32(limitStr); parseErr == nil {
					limit = parsedLimit
				}
			}
			if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
				if parsedOffset, parseErr := parseInt32(offsetStr); parseErr == nil {
					offset = parsedOffset
				}
			}

			response, err = serviceInterface.List(r.Context(), limit, offset)
			if err != nil {
				s.handleServiceError(w, err, contentType)
				return
			}
		} else {
			s.logger.Warn("Service doesn't implement ServiceInterface", "service", serviceName)
			response = s.createMockResponse("list", tableName, contentType)
		}
	} else {
		response = s.createMockResponse("list", tableName, contentType)
		s.logger.Debug("No service found, using mock response", "table", tableName)
	}

	s.serializeResponse(w, response, contentType)
}

func (s *DualServer) handleGetRoute(w http.ResponseWriter, r *http.Request, tableName string) {
	contentType := s.detectContentType(r)

	id := s.extractIDFromPath(r.URL.Path, 2)
	if id == "" {
		s.handleServiceError(w, fmt.Errorf("missing ID in path"), contentType)
		return
	}

	serviceName := s.toServiceName(tableName)
	service, exists := s.serviceRegistry.GetService(serviceName)

	var response any
	var err error

	if exists {
		if serviceInterface, ok := service.(ServiceInterface); ok {
			response, err = serviceInterface.Get(r.Context(), id)
			if err != nil {
				s.handleServiceError(w, err, contentType)
				return
			}
		} else {
			response = s.createMockResponse("get", tableName, contentType)
		}
	} else {
		response = s.createMockResponse("get", tableName, contentType)
		s.logger.Debug("No service found, using mock response", "table", tableName, "id", id)
	}

	s.serializeResponse(w, response, contentType)
}

func (s *DualServer) handleCreateRoute(w http.ResponseWriter, r *http.Request, tableName string) {
	contentType := s.detectContentType(r)

	serviceName := s.toServiceName(tableName)
	service, exists := s.serviceRegistry.GetService(serviceName)

	var response any
	var err error

	if exists {
		if serviceInterface, ok := service.(ServiceInterface); ok {
			response, err = serviceInterface.Create(r.Context(), map[string]any{
				"request_body": "would_be_deserialized",
				"headers":      r.Header,
			})
			if err != nil {
				s.handleServiceError(w, err, contentType)
				return
			}
		} else {
			response = s.createMockResponse("create", tableName, contentType)
		}
	} else {
		response = s.createMockResponse("create", tableName, contentType)
		s.logger.Debug("No service found, using mock response", "table", tableName)
	}

	s.serializeResponse(w, response, contentType)
}

func (s *DualServer) handleUpdateRoute(w http.ResponseWriter, r *http.Request, tableName string) {
	contentType := s.detectContentType(r)

	id := s.extractIDFromPath(r.URL.Path, 2)
	if id == "" {
		s.handleServiceError(w, fmt.Errorf("missing ID in path"), contentType)
		return
	}

	serviceName := s.toServiceName(tableName)
	service, exists := s.serviceRegistry.GetService(serviceName)

	var response any
	var err error

	if exists {
		if serviceInterface, ok := service.(ServiceInterface); ok {
			params := map[string]any{
				"id":           id,
				"request_body": "would_be_deserialized",
				"headers":      r.Header,
			}
			response, err = serviceInterface.Update(r.Context(), params)
			if err != nil {
				s.handleServiceError(w, err, contentType)
				return
			}
		} else {
			response = s.createMockResponse("update", tableName, contentType)
		}
	} else {
		response = s.createMockResponse("update", tableName, contentType)
		s.logger.Debug("No service found, using mock response", "table", tableName, "id", id)
	}

	s.serializeResponse(w, response, contentType)
}

func (s *DualServer) handleDeleteRoute(w http.ResponseWriter, r *http.Request, tableName string) {
	contentType := s.detectContentType(r)

	id := s.extractIDFromPath(r.URL.Path, 2)
	if id == "" {
		s.handleServiceError(w, fmt.Errorf("missing ID in path"), contentType)
		return
	}

	serviceName := s.toServiceName(tableName)
	service, exists := s.serviceRegistry.GetService(serviceName)

	if exists {
		if serviceInterface, ok := service.(ServiceInterface); ok {
			err := serviceInterface.Delete(r.Context(), id)
			if err != nil {
				s.handleServiceError(w, err, contentType)
				return
			}
		}
	} else {
		s.logger.Debug("No service found, using mock response", "table", tableName, "id", id)
	}

	response := map[string]any{
		"message":   fmt.Sprintf("Delete %s successful", tableName),
		"operation": "delete",
		"id":        id,
		"success":   true,
	}

	s.serializeResponse(w, response, contentType)
}

func (s *DualServer) handleServiceError(w http.ResponseWriter, err error, contentType string) {
	s.logger.Error("Service error", "error", err)

	statusCode := http.StatusInternalServerError
	errorMsg := "Internal server error"

	if strings.Contains(err.Error(), "not found") {
		statusCode = http.StatusNotFound
		errorMsg = "Resource not found"
	} else if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "bad request") {
		statusCode = http.StatusBadRequest
		errorMsg = "Invalid request"
	}

	response := map[string]any{
		"error":   errorMsg,
		"message": err.Error(),
		"status":  statusCode,
	}

	s.serializeResponseWithStatus(w, response, contentType, statusCode)
}

func (s *DualServer) createMockResponse(operation, tableName, contentType string) any {
	opTitle := operation
	if len(operation) > 0 {
		opTitle = strings.ToUpper(string(operation[0])) + operation[1:]
	}
	return map[string]any{
		"message":   fmt.Sprintf("%s %s", opTitle, tableName),
		"operation": operation,
		"table":     tableName,
		"format":    contentType,
		"mock":      true,
	}
}

func (s *DualServer) detectContentType(r *http.Request) string {
	acceptHeader := r.Header.Get("Accept")
	contentType := s.contentNeg.DetectContentType(acceptHeader)
	if contentType == "" {
		contentType = "application/json"
	}
	return contentType
}

func (s *DualServer) extractIDFromPath(path string, index int) string {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) < index+1 {
		return ""
	}
	return pathParts[index]
}

func parseInt32(s string) (int32, error) {
	val, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(val), nil
}

func (s *DualServer) serializeResponseWithStatus(w http.ResponseWriter, response any, contentType string, statusCode int) {
	data, err := s.contentNeg.SerializeResponse(response, contentType)
	if err != nil {
		http.Error(w, "Serialization failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
	if _, err := w.Write(data); err != nil {
		s.logger.Warn("failed to write response", "error", err)
	}
}

func (s *DualServer) serializeResponse(w http.ResponseWriter, response any, contentType string) {
	data, err := s.contentNeg.SerializeResponse(response, contentType)
	if err != nil {
		http.Error(w, "Serialization failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		s.logger.Warn("failed to write response", "error", err)
	}
}

func (s *DualServer) toServiceName(tableName string) string {
	parts := strings.Split(tableName, "_")
	var result string
	for _, part := range parts {
		if part != "" && len(part) > 0 {
			result += strings.ToUpper(string(part[0])) + part[1:]
		}
	}
	return result + "Service"
}

func (s *DualServer) docsHandler(w http.ResponseWriter, r *http.Request) {
	data, err := docsFS.ReadFile("static/docs/index.html")
	if err != nil {
		http.Error(w, "Failed to load SwaggerUI", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		s.logger.Warn("failed to write docs response", "error", err)
	}
}

func (s *DualServer) docsOpenAPIHandler(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join(s.projectDir, "gen/openapi/openapi.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "OpenAPI documentation not found. Run 'apiright gen' first.", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to read OpenAPI documentation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		s.logger.Warn("failed to write docs response", "error", err)
	}
}

func (s *DualServer) docsCSSHandler(w http.ResponseWriter, r *http.Request) {
	data, err := docsFS.ReadFile("static/docs/swagger-ui.css")
	if err != nil {
		http.Error(w, "Failed to load CSS", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/css")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		s.logger.Warn("failed to write docs CSS response", "error", err)
	}
}

func (s *DualServer) docsJSHandler(w http.ResponseWriter, r *http.Request) {
	data, err := docsFS.ReadFile("static/docs/swagger-ui-bundle.js")
	if err != nil {
		http.Error(w, "Failed to load JavaScript", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/javascript")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		s.logger.Warn("failed to write docs JS response", "error", err)
	}
}

func (s *DualServer) docsStandalonePresetHandler(w http.ResponseWriter, r *http.Request) {
	data, err := docsFS.ReadFile("static/docs/swagger-ui-standalone-preset.js")
	if err != nil {
		http.Error(w, "Failed to load standalone preset", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/javascript")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		s.logger.Warn("failed to write docs standalone preset response", "error", err)
	}
}

func (s *DualServer) docsJSONRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.config.DocsPath, http.StatusMovedPermanently)
}
