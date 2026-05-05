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
	storagerepo "arkive/core/repositories/storage"
	usagerepo "arkive/core/repositories/usage"
	usersrepo "arkive/core/repositories/users"
	"arkive/pkg/storage"
	"arkive/pkg/validation"
)

const (
	FileStatusPending   = "pending"
	FileStatusUploading = "uploading"
	FileStatusComplete  = "complete"
	FileStatusFailed    = "failed"
	FileStatusAborted   = "aborted"
)

type Service struct {
	db                   database.PgPool
	storageRepo          *storagerepo.Repository
	fileRepo             *filerepo.Repository
	usageRepo            *usagerepo.Repository
	userRepo             *usersrepo.Repository
	storage              storage.Provider
	uploadExpires        time.Duration
	downloadExpire       time.Duration
	shareDownloadExpire  time.Duration
	maxFileSizeBytes     int64
	maxUploadConcurrency int
	maxQueueItems        int
}

type Config struct {
	UploadExpires        time.Duration
	DownloadExpire       time.Duration
	ShareDownloadExpire  time.Duration
	MaxFileSizeBytes     int64
	MaxUploadConcurrency int
	MaxQueueItems        int
}

func NewService(
	db database.PgPool,
	storageRepo *storagerepo.Repository,
	fileRepo *filerepo.Repository,
	usageRepo *usagerepo.Repository,
	userRepo *usersrepo.Repository,
	storageProvider storage.Provider,
	cfg Config,
) *Service {
	return &Service{
		db:                   db,
		storageRepo:          storageRepo,
		fileRepo:             fileRepo,
		usageRepo:            usageRepo,
		userRepo:             userRepo,
		storage:              storageProvider,
		uploadExpires:        cfg.UploadExpires,
		downloadExpire:       cfg.DownloadExpire,
		shareDownloadExpire:  cfg.ShareDownloadExpire,
		maxFileSizeBytes:     cfg.MaxFileSizeBytes,
		maxUploadConcurrency: cfg.MaxUploadConcurrency,
		maxQueueItems:        cfg.MaxQueueItems,
	}
}

func (s *Service) validateStartInput(userID, filename string, sizeBytes int64) (validation.Errors, error) {
	validationErrors := validation.New()
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUnauthorized
	}
	if strings.TrimSpace(filename) == "" {
		validationErrors.Add("filename", ErrFilenameRequired.Error())
	}
	if sizeBytes <= 0 {
		validationErrors.Add("size", ErrFileSizeRequired.Error())
	}
	if sizeBytes > s.maxFileSizeBytes {
		validationErrors.Add("size", ErrFileTooLarge.Error())
	}
	if validationErrors.HasAny() {
		return validationErrors, nil
	}
	return nil, nil
}

func isExpired(expiresAt *time.Time) bool {
	if expiresAt == nil {
		return false
	}
	return time.Now().After(*expiresAt)
}

func (s *Service) beginUploadTx(ctx context.Context, userID, objectKey, filename, contentType string, sizeBytes int64) (pgx.Tx, models.File, validation.Errors, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, models.File{}, nil, err
	}

	file, err := s.fileRepo.CreateFile(ctx, tx, models.File{
		UserID:      userID,
		ObjectKey:   objectKey,
		Filename:    filename,
		ContentType: contentType,
		SizeBytes:   sizeBytes,
		Status:      FileStatusPending,
		ExpiresAt:   expiresAtPtr(time.Now().Add(s.uploadExpires)),
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, models.File{}, nil, err
	}

	reserved, err := s.storageRepo.ReserveStorage(ctx, tx, userID, sizeBytes)
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, models.File{}, nil, err
	}
	if !reserved {
		_ = tx.Rollback(ctx)
		return nil, models.File{}, nil, ErrQuotaExceeded
	}

	updated, err := s.fileRepo.UpdateFileStatusIf(ctx, tx, file.ID, FileStatusUploading, []string{FileStatusPending})
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, models.File{}, nil, err
	}
	if !updated {
		_ = tx.Rollback(ctx)
		return nil, models.File{}, nil, ErrUploadFailed
	}

	return tx, file, nil, nil
}

func (s *Service) finalizeFileCompletion(ctx context.Context, tx pgx.Tx, userID string, file models.File, actualSize int64) error {
	if actualSize <= 0 {
		return ErrUploadFailed
	}
	if actualSize > s.maxFileSizeBytes {
		return ErrFileTooLarge
	}

	if err := s.fileRepo.UpdateFileSize(ctx, tx, file.ID, actualSize); err != nil {
		return err
	}
	if err := s.fileRepo.UpdateFileStatus(ctx, tx, file.ID, FileStatusComplete); err != nil {
		return err
	}
	if err := s.fileRepo.ClearFileExpiry(ctx, tx, file.ID); err != nil {
		return err
	}

	if actualSize > file.SizeBytes {
		extra := actualSize - file.SizeBytes
		reserved, err := s.storageRepo.ReserveStorage(ctx, tx, userID, extra)
		if err != nil {
			return err
		}
		if !reserved {
			return ErrQuotaExceeded
		}
	}

	committed, err := s.storageRepo.CommitStorage(ctx, tx, userID, actualSize)
	if err != nil {
		return err
	}
	if !committed {
		return ErrUploadFailed
	}

	if actualSize < file.SizeBytes {
		extra := file.SizeBytes - actualSize
		released, err := s.storageRepo.ReleaseReservedStorage(ctx, tx, userID, extra)
		if err != nil {
			return err
		}
		if !released {
			return ErrUploadFailed
		}
	}

	return nil
}

func (s *Service) markUploadFailed(ctx context.Context, userID string, file models.File) {
	updated, err := s.fileRepo.UpdateFileStatusIf(ctx, s.db, file.ID, FileStatusFailed, []string{FileStatusPending, FileStatusUploading})
	if err != nil || !updated {
		return
	}
	_, _ = s.storageRepo.ReleaseReservedStorage(ctx, s.db, userID, file.SizeBytes)
}

func (s *Service) StartUpload(ctx context.Context, userID, filename string, sizeBytes int64, contentType string) (models.UploadStartResponse, validation.Errors, error) {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return models.UploadStartResponse{}, nil, err
	}
	if err := s.TouchUserActivity(ctx, userID); err != nil {
		return models.UploadStartResponse{}, nil, err
	}

	inFlight, err := s.fileRepo.CountInFlightForUser(ctx, s.db, userID)
	if err != nil {
		return models.UploadStartResponse{}, nil, err
	}
	validationErrors := validation.New()
	if inFlight >= int64(s.maxQueueItems) {
		validationErrors.Add("queue", ErrQueueLimitReached.Error())
	}
	if inFlight >= int64(s.maxUploadConcurrency) {
		validationErrors.Add("queue", ErrConcurrentLimit.Error())
	}
	if validationErrors.HasAny() {
		return models.UploadStartResponse{}, validationErrors, nil
	}

	resp, validationErrors, err := s.StartSingleUpload(ctx, userID, filename, sizeBytes, contentType)
	if err != nil || (validationErrors != nil && validationErrors.HasAny()) {
		return models.UploadStartResponse{}, validationErrors, err
	}
	return models.UploadStartResponse{
		UploadID:  resp.FileID,
		FileID:    resp.FileID,
		ObjectKey: resp.ObjectKey,
		Mode:      "single",
		UploadURL: resp.UploadURL,
	}, nil, nil
}

func (s *Service) MonthlyUsage(ctx context.Context, userID string) (int64, error) {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return 0, err
	}
	cutoff := time.Now().Add(-30 * 24 * time.Hour)
	return s.usageRepo.SumUsageSince(ctx, s.db, userID, cutoff)
}

func (s *Service) CountActiveFiles(ctx context.Context, userID string) (int64, error) {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return 0, err
	}
	return s.fileRepo.CountActiveFilesForUser(ctx, s.db, userID)
}

func (s *Service) CountArchivedFiles(ctx context.Context, userID string) (int64, error) {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return 0, err
	}
	return s.fileRepo.CountArchivedFilesForUser(ctx, s.db, userID)
}

func (s *Service) StartSingleUpload(ctx context.Context, userID, filename string, sizeBytes int64, contentType string) (models.SingleStartResponse, validation.Errors, error) {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return models.SingleStartResponse{}, nil, err
	}
	filename = strings.TrimSpace(filename)
	contentType = strings.TrimSpace(contentType)

	validationErrors, err := s.validateStartInput(userID, filename, sizeBytes)
	if err != nil {
		return models.SingleStartResponse{}, nil, err
	}
	if validationErrors != nil && validationErrors.HasAny() {
		return models.SingleStartResponse{}, validationErrors, nil
	}

	objectKey, err := storage.BuildObjectKey(userID)
	if err != nil {
		return models.SingleStartResponse{}, nil, err
	}

	tx, file, valErrors, err := s.beginUploadTx(ctx, userID, objectKey, filename, contentType, sizeBytes)
	if err != nil {
		return models.SingleStartResponse{}, nil, err
	}
	if valErrors != nil && valErrors.HasAny() {
		return models.SingleStartResponse{}, valErrors, nil
	}

	uploadURL, err := s.storage.PresignUpload(ctx, objectKey, contentType, s.uploadExpires)
	if err != nil {
		_ = tx.Rollback(ctx)
		s.markUploadFailed(ctx, userID, file)
		return models.SingleStartResponse{}, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.SingleStartResponse{}, nil, err
	}

	return models.SingleStartResponse{
		FileID:    file.ID,
		ObjectKey: objectKey,
		UploadURL: uploadURL,
	}, nil, nil
}

func (s *Service) CompleteSingleUpload(ctx context.Context, userID, fileID string) error {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return err
	}
	fileID, err = validateUploadID(fileID)
	if err != nil {
		return err
	}

	file, err := s.fileRepo.GetFileForUser(ctx, s.db, fileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if file.Status == FileStatusComplete {
		return nil
	}
	if file.Status == FileStatusAborted || file.Status == FileStatusFailed {
		return ErrUploadCancelled
	}
	if isExpired(file.ExpiresAt) {
		_ = s.AbortSingleUpload(ctx, userID, file.ID)
		return ErrUploadCancelled
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	if err := s.finalizeFileCompletion(ctx, tx, userID, file, file.SizeBytes); err != nil {
		_ = tx.Rollback(ctx)
		s.markUploadFailed(ctx, userID, file)
		_ = s.storage.DeleteObject(ctx, file.ObjectKey)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	s.scheduleVideoMetadata(file)
	return nil
}

func (s *Service) AbortSingleUpload(ctx context.Context, userID, fileID string) error {
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
	if file.Status == FileStatusComplete || file.Status == FileStatusAborted || file.Status == FileStatusFailed {
		_ = tx.Rollback(ctx)
		return nil
	}

	updated, err := s.fileRepo.UpdateFileStatusIf(ctx, tx, fileID, FileStatusAborted, []string{FileStatusPending, FileStatusUploading})
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if !updated {
		_ = tx.Rollback(ctx)
		return nil
	}

	released, err := s.storageRepo.ReleaseReservedStorage(ctx, tx, userID, file.SizeBytes)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if !released {
		_ = tx.Rollback(ctx)
		return ErrUploadFailed
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	_ = s.storage.DeleteObject(ctx, file.ObjectKey)
	return nil
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
	if file.Status != FileStatusComplete {
		if file.Status == FileStatusFailed || file.Status == FileStatusAborted {
			return "", ErrUploadCancelled
		}
		return "", ErrNotFound
	}
	if isExpired(file.ExpiresAt) {
		return "", ErrNotFound
	}

	return s.storage.PresignDownload(ctx, file.ObjectKey, file.Filename, "attachment", s.downloadExpire)
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

	switch file.Status {
	case FileStatusComplete:
		if err := s.storageRepo.DecreaseUsedStorage(ctx, tx, userID, file.SizeBytes); err != nil {
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

	_ = s.storage.DeleteObject(ctx, file.ObjectKey)
	return nil
}

func (s *Service) ListPendingUploads(ctx context.Context, userID string) ([]models.File, error) {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return nil, err
	}
	return s.fileRepo.ListPendingForUser(ctx, s.db, userID)
}

type FileList struct {
	Files      []models.File
	TotalFiles int
}

func (s *Service) ListCompletedUploads(ctx context.Context, userID, sort string, page, pageSize int) (FileList, error) {
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
	files, err := s.fileRepo.ListCompletedForUser(ctx, s.db, userID, sort, page, pageSize)
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
