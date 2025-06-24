package core

import (
	"fmt"

	"github.com/bata94/apiright/pkg/logger"
)

type Middleware func(Handler) Handler

func LogMiddleware(logger logger.Logger) Middleware {
	return func(next Handler) Handler {
		return func(c *Ctx) error {
			go func(c *Ctx) {
				select {
					case <-c.conClosed:
						duration := c.conEnded.Sub(c.conStarted)
						// TODO: use tabs and colors to make logs more appealing
						infoLog := fmt.Sprintf("[%d] <%d ms> | [%s] %s - %s\n", c.Response.StatusCode, duration.Microseconds(), c.Request.Method, c.Request.RequestURI, c.Request.RemoteAddr)

						if c.Response.StatusCode >= 400 {
							// TODO: add the error Msg here
							logger.Error(infoLog)
						} else {
							logger.Info(infoLog)
						}
				}
			} (c)

			return next(c)
		}
	}
}
