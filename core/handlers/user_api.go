package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/services/auth"
	"arkive/pkg/apierror"
	"arkive/pkg/errs"
)

func APIMe(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}

		user, err := svc.GetUserByID(c.Request.Context(), userID.(string))
		if err != nil {
			if err == auth.ErrInvalidInput {
				apierror.Unauthorized(c)
				return
			}
			_ = c.Error(errs.WithStack(err))
			apierror.Internal(c, "Could not load user")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         user.ID,
			"brand_name": user.BrandName,
			"email":      user.Email,
		})
	}
}
