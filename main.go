package main

import (
	"context"
	"log"
	"strings"
	"time"

	"arkive/core/config"
	"arkive/core/database"
	filerepo "arkive/core/repositories/files"
	foldersrepo "arkive/core/repositories/folders"
	settingsrepo "arkive/core/repositories/settings"
	storagerepo "arkive/core/repositories/storage"
	uploadrepo "arkive/core/repositories/uploads"
	usersrepo "arkive/core/repositories/users"
	"arkive/core/router"
	folderssvc "arkive/core/services/folders"
	settingssvc "arkive/core/services/settings"
	"arkive/core/services/storageprovider"
	"arkive/core/services/uploads"
	"arkive/migrations"

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
	localStorage := settingssvc.NewLocalStorage(db, settingsRepo)
	storageProvider := storageprovider.New(db, settingsRepo, localStorage)
	folderRepo := foldersrepo.New()
	fileRepo := filerepo.New()
	folderService := folderssvc.NewService(db, folderRepo, fileRepo)

	uploadService := uploads.NewService(
		db,
		storagerepo.New(),
		folderRepo,
		fileRepo,
		settingsRepo,
		uploadrepo.New(),
		usersrepo.New(),
		storageProvider,
		uploads.Config{
			UploadExpires:       15 * time.Minute,
			DownloadExpire:      3 * time.Hour,
			ShareDownloadExpire: 30 * time.Minute,
		})

	if strings.EqualFold(cfg.Env, "dev") {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := router.New(db, cfg, uploadService, folderService, localStorage)

	if err := r.Run(cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
