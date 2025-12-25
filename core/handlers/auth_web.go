package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/services/auth"
	"arkive/core/web"
	"arkive/core/web/pages"
	"arkive/pkg/cookies"
)

func WebLoginGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		webPage := pages.LoginPage()
		web.Render(c, webPage)
	}
}

func WebSignupGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		webPage := pages.SignupPage()
		web.Render(c, webPage)
	}
}

func WebLoginPost(svc *auth.Service) gin.HandlerFunc {
	type loginForm struct {
		Email    string `form:"email"`
		Password string `form:"password"`
	}

	return func(c *gin.Context) {
		var form loginForm
		if err := c.ShouldBind(&form); err != nil {
			web.Render(c, pages.LoginPage())
			return
		}

		sessionID, expiresAt, err := svc.WebLogin(c.Request.Context(), form.Email, form.Password)
		if err != nil {
			switch err {
			case auth.ErrInvalidInput, auth.ErrInvalidCredentials:
				web.Render(c, pages.LoginPage())
			default:
				c.Status(http.StatusInternalServerError)
			}
			return
		}

		cookies.SetSession(c, sessionID, expiresAt, false)
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}

func WebSignupPost(svc *auth.Service) gin.HandlerFunc {
	type signupForm struct {
		BrandName string `form:"brand_name"`
		Email     string `form:"email"`
		Password  string `form:"password"`
	}

	return func(c *gin.Context) {
		var form signupForm
		if err := c.ShouldBind(&form); err != nil {
			web.Render(c, pages.SignupPage())
			return
		}

		sessionID, expiresAt, err := svc.WebSignup(c.Request.Context(), form.BrandName, form.Email, form.Password)
		if err != nil {
			switch err {
			case auth.ErrInvalidInput, auth.ErrEmailExists, auth.ErrBrandNameExists:
				web.Render(c, pages.SignupPage())
			default:
				c.Status(http.StatusInternalServerError)
			}
			return
		}

		cookies.SetSession(c, sessionID, expiresAt, false)
		c.Redirect(http.StatusSeeOther, "/dashboard")
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
			c.Status(http.StatusInternalServerError)
			return
		}

		cookies.ClearSession(c, false)
		c.Redirect(http.StatusSeeOther, "/")
	}
}
