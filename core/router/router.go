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
	usersrepo "arkive/core/repositories/users"
	"arkive/core/services/auth"
	"arkive/core/services/shares"
	"arkive/core/services/uploads"
	"arkive/core/web"
	"arkive/pkg/email"
)

func New(db database.PgPool, cfg config.Config, uploadService *uploads.Service) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.ErrorLogger())

	authService := auth.NewService(db, authrepo.New(), sessionrepo.New(), usersrepo.New(), auth.Config{
		SessionTTL: cfg.SessionTTL,
	})
	if cfg.PostmarkServerToken != "" {
		sender, err := email.NewPostmarkSender(cfg.PostmarkServerToken)
		if err != nil {
			panic("email sender failed: " + err.Error())
		}
		authService.ConfigureEmailVerification(sender, auth.VerifyConfig{PublicBaseURL: cfg.PublicBaseURL})
	}
	shareService := shares.NewService(db, filerepo.New(), sharerepo.New())

	r.StaticFS("/static", web.StaticFS("static"))
	r.StaticFS("/web/pages", web.StaticFS("pages"))
	r.GET("/favicon.ico", handlers.FaviconICO())
	r.GET("/robots.txt", handlers.RobotsTxt())
	r.GET("/sitemap.xml", handlers.SitemapXML())
	r.GET("/", handlers.WebHome())
	r.GET("/pricing", handlers.WebPricing())
	r.GET("/contact", handlers.WebContact())
	r.GET("/secure-file-sharing", handlers.WebSecureFileSharing())
	r.GET("/share-large-files", handlers.WebShareLargeFiles())
	r.GET("/file-sharing-without-login", handlers.WebFileSharingWithoutLogin())
	r.GET("/drop-pages", handlers.WebDropPages())
	r.GET("/privacy", handlers.WebPrivacy())
	r.GET("/cookies", handlers.WebCookie())
	r.GET("/terms", handlers.WebTerms())
	r.GET("/aup", handlers.WebAUP())
	r.GET("/abuse", handlers.WebAbuse())
	r.GET("/dmca", handlers.WebAbuse())
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
	r.GET("/login", handlers.WebLoginGet(authService))
	r.POST("/login", handlers.WebLoginPost(authService))
	r.GET("/signup", handlers.WebSignupGet(authService))
	r.POST("/signup", handlers.WebSignupPost(authService))
	r.GET("/verify-email", handlers.WebVerifyEmail(authService))
	protected := r.Group("/")
	protected.Use(middleware.RequireSessionRedirect(authService))
	protected.GET("/dashboard", handlers.WebDashboard(uploadService))
	protected.GET("/files", handlers.WebFiles(uploadService))
	protected.GET("/files/:id/view", handlers.WebFileView(uploadService))
	protected.GET("/shares", handlers.WebShares(shareService, uploadService))
	protected.GET("/settings", handlers.WebSettings(uploadService))
	protected.POST("/logout", handlers.WebLogout(authService))

	api := r.Group("/api")
	{
		api.GET("/me", middleware.RequireSessionJSON(authService), handlers.APIMe(authService))
		api.GET("/health", handlers.Health(db))
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
		apiUploads.POST("/:id/next", middleware.RateLimit(middleware.RateLimitConfig{
			RequestsPerMinute: 120,
			Burst:             240,
			KeyPrefix:         "upload:next",
		}), handlers.APIUploadNext(uploadService))
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
		apiShares.POST("/:id/revoke", middleware.RateLimit(middleware.RateLimitConfig{
			RequestsPerMinute: 20,
			Burst:             40,
			KeyPrefix:         "share:revoke",
		}), handlers.APIRevokeShare(shareService))
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
