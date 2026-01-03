package middleware

import (
	"log"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

type stackProvider interface {
	Stack() []byte
}

func ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		status := c.Writer.Status()
		if len(c.Errors) == 0 {
			if status >= 500 {
				log.Printf("unhandled %d: %s %s\n%s", status, c.Request.Method, c.Request.URL.Path, debug.Stack())
			}
			return
		}

		for _, err := range c.Errors {
			stack := debug.Stack()
			if provider, ok := err.Err.(stackProvider); ok {
				stack = provider.Stack()
			}
			log.Printf("request error: %s %s: %v\n%s", c.Request.Method, c.Request.URL.Path, err.Err, stack)
		}
	}
}
