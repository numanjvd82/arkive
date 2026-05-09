package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"arkive/core/services/uploads"
	"arkive/core/web"
	"arkive/core/web/pages"
	appcontext "arkive/pkg/context"
	"arkive/pkg/errs"
	"arkive/pkg/pagination"
)

func WebFiles(uploadService *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}
		if err := uploadService.TouchUserActivity(c.Request.Context(), user.ID); err != nil {
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		page := pagination.ParsePageParam(c.DefaultQuery("page", "1"))
		pageSize := pagination.ParsePageSizeParam(c.DefaultQuery("pageSize", "25"))
		query := c.Request.URL.Query()
		query.Del("path")
		query.Del("sort")
		contents, err := uploadService.ListCompletedUploads(c.Request.Context(), user.ID, page, pageSize)
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		archivedCount, err := uploadService.CountArchivedFiles(c.Request.Context(), user.ID)
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		web.Render(c, pages.FilesPage(pages.FilesPageProps{
			Ctx:           pages.ContextWithUser(user),
			Files:         contents.Files,
			Query:         query,
			Page:          page,
			PageSize:      pageSize,
			TotalFiles:    contents.TotalFiles,
			ArchivedCount: archivedCount,
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
		if err := uploadService.TouchUserActivity(c.Request.Context(), user.ID); err != nil {
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		fileID := strings.TrimSpace(c.Param("id"))
		if fileID == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		file, err := uploadService.GetFileForDisplay(c.Request.Context(), user.ID, fileID)
		if err != nil {
			switch err {
			case uploads.ErrNotFound, uploads.ErrUploadCancelled:
				c.AbortWithStatus(http.StatusNotFound)
			case uploads.ErrUnauthorized, uploads.ErrInvalidInput:
				c.AbortWithStatus(http.StatusBadRequest)
			default:
				_ = c.Error(errs.WithStack(err))
				c.AbortWithStatus(http.StatusInternalServerError)
			}
			return
		}

		web.Render(c, pages.MediaViewPage(pages.MediaViewPageProps{
			Ctx:  pages.ContextWithUser(user),
			File: file,
		}))
	}
}
