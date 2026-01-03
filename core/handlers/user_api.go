package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/services/auth"
	"arkive/pkg/errs"
)

func APIMe(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		user, err := svc.GetUserByID(c.Request.Context(), userID.(string))
		if err != nil {
			if err == auth.ErrInvalidInput {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
			_ = c.Error(errs.WithStack(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         user.ID,
			"brand_name": user.BrandName,
			"email":      user.Email,
		})
	}
}
