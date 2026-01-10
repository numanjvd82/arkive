package handlers

import (
	"github.com/gin-gonic/gin"

	"arkive/core/web"
)

func RobotsTxt() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.FileFromFS("robots.txt", web.StaticFS("static"))
	}
}

func SitemapXML() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/xml; charset=utf-8")
		c.FileFromFS("sitemap.xml", web.StaticFS("static"))
	}
}
