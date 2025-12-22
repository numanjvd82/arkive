package router

import (
	"github.com/gin-gonic/gin"

	"arkive/core/database"
	"arkive/core/handlers"
)

func New(db database.PgExecutor) *gin.Engine {
	r := gin.Default()

	r.Static("/static", "./core/web/static")
	r.GET("/", handlers.WebHome())

	api := r.Group("/api")
	{
		api.GET("/health", handlers.Health(db))
	}

	return r
}
