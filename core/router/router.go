package router

import (
	"github.com/gin-gonic/gin"

	"arkive/core/database"
	"arkive/core/handlers"
)

func New(db database.PgExecutor) *gin.Engine {
	r := gin.Default()

	r.LoadHTMLGlob("web/templates/*.html")
	r.Static("/statics", "./web/statics")
	r.GET("/", handlers.WebHome())
	r.GET("/login", handlers.WebLoginGet())
	r.POST("/login", handlers.WebLoginPost())
	r.GET("/signup", handlers.WebSignupGet())
	r.POST("/signup", handlers.WebSignupPost())

	api := r.Group("/api")
	{
		api.GET("/health", handlers.Health(db))
	}

	return r
}
