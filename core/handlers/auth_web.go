package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"arkive/core/services/auth"
	"arkive/core/web"
	"arkive/core/web/pages"
	appcontext "arkive/pkg/context"
	"arkive/pkg/cookies"
	"arkive/pkg/errs"
	"arkive/pkg/validation"
)

func WebLoginGet(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
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
			Ctx:            pages.PageContext{},
			Message:        msg,
			GoogleClientID: svc.GoogleClientID(),
		})
		web.Render(c, webPage)
	}
}

func WebSignupGet(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok, err := appcontext.LoadUser(c, svc); err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		} else if ok {
			c.Redirect(http.StatusSeeOther, "/dashboard")
			return
		}

		webPage := pages.SignupPage(pages.SignupPageProps{
			Ctx:            pages.PageContext{},
			GoogleClientID: svc.GoogleClientID(),
		})
		web.Render(c, webPage)
	}
}

func WebLoginPost(svc *auth.Service) gin.HandlerFunc {
	type loginForm struct {
		Email    string `form:"email"`
		Password string `form:"password"`
	}

	return func(c *gin.Context) {
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
				Email:          strings.TrimSpace(c.PostForm("email")),
				Message:        "",
				GoogleClientID: svc.GoogleClientID(),
			}))
			return
		}

		sessionID, expiresAt, validationErrors, err := svc.WebLogin(c.Request.Context(), form.Email, form.Password, c.ClientIP())
		if len(validationErrors) > 0 {
			web.Render(c, pages.LoginPage(pages.LoginPageProps{
				Ctx:            pages.PageContext{},
				Errors:         validationErrors,
				Email:          strings.TrimSpace(form.Email),
				Message:        "",
				GoogleClientID: svc.GoogleClientID(),
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

func WebSignupPost(svc *auth.Service) gin.HandlerFunc {
	type signupForm struct {
		BrandName       string `form:"brand_name"`
		Email           string `form:"email"`
		Password        string `form:"password"`
		ConfirmPassword string `form:"confirm_password"`
	}

	return func(c *gin.Context) {
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
				BrandName:      strings.TrimSpace(c.PostForm("brand_name")),
				Email:          strings.TrimSpace(c.PostForm("email")),
				GoogleClientID: svc.GoogleClientID(),
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
				Ctx:            pages.PageContext{},
				Errors:         validationErrors,
				BrandName:      strings.TrimSpace(form.BrandName),
				Email:          strings.TrimSpace(form.Email),
				GoogleClientID: svc.GoogleClientID(),
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

func WebGoogleLogin(svc *auth.Service) gin.HandlerFunc {
	type payload struct {
		Credential string `json:"credential"`
	}

	return func(c *gin.Context) {
		var body payload
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		sessionID, expiresAt, err := svc.WebGoogleLogin(c.Request.Context(), body.Credential, c.ClientIP())
		if err != nil {
			switch err {
			case auth.ErrInvalidInput, auth.ErrGoogleTokenInvalid, auth.ErrGoogleEmailNotVerified, auth.ErrGoogleClientNotConfigured:
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			case auth.ErrGoogleEmailHasPassword:
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to sign in"})
			}
			return
		}

		cookies.SetSession(c, sessionID, expiresAt, false)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}
