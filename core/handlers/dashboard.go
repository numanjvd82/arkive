package handlers

import (
	"github.com/gin-gonic/gin"

	"arkive/core/web"
	"arkive/core/web/pages"
)

func WebDashboard() gin.HandlerFunc {
	return func(c *gin.Context) {
		web.Render(c, pages.DashboardPage())
	}
}
