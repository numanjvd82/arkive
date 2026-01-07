package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/services/uploads"
	"arkive/core/web"
	"arkive/core/web/pages"
	appcontext "arkive/pkg/context"
)

func WebSettings(uploadService *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		monthlyUsageBytes, err := uploadService.MonthlyUsage(c.Request.Context(), user.ID)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		fullSpeedLimit := uploads.MonthlyFullSpeedBytes
		if user.IsPremium {
			fullSpeedLimit = 0
		}

		web.Render(c, pages.SettingsPage(pages.SettingsPageProps{
			Ctx:               pages.ContextWithUser(user),
			MonthlyUsageBytes: monthlyUsageBytes,
			FullSpeedLimit:    fullSpeedLimit,
		}))
	}
}
