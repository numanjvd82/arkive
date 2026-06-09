package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"arkive/core/models"
	filessvc "arkive/core/services/files"
	settingssvc "arkive/core/services/settings"
	"arkive/core/services/uploads"
	"arkive/pkg/apierror"
	"arkive/pkg/errs"
)

type uploadStartRequest struct {
	OriginalSize      int64   `json:"originalSize"`
	FileChunkSize     int64   `json:"fileChunkSize"`
	TotalChunks       int     `json:"totalChunks"`
	UploadPartSize    int64   `json:"uploadPartSize"`
	UploadPartCount   int     `json:"uploadPartCount"`
	EncryptionVersion int16   `json:"encryptionVersion"`
	FolderID          *string `json:"folderId"`
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
	EncryptedMetadata string               `json:"encryptedMetadata"`
	EncryptedFileKey  string               `json:"encryptedFileKey"`
	EncryptedManifest string               `json:"encryptedManifest"`
	EncryptedHash     string               `json:"encryptedHash"`
	SearchTokens      []searchTokenRequest `json:"searchTokens"`
	HasThumbnail      bool                 `json:"hasThumbnail"`
	ThumbnailMime     string               `json:"thumbnailMime"`
	ThumbnailWidth    int                  `json:"thumbnailWidth"`
	ThumbnailHeight   int                  `json:"thumbnailHeight"`
}

type searchTokenRequest struct {
	Token  string `json:"token"`
	Field  string `json:"field"`
	Weight int    `json:"weight"`
}

func decodeSearchTokens(input []searchTokenRequest) ([]models.FileSearchToken, error) {
	if len(input) == 0 {
		return nil, filessvc.ErrInvalidInput
	}
	tokens := make([]models.FileSearchToken, 0, len(input))
	for _, item := range input {
		tokenHash, err := filessvc.DecodeSearchTokenString(item.Token)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, models.FileSearchToken{
			TokenHash: tokenHash,
			Field:     strings.TrimSpace(item.Field),
			Weight:    item.Weight,
		})
	}
	return tokens, nil
}

type thumbnailPresignRequest struct {
	EncryptedSize int64  `json:"encryptedSize"`
	Mime          string `json:"mime"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
}

func APIUploadLimits(settingsService *settingssvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		settings, err := settingsService.UploadSettings(c.Request.Context())
		if err != nil {
			settings = settingssvc.DefaultUploadSettings()
		}

		defaults := settingssvc.DefaultUploadSettings()
		if settings.MaxQueueItems <= 0 {
			settings.MaxQueueItems = defaults.MaxQueueItems
		}
		if settings.PartConcurrency <= 0 {
			settings.PartConcurrency = defaults.PartConcurrency
		}
		if settings.StaleUploadHours <= 0 {
			settings.StaleUploadHours = defaults.StaleUploadHours
		}

		c.JSON(http.StatusOK, gin.H{
			"maxQueueItems":    settings.MaxQueueItems,
			"partConcurrency":  settings.PartConcurrency,
			"staleUploadHours": settings.StaleUploadHours,
		})
	}
}

func APIUploadStart(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req uploadStartRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			apierror.InvalidPayload(c)
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}

		resp, validationErrors, err := svc.StartMultipartUploadSession(c.Request.Context(), userID.(string), uploads.MultipartUploadStartInput{
			OriginalSize:      req.OriginalSize,
			FileChunkSize:     req.FileChunkSize,
			TotalChunks:       req.TotalChunks,
			UploadPartSize:    req.UploadPartSize,
			UploadPartCount:   req.UploadPartCount,
			EncryptionVersion: req.EncryptionVersion,
			FolderID:          req.FolderID,
		})
		if err != nil {
			var limitErr *uploads.StorageLimitExceededError
			switch err {
			case uploads.ErrUnauthorized:
				apierror.Unauthorized(c)
			case uploads.ErrNotFound:
				apierror.Write(c, http.StatusNotFound, "folder_not_found", "Folder not found", nil)
			case uploads.ErrInvalidInput:
				apierror.InvalidPayload(c)
			default:
				if errors.As(err, &limitErr) {
					apierror.Write(c, http.StatusForbidden, "storage_limit_exceeded", "Storage limit exceeded", gin.H{
						"maxBytes":       limitErr.MaxBytes,
						"usedBytes":      limitErr.UsedBytes,
						"requestedBytes": limitErr.RequestedBytes,
					})
					return
				}
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Upload start failed")
			}
			return
		}
		if validationErrors != nil && validationErrors.HasAny() {
			apierror.Validation(c, validationErrors)
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
			apierror.InvalidPayload(c)
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}

		url, err := svc.PresignMultipartUploadPart(c.Request.Context(), userID.(string), uploadSessionID, partNumber)
		if err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				apierror.Unauthorized(c)
			case uploads.ErrNotFound:
				apierror.Write(c, http.StatusNotFound, "upload_not_found", "Upload not found", nil)
			case uploads.ErrInvalidInput:
				apierror.InvalidPayload(c)
			case uploads.ErrUploadCancelled:
				apierror.Write(c, http.StatusConflict, "upload_cancelled", "Upload cancelled", nil)
			default:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Part presign failed")
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
			apierror.InvalidPayload(c)
			return
		}

		var req uploadPartPresignBatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			apierror.InvalidPayload(c)
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}

		urls, err := svc.PresignMultipartUploadParts(c.Request.Context(), userID.(string), uploadSessionID, req.Parts)
		if err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				apierror.Unauthorized(c)
			case uploads.ErrNotFound:
				apierror.Write(c, http.StatusNotFound, "upload_not_found", "Upload not found", nil)
			case uploads.ErrInvalidInput:
				apierror.InvalidPayload(c)
			case uploads.ErrUploadCancelled:
				apierror.Write(c, http.StatusConflict, "upload_cancelled", "Upload cancelled", nil)
			default:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Part presign failed")
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
			apierror.InvalidPayload(c)
			return
		}

		var req uploadPartRecordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			apierror.InvalidPayload(c)
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}

		if err := svc.RecordMultipartUploadPart(c.Request.Context(), userID.(string), uploadSessionID, uploads.UploadPartRecordInput{
			PartNumber:    req.PartNumber,
			EncryptedHash: req.EncryptedHash,
			ETag:          req.ETag,
		}); err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				apierror.Unauthorized(c)
			case uploads.ErrNotFound:
				apierror.Write(c, http.StatusNotFound, "upload_not_found", "Upload not found", nil)
			case uploads.ErrInvalidInput:
				apierror.Write(c, http.StatusBadRequest, "invalid_upload_part", "Invalid upload part", nil)
			case uploads.ErrUploadCancelled:
				apierror.Write(c, http.StatusConflict, "upload_cancelled", "Upload cancelled", nil)
			default:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Record part failed")
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
			apierror.InvalidPayload(c)
			return
		}
		var req uploadCompleteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			apierror.InvalidPayload(c)
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}
		searchTokens, err := decodeSearchTokens(req.SearchTokens)
		if err != nil {
			apierror.InvalidPayload(c)
			return
		}

		if err := svc.CompleteMultipartUploadSession(c.Request.Context(), userID.(string), uploadSessionID, uploads.MultipartUploadCompleteInput{
			EncryptedMetadata: req.EncryptedMetadata,
			EncryptedFileKey:  req.EncryptedFileKey,
			EncryptedManifest: req.EncryptedManifest,
			EncryptedHash:     req.EncryptedHash,
			SearchTokens:      searchTokens,
			HasThumbnail:      req.HasThumbnail,
			ThumbnailMime:     req.ThumbnailMime,
			ThumbnailWidth:    req.ThumbnailWidth,
			ThumbnailHeight:   req.ThumbnailHeight,
		}); err != nil {
			var limitErr *uploads.StorageLimitExceededError
			switch err {
			case uploads.ErrUnauthorized:
				apierror.Unauthorized(c)
			case uploads.ErrNotFound:
				apierror.Write(c, http.StatusNotFound, "upload_not_found", "Upload not found", nil)
			case uploads.ErrUploadCancelled:
				apierror.Write(c, http.StatusConflict, "upload_cancelled", "Upload cancelled", nil)
			case uploads.ErrFileTooLarge:
				apierror.Write(c, http.StatusBadRequest, "file_too_large", "File is too large", nil)
			case uploads.ErrPartsRequired:
				apierror.Write(c, http.StatusConflict, "missing_upload_parts", "Missing upload parts", nil)
			default:
				if errors.As(err, &limitErr) {
					apierror.Write(c, http.StatusForbidden, "storage_limit_exceeded", "Storage limit exceeded", gin.H{
						"maxBytes":       limitErr.MaxBytes,
						"usedBytes":      limitErr.UsedBytes,
						"requestedBytes": limitErr.RequestedBytes,
					})
					return
				}
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Complete failed")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func APIThumbnailPresign(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadSessionID := c.Param("id")
		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}

		var req thumbnailPresignRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			apierror.InvalidPayload(c)
			return
		}

		url, err := svc.PresignThumbnailUpload(c.Request.Context(), userID.(string), uploadSessionID, uploads.ThumbnailUploadInput{
			EncryptedSize: req.EncryptedSize,
			Mime:          req.Mime,
			Width:         req.Width,
			Height:        req.Height,
		})
		if err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				apierror.Unauthorized(c)
			case uploads.ErrNotFound:
				apierror.Write(c, http.StatusNotFound, "file_not_found", "File not found", nil)
			case uploads.ErrInvalidInput:
				apierror.InvalidPayload(c)
			case uploads.ErrFileTooLarge:
				apierror.Write(c, http.StatusBadRequest, "thumbnail_too_large", "Thumbnail is too large", nil)
			case uploads.ErrUploadCancelled:
				apierror.Write(c, http.StatusConflict, "upload_cancelled", "Upload cancelled", nil)
			default:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Thumbnail presign failed")
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}

func APIUploadCancel(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadSessionID := c.Param("id")
		if uploadSessionID == "" {
			apierror.InvalidPayload(c)
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}

		if err := svc.AbortMultipartUploadSession(c.Request.Context(), userID.(string), uploadSessionID); err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				apierror.Unauthorized(c)
			case uploads.ErrNotFound:
				apierror.Write(c, http.StatusNotFound, "upload_not_found", "Upload not found", nil)
			case uploads.ErrInvalidInput:
				apierror.InvalidPayload(c)
			default:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Cancel failed")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func APIThumbnailRedirect(svc *filessvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID := c.Param("id")
		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}

		url, err := svc.PresignThumbnailDownload(c.Request.Context(), userID.(string), fileID)
		if err != nil {
			switch err {
			case filessvc.ErrUnauthorized:
				apierror.Unauthorized(c)
			case filessvc.ErrNotFound:
				apierror.Write(c, http.StatusNotFound, "thumbnail_not_found", "Thumbnail not found", nil)
			case filessvc.ErrInvalidInput:
				apierror.InvalidPayload(c)
			default:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Thumbnail failed")
			}
			return
		}

		c.Redirect(http.StatusFound, url)
	}
}

func APIFileRecord(svc *filessvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID := c.Param("id")
		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}
		record, err := svc.GetEncryptedFileRecord(c.Request.Context(), userID.(string), fileID)
		if err != nil {
			switch err {
			case filessvc.ErrUnauthorized:
				apierror.Unauthorized(c)
			case filessvc.ErrNotFound:
				apierror.Write(c, http.StatusNotFound, "file_not_found", "File not found", nil)
			case filessvc.ErrInvalidInput:
				apierror.InvalidPayload(c)
			default:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "File record failed")
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
