package files

import (
	"strings"
	"time"

	"arkive/core/database"
	"arkive/core/models"
	filerepo "arkive/core/repositories/files"
	storagerepo "arkive/core/repositories/storage"
	usersrepo "arkive/core/repositories/users"
	"arkive/pkg/storage"
)

type Service struct {
	db                  database.PgPool
	storageRepo         *storagerepo.Repository
	fileRepo            *filerepo.Repository
	userRepo            *usersrepo.Repository
	storage             storage.Provider
	downloadExpire      time.Duration
	shareDownloadExpire time.Duration
}

type Config struct {
	DownloadExpire      time.Duration
	ShareDownloadExpire time.Duration
}

func NewService(
	db database.PgPool,
	storageRepo *storagerepo.Repository,
	fileRepo *filerepo.Repository,
	userRepo *usersrepo.Repository,
	storageProvider storage.Provider,
	cfg Config,
) *Service {
	return &Service{
		db:                  db,
		storageRepo:         storageRepo,
		fileRepo:            fileRepo,
		userRepo:            userRepo,
		storage:             storageProvider,
		downloadExpire:      cfg.DownloadExpire,
		shareDownloadExpire: cfg.ShareDownloadExpire,
	}
}

func validateUserID(userID string) (string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", ErrUnauthorized
	}
	return userID, nil
}

func validateUploadID(uploadID string) (string, error) {
	uploadID = strings.TrimSpace(uploadID)
	if uploadID == "" {
		return "", ErrInvalidInput
	}
	return uploadID, nil
}

func isExpired(expiresAt *time.Time) bool {
	if expiresAt == nil {
		return false
	}
	return time.Now().After(*expiresAt)
}

func totalStoredSize(file models.File) int64 {
	total := file.ActualEncryptedSize
	if file.ThumbnailSizeBytes > 0 {
		total += file.ThumbnailSizeBytes
	}
	return total
}
