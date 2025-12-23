package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	g "maragu.dev/gomponents"
)

type Page struct {
	Title string
	CSS   []string
	JS    []string
	Body  g.Node
}

func Render(c *gin.Context, page Page) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	var node g.Node
	if page.Body == nil {
		node = Layout(LayoutData{Title: page.Title, CSS: page.CSS, JS: page.JS})
	} else {
		node = Layout(LayoutData{Title: page.Title, CSS: page.CSS, JS: page.JS}, page.Body)
	}
	if err := node.Render(c.Writer); err != nil {
		c.Status(http.StatusInternalServerError)
	}
}
