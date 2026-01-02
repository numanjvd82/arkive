package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/models"
	"arkive/core/services/uploads"
)

type uploadStartRequest struct {
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
}

type uploadNextRequest struct {
	UploadedParts []int32 `json:"uploadedParts"`
}

type uploadCompleteRequest struct {
	Parts []models.CompletedPartInput `json:"parts"`
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": "upload start failed"})
			}
			return
		}
		if validationErrors != nil && validationErrors.HasAny() {
			c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"uploadId":   resp.UploadID,
			"fileId":     resp.FileID,
			"objectKey":  resp.ObjectKey,
			"mode":       resp.Mode,
			"chunkSize":  resp.ChunkSize,
			"totalParts": resp.TotalParts,
			"uploadUrl":  resp.UploadURL,
		})
	}
}

func APIUploadNext(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadID := c.Param("id")
		if uploadID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		var req uploadNextRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		resp, err := svc.NextUpload(c.Request.Context(), userID.(string), uploadID, req.UploadedParts)
		if err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
			case uploads.ErrUploadCancelled:
				c.JSON(http.StatusConflict, gin.H{"error": "upload cancelled"})
			case uploads.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			case uploads.ErrNoNextPart:
				c.Status(http.StatusNoContent)
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "next failed"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"uploadId":      resp.UploadID,
			"fileId":        resp.FileID,
			"mode":          resp.Mode,
			"nextPart":      resp.NextPart,
			"url":           resp.URL,
			"chunkSize":     resp.ChunkSize,
			"totalParts":    resp.TotalParts,
			"uploadedParts": resp.UploadedParts,
			"throttleMs":    resp.ThrottleMs,
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

		var req uploadCompleteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if err := svc.CompleteUpload(c.Request.Context(), userID.(string), uploadID, req.Parts); err != nil {
			var missingErr uploads.MissingPartsError
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
			case uploads.ErrInvalidInput, uploads.ErrPartsRequired:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
				if errors.As(err, &missingErr) {
					c.JSON(http.StatusConflict, gin.H{
						"error":        "missing parts",
						"missingParts": missingErr.Missing,
					})
					return
				}
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

		if err := svc.CancelUpload(c.Request.Context(), userID.(string), uploadID); err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
			case uploads.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": "download failed"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"url": url})
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": "media failed"})
			}
			return
		}

		c.Redirect(http.StatusFound, url)
	}
}
