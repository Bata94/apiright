package core

import (
	"fmt"
	"net/http"
)

type Response any

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
		err = fmt.Errorf("Error in HanlderFunc: %w", err)
		log.Errorf("Handler error: %v", err)
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
		w.Write([]byte(c.Response.Message))
	} else {
		w.Write(c.Response.Data)
	}
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
