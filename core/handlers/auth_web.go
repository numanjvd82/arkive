package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"arkive/core/services/auth"
	"arkive/core/services/setup"
	"arkive/core/web"
	"arkive/core/web/pages"
	appcontext "arkive/pkg/context"
	"arkive/pkg/cookies"
	"arkive/pkg/errs"
)

func WebLoginGet(svc *auth.Service, setupSvc *setup.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if redirectIfUninitialized(c, setupSvc) {
			return
		}
		if _, ok, err := appcontext.LoadUser(c, svc); err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		} else if ok {
			c.Redirect(http.StatusSeeOther, "/dashboard")
			return
		}
		ok, err := hasRecoveryPending(c, setupSvc)
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if ok {
			c.Redirect(http.StatusSeeOther, "/setup/recovery-key")
			return
		}

		msg := ""
		switch strings.TrimSpace(c.Query("msg")) {
		case "account-created":
			msg = "Account created. You can log in now."
		}
		webPage := pages.LoginPage(pages.LoginPageProps{
			Ctx:     pages.PageContext{},
			Message: msg,
		})
		web.Render(c, webPage)
	}
}

func WebLockGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		web.Render(c, pages.LockPage(pages.LockPageProps{
			Ctx:  pages.ContextWithUser(user),
			Next: sanitizeNextPath(c.Query("next")),
		}))
	}
}

func sanitizeNextPath(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") {
		return ""
	}
	return value
}

func redirectIfUninitialized(c *gin.Context, setupSvc *setup.Service) bool {
	initialized, err := setupSvc.IsInitialized(c.Request.Context())
	if err != nil {
		_ = c.Error(errs.WithStack(err))
		c.Status(http.StatusInternalServerError)
		return true
	}
	if !initialized {
		c.Redirect(http.StatusSeeOther, "/setup")
		return true
	}
	return false
}

func WebLogout(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := ""
		cookie, err := c.Request.Cookie(cookies.SessionName)
		if err == nil && cookie.Value != "" {
			sessionID = cookie.Value
		}

		if err := svc.LogoutSession(c.Request.Context(), sessionID); err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		cookies.ClearSession(c, false)
		c.Redirect(http.StatusSeeOther, "/")
	}
}
