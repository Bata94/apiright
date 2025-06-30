package core

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewApiResponse(t *testing.T) {
	resp := NewApiResponse()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected default status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if resp.Headers != nil {
		t.Errorf("Expected headers to be nil, got %v", resp.Headers)
	}
}

func TestApiResponse_AddHeader(t *testing.T) {
	resp := NewApiResponse()
	resp.AddHeader("Content-Type", "application/json")

	if resp.Headers == nil {
		t.Fatal("Headers map was not initialized")
	}
	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type header to be application/json, got %s", resp.Headers["Content-Type"])
	}

	resp.AddHeader("X-Custom-Header", "test-value")
	if resp.Headers["X-Custom-Header"] != "test-value" {
		t.Errorf("Expected X-Custom-Header to be test-value, got %s", resp.Headers["X-Custom-Header"])
	}
}

func TestApiResponse_SetStatus(t *testing.T) {
	resp := NewApiResponse()
	resp.SetStatus(http.StatusNotFound)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestApiResponse_SetMessage(t *testing.T) {
	resp := NewApiResponse()
	resp.SetMessage("Test Message")

	if resp.Message != "Test Message" {
		t.Errorf("Expected message \"Test Message\", got %s", resp.Message)
	}
	// If status code was not set before, it should be 200
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	resp = NewApiResponse()
	resp.SetStatus(http.StatusAccepted)
	resp.SetMessage("Another Message")
	// If status code was set before, it should not change
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("Expected status code %d, got %d", http.StatusAccepted, resp.StatusCode)
	}
}

func TestApiResponse_SetData(t *testing.T) {
	resp := NewApiResponse()
	data := []byte("{\"key\":\"value\"}")
	resp.SetData(data)

	if !bytes.Equal(resp.Data, data) {
		t.Errorf("Expected data %s, got %s", string(data), string(resp.Data))
	}
}

func TestApiResponse_SendingReturn(t *testing.T) {
	var rec *httptest.ResponseRecorder
	var req *http.Request

	// Test case 1: Successful response with message
	t.Run("SuccessWithMessage", func(t *testing.T) {
		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := NewCtx(rec, req)
		ctx.Response.SetStatus(http.StatusCreated)
		ctx.Response.SetMessage("Resource created")
		ctx.Response.AddHeader("X-Test-Header", "test-value")

		ctx.SendingReturn(rec, nil)

		if rec.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
		}
		if rec.Body.String() != "Resource created" {
			t.Errorf("Expected body \"Resource created\", got %s", rec.Body.String())
		}
		if rec.Header().Get("X-Test-Header") != "test-value" {
			t.Errorf("Expected X-Test-Header \"test-value\", got %s", rec.Header().Get("X-Test-Header"))
		}
	})

	// Test case 2: Successful response with data
	t.Run("SuccessWithData", func(t *testing.T) {
		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := NewCtx(rec, req)
		ctx.Response.SetStatus(http.StatusOK)
		data := map[string]string{"status": "ok"}
		jsonData, _ := json.Marshal(data)
		ctx.Response.SetData(jsonData)

		ctx.SendingReturn(rec, nil)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if !bytes.Equal(rec.Body.Bytes(), jsonData) {
			t.Errorf("Expected body %s, got %s", string(jsonData), rec.Body.String())
		}
	})

	// Test case 3: Error in handler
	t.Run("ErrorInHandler", func(t *testing.T) {
						rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := NewCtx(rec, req)
		ctx.Response.SetStatus(http.StatusOK) // Ensure default status
		// Simulate an error from the handler
		err := http.ErrBodyReadAfterClose

		ctx.SendingReturn(rec, err)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
		if rec.Body.String() != "error in HanlderFunc: http: invalid Read on closed Body" {
			t.Errorf("Expected error message, got %s", rec.Body.String())
		}
	})

	// Test case 4: Error in handler with pre-set status
	t.Run("ErrorInHandlerWithPresetStatus", func(t *testing.T) {
		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := NewCtx(rec, req)
		ctx.Response.SetStatus(http.StatusBadRequest) // Pre-set status
		// Simulate an error from the handler
		err := http.ErrBodyReadAfterClose

		ctx.SendingReturn(rec, err)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
		if rec.Body.String() != "error in HanlderFunc: http: invalid Read on closed Body" {
			t.Errorf("Expected error message, got %s", rec.Body.String())
		}
	})
}
