package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/models"
	"arkive/core/services/uploads"
)

type multipartStartRequest struct {
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
}

type multipartPartURLRequest struct {
	MultipartID string `json:"multipartId"`
	PartNumber  int32  `json:"partNumber"`
}

type multipartCompleteRequest struct {
	MultipartID string                      `json:"multipartId"`
	Parts       []models.CompletedPartInput `json:"parts"`
}

type multipartAbortRequest struct {
	MultipartID string `json:"multipartId"`
}

type singleStartRequest struct {
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
}

type singleCompleteRequest struct {
	FileID string `json:"fileId"`
}

type singleAbortRequest struct {
	FileID string `json:"fileId"`
}

type uploadAbortRequest struct {
	FileID string `json:"fileId"`
}

type multipartResumeRequest struct {
	FileID string `form:"fileId"`
}

func APIMultipartStart(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req multipartStartRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		resp, validationErrors, err := svc.StartMultipart(c.Request.Context(), userID.(string), req.Filename, req.Size, req.ContentType)
		if err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "multipart start failed"})
			}
			return
		}
		if validationErrors != nil && validationErrors.HasAny() {
			c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"fileId":      resp.FileID,
			"multipartId": resp.MultipartID,
			"objectKey":   resp.ObjectKey,
			"chunkSize":   resp.ChunkSize,
			"totalParts":  resp.TotalParts,
		})
	}
}

func APIMultipartResume(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req multipartResumeRequest
		if err := c.ShouldBindQuery(&req); err != nil || req.FileID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		resp, err := svc.ResumeMultipart(c.Request.Context(), userID.(string), req.FileID)
		if err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
			case uploads.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "resume failed"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"fileId":        resp.FileID,
			"multipartId":   resp.MultipartID,
			"filename":      resp.Filename,
			"sizeBytes":     resp.SizeBytes,
			"chunkSize":     resp.ChunkSize,
			"totalParts":    resp.TotalParts,
			"uploadedParts": resp.UploadedParts,
		})
	}
}

func APISingleStart(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req singleStartRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		resp, validationErrors, err := svc.StartSingleUpload(c.Request.Context(), userID.(string), req.Filename, req.Size, req.ContentType)
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
			"fileId":    resp.FileID,
			"objectKey": resp.ObjectKey,
			"uploadUrl": resp.UploadURL,
			"expiresAt": resp.ExpiresAt,
		})
	}
}

func APIMultipartPartURL(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req multipartPartURLRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		url, err := svc.PresignPart(c.Request.Context(), userID.(string), req.MultipartID, req.PartNumber)
		if err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
			case uploads.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "part url failed"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}

func APIMultipartComplete(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req multipartCompleteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if err := svc.CompleteMultipart(c.Request.Context(), userID.(string), req.MultipartID, req.Parts); err != nil {
			var missingErr uploads.MissingPartsError
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
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

func APIMultipartAbort(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req multipartAbortRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if err := svc.AbortMultipart(c.Request.Context(), userID.(string), req.MultipartID); err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
			case uploads.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "abort failed"})
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func APISingleComplete(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req singleCompleteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if err := svc.CompleteSingleUpload(c.Request.Context(), userID.(string), req.FileID); err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			case uploads.ErrQuotaExceeded:
				c.JSON(http.StatusForbidden, gin.H{"error": "quota exceeded"})
			case uploads.ErrFileTooLarge:
				c.JSON(http.StatusBadRequest, gin.H{"error": "file is too large"})
			case uploads.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "complete failed"})
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func APISingleAbort(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req singleAbortRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if err := svc.AbortSingleUpload(c.Request.Context(), userID.(string), req.FileID); err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			case uploads.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "abort failed"})
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func APIUploadAbort(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req uploadAbortRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if err := svc.AbortUploadByFile(c.Request.Context(), userID.(string), req.FileID); err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			case uploads.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "abort failed"})
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
