package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/services/auth"
	"arkive/core/web"
	"arkive/core/web/pages"
	appcontext "arkive/pkg/context"
)

func WebDashboard(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ok, err := appcontext.LoadUser(c, svc)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		if !ok {
			c.Status(http.StatusForbidden)
			web.Render(c, pages.ForbiddenPage())
			return
		}

		web.Render(c, pages.DashboardPage())
	}
}
