package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/pkg/apierror"
)

func LimitBody(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBytes <= 0 {
			c.Next()
			return
		}
		if c.Request.ContentLength > maxBytes {
			apierror.RequestTooLarge(c)
			c.Abort()
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
