package main

import (
	"context"
	"log"
	"strings"
	"time"

	"arkive/core/config"
	"arkive/core/database"
	filerepo "arkive/core/repositories/files"
	folderrepo "arkive/core/repositories/folders"
	storagerepo "arkive/core/repositories/storage"
	uploadrepo "arkive/core/repositories/uploads"
	usagerepo "arkive/core/repositories/usage"
	usersrepo "arkive/core/repositories/users"
	"arkive/core/router"
	"arkive/core/services/uploads"
	"arkive/migrations"
	"arkive/pkg/jobs"
	"arkive/pkg/storage/s3client"

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

	storageClient, err := s3client.New(context.Background(), s3client.Config{
		AccessKeyID:     cfg.S3AccessKeyID,
		SecretAccessKey: cfg.S3SecretAccessKey,
		SessionToken:    cfg.S3SessionToken,
		Bucket:          cfg.S3Bucket,
		Endpoint:        cfg.S3Endpoint,
		Region:          cfg.S3Region,
	})
	if err != nil {
		log.Fatalf("s3 client failed: %v", err)
	}

	uploadService := uploads.NewService(
		db,
		storagerepo.New(),
		filerepo.New(),
		folderrepo.New(),
		uploadrepo.New(),
		usagerepo.New(),
		usersrepo.New(),
		storageClient,
		uploads.Config{
			UploadExpires:       15 * time.Minute,
			DownloadExpire:      3 * time.Hour,
			ShareDownloadExpire: 30 * time.Minute,
			MaxFileSizeBytes:    cfg.MaxFileSizeBytes,
			MaxUploadConcurrency: cfg.MaxUploadConcurrency,
			MaxQueueItems:       cfg.MaxQueueItems,
		})

	cleanupCron, err := jobs.StartUploadCleanup(uploadService)
	if err != nil {
		log.Fatalf("cleanup cron failed: %v", err)
	}
	defer cleanupCron.Stop()

	retentionCron, err := jobs.StartRetentionCleanup(uploadService)
	if err != nil {
		log.Fatalf("retention cron failed: %v", err)
	}
	defer retentionCron.Stop()

	if strings.EqualFold(cfg.Env, "dev") {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := router.New(db, cfg, uploadService)

	if err := r.Run(cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
