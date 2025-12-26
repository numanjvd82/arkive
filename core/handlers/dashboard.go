package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/web"
	"arkive/core/web/pages"
	appcontext "arkive/pkg/context"
)

func WebDashboard() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		web.Render(c, pages.DashboardPage(pages.DashboardPageProps{
			Ctx: pages.ContextWithUser(user),
		}))
	}
}
