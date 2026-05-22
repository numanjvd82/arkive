package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/services/auth"
	"arkive/core/web"
	"arkive/core/web/pages"
	"arkive/pkg/apierror"
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
			c.Status(http.StatusForbidden)
			web.Render(c, pages.ForbiddenPage(pages.ForbiddenPageProps{
				Ctx: pages.PageContext{},
			}))
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
			apierror.Internal(c, "Session lookup failed")
			c.Abort()
			return
		}
		if !ok {
			apierror.Unauthorized(c)
			c.Abort()
			return
		}
		c.Set("user_id", user.ID)
		c.Next()
	}
}
