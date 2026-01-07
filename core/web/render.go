package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	g "maragu.dev/gomponents"

	"arkive/core/models"
)

type Page struct {
	Title      string
	CSS        []string
	JS         []string
	Body       g.Node
	HideNav    bool
	AuthLayout bool
	User       *models.User
}

func Render(c *gin.Context, page Page) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	var node g.Node
	if page.AuthLayout {
		if page.Body == nil {
			node = AuthLayout(LayoutData{Title: page.Title, CSS: page.CSS, JS: page.JS, User: page.User})
		} else {
			node = AuthLayout(LayoutData{Title: page.Title, CSS: page.CSS, JS: page.JS, User: page.User}, page.Body)
		}
	} else if page.Body == nil {
		node = Layout(LayoutData{Title: page.Title, CSS: page.CSS, JS: page.JS, HideNav: page.HideNav, User: page.User})
	} else {
		node = Layout(LayoutData{Title: page.Title, CSS: page.CSS, JS: page.JS, HideNav: page.HideNav, User: page.User}, page.Body)
	}
	if err := node.Render(c.Writer); err != nil {
		c.Status(http.StatusInternalServerError)
	}
}
