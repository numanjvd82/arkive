package router

import (
	"github.com/gin-gonic/gin"

	"arkive/core/config"
	"arkive/core/database"
	"arkive/core/handlers"
	"arkive/core/middleware"
	"arkive/core/services/auth"
)

func New(db database.PgExecutor, cfg config.Config) *gin.Engine {
	r := gin.Default()

	authService := auth.NewService(db, auth.Config{
		JWTSecret:  cfg.JWTSecret,
		AccessTTL:  cfg.AccessTTL,
		RefreshTTL: cfg.RefreshTTL,
		SessionTTL: cfg.SessionTTL,
	})

	r.Static("/static", "./core/web/static")
	r.Static("/web/pages", "./core/web/pages")
	r.GET("/", handlers.WebHome())
	r.GET("/login", handlers.WebLoginGet())
	r.POST("/login", handlers.WebLoginPost(authService))
	r.GET("/signup", handlers.WebSignupGet())
	r.POST("/signup", handlers.WebSignupPost(authService))
	r.GET("/dashboard", handlers.WebDashboard())
	r.POST("/logout", handlers.WebLogout(authService))

	api := r.Group("/api")
	{
		api.POST("/auth/login", handlers.APILogin(authService))
		api.POST("/auth/refresh", handlers.APIRefresh(authService))
		api.POST("/auth/logout", handlers.APILogout(authService))

		api.GET("/me", middleware.RequireAccessToken(authService), handlers.APIMe(authService))
		api.GET("/health", handlers.Health(db))
	}

	return r
}
