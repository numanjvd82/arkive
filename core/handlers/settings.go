package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/services/uploads"
	appcontext "arkive/pkg/context"
	"arkive/core/web"
	pages "arkive/core/web/pages"
)

func WebSettings(uploadService *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if err := uploadService.TouchUserActivity(c.Request.Context(), user.ID); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		web.Render(c, pages.SettingsPage(pages.SettingsPageProps{
			Ctx: pages.ContextWithUser(user),
		}))
	}
}
