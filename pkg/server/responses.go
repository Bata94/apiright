package server

import (
	"net/http"
	"time"

	"github.com/bata94/apiright/pkg/core"
)

func (s *DualServer) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contentType := s.detectContentType(r)

	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   core.Version,
	}

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

func (s *DualServer) handleDefaultRoute(w http.ResponseWriter, r *http.Request) {
	contentType := s.detectContentType(r)

	response := map[string]interface{}{
		"message":  "APIRight server is running",
		"version":  core.Version,
		"services": len(s.services),
	}

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
