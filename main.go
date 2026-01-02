package main

import (
	"context"
	"log"
	"strings"
	"time"

	"arkive/core/config"
	"arkive/core/database"
	filerepo "arkive/core/repositories/files"
	storagerepo "arkive/core/repositories/storage"
	uploadrepo "arkive/core/repositories/uploads"
	usagerepo "arkive/core/repositories/usage"
	"arkive/core/router"
	"arkive/core/services/uploads"
	"arkive/migrations"
	"arkive/pkg/jobs"
	"arkive/pkg/storage/r2"

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

	if err := migrations.Run(context.Background(), db, "migrations"); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	r2Client, err := r2.New(context.Background(), r2.Config{
		AccessKeyID:     cfg.R2AccessKeyID,
		SecretAccessKey: cfg.R2SecretAccessKey,
		SessionToken:    cfg.R2SessionToken,
		Bucket:          cfg.R2Bucket,
		Endpoint:        cfg.R2Endpoint,
		Region:          cfg.R2Region,
	})
	if err != nil {
		log.Fatalf("r2 client failed: %v", err)
	}

	uploadService := uploads.NewService(db, storagerepo.New(), filerepo.New(), uploadrepo.New(), usagerepo.New(), r2Client, uploads.Config{
		Bucket:              cfg.R2Bucket,
		UploadExpires:       15 * time.Minute,
		DownloadExpire:      3 * time.Hour,
		ShareDownloadExpire: 30 * time.Minute,
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

	r := router.New(db, cfg, uploadService)

	if err := r.Run(cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
