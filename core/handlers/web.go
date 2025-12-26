package handlers

import (
	"arkive/core/web"
	pages "arkive/core/web/pages"

	"github.com/gin-gonic/gin"
)

func WebHome() gin.HandlerFunc {
	return func(c *gin.Context) {
		web.Render(c, pages.HomePage(pages.HomePageProps{Ctx: pages.PageContext{}}))
	}
}
