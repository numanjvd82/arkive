package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/services/uploads"
	"arkive/core/web"
	"arkive/core/web/pages"
	appcontext "arkive/pkg/context"
	"arkive/pkg/format"
)

func WebSettings(uploadService *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if err := uploadService.TouchUserActivity(c.Request.Context(), user.ID, user.IsPremium); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		fileCount, err := uploadService.CountActiveFiles(c.Request.Context(), user.ID)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		fileLimitLabel := format.Commas(uploads.FreeFileLimit)
		if user.IsPremium {
			fileLimitLabel = "Unlimited"
		}

		web.Render(c, pages.SettingsPage(pages.SettingsPageProps{
			Ctx:            pages.ContextWithUser(user),
			FileCount:      fileCount,
			FileLimitLabel: fileLimitLabel,
		}))
	}
}
