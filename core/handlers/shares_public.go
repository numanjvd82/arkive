package handlers

import (
	"encoding/base64"
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

		renderShareLanding(c, uploadService, token, share, file)
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
			renderShareLanding(c, uploadService, token, share, file)
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

		renderShareLanding(c, uploadService, token, share, file)
	}
}

func APIPublicShareRecord(shareService *shares.Service, uploadService *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.Param("token"))
		if token == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
			return
		}

		share, err := shareService.GetShareByToken(c.Request.Context(), token)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
				return
			}
			_ = c.Error(errs.WithStack(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "share lookup failed"})
			return
		}
		if share.Status != shares.ShareStatusActive {
			c.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
			return
		}
		if share.PasswordHash != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "share access denied"})
			return
		}
		if share.ExpiresAt != nil && !share.ExpiresAt.After(time.Now()) {
			c.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
			return
		}

		file, err := uploadService.GetFileForShare(c.Request.Context(), share.FileID)
		if err != nil {
			switch err {
			case uploads.ErrNotFound, uploads.ErrUploadCancelled:
				c.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
			case uploads.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "share lookup failed"})
			}
			return
		}

		sourceURL, err := uploadService.PresignShareSourceForFile(c.Request.Context(), file)
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "share source failed"})
			return
		}

		record, err := shareService.GetPublicShareRecord(c.Request.Context(), token)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
				return
			}
			_ = c.Error(errs.WithStack(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "share record failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"shareId":                  record.ShareID,
			"token":                    record.Token,
			"fileId":                   record.FileID,
			"vaultId":                  record.VaultID,
			"encryptionVersion":        record.EncryptionVersion,
			"chunkSize":                record.ChunkSize,
			"totalChunks":              record.TotalChunks,
			"plaintextSize":            record.PlaintextSize,
			"encryptedHash":            base64.StdEncoding.EncodeToString(record.EncryptedHash),
			"encryptedMetadata":        base64.StdEncoding.EncodeToString(record.EncryptedMetadata),
			"encryptedManifest":        base64.StdEncoding.EncodeToString(record.EncryptedManifest),
			"encryptedFileKeyForShare": base64.StdEncoding.EncodeToString(record.EncryptedFileKeyForShare),
			"sourceUrl":                sourceURL,
			"shareFileKeyAad":          "arkive:share-file-key:v1:" + record.FileID + ":" + record.Token,
			"metadataAad":              "arkive:file-metadata:v1:" + record.VaultID + ":" + record.FileID,
			"manifestAad":              "arkive:file-manifest:v1:" + record.VaultID + ":" + record.FileID,
		})
	}
}

func renderShareLanding(c *gin.Context, uploadService *uploads.Service, token string, share models.Share, file models.File) {
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
		Token:       token,
		File:        file,
		ViewURL:     viewURL,
		DownloadURL: downloadURL,
		IsImage:     isImage,
		IsVideo:     isVideo,
		Viewable:    viewable && viewURL != "",
		ShareURL:    shareURL,
		SharedAt:    share.CreatedAt,
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
