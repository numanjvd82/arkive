package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	filessvc "arkive/core/services/files"
	"arkive/pkg/errs"
)

type bulkDeleteFilesRequest struct {
	FileIDs []string `json:"fileIds"`
}

func APIDeleteFile(svc *filessvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID := c.Param("id")
		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if err := svc.DeleteFile(c.Request.Context(), userID.(string), fileID); err != nil {
			switch err {
			case filessvc.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case filessvc.ErrUploadCancelled:
				c.JSON(http.StatusConflict, gin.H{"error": "upload in progress"})
			case filessvc.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			case filessvc.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func APIBulkDeleteFiles(svc *filessvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req bulkDeleteFilesRequest
		if err := c.ShouldBindJSON(&req); err != nil || len(req.FileIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		deletedCount, err := svc.DeleteFiles(c.Request.Context(), userID.(string), req.FileIDs)
		if err != nil {
			switch err {
			case filessvc.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case filessvc.ErrUploadCancelled:
				c.JSON(http.StatusConflict, gin.H{"error": "upload in progress"})
			case filessvc.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			case filessvc.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"deleted": deletedCount})
	}
}
