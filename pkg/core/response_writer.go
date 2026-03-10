package core

import (
	"net/http"
)

type ResponseWriter struct {
	w http.ResponseWriter
}

func (rw *ResponseWriter) Header() http.Header {
	return rw.w.Header()
}

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.w.WriteHeader(statusCode)
}

func (rw *ResponseWriter) Write(p []byte) (int, error) {
	return rw.w.Write(p)
}

func (rw *ResponseWriter) Unwrap() http.ResponseWriter {
	return rw.w
}

func (rw *ResponseWriter) Flush() {
	if f, ok := rw.w.(http.Flusher); ok {
		f.Flush()
	}
}

func (rw *ResponseWriter) Hijack() error {
	if h, ok := rw.w.(http.Hijacker); ok {
		_, _, err := h.Hijack()
		return err
	}
	return nil
}
