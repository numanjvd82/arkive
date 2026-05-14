package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"arkive/core/services/uploads"
	"arkive/pkg/errs"
)

type uploadStartRequest struct {
	OriginalSize      int64 `json:"originalSize"`
	FileChunkSize     int64 `json:"fileChunkSize"`
	TotalChunks       int   `json:"totalChunks"`
	UploadPartSize    int64 `json:"uploadPartSize"`
	UploadPartCount   int   `json:"uploadPartCount"`
	EncryptionVersion int16 `json:"encryptionVersion"`
}

type uploadPartRecordRequest struct {
	PartNumber    int    `json:"partNumber"`
	EncryptedHash string `json:"encryptedHash"`
	ETag          string `json:"etag"`
}

type uploadPartPresignBatchRequest struct {
	Parts []int `json:"parts"`
}

type uploadCompleteRequest struct {
	EncryptedMetadata string `json:"encryptedMetadata"`
	EncryptedFileKey  string `json:"encryptedFileKey"`
	EncryptedManifest string `json:"encryptedManifest"`
	EncryptedHash     string `json:"encryptedHash"`
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
			OriginalSize:      req.OriginalSize,
			FileChunkSize:     req.FileChunkSize,
			TotalChunks:       req.TotalChunks,
			UploadPartSize:    req.UploadPartSize,
			UploadPartCount:   req.UploadPartCount,
			EncryptionVersion: req.EncryptionVersion,
		})
		if err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrQuotaExceeded:
				c.JSON(http.StatusForbidden, gin.H{"error": "quota exceeded"})
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
			"vaultId":          resp.VaultID,
			"uploadSessionId":  resp.UploadSessionID,
			"providerUploadId": resp.ProviderUploadID,
			"fileChunkSize":    resp.FileChunkSize,
			"totalChunks":      resp.TotalChunks,
			"uploadPartSize":   resp.UploadPartSize,
			"uploadPartCount":  resp.UploadPartCount,
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

func APIUploadPartPresignBatch(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadSessionID := c.Param("id")
		if uploadSessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		var req uploadPartPresignBatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		urls, err := svc.PresignMultipartUploadParts(c.Request.Context(), userID.(string), uploadSessionID, req.Parts)
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

		c.JSON(http.StatusOK, gin.H{"urls": urls})
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

		if err := svc.CompleteMultipartUploadSession(c.Request.Context(), userID.(string), uploadSessionID, uploads.MultipartUploadCompleteInput{
			EncryptedMetadata: req.EncryptedMetadata,
			EncryptedFileKey:  req.EncryptedFileKey,
			EncryptedManifest: req.EncryptedManifest,
			EncryptedHash:     req.EncryptedHash,
		}); err != nil {
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

func APIFileRecord(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID := c.Param("id")
		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		record, err := svc.GetEncryptedFileRecord(c.Request.Context(), userID.(string), fileID)
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": "file record failed"})
			}
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"fileId":            record.FileID,
			"vaultId":           record.VaultID,
			"encryptionVersion": record.EncryptionVersion,
			"chunkSize":         record.ChunkSize,
			"totalChunks":       record.TotalChunks,
			"plaintextSize":     record.PlaintextSize,
			"encryptedHash":     record.EncryptedHash,
			"encryptedMetadata": record.EncryptedMetadata,
			"encryptedFileKey":  record.EncryptedFileKey,
			"encryptedManifest": record.EncryptedManifest,
			"sourceUrl":         record.SourceURL,
		})
	}
}
