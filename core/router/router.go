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
	"arkive/core/services/uploads"
	"arkive/core/web"
)

func New(db database.PgPool, cfg config.Config, uploadService *uploads.Service) *gin.Engine {
	r := gin.Default()

	authService := auth.NewService(db, authrepo.New(), sessionrepo.New(), auth.Config{
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
		apiFiles.GET("/:id/download", handlers.APIDownloadFile(uploadService))
		apiFiles.DELETE("/:id", handlers.APIDeleteFile(uploadService))
	}

	return r
}
