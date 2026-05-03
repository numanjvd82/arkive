package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/services/uploads"
	"arkive/pkg/errs"
)

type uploadStartRequest struct {
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
}

func APIUploadStart(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req uploadStartRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		resp, validationErrors, err := svc.StartUpload(c.Request.Context(), userID.(string), req.Filename, req.Size, req.ContentType)
		if err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "upload start failed"})
			}
			return
		}
		if validationErrors != nil && validationErrors.HasAny() {
			c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"uploadId":  resp.UploadID,
			"fileId":    resp.FileID,
			"objectKey": resp.ObjectKey,
			"mode":      resp.Mode,
			"uploadUrl": resp.UploadURL,
		})
	}
}

func APIUploadComplete(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadID := c.Param("id")
		if uploadID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if err := svc.CompleteSingleUpload(c.Request.Context(), userID.(string), uploadID); err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
			case uploads.ErrUploadCancelled:
				c.JSON(http.StatusConflict, gin.H{"error": "upload cancelled"})
			case uploads.ErrQuotaExceeded:
				c.JSON(http.StatusForbidden, gin.H{"error": "quota exceeded"})
			case uploads.ErrFileTooLarge:
				c.JSON(http.StatusBadRequest, gin.H{"error": "file is too large"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "complete failed"})
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func APIUploadCancel(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadID := c.Param("id")
		if uploadID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if err := svc.AbortSingleUpload(c.Request.Context(), userID.(string), uploadID); err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
			case uploads.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "cancel failed"})
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func APIDownloadFile(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID := c.Param("id")
		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if err := svc.TouchUserActivity(c.Request.Context(), userID.(string)); err != nil {
			_ = c.Error(errs.WithStack(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "download failed"})
			return
		}

		url, err := svc.PresignDownload(c.Request.Context(), userID.(string), fileID)
		if err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			case uploads.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "download failed"})
			}
			return
		}

		c.Redirect(http.StatusFound, url)
	}
}

func APIMediaRedirect(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID := c.Param("id")
		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if err := svc.TouchUserActivity(c.Request.Context(), userID.(string)); err != nil {
			_ = c.Error(errs.WithStack(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "media failed"})
			return
		}

		url, err := svc.PresignView(c.Request.Context(), userID.(string), fileID)
		if err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			case uploads.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "media failed"})
			}
			return
		}

		c.Redirect(http.StatusFound, url)
	}
}
