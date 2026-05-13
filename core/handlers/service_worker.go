package handlers

import (
	"github.com/gin-gonic/gin"

	"arkive/core/web"
)

func ServiceWorkerJS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/javascript; charset=utf-8")
		c.Header("Cache-Control", "no-cache")
		c.FileFromFS("sw.js", web.StaticFS("static"))
	}
}
