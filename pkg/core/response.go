package core

import (
	"fmt"
	"net/http"
)

type Response any

type ErrorResponse struct {
	Error   string `json:"error" xml:"error" description:"Error message"`
	Code    int    `json:"code" xml:"code" description:"Error code"`
	Details string `json:"details,omitempty" xml:"details,omitempty" description:"Additional error details"`
}

type ApiResponse struct {
	Headers           map[string]string
	StatusCode        int    `json:"statusCode" xml:"statusCode"`
	InternalErrorCode int    ``
	Message           string `json:"message" xml:"message"`
	Data              []byte `json:"data,omitempty" xml:"data,omitempty"`
}

func NewApiResponse() ApiResponse {
	return ApiResponse{
		StatusCode: http.StatusOK,
	}
}

func (r *ApiResponse) AddHeader(k, v string) {
	r.Headers = map[string]string{
		k: v,
	}
}

func (r *ApiResponse) SendingReturn(w http.ResponseWriter, c *Ctx, err error) {
	if err != nil {
		err = fmt.Errorf("error in HanlderFunc: %w", err)
		log.Errorf("handler error: %v", err)
		c.Response.SetMessage(err.Error())
		c.Response.SetStatus(http.StatusInternalServerError)
	}

	for k, v := range c.Response.Headers {
		if _, ok := w.Header()[k]; !ok {
			w.Header().Add(k, v)
		} else {
			w.Header().Set(k, v)
		}
	}

	w.WriteHeader(c.Response.StatusCode)

	if c.Response.Data == nil {
		_, _ = w.Write([]byte(c.Response.Message))
	} else {
		_, _ = w.Write(c.Response.Data)
	}

	c.Close()
}

func (r *ApiResponse) SetStatus(code int) {
	r.StatusCode = code
}

func (r *ApiResponse) SetMessage(msg string) {
	r.Message = msg
}

func (r *ApiResponse) SetData(data []byte) {
	r.Data = data
}
