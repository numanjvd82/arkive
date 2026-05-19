package uploads

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"arkive/core/database"
	"arkive/core/models"
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
	db                  database.PgPool
	storageRepo         *storagerepo.Repository
	folderRepo          *folderrepo.Repository
	fileRepo            *filerepo.Repository
	settingsRepo        *settingsrepo.Repository
	uploadRepo          *uploadrepo.Repository
	userRepo            *usersrepo.Repository
	storage             storage.Provider
	uploadExpires       time.Duration
	downloadExpire      time.Duration
	shareDownloadExpire time.Duration
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
		db:                  db,
		storageRepo:         storageRepo,
		folderRepo:          folderRepo,
		fileRepo:            fileRepo,
		settingsRepo:        settingsRepo,
		uploadRepo:          uploadRepo,
		userRepo:            userRepo,
		storage:             storageProvider,
		uploadExpires:       cfg.UploadExpires,
		downloadExpire:      cfg.DownloadExpire,
		shareDownloadExpire: cfg.ShareDownloadExpire,
	}
}

func isExpired(expiresAt *time.Time) bool {
	if expiresAt == nil {
		return false
	}
	return time.Now().After(*expiresAt)
}

func (s *Service) CountArchivedFiles(ctx context.Context, userID string) (int64, error) {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return 0, err
	}
	return s.fileRepo.CountArchivedFilesForUser(ctx, s.db, userID)
}

func (s *Service) PresignDownload(ctx context.Context, userID, fileID string) (string, error) {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return "", err
	}
	fileID, err = validateUploadID(fileID)
	if err != nil {
		return "", err
	}

	file, err := s.fileRepo.GetFileForUser(ctx, s.db, fileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	if file.UploadStatus != FileStatusComplete {
		if file.UploadStatus == FileStatusFailed || file.UploadStatus == FileStatusAborted {
			return "", ErrUploadCancelled
		}
		return "", ErrNotFound
	}
	if isExpired(file.ExpiresAt) {
		return "", ErrNotFound
	}

	objectKey, err := storage.BuildObjectKey(userID, file.ID)
	if err != nil {
		return "", err
	}
	return s.storage.PresignDownload(ctx, objectKey, file.ID, "attachment", s.downloadExpire)
}

func (s *Service) DeleteFile(ctx context.Context, userID, fileID string) error {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return err
	}
	fileID, err = validateUploadID(fileID)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	file, err := s.fileRepo.GetFileForUser(ctx, tx, fileID, userID)
	if err != nil {
		_ = tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	switch file.UploadStatus {
	case FileStatusComplete:
		if err := s.storageRepo.DecreaseUsedStorage(ctx, tx, userID, totalStoredSize(file)); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	case FileStatusPending, FileStatusUploading:
		_ = tx.Rollback(ctx)
		return ErrUploadCancelled
	}

	deleted, err := s.fileRepo.DeleteFileForUser(ctx, tx, fileID, userID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if !deleted {
		_ = tx.Rollback(ctx)
		return ErrNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	if objectKey, keyErr := storage.BuildObjectKey(userID, file.ID); keyErr == nil {
		_ = s.storage.DeleteObject(ctx, objectKey)
	}
	if thumbnailKey, keyErr := storage.BuildThumbnailObjectKey(userID, file.ID); keyErr == nil {
		_ = s.storage.DeleteObject(ctx, thumbnailKey)
	}
	return nil
}

func (s *Service) DeleteFiles(ctx context.Context, userID string, fileIDs []string) (int, error) {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return 0, err
	}
	if len(fileIDs) == 0 {
		return 0, ErrInvalidInput
	}

	uniqueIDs := make([]string, 0, len(fileIDs))
	seen := make(map[string]struct{}, len(fileIDs))
	for _, rawID := range fileIDs {
		fileID, validateErr := validateUploadID(rawID)
		if validateErr != nil {
			return 0, validateErr
		}
		if _, exists := seen[fileID]; exists {
			continue
		}
		seen[fileID] = struct{}{}
		uniqueIDs = append(uniqueIDs, fileID)
	}
	if len(uniqueIDs) == 0 {
		return 0, ErrInvalidInput
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}

	files := make([]models.File, 0, len(uniqueIDs))
	for _, fileID := range uniqueIDs {
		file, getErr := s.fileRepo.GetFileForUser(ctx, tx, fileID, userID)
		if getErr != nil {
			_ = tx.Rollback(ctx)
			if errors.Is(getErr, pgx.ErrNoRows) {
				return 0, ErrNotFound
			}
			return 0, getErr
		}

		switch file.UploadStatus {
		case FileStatusComplete:
			if err := s.storageRepo.DecreaseUsedStorage(ctx, tx, userID, totalStoredSize(file)); err != nil {
				_ = tx.Rollback(ctx)
				return 0, err
			}
		case FileStatusPending, FileStatusUploading:
			_ = tx.Rollback(ctx)
			return 0, ErrUploadCancelled
		}

		deleted, deleteErr := s.fileRepo.DeleteFileForUser(ctx, tx, fileID, userID)
		if deleteErr != nil {
			_ = tx.Rollback(ctx)
			return 0, deleteErr
		}
		if !deleted {
			_ = tx.Rollback(ctx)
			return 0, ErrNotFound
		}
		files = append(files, file)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	for _, file := range files {
		if objectKey, keyErr := storage.BuildObjectKey(userID, file.ID); keyErr == nil {
			_ = s.storage.DeleteObject(ctx, objectKey)
		}
		if thumbnailKey, keyErr := storage.BuildThumbnailObjectKey(userID, file.ID); keyErr == nil {
			_ = s.storage.DeleteObject(ctx, thumbnailKey)
		}
	}
	return len(files), nil
}

type FileList struct {
	Files      []models.File
	TotalFiles int
}

func (s *Service) ListCompletedUploads(ctx context.Context, userID string, page, pageSize int) (FileList, error) {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return FileList{}, err
	}
	totalFiles, err := s.fileRepo.CountCompletedForUser(ctx, s.db, userID)
	if err != nil {
		return FileList{}, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 25
	}
	if totalFiles > 0 {
		totalPages := (totalFiles + pageSize - 1) / pageSize
		if page > totalPages {
			page = totalPages
		}
	}
	files, err := s.fileRepo.ListCompletedForUser(ctx, s.db, userID, page, pageSize)
	if err != nil {
		return FileList{}, err
	}

	return FileList{
		Files:      files,
		TotalFiles: totalFiles,
	}, nil
}

func (s *Service) SearchCompletedUploads(ctx context.Context, userID, query string, limit int) ([]models.File, error) {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return nil, err
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return []models.File{}, nil
	}
	return s.fileRepo.SearchCompletedForUser(ctx, s.db, userID, query, limit)
}
