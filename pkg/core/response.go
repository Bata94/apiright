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
	defer func() {
		log.Debug("Closing Connection")
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

	log.Debug("Adding/Setting Return Headers")
	for k, v := range c.Response.Headers {
		if _, ok := w.Header()[k]; !ok {
			w.Header().Add(k, v)
		} else {
			w.Header().Set(k, v)
		}
	}

	log.Debug("WriteHeaders")
	w.WriteHeader(c.Response.StatusCode)

	log.Debug("Writing Body")
	if c.Response.Data == nil {
		log.Debug("Write MSG")
		_, err = w.Write([]byte(c.Response.Message))
	} else {
		log.Debug("Write Data")
		_, err = w.Write(c.Response.Data)
	}

	if err != nil {
    log.Errorf("error writing body: %v", err)
		panic(err)
  }
}

func (r *ApiResponse) SetStatus(code int) {
	r.StatusCode = code
}

func (r *ApiResponse) SetMessage(msg string) {
	r.Message = msg
	if r.StatusCode == 0 {
		log.Debug("StatusCode not set before Message, setting to 200")
		r.StatusCode = 200
	}
}

func (r *ApiResponse) SetData(data []byte) {
	r.Data = data
}
