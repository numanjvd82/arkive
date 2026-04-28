package handlers

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"arkive/core/services/auth"
	"arkive/core/web"
	"arkive/core/web/pages"
	"arkive/pkg/errs"
)

func WebVerifyEmail(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.Query("token"))
		err := svc.VerifyEmail(c.Request.Context(), token)
		if err != nil {
			success := false
			message := auth.ErrVerifyTokenInvalid.Error()
			if !errors.Is(err, auth.ErrVerifyTokenInvalid) {
				_ = c.Error(errs.WithStack(err))
				message = "We couldn't verify your email right now. Please try again."
			}
			web.Render(c, pages.VerifyEmailPage(pages.VerifyEmailPageProps{Ctx: pages.PageContext{}, Success: success, Message: message}))
			return
		}

		web.Render(c, pages.VerifyEmailPage(pages.VerifyEmailPageProps{Ctx: pages.PageContext{}, Success: true, Message: ""}))
	}
}
