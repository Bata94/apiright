package core

import (
	"fmt"

	"github.com/bata94/apiright/pkg/logger"
)

type Middleware func(Handler) Handler

func LogMiddleware(logger logger.Logger) Middleware {
	return func(next Handler) Handler {
		return func(c *Ctx) error {
			// logger.Infof("[${time}] ${ip} ${status} - ${latency} ${method} ${path} ${error}\n")
			fmt.Printf("[%s] %s - %s\n", c.Request.Method, c.Request.RequestURI, c.Request.RemoteAddr)

			return next(c)
		}
	}
}
