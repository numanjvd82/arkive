package handlers

import (
	"net/http"
	"strings"

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
	return func(c *gin.Context) {
		email := strings.TrimSpace(c.PostForm("email"))
		password := strings.TrimSpace(c.PostForm("password"))
		if email == "" || password == "" {
			web.Render(c, pages.LoginPage())
			return
		}

		user, err := svc.Authenticate(c.Request.Context(), email, password)
		if err != nil {
			web.Render(c, pages.LoginPage())
			return
		}

		sessionID, expiresAt, err := svc.CreateSession(c.Request.Context(), user.ID)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		cookies.SetSession(c, sessionID, expiresAt, false)
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}

func WebSignupPost(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		brandName := strings.TrimSpace(c.PostForm("brand_name"))
		email := strings.TrimSpace(c.PostForm("email"))
		password := strings.TrimSpace(c.PostForm("password"))
		if brandName == "" || email == "" || password == "" {
			web.Render(c, pages.SignupPage())
			return
		}

		user, err := svc.CreateUser(c.Request.Context(), brandName, email, password)
		if err != nil {
			message := "Unable to create account."
			switch err {
			case auth.ErrEmailExists:
				message = "Email already in use."
			case auth.ErrBrandNameExists:
				message = "Brand name already taken."
			}
			web.Render(c, pages.SignupPage())
			return
		}

		sessionID, expiresAt, err := svc.CreateSession(c.Request.Context(), user.ID)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		cookies.SetSession(c, sessionID, expiresAt, false)
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}

func WebLogout(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie(cookies.SessionName)
		if err == nil && cookie.Value != "" {
			_ = svc.DeleteSession(c.Request.Context(), cookie.Value)
		}

		cookies.ClearSession(c, false)
		c.Redirect(http.StatusSeeOther, "/")
	}
}
