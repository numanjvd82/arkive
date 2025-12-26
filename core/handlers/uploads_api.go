package handlers

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

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
	MultipartID string                       `json:"multipartId"`
	Parts       []uploads.CompletedPartInput `json:"parts"`
}

type multipartAbortRequest struct {
	MultipartID string `json:"multipartId"`
}

type shareCreateRequest struct {
	ExpiresAt string `json:"expiresAt"`
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
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
			case uploads.ErrInvalidInput, uploads.ErrPartsRequired:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
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

func APIShareCreate(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID := c.Param("id")
		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req shareCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		var expiresAt *time.Time
		if req.ExpiresAt != "" {
			parsed, err := time.Parse(time.RFC3339, req.ExpiresAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expiresAt"})
				return
			}
			expiresAt = &parsed
		}

		token, err := svc.CreateShare(c.Request.Context(), userID.(string), fileID, expiresAt)
		if err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			case uploads.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "share creation failed"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"url":   "/share/" + token,
		})
	}
}

func ShareDownload(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		url, err := svc.PresignShareDownload(c.Request.Context(), token)
		if err != nil {
			switch err {
			case uploads.ErrNotFound:
				c.Status(http.StatusNotFound)
			case uploads.ErrInvalidInput:
				c.Status(http.StatusBadRequest)
			default:
				c.Status(http.StatusInternalServerError)
			}
			return
		}
		c.Redirect(http.StatusFound, url)
	}
}
