package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	g "maragu.dev/gomponents"
)

func Render(c *gin.Context, node g.Node) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := node.Render(c.Writer); err != nil {
		c.Status(http.StatusInternalServerError)
	}
}
