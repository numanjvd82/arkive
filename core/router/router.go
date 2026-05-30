package router

import (
	"context"

	"github.com/gin-gonic/gin"

	"arkive/core/config"
	"arkive/core/database"
	"arkive/core/handlers"
	"arkive/core/middleware"
	"arkive/core/models"
	authrepo "arkive/core/repositories/auth"
	filerepo "arkive/core/repositories/files"
	sessionrepo "arkive/core/repositories/session"
	settingsrepo "arkive/core/repositories/settings"
	sharerepo "arkive/core/repositories/shares"
	usersrepo "arkive/core/repositories/users"
	"arkive/core/services/auth"
	filessvc "arkive/core/services/files"
	folderssvc "arkive/core/services/folders"
	settingssvc "arkive/core/services/settings"
	"arkive/core/services/setup"
	"arkive/core/services/shares"
	"arkive/core/services/uploads"
	"arkive/core/web"
	"arkive/pkg/mailer"
	"arkive/pkg/storage/localclient"
)

func New(db database.PgPool, cfg config.Config, uploadService *uploads.Service, filesService *filessvc.Service, folderService *folderssvc.Service, localStorage *localclient.Client) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.ErrorLogger())
	r.Use(middleware.SecurityHeaders())

	authRepo := authrepo.New()
	usersRepo := usersrepo.New()
	authService := auth.NewService(db, authRepo, sessionrepo.New(), usersRepo, auth.Config{
		SessionTTL: cfg.SessionTTL,
	})
	settingsRepo := settingsrepo.New()
	settingsService := settingssvc.NewService(db, settingsRepo)
	setupService := setup.NewService(db, authRepo, usersRepo, settingsRepo)
	emailSettings, err := settingsService.EmailSettings(context.Background())
	if err != nil {
		emailSettings = models.EmailSettings{Provider: "noop"}
	}
	mailerProvider, err := mailer.NewMailerFromConfig(mailer.Config{
		Provider: emailSettings.Provider,
		From:     emailSettings.From,
		SMTPHost: emailSettings.SMTPHost,
		SMTPPort: emailSettings.SMTPPort,
		SMTPUser: emailSettings.SMTPUser,
		SMTPPass: emailSettings.SMTPPass,
	})
	if err != nil {
		panic("mailer setup failed: " + err.Error())
	}
	authService.SetMailer(mailerProvider, emailSettings.PublicBaseURL)
	shareService := shares.NewService(db, filerepo.New(), sharerepo.New())

	r.StaticFS("/static", web.StaticFS("static"))
	r.StaticFS("/web/pages", web.StaticFS("pages"))
	r.GET("/sw.js", handlers.ServiceWorkerJS())
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
	r.GET("/s/:token", handlers.PublicShareView(shareService, filesService, cfg.CookieSecret))
	r.POST("/s/:token", handlers.PublicShareUnlock(shareService, filesService, cfg.CookieSecret))
	r.GET("/login", handlers.WebLoginGet(authService, setupService))
	protected := r.Group("/")
	protected.Use(middleware.RequireSessionRedirect(authService))
	protected.GET("/lock", handlers.WebLockGet())
	protected.GET("/dashboard", handlers.WebDashboard(filesService, folderService, settingsService))
	protected.GET("/files", handlers.WebFiles(filesService, folderService))
	protected.GET("/folders/:id", handlers.WebFolder(filesService, folderService))
	protected.GET("/files/:id/view", handlers.WebFileView(filesService))
	protected.GET("/shares", handlers.WebShares(shareService, filesService))
	protected.GET("/settings", handlers.WebSettings(filesService, settingsService))
	protected.POST("/settings/storage", handlers.WebSettingsStoragePost(settingsService))
	protected.POST("/settings/email", handlers.WebSettingsEmailPost(settingsService))
	protected.POST("/settings/uploads", handlers.WebSettingsUploadPost(settingsService))
	protected.POST("/logout", handlers.WebLogout(authService))

	api := r.Group("/api")
	{
		api.POST("/auth/login", handlers.APILogin(authService))
		api.POST("/auth/unlock", middleware.RequireSessionJSON(authService), handlers.APIUnlockVault(authService))
		api.GET("/me", middleware.RequireSessionJSON(authService), handlers.APIMe(authService))
		api.GET("/health", handlers.Health(db))
		api.GET("/public/shares/:token", handlers.APIPublicShareRecord(shareService, filesService, cfg.CookieSecret))
		api.POST("/public/shares/:token/consume", handlers.APIPublicShareConsume(shareService, filesService, cfg.CookieSecret))
		api.POST("/search", middleware.RequireSessionJSON(authService), handlers.APISearch(filesService))
	}

	apiUploads := api.Group("/uploads")
	apiUploads.Use(middleware.RequireSessionJSON(authService))
	apiUploads.Use(middleware.LimitBody(2 << 20))
	{
		apiUploads.POST("/start", handlers.APIUploadStart(uploadService))
		apiUploads.POST("/:id/parts", handlers.APIUploadPartRecord(uploadService))
		apiUploads.POST("/:id/parts/presign", handlers.APIUploadPartPresignBatch(uploadService))
		apiUploads.POST("/:id/parts/:part/presign", handlers.APIUploadPartPresign(uploadService))
		apiUploads.POST("/:id/thumbnail/presign", handlers.APIThumbnailPresign(uploadService))
		apiUploads.POST("/:id/complete", handlers.APIUploadComplete(uploadService))
		apiUploads.POST("/:id/cancel", handlers.APIUploadCancel(uploadService))
	}

	apiFiles := api.Group("/files")
	apiFiles.Use(middleware.RequireSessionJSON(authService))
	{
		apiFiles.GET("/:id/record", handlers.APIFileRecord(filesService))
		apiFiles.GET("/:id/thumbnail", handlers.APIThumbnailRedirect(filesService))
		apiFiles.GET("/:id/share", handlers.APIGetShareForFile(shareService))
		apiFiles.GET("/:id/download", handlers.APIDownloadFile(filesService))
		apiFiles.GET("/:id/media", handlers.APIMediaRedirect(filesService))
		apiFiles.POST("/:id/share", handlers.APICreateShare(shareService))
	}

	apiShares := api.Group("/shares")
	apiShares.Use(middleware.RequireSessionJSON(authService))
	{
		apiShares.PATCH("/:id", handlers.APIUpdateShare(shareService))
		apiShares.POST("/:id/revoke", handlers.APIRevokeShare(shareService))
		apiShares.POST("/:id/activate", handlers.APIActivateShare(shareService))
		apiShares.GET("/:id/crypto-record", handlers.APIGetShareCryptoRecord(shareService))
		apiShares.DELETE("/:id", handlers.APIDeleteShare(shareService))
	}

	apiFolders := api.Group("")
	apiFolders.Use(middleware.RequireSessionJSON(authService))
	{
		apiFolders.POST("/folders", handlers.APICreateFolder(folderService))
		apiFolders.GET("/folders/root/entries", handlers.APIListRootFolderEntries(folderService))
		apiFolders.GET("/folders/:id/entries", handlers.APIListFolderEntries(folderService))
		apiFolders.POST("/entries/delete", handlers.APIDeleteEntries(folderService))
		apiFolders.POST("/entries/move", handlers.APIMoveEntries(folderService))
		apiFolders.POST("/entries/rename", handlers.APIRenameEntry(folderService, filesService))
	}

	r.NoRoute(handlers.WebNotFound())

	return r
}
