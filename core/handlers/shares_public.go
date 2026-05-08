package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"arkive/core/models"
	"arkive/core/services/shares"
	"arkive/core/services/uploads"
	"arkive/core/web"
	"arkive/core/web/pages"
	"arkive/pkg/errs"
)

func PublicShareView(shareService *shares.Service, uploadService *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.Param("token"))
		if token == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		share, err := shareService.GetShareByToken(c.Request.Context(), token)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if share.Status != shares.ShareStatusActive {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		if share.ExpiresAt != nil && !share.ExpiresAt.After(time.Now()) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		file, err := uploadService.GetFileForShare(c.Request.Context(), share.FileID)
		if err != nil {
			switch err {
			case uploads.ErrNotFound, uploads.ErrUploadCancelled:
				c.AbortWithStatus(http.StatusNotFound)
			case uploads.ErrInvalidInput:
				c.AbortWithStatus(http.StatusBadRequest)
			default:
				_ = c.Error(errs.WithStack(err))
				c.AbortWithStatus(http.StatusInternalServerError)
			}
			return
		}

		if share.PasswordHash != nil {
			web.Render(c, pages.PublicSharePassword(pages.PublicSharePasswordProps{
				Token: token,
				File:  file,
			}))
			return
		}

			renderShareLanding(c, uploadService, token, file)
	}
}

func PublicShareUnlock(shareService *shares.Service, uploadService *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.Param("token"))
		if token == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		share, err := shareService.GetShareByToken(c.Request.Context(), token)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if share.Status != shares.ShareStatusActive {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		if share.ExpiresAt != nil && !share.ExpiresAt.After(time.Now()) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		file, err := uploadService.GetFileForShare(c.Request.Context(), share.FileID)
		if err != nil {
			switch err {
			case uploads.ErrNotFound, uploads.ErrUploadCancelled:
				c.AbortWithStatus(http.StatusNotFound)
			case uploads.ErrInvalidInput:
				c.AbortWithStatus(http.StatusBadRequest)
			default:
				_ = c.Error(errs.WithStack(err))
				c.AbortWithStatus(http.StatusInternalServerError)
			}
			return
		}

		if share.PasswordHash == nil {
							renderShareLanding(c, uploadService, token, file)
			return
		}

		password := strings.TrimSpace(c.PostForm("password"))
		if password == "" || bcrypt.CompareHashAndPassword([]byte(*share.PasswordHash), []byte(password)) != nil {
			c.Status(http.StatusUnauthorized)
			web.Render(c, pages.PublicSharePassword(pages.PublicSharePasswordProps{
				Token:   token,
				File:    file,
				Message: "Password is incorrect.",
			}))
			return
		}

			renderShareLanding(c, uploadService, token, file)
	}
}

func renderShareLanding(c *gin.Context, uploadService *uploads.Service, token string, file models.File) {
	viewURL := ""
	isImage := false
	isVideo := false
	viewable := false
	downloadURL, err := uploadService.PresignShareDownloadForFile(c.Request.Context(), file)
	if err != nil {
		_ = c.Error(errs.WithStack(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	shareURL := buildShareURL(c, token)
	web.Render(c, pages.PublicShareViewPage(pages.PublicShareViewProps{
		File:        file,
		ViewURL:     viewURL,
		DownloadURL: downloadURL,
		IsImage:     isImage,
		IsVideo:     isVideo,
		Viewable:    viewable && viewURL != "",
		ShareURL:    shareURL,
	}))
}

func buildShareURL(c *gin.Context, token string) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if forwarded := c.GetHeader("X-Forwarded-Proto"); forwarded != "" {
		scheme = strings.ToLower(strings.TrimSpace(strings.Split(forwarded, ",")[0]))
	}
	host := c.Request.Host
	return scheme + "://" + host + "/s/" + token
}
