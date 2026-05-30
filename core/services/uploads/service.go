package uploads

import (
	"context"
	"time"

	"arkive/core/database"
	filerepo "arkive/core/repositories/files"
	folderrepo "arkive/core/repositories/folders"
	settingsrepo "arkive/core/repositories/settings"
	storagerepo "arkive/core/repositories/storage"
	uploadrepo "arkive/core/repositories/uploads"
	usersrepo "arkive/core/repositories/users"
	"arkive/pkg/storage"
)

const (
	FileStatusPending   = "pending"
	FileStatusUploading = "uploading"
	FileStatusComplete  = "complete"
	FileStatusFailed    = "failed"
	FileStatusAborted   = "aborted"
)

type Service struct {
	db            database.PgPool
	storageRepo   *storagerepo.Repository
	folderRepo    *folderrepo.Repository
	fileRepo      *filerepo.Repository
	settingsRepo  *settingsrepo.Repository
	uploadRepo    *uploadrepo.Repository
	userRepo      *usersrepo.Repository
	storage       storage.Provider
	uploadExpires time.Duration
}

type Config struct {
	UploadExpires       time.Duration
	DownloadExpire      time.Duration
	ShareDownloadExpire time.Duration
}

func NewService(
	db database.PgPool,
	storageRepo *storagerepo.Repository,
	folderRepo *folderrepo.Repository,
	fileRepo *filerepo.Repository,
	settingsRepo *settingsrepo.Repository,
	uploadRepo *uploadrepo.Repository,
	userRepo *usersrepo.Repository,
	storageProvider storage.Provider,
	cfg Config,
) *Service {
	return &Service{
		db:            db,
		storageRepo:   storageRepo,
		folderRepo:    folderRepo,
		fileRepo:      fileRepo,
		settingsRepo:  settingsRepo,
		uploadRepo:    uploadRepo,
		userRepo:      userRepo,
		storage:       storageProvider,
		uploadExpires: cfg.UploadExpires,
	}
}

func isExpired(expiresAt *time.Time) bool {
	if expiresAt == nil {
		return false
	}
	return time.Now().After(*expiresAt)
}

func (s *Service) uploadExpiry(ctx context.Context) time.Duration {
	settings, err := s.settingsRepo.GetUploadSettings(ctx, s.db)
	if err != nil || settings.StaleUploadHours <= 0 {
		return s.uploadExpires
	}
	return time.Duration(settings.StaleUploadHours) * time.Hour
}
