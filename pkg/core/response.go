package core

import (
	"fmt"
	"net/http"
)

// Response is a generic response type.
type Response any

// ErrorResponse is the response for errors.
type ErrorResponse struct {
	Error   string `json:"error" xml:"error" description:"Error message"`
	Code    int    `json:"code" xml:"code" description:"Error code"`
	Details string `json:"details,omitempty" xml:"details,omitempty" description:"Additional error details"`
}

// ApiResponse is the standard API response.
type ApiResponse struct {
	Headers           map[string]string
	StatusCode        int    `json:"statusCode" xml:"statusCode"`
	InternalErrorCode int    ``
	Message           string `json:"message" xml:"message"`
	Data              []byte `json:"data,omitempty" xml:"data,omitempty"`
}

// NewApiResponse creates a new ApiResponse.
func NewApiResponse() *ApiResponse {
	return &ApiResponse{
		StatusCode: http.StatusOK,
	}
}

// AddHeader adds a header to the response.
func (r *ApiResponse) AddHeader(k, v string) {
	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}
	r.Headers[k] = v
}

// SendingReturn sends the response to the client.
func (c *Ctx) SendingReturn(w http.ResponseWriter, err error) {
	defer func() {
		c.Close()
	}()

	if err != nil {
		err = fmt.Errorf("error in HanlderFunc: %w", err)
		log.Errorf("handler error: %v", err)
		c.Response.SetMessage(err.Error())
		// Only set status to 500 if no status has been set yet (still default 200)
		if c.Response.StatusCode == http.StatusOK {
			c.Response.SetStatus(http.StatusInternalServerError)
		}
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
		_, err = w.Write([]byte(c.Response.Message))
	} else {
		_, err = w.Write(c.Response.Data)
	}

	if err != nil {
		log.Errorf("error writing body: %v", err)
		panic(err)
	}
}

// SetStatus sets the status code of the response.
func (r *ApiResponse) SetStatus(code int) {
	r.StatusCode = code
}

// SetMessage sets the message of the response.
func (r *ApiResponse) SetMessage(msg string) {
	r.Message = msg
}

// SetMessage sets the message of the response.
func (r *ApiResponse) SetMessagef(msg string, a ...any) {
	r.Message = fmt.Sprintf(msg, a...)
}

// SetData sets the data of the response.
func (r *ApiResponse) SetData(data []byte) {
	r.Data = data
}
