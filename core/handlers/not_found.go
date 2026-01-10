package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/web"
	"arkive/core/web/pages"
)

func WebNotFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Status(http.StatusNotFound)
		web.Render(c, pages.NotFoundPage(pages.NotFoundPageProps{
			Ctx:  pages.PageContext{},
			Path: c.Request.URL.Path,
		}))
	}
}
