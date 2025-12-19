package router

import (
	"github.com/gin-gonic/gin"

	"arkive/internal/database"
	"arkive/internal/handlers"
)

func New(db database.PgExecutor) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		api.GET("/health", handlers.Health(db))
	}

	return r
}
