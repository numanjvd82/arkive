package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/services/uploads"
)

func APIDeleteFile(svc *uploads.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID := c.Param("id")
		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if err := svc.DeleteFile(c.Request.Context(), userID.(string), fileID); err != nil {
			switch err {
			case uploads.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case uploads.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			case uploads.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}
