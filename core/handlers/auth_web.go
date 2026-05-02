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
	"arkive/pkg/validation"
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

		msg := ""
		switch strings.TrimSpace(c.Query("msg")) {
		case "check-your-email":
			if svc.EmailVerificationEnabled() {
				msg = "Account created. Check your email to confirm your address."
			}
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

func WebSignupGet(svc *auth.Service, setupSvc *setup.Service) gin.HandlerFunc {
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

		webPage := pages.SignupPage(pages.SignupPageProps{
			Ctx: pages.PageContext{},
		})
		web.Render(c, webPage)
	}
}

func WebLoginPost(svc *auth.Service, setupSvc *setup.Service) gin.HandlerFunc {
	type loginForm struct {
		Email    string `form:"email"`
		Password string `form:"password"`
	}

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

		var form loginForm
		if err := c.ShouldBind(&form); err != nil {
			web.Render(c, pages.LoginPage(pages.LoginPageProps{
				Ctx: pages.PageContext{},
				Errors: validation.Errors{
					validation.GeneralKey: "Please fill out the form.",
				},
				Email:   strings.TrimSpace(c.PostForm("email")),
				Message: "",
			}))
			return
		}

		sessionID, expiresAt, validationErrors, err := svc.WebLogin(c.Request.Context(), form.Email, form.Password, c.ClientIP())
		if len(validationErrors) > 0 {
			web.Render(c, pages.LoginPage(pages.LoginPageProps{
				Ctx:     pages.PageContext{},
				Errors:  validationErrors,
				Email:   strings.TrimSpace(form.Email),
				Message: "",
			}))
			return
		}
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		cookies.SetSession(c, sessionID, expiresAt, false)
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}

func WebSignupPost(svc *auth.Service, setupSvc *setup.Service) gin.HandlerFunc {
	type signupForm struct {
		BrandName       string `form:"brand_name"`
		Email           string `form:"email"`
		Password        string `form:"password"`
		ConfirmPassword string `form:"confirm_password"`
	}

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

		var form signupForm
		if err := c.ShouldBind(&form); err != nil {
			web.Render(c, pages.SignupPage(pages.SignupPageProps{
				Ctx: pages.PageContext{},
				Errors: validation.Errors{
					validation.GeneralKey: "Please fill out the form.",
				},
				BrandName: strings.TrimSpace(c.PostForm("brand_name")),
				Email:     strings.TrimSpace(c.PostForm("email")),
			}))
			return
		}

		validationErrors, err := svc.WebSignup(
			c.Request.Context(),
			form.BrandName,
			form.Email,
			form.Password,
			form.ConfirmPassword,
		)
		if len(validationErrors) > 0 {
			web.Render(c, pages.SignupPage(pages.SignupPageProps{
				Ctx:       pages.PageContext{},
				Errors:    validationErrors,
				BrandName: strings.TrimSpace(form.BrandName),
				Email:     strings.TrimSpace(form.Email),
			}))
			return
		}
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		if svc.EmailVerificationEnabled() {
			c.Redirect(http.StatusSeeOther, "/login?msg=check-your-email")
			return
		}
		c.Redirect(http.StatusSeeOther, "/login?msg=account-created")
	}
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
