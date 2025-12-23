package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"arkive/core/services/auth"
	"arkive/core/web"
	"arkive/core/web/pages"
)

const sessionCookieName = "arkive_session"

func WebLoginGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		webPage := pages.LoginPage(pages.AuthPageData{
			Title: "Arkive · Login",
			CSS:   "/static/pages/auth.css",
		})
		web.Render(c, webPage)
	}
}

func WebSignupGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		webPage := pages.SignupPage(pages.AuthPageData{
			Title: "Arkive · Sign Up",
			CSS:   "/static/pages/auth.css",
		})
		web.Render(c, webPage)
	}
}

func WebLoginPost(svc *auth.Service, cookieSecure bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := strings.TrimSpace(c.PostForm("email"))
		password := strings.TrimSpace(c.PostForm("password"))
		if email == "" || password == "" {
			web.Render(c, pages.LoginPage(pages.AuthPageData{
				Title: "Arkive · Login",
				CSS:   "/static/pages/auth.css",
				Error: "Email and password are required.",
				Email: email,
			}))
			return
		}

		user, err := svc.Authenticate(c.Request.Context(), email, password)
		if err != nil {
			web.Render(c, pages.LoginPage(pages.AuthPageData{
				Title: "Arkive · Login",
				CSS:   "/static/pages/auth.css",
				Error: "Invalid email or password.",
				Email: email,
			}))
			return
		}

		sessionID, expiresAt, err := svc.CreateSession(c.Request.Context(), user.ID)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		setSessionCookie(c, sessionID, expiresAt, cookieSecure)
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}

func WebSignupPost(svc *auth.Service, cookieSecure bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		brandName := strings.TrimSpace(c.PostForm("brand_name"))
		email := strings.TrimSpace(c.PostForm("email"))
		password := strings.TrimSpace(c.PostForm("password"))
		if brandName == "" || email == "" || password == "" {
			web.Render(c, pages.SignupPage(pages.AuthPageData{
				Title: "Arkive · Sign Up",
				CSS:   "/static/pages/auth.css",
				Error: "Brand name, email, and password are required.",
				Name:  brandName,
				Email: email,
			}))
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
			web.Render(c, pages.SignupPage(pages.AuthPageData{
				Title: "Arkive · Sign Up",
				CSS:   "/static/pages/auth.css",
				Error: message,
				Name:  brandName,
				Email: email,
			}))
			return
		}

		sessionID, expiresAt, err := svc.CreateSession(c.Request.Context(), user.ID)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		setSessionCookie(c, sessionID, expiresAt, cookieSecure)
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}

func WebLogout(svc *auth.Service, cookieSecure bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie(sessionCookieName)
		if err == nil && cookie.Value != "" {
			_ = svc.DeleteSession(c.Request.Context(), cookie.Value)
		}

		clearSessionCookie(c, cookieSecure)
		c.Redirect(http.StatusSeeOther, "/")
	}
}

func setSessionCookie(c *gin.Context, sessionID string, expiresAt time.Time, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})
}

func clearSessionCookie(c *gin.Context, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}
