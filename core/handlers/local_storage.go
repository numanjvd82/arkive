package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/pkg/storage/localclient"
)

func LocalStorageUpload(client *localclient.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPut {
			c.Status(http.StatusMethodNotAllowed)
			return
		}
		client.ServeUpload(c.Writer, c.Request, c.Param("token"))
	}
}

func LocalStorageDownload(client *localclient.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.ServeDownload(c.Writer, c.Request, c.Param("token"))
	}
}
