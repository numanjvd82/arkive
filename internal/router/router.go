package router

import (
	"github.com/gin-gonic/gin"

	"arkive/internal/handlers"
)

func New() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		api.GET("/health", handlers.Health)
	}

	return r
}
