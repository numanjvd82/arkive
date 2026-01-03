package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/services/auth"
	appcontext "arkive/pkg/context"
	"arkive/pkg/errs"
)

func RequireSessionRedirect(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ok, err := appcontext.LoadUser(c, svc)
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if !ok {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireSessionJSON(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok, err := appcontext.LoadUser(c, svc)
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "session lookup failed"})
			return
		}
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("user_id", user.ID)
		c.Next()
	}
}
