package main

import (
	"context"
	"log"
	"strings"
	"time"

	"arkive/core/config"
	"arkive/core/database"
	"arkive/core/models"
	filerepo "arkive/core/repositories/files"
	folderrepo "arkive/core/repositories/folders"
	settingsrepo "arkive/core/repositories/settings"
	storagerepo "arkive/core/repositories/storage"
	uploadrepo "arkive/core/repositories/uploads"
	usagerepo "arkive/core/repositories/usage"
	usersrepo "arkive/core/repositories/users"
	"arkive/core/router"
	"arkive/core/services/storageprovider"
	"arkive/core/services/uploads"
	"arkive/migrations"
	"arkive/pkg/jobs"
	"arkive/pkg/storage/localclient"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: no .env file loaded: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connection failed: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("db ping failed: %v", err)
	}

	if err := migrations.Run(context.Background(), db); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	settingsRepo := settingsrepo.New()
	if err := seedStorageSettings(context.Background(), db, settingsRepo, cfg); err != nil {
		log.Fatalf("storage settings seed failed: %v", err)
	}
	localStorage := localclient.New(func(ctx context.Context) (string, error) {
		settings, err := settingsRepo.GetStorageSettings(ctx, db)
		if err != nil {
			return "", err
		}
		return settings.LocalPath, nil
	})
	storageProvider := storageprovider.New(db, settingsRepo, localStorage)

	uploadService := uploads.NewService(
		db,
		storagerepo.New(),
		filerepo.New(),
		folderrepo.New(),
		uploadrepo.New(),
		usagerepo.New(),
		usersrepo.New(),
		storageProvider,
		uploads.Config{
			UploadExpires:        15 * time.Minute,
			DownloadExpire:       3 * time.Hour,
			ShareDownloadExpire:  30 * time.Minute,
			MaxFileSizeBytes:     cfg.MaxFileSizeBytes,
			MaxUploadConcurrency: cfg.MaxUploadConcurrency,
			MaxQueueItems:        cfg.MaxQueueItems,
		})

	cleanupCron, err := jobs.StartUploadCleanup(uploadService)
	if err != nil {
		log.Fatalf("cleanup cron failed: %v", err)
	}
	defer cleanupCron.Stop()

	if strings.EqualFold(cfg.Env, "dev") {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := router.New(db, cfg, uploadService, localStorage)

	if err := r.Run(cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func seedStorageSettings(ctx context.Context, db database.PgPool, settingsRepo *settingsrepo.Repository, cfg config.Config) error {
	exists, err := settingsRepo.HasStorageSettings(ctx, db)
	if err != nil || exists {
		return err
	}
	if cfg.S3AccessKeyID == "" || cfg.S3SecretAccessKey == "" || cfg.S3Bucket == "" || cfg.S3Endpoint == "" {
		return nil
	}
	return settingsRepo.SaveStorageSettings(ctx, db, models.StorageSettings{
		Provider:          "s3",
		MaxStorageBytes:   0,
		S3AccessKeyID:     cfg.S3AccessKeyID,
		S3SecretAccessKey: cfg.S3SecretAccessKey,
		S3SessionToken:    cfg.S3SessionToken,
		S3Bucket:          cfg.S3Bucket,
		S3Endpoint:        cfg.S3Endpoint,
		S3Region:          cfg.S3Region,
	})
}
