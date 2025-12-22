package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func WebHome() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title":   "Arkive",
			"PageCSS": "/statics/pages/home.css",
		})
	}
}

func WebLoginGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Title":   "Arkive · Login",
			"PageCSS": "/statics/pages/auth.css",
		})
	}
}

func WebLoginPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		email := strings.TrimSpace(c.PostForm("email"))
		password := strings.TrimSpace(c.PostForm("password"))
		if email == "" || password == "" {
			c.HTML(http.StatusOK, "login.html", gin.H{
				"Title":   "Arkive · Login",
				"PageCSS": "/statics/pages/auth.css",
				"Error":   "Email and password are required.",
				"Email":   email,
			})
			return
		}

		c.HTML(http.StatusOK, "login.html", gin.H{
			"Title":   "Arkive · Login",
			"PageCSS": "/statics/pages/auth.css",
			"Success": "Login submitted. We'll continue soon.",
			"Email":   email,
		})
	}
}

func WebSignupGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "signup.html", gin.H{
			"Title":   "Arkive · Sign Up",
			"PageCSS": "/statics/pages/auth.css",
		})
	}
}

func WebSignupPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := strings.TrimSpace(c.PostForm("name"))
		email := strings.TrimSpace(c.PostForm("email"))
		password := strings.TrimSpace(c.PostForm("password"))
		if name == "" || email == "" || password == "" {
			c.HTML(http.StatusOK, "signup.html", gin.H{
				"Title":   "Arkive · Sign Up",
				"PageCSS": "/statics/pages/auth.css",
				"Error":   "Name, email, and password are required.",
				"Name":    name,
				"Email":   email,
			})
			return
		}

		c.HTML(http.StatusOK, "signup.html", gin.H{
			"Title":   "Arkive · Sign Up",
			"PageCSS": "/statics/pages/auth.css",
			"Success": "Signup submitted. We'll be in touch soon.",
			"Name":    name,
			"Email":   email,
		})
	}
}
