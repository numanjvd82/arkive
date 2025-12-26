package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/models"
	"arkive/core/services/uploads"
	"arkive/core/web"
	"arkive/core/web/pages"
	appcontext "arkive/pkg/context"
)

func WebFiles(uploadService *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userValue, ok := c.Get(appcontext.UserKey)
		if !ok {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		user, ok := userValue.(models.User)
		if !ok || user.ID == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		files, err := uploadService.ListPendingUploads(c.Request.Context(), user.ID)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		web.Render(c, pages.FilesPage(files))
	}
}
