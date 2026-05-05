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
	settingsrepo "arkive/core/repositories/settings"
	sharerepo "arkive/core/repositories/shares"
	usersrepo "arkive/core/repositories/users"
	"arkive/core/services/auth"
	settingssvc "arkive/core/services/settings"
	"arkive/core/services/setup"
	"arkive/core/services/shares"
	"arkive/core/services/uploads"
	"arkive/core/web"
	"arkive/pkg/mailer"
	"arkive/pkg/storage/localclient"
)

func New(db database.PgPool, cfg config.Config, uploadService *uploads.Service, localStorage *localclient.Client) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.ErrorLogger())

	authRepo := authrepo.New()
	usersRepo := usersrepo.New()
	authService := auth.NewService(db, authRepo, sessionrepo.New(), usersRepo, auth.Config{
		SessionTTL: cfg.SessionTTL,
	})
	settingsRepo := settingsrepo.New()
	settingsService := settingssvc.NewService(db, settingsRepo, usersRepo)
	setupService := setup.NewService(db, authRepo, usersRepo, settingsRepo)
	mailerProvider, err := mailer.NewMailerFromConfig(mailer.Config{
		Provider: cfg.EmailProvider,
		From:     cfg.EmailFrom,
		SMTPHost: cfg.SMTPHost,
		SMTPPort: cfg.SMTPPort,
		SMTPUser: cfg.SMTPUser,
		SMTPPass: cfg.SMTPPass,
	})
	if err != nil {
		panic("mailer setup failed: " + err.Error())
	}
	authService.SetMailer(mailerProvider, cfg.PublicBaseURL)
	shareService := shares.NewService(db, filerepo.New(), sharerepo.New())

	r.StaticFS("/static", web.StaticFS("static"))
	r.StaticFS("/web/pages", web.StaticFS("pages"))
	r.GET("/favicon.ico", handlers.FaviconICO())
	r.GET("/robots.txt", handlers.RobotsTxt())
	r.GET("/sitemap.xml", handlers.SitemapXML())
	r.PUT("/local-storage/upload/:token", handlers.LocalStorageUpload(localStorage))
	r.GET("/local-storage/download/:token", handlers.LocalStorageDownload(localStorage))
	r.GET("/", handlers.WebRoot(authService, setupService))
	r.GET("/setup", handlers.WebSetupGet(setupService))
	r.POST("/setup", handlers.WebSetupPost(setupService))
	r.GET("/setup/recovery-key", handlers.WebSetupRecoveryGet(setupService))
	r.POST("/setup/recovery-key", handlers.WebSetupRecoveryPost(setupService))
	r.GET("/s/:token", middleware.RateLimit(middleware.RateLimitConfig{
		RequestsPerMinute: 2,
		Burst:             2,
		KeyPrefix:         "share:public:get",
	}), handlers.PublicShareView(shareService, uploadService))
	r.POST("/s/:token", middleware.RateLimit(middleware.RateLimitConfig{
		RequestsPerMinute: 2,
		Burst:             2,
		KeyPrefix:         "share:public:post",
	}), handlers.PublicShareUnlock(shareService, uploadService))
	r.GET("/login", handlers.WebLoginGet(authService, setupService))
	r.POST("/login", handlers.WebLoginPost(authService, setupService))
	r.GET("/verify-email", handlers.WebVerifyEmail(authService))
	protected := r.Group("/")
	protected.Use(middleware.RequireSessionRedirect(authService))
	protected.GET("/dashboard", handlers.WebDashboard(uploadService))
	protected.GET("/files", handlers.WebFiles(uploadService))
	protected.GET("/files/:id/view", handlers.WebFileView(uploadService))
	protected.GET("/shares", handlers.WebShares(shareService, uploadService))
	protected.GET("/settings", handlers.WebSettings(uploadService, settingsService))
	protected.POST("/settings/storage", handlers.WebSettingsStoragePost(settingsService))
	protected.POST("/logout", handlers.WebLogout(authService))

	api := r.Group("/api")
	{
		api.GET("/me", middleware.RequireSessionJSON(authService), handlers.APIMe(authService))
		api.GET("/health", handlers.Health(db))
		api.GET("/search", middleware.RequireSessionJSON(authService), handlers.APISearch(uploadService, shareService))
	}

	apiUploads := api.Group("/uploads")
	apiUploads.Use(middleware.RequireSessionJSON(authService))
	apiUploads.Use(middleware.LimitBody(2 << 20))
	{
		apiUploads.POST("/start", middleware.RateLimit(middleware.RateLimitConfig{
			RequestsPerMinute: 6,
			Burst:             10,
			KeyPrefix:         "upload:start",
		}), handlers.APIUploadStart(uploadService))
		apiUploads.POST("/:id/complete", middleware.RateLimit(middleware.RateLimitConfig{
			RequestsPerMinute: 10,
			Burst:             20,
			KeyPrefix:         "upload:complete",
		}), handlers.APIUploadComplete(uploadService))
		apiUploads.POST("/:id/cancel", middleware.RateLimit(middleware.RateLimitConfig{
			RequestsPerMinute: 10,
			Burst:             20,
			KeyPrefix:         "upload:cancel",
		}), handlers.APIUploadCancel(uploadService))
	}

	apiFiles := api.Group("/files")
	apiFiles.Use(middleware.RequireSessionJSON(authService))
	{
		apiFiles.GET("/:id/share", middleware.RateLimit(middleware.RateLimitConfig{
			RequestsPerMinute: 60,
			Burst:             120,
			KeyPrefix:         "share:get",
		}), handlers.APIGetShareForFile(shareService))
		apiFiles.GET("/:id/download", middleware.RateLimit(middleware.RateLimitConfig{
			RequestsPerMinute: 60,
			Burst:             120,
			KeyPrefix:         "file:download",
		}), handlers.APIDownloadFile(uploadService))
		apiFiles.GET("/:id/media", middleware.RateLimit(middleware.RateLimitConfig{
			RequestsPerMinute: 60,
			Burst:             120,
			KeyPrefix:         "file:media",
		}), handlers.APIMediaRedirect(uploadService))
		apiFiles.DELETE("/:id", middleware.RateLimit(middleware.RateLimitConfig{
			RequestsPerMinute: 30,
			Burst:             60,
			KeyPrefix:         "file:delete",
		}), handlers.APIDeleteFile(uploadService))
		apiFiles.POST("/:id/share", middleware.RateLimit(middleware.RateLimitConfig{
			RequestsPerMinute: 6,
			Burst:             12,
			KeyPrefix:         "share:create",
		}), handlers.APICreateShare(shareService))
	}

	apiShares := api.Group("/shares")
	apiShares.Use(middleware.RequireSessionJSON(authService))
	{
		apiShares.PATCH("/:id", middleware.RateLimit(middleware.RateLimitConfig{
			RequestsPerMinute: 20,
			Burst:             40,
			KeyPrefix:         "share:update",
		}), handlers.APIUpdateShare(shareService))
		apiShares.DELETE("/:id", middleware.RateLimit(middleware.RateLimitConfig{
			RequestsPerMinute: 20,
			Burst:             40,
			KeyPrefix:         "share:delete",
		}), handlers.APIDeleteShare(shareService))
	}

	r.NoRoute(handlers.WebNotFound())

	return r
}
