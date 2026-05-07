package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"arkive/core/services/uploads"
	"arkive/pkg/errs"
)

type uploadStartRequest struct {
	EncryptedMetadata string `json:"encryptedMetadata"`
	EncryptedFileKey  string `json:"encryptedFileKey"`
	OriginalSize      int64  `json:"originalSize"`
	PartSize          int64  `json:"partSize"`
	TotalParts        int    `json:"totalParts"`
	EncryptionVersion int16  `json:"encryptionVersion"`
}

type uploadPartRecordRequest struct {
	PartNumber    int    `json:"partNumber"`
	EncryptedSize int64  `json:"encryptedSize"`
	EncryptedHash string `json:"encryptedHash"`
	ETag          string `json:"etag"`
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

		resp, validationErrors, err := svc.StartMultipartUploadSession(c.Request.Context(), userID.(string), uploads.MultipartUploadStartInput{
			EncryptedMetadata: req.EncryptedMetadata,
			EncryptedFileKey:  req.EncryptedFileKey,
			OriginalSize:      req.OriginalSize,
			PartSize:          req.PartSize,
			TotalParts:        req.TotalParts,
			EncryptionVersion: req.EncryptionVersion,
		})
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
			"fileId":           resp.FileID,
			"uploadSessionId":  resp.UploadSessionID,
			"providerUploadId": resp.ProviderUploadID,
			"partSize":         resp.PartSize,
			"totalParts":       resp.TotalParts,
		})
	}
}

func APIUploadPartPresign(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadSessionID := c.Param("id")
		partNumber, err := strconv.Atoi(c.Param("part"))
		if err != nil || uploadSessionID == "" || partNumber <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		url, err := svc.PresignMultipartUploadPart(c.Request.Context(), userID.(string), uploadSessionID, partNumber)
		if err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
			case uploads.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			case uploads.ErrUploadCancelled:
				c.JSON(http.StatusConflict, gin.H{"error": "upload cancelled"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "part presign failed"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}

func APIUploadPartRecord(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadSessionID := c.Param("id")
		if uploadSessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		var req uploadPartRecordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if err := svc.RecordMultipartUploadPart(c.Request.Context(), userID.(string), uploadSessionID, uploads.UploadPartRecordInput{
			PartNumber:    req.PartNumber,
			EncryptedSize: req.EncryptedSize,
			EncryptedHash: req.EncryptedHash,
			ETag:          req.ETag,
		}); err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
			case uploads.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			case uploads.ErrUploadCancelled:
				c.JSON(http.StatusConflict, gin.H{"error": "upload cancelled"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "record part failed"})
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func APIUploadComplete(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadSessionID := c.Param("id")
		if uploadSessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if err := svc.CompleteMultipartUploadSession(c.Request.Context(), userID.(string), uploadSessionID); err != nil {
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
			case uploads.ErrPartsRequired:
				c.JSON(http.StatusConflict, gin.H{"error": "missing parts"})
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
		uploadSessionID := c.Param("id")
		if uploadSessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if err := svc.AbortMultipartUploadSession(c.Request.Context(), userID.(string), uploadSessionID); err != nil {
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
