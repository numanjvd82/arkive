package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"arkive/core/services/uploads"
	"arkive/core/web"
	"arkive/core/web/pages"
	appcontext "arkive/pkg/context"
)

func WebFiles(uploadService *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		files, err := uploadService.ListCompletedUploads(c.Request.Context(), user.ID)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		web.Render(c, pages.FilesPage(pages.FilesPageProps{
			Ctx:   pages.ContextWithUser(user),
			Files: files,
		}))
	}
}

func WebFileView(uploadService *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		fileID := strings.TrimSpace(c.Param("id"))
		if fileID == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		file, viewURL, err := uploadService.GetFileForView(c.Request.Context(), user.ID, fileID)
		if err != nil {
			switch err {
			case uploads.ErrNotFound, uploads.ErrUploadCancelled:
				c.AbortWithStatus(http.StatusNotFound)
			case uploads.ErrUnauthorized, uploads.ErrInvalidInput:
				c.AbortWithStatus(http.StatusBadRequest)
			default:
				c.AbortWithStatus(http.StatusInternalServerError)
			}
			return
		}

		contentType := strings.ToLower(strings.TrimSpace(file.ContentType))
		isImage := strings.HasPrefix(contentType, "image/")
		isVideo := strings.HasPrefix(contentType, "video/")

		web.Render(c, pages.MediaViewPage(pages.MediaViewPageProps{
			Ctx:      pages.ContextWithUser(user),
			File:     file,
			ViewURL:  viewURL,
			IsImage:  isImage,
			IsVideo:  isVideo,
			Viewable: isImage || isVideo,
		}))
	}
}
