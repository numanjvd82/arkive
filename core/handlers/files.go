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
	"arkive/pkg/video"
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

		folderPath := uploads.NormalizeFolderPath(c.Query("path"))
		sort := strings.TrimSpace(c.DefaultQuery("sort", "updated_desc"))
		page := pagination.ParsePageParam(c.DefaultQuery("page", "1"))
		pageSize := pagination.ParsePageSizeParam(c.DefaultQuery("pageSize", "25"))
		contents, err := uploadService.ListFolderContents(c.Request.Context(), user.ID, folderPath, sort, page, pageSize)
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
			FolderPath:    folderPath,
			Folders:       contents.Folders,
			Files:         contents.Files,
			Query:         c.Request.URL.Query(),
			Sort:          sort,
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

		contentType := strings.ToLower(strings.TrimSpace(file.ContentType))
		isImage := strings.HasPrefix(contentType, "image/")
		isVideo := strings.HasPrefix(contentType, "video/")
		largeVideo := isVideo && video.IsLarge(
			file.SizeBytes,
			file.VideoDurationSeconds,
			file.VideoWidth,
			file.VideoHeight,
		)

		viewURL := ""
		if isImage || isVideo {
			viewURL = "/api/files/" + file.ID + "/media"
		}

		web.Render(c, pages.MediaViewPage(pages.MediaViewPageProps{
			Ctx:      pages.ContextWithUser(user),
			File:     file,
			ViewURL:  viewURL,
			IsImage:  isImage,
			IsVideo:  isVideo,
			Viewable: isImage || isVideo,
			Large:    largeVideo,
		}))
	}
}
