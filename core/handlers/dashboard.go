package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"arkive/core/models"
	filessvc "arkive/core/services/files"
	folderssvc "arkive/core/services/folders"
	settingssvc "arkive/core/services/settings"
	"arkive/core/web"
	"arkive/core/web/pages"
	appcontext "arkive/pkg/context"
	"arkive/pkg/errs"
)

func WebDashboard(filesService *filessvc.Service, folderService *folderssvc.Service, settingsService *settingssvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if err := filesService.TouchUserActivity(c.Request.Context(), user.ID); err != nil {
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		list, err := filesService.ListCompletedUploads(c.Request.Context(), user.ID, 1, 4)
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		storageSettings, err := settingsService.StorageSettings(c.Request.Context())
		if err != nil {
			storageSettings = models.StorageSettings{}
		}

		var currentFolder *string
		if folderID := strings.TrimSpace(c.Query("folder")); folderID != "" {
			if _, err := folderService.ValidateFolderAccess(c.Request.Context(), user.ID, folderID); err == nil {
				currentFolder = &folderID
			}
		}

		web.Render(c, pages.DashboardPage(pages.DashboardPageProps{
			Ctx:             pages.ContextWithUser(user),
			RecentFiles:     list.Files,
			TotalFiles:      list.TotalFiles,
			CurrentFolder:   currentFolder,
			StorageSettings: storageSettings,
		}))
	}
}
