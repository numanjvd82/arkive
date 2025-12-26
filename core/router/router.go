package router

import (
	"github.com/gin-gonic/gin"

	"arkive/core/config"
	"arkive/core/database"
	"arkive/core/handlers"
	"arkive/core/middleware"
	authrepo "arkive/core/repositories/auth"
	sessionrepo "arkive/core/repositories/session"
	"arkive/core/services/auth"
	jwtservice "arkive/core/services/jwt"
	"arkive/core/web"
)

func New(db database.PgPool, cfg config.Config) *gin.Engine {
	r := gin.Default()

	jwtSvc := jwtservice.New(cfg.JWTSecret)
	authService := auth.NewService(db, authrepo.New(), sessionrepo.New(), jwtSvc, auth.Config{
		AccessTTL:  cfg.AccessTTL,
		RefreshTTL: cfg.RefreshTTL,
		SessionTTL: cfg.SessionTTL,
	})

	r.StaticFS("/static", web.StaticFS("static"))
	r.StaticFS("/web/pages", web.StaticFS("pages"))
	r.GET("/", handlers.WebHome())
	r.GET("/login", handlers.WebLoginGet(authService))
	r.POST("/login", handlers.WebLoginPost(authService))
	r.GET("/signup", handlers.WebSignupGet(authService))
	r.POST("/signup", handlers.WebSignupPost(authService))

	protected := r.Group("/")
	protected.Use(middleware.RequireSessionRedirect(authService))
	protected.GET("/dashboard", handlers.WebDashboard())
	protected.POST("/logout", handlers.WebLogout(authService))

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
