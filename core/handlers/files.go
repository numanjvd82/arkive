package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	filessvc "arkive/core/services/files"
	folderssvc "arkive/core/services/folders"
	"arkive/core/web"
	"arkive/core/web/pages"
	appcontext "arkive/pkg/context"
	"arkive/pkg/errs"
	"arkive/pkg/pagination"
)

const filesViewCookieName = "arkive_files_view"

func resolveFilesViewMode(c *gin.Context) string {
	viewMode := strings.TrimSpace(c.Query("view"))
	if viewMode == "grid" || viewMode == "list" {
		c.SetCookie(filesViewCookieName, viewMode, 60*60*24*365, "/", "", false, false)
		return viewMode
	}
	if cookieValue, err := c.Cookie(filesViewCookieName); err == nil {
		cookieValue = strings.TrimSpace(cookieValue)
		if cookieValue == "grid" || cookieValue == "list" {
			return cookieValue
		}
	}
	return "list"
}

func filesPageQuery(c *gin.Context) url.Values {
	query := c.Request.URL.Query()
	query.Del("path")
	query.Del("sort")
	return query
}

func WebFiles(filesService *filessvc.Service, folderService *folderssvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}
		if err := filesService.TouchUserActivity(c.Request.Context(), user.ID); err != nil {
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		page := pagination.ParsePageParam(c.DefaultQuery("page", "1"))
		pageSize := pagination.ParsePageSizeParam(c.DefaultQuery("pageSize", "25"))
		viewMode := resolveFilesViewMode(c)
		query := filesPageQuery(c)
		contents, err := folderService.ListEntries(c.Request.Context(), folderssvc.ListEntriesInput{
			UserID:   user.ID,
			FolderID: nil,
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		archivedCount, err := filesService.CountArchivedFiles(c.Request.Context(), user.ID)
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		web.Render(c, pages.FilesPage(pages.FilesPageProps{
			Ctx:           pages.ContextWithUser(user),
			Path:          contents.Path,
			Folders:       contents.Folders,
			Files:         contents.Files,
			Query:         query,
			ViewMode:      viewMode,
			CurrentFolder: nil,
			Page:          page,
			PageSize:      pageSize,
			TotalEntries:  contents.TotalEntries,
			ArchivedCount: archivedCount,
		}))
	}
}

func WebFolder(filesService *filessvc.Service, folderService *folderssvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}
		if err := filesService.TouchUserActivity(c.Request.Context(), user.ID); err != nil {
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		folderID := strings.TrimSpace(c.Param("id"))
		if folderID == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		page := pagination.ParsePageParam(c.DefaultQuery("page", "1"))
		pageSize := pagination.ParsePageSizeParam(c.DefaultQuery("pageSize", "25"))
		viewMode := resolveFilesViewMode(c)
		query := filesPageQuery(c)

		entries, err := folderService.ListEntries(c.Request.Context(), folderssvc.ListEntriesInput{
			UserID:   user.ID,
			FolderID: &folderID,
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			if err == folderssvc.ErrNotFound {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		archivedCount, err := filesService.CountArchivedFiles(c.Request.Context(), user.ID)
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		web.Render(c, pages.FilesPage(pages.FilesPageProps{
			Ctx:           pages.ContextWithUser(user),
			Path:          entries.Path,
			Folders:       entries.Folders,
			Files:         entries.Files,
			Query:         query,
			ViewMode:      viewMode,
			CurrentFolder: &folderID,
			Page:          page,
			PageSize:      pageSize,
			TotalEntries:  entries.TotalEntries,
			ArchivedCount: archivedCount,
		}))
	}
}

func WebFileView(filesService *filessvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}
		if err := filesService.TouchUserActivity(c.Request.Context(), user.ID); err != nil {
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		fileID := strings.TrimSpace(c.Param("id"))
		if fileID == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		file, err := filesService.GetFileForDisplay(c.Request.Context(), user.ID, fileID)
		if err != nil {
			switch err {
			case filessvc.ErrNotFound, filessvc.ErrUploadCancelled:
				c.AbortWithStatus(http.StatusNotFound)
			case filessvc.ErrUnauthorized, filessvc.ErrInvalidInput:
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
