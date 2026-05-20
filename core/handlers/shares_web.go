package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	filessvc "arkive/core/services/files"
	"arkive/core/services/shares"
	"arkive/core/web"
	"arkive/core/web/pages"
	appcontext "arkive/pkg/context"
	"arkive/pkg/errs"
)

func WebShares(shareService *shares.Service, filesService *filessvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if err := filesService.TouchUserActivity(c.Request.Context(), user.ID); err != nil {
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		items, err := shareService.ListSharesForUser(c.Request.Context(), user.ID)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		web.Render(c, pages.SharesPage(pages.SharesPageProps{
			Ctx:    pages.ContextWithUser(user),
			Shares: items,
		}))
	}
}
