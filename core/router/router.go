package router

import (
	"github.com/gin-gonic/gin"

	"arkive/core/config"
	"arkive/core/database"
	"arkive/core/handlers"
	"arkive/core/middleware"
	authrepo "arkive/core/repositories/auth"
	filerepo "arkive/core/repositories/files"
	sessionrepo "arkive/core/repositories/session"
	sharerepo "arkive/core/repositories/shares"
	"arkive/core/services/auth"
	"arkive/core/services/shares"
	"arkive/core/services/uploads"
	"arkive/core/web"
)

func New(db database.PgPool, cfg config.Config, uploadService *uploads.Service) *gin.Engine {
	r := gin.Default()

	authService := auth.NewService(db, authrepo.New(), sessionrepo.New(), auth.Config{
		SessionTTL:     cfg.SessionTTL,
		GoogleClientID: cfg.GoogleClientID,
	})
	shareService := shares.NewService(db, filerepo.New(), sharerepo.New())

	r.StaticFS("/static", web.StaticFS("static"))
	r.StaticFS("/web/pages", web.StaticFS("pages"))
	r.GET("/", handlers.WebHome())
	r.GET("/s/:token", handlers.PublicShareView(shareService, uploadService))
	r.POST("/s/:token", handlers.PublicShareUnlock(shareService, uploadService))
	r.GET("/login", handlers.WebLoginGet(authService))
	r.POST("/login", handlers.WebLoginPost(authService))
	r.GET("/signup", handlers.WebSignupGet(authService))
	r.POST("/signup", handlers.WebSignupPost(authService))
	r.POST("/auth/google", handlers.WebGoogleLogin(authService))

	protected := r.Group("/")
	protected.Use(middleware.RequireSessionRedirect(authService))
	protected.GET("/dashboard", handlers.WebDashboard())
	protected.GET("/files", handlers.WebFiles(uploadService))
	protected.GET("/files/:id/view", handlers.WebFileView(uploadService))
	protected.POST("/logout", handlers.WebLogout(authService))

	api := r.Group("/api")
	{
		api.GET("/me", middleware.RequireSessionJSON(authService), handlers.APIMe(authService))
		api.GET("/health", handlers.Health(db))
	}

	apiUploads := api.Group("/uploads")
	apiUploads.Use(middleware.RequireSessionJSON(authService))
	{
		apiUploads.POST("/start", handlers.APIUploadStart(uploadService))
		apiUploads.POST("/:id/next", handlers.APIUploadNext(uploadService))
		apiUploads.POST("/:id/complete", handlers.APIUploadComplete(uploadService))
		apiUploads.POST("/:id/cancel", handlers.APIUploadCancel(uploadService))
	}

	apiFiles := api.Group("/files")
	apiFiles.Use(middleware.RequireSessionJSON(authService))
	{
		apiFiles.GET("/:id/share", handlers.APIGetShareForFile(shareService))
		apiFiles.GET("/:id/download", handlers.APIDownloadFile(uploadService))
		apiFiles.GET("/:id/media", handlers.APIMediaRedirect(uploadService))
		apiFiles.DELETE("/:id", handlers.APIDeleteFile(uploadService))
		apiFiles.POST("/:id/share", handlers.APICreateShare(shareService))
	}

	apiShares := api.Group("/shares")
	apiShares.Use(middleware.RequireSessionJSON(authService))
	{
		apiShares.POST("/:id/revoke", handlers.APIRevokeShare(shareService))
		apiShares.PATCH("/:id", handlers.APIUpdateShare(shareService))
		apiShares.DELETE("/:id", handlers.APIDeleteShare(shareService))
	}

	return r
}
