package uploads

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/jackc/pgx/v5"

	"arkive/core/database"
	"arkive/core/models"
	filerepo "arkive/core/repositories/files"
	storagerepo "arkive/core/repositories/storage"
	uploadrepo "arkive/core/repositories/uploads"
	"arkive/pkg/storage"
	"arkive/pkg/storage/r2"
	"arkive/pkg/validation"
)

const (
	FileStatusPending   = "pending"
	FileStatusUploading = "uploading"
	FileStatusComplete  = "complete"
	FileStatusFailed    = "failed"
	FileStatusAborted   = "aborted"

	MultipartStatusInitiated = "initiated"
	MultipartStatusUploading = "uploading"
	MultipartStatusCompleted = "completed"
	MultipartStatusAborted   = "aborted"
	MultipartStatusFailed    = "failed"
)

const (
	MaxFileSizeBytes        int64 = 1 * 1024 * 1024 * 1024
	MultipartThresholdBytes int64 = 200 * 1024 * 1024
)

type Service struct {
	db             database.PgPool
	storageRepo    *storagerepo.Repository
	fileRepo       *filerepo.Repository
	uploadRepo     *uploadrepo.Repository
	r2             *r2.Client
	bucket         string
	uploadExpires  time.Duration
	downloadExpire time.Duration
}

type Config struct {
	Bucket         string
	UploadExpires  time.Duration
	DownloadExpire time.Duration
}

func NewService(
	db database.PgPool,
	storageRepo *storagerepo.Repository,
	fileRepo *filerepo.Repository,
	uploadRepo *uploadrepo.Repository,
	r2Client *r2.Client,
	cfg Config,
) *Service {
	return &Service{
		db:             db,
		storageRepo:    storageRepo,
		fileRepo:       fileRepo,
		uploadRepo:     uploadRepo,
		r2:             r2Client,
		bucket:         cfg.Bucket,
		uploadExpires:  cfg.UploadExpires,
		downloadExpire: cfg.DownloadExpire,
	}
}

func (s *Service) StartMultipart(ctx context.Context, userID, filename string, sizeBytes int64, contentType string) (models.MultipartStartResponse, validation.Errors, error) {
	userID = strings.TrimSpace(userID)
	filename = strings.TrimSpace(filename)
	contentType = strings.TrimSpace(contentType)
	validationErrors := validation.New()
	if userID == "" {
		return models.MultipartStartResponse{}, nil, ErrUnauthorized
	}
	if filename == "" {
		validationErrors.Add("filename", ErrFilenameRequired.Error())
	}
	if sizeBytes <= 0 {
		validationErrors.Add("size", ErrFileSizeRequired.Error())
	}
	if sizeBytes > 0 {
		switch {
		case sizeBytes > MaxFileSizeBytes:
			validationErrors.Add("size", ErrFileTooLarge.Error())
		case sizeBytes <= MultipartThresholdBytes:
			validationErrors.Add("size", ErrFileTooSmall.Error())
		}
	}
	if validationErrors.HasAny() {
		return models.MultipartStartResponse{}, validationErrors, nil
	}

	chunkSize := storage.ChooseChunkSize(sizeBytes, MaxFileSizeBytes)
	partCount := storage.TotalParts(sizeBytes, chunkSize)

	objectKey, err := storage.BuildObjectKey(userID)
	if err != nil {
		return models.MultipartStartResponse{}, nil, err
	}

	uploadID, err := s.r2.CreateMultipartUpload(ctx, objectKey, contentType)
	if err != nil {
		return models.MultipartStartResponse{}, nil, err
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		_ = s.r2.AbortMultipartUpload(ctx, objectKey, uploadID)
		return models.MultipartStartResponse{}, nil, err
	}

	reserved, err := s.storageRepo.ReserveStorage(ctx, tx, userID, sizeBytes)
	if err != nil {
		_ = tx.Rollback(ctx)
		_ = s.r2.AbortMultipartUpload(ctx, objectKey, uploadID)
		return models.MultipartStartResponse{}, nil, err
	}
	if !reserved {
		_ = tx.Rollback(ctx)
		_ = s.r2.AbortMultipartUpload(ctx, objectKey, uploadID)
		validationErrors.Add("size", ErrQuotaExceeded.Error())
		return models.MultipartStartResponse{}, validationErrors, nil
	}

	file, err := s.fileRepo.CreateFile(ctx, tx, models.File{
		UserID:      userID,
		Bucket:      s.bucket,
		ObjectKey:   objectKey,
		Filename:    filename,
		ContentType: contentType,
		SizeBytes:   sizeBytes,
		Status:      FileStatusUploading,
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		_ = s.r2.AbortMultipartUpload(ctx, objectKey, uploadID)
		return models.MultipartStartResponse{}, nil, err
	}

	multipart, err := s.uploadRepo.CreateMultipart(ctx, tx, models.MultipartUpload{
		FileID:        file.ID,
		UploadID:      uploadID,
		Bucket:        s.bucket,
		ObjectKey:     objectKey,
		ChunkSize:     chunkSize,
		TotalParts:    partCount,
		UploadedParts: []byte("[]"),
		Status:        MultipartStatusInitiated,
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		_ = s.r2.AbortMultipartUpload(ctx, objectKey, uploadID)
		return models.MultipartStartResponse{}, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		_ = s.r2.AbortMultipartUpload(ctx, objectKey, uploadID)
		return models.MultipartStartResponse{}, nil, err
	}

	return models.MultipartStartResponse{
		FileID:      file.ID,
		MultipartID: multipart.ID,
		ObjectKey:   objectKey,
		ChunkSize:   chunkSize,
		TotalParts:  partCount,
	}, nil, nil
}

func (s *Service) StartSingleUpload(ctx context.Context, userID, filename string, sizeBytes int64, contentType string) (models.SingleStartResponse, validation.Errors, error) {
	userID = strings.TrimSpace(userID)
	filename = strings.TrimSpace(filename)
	contentType = strings.TrimSpace(contentType)
	validationErrors := validation.New()
	if userID == "" {
		return models.SingleStartResponse{}, nil, ErrUnauthorized
	}
	if filename == "" {
		validationErrors.Add("filename", ErrFilenameRequired.Error())
	}
	if sizeBytes <= 0 {
		validationErrors.Add("size", ErrFileSizeRequired.Error())
	}
	if sizeBytes > 0 {
		switch {
		case sizeBytes > MaxFileSizeBytes:
			validationErrors.Add("size", ErrFileTooLarge.Error())
		case sizeBytes > MultipartThresholdBytes:
			validationErrors.Add("size", ErrMultipartRequired.Error())
		}
	}
	if validationErrors.HasAny() {
		return models.SingleStartResponse{}, validationErrors, nil
	}

	objectKey, err := storage.BuildObjectKey(userID)
	if err != nil {
		return models.SingleStartResponse{}, nil, err
	}

	uploadURL, err := s.r2.PresignUpload(ctx, objectKey, contentType, s.uploadExpires)
	if err != nil {
		return models.SingleStartResponse{}, nil, err
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.SingleStartResponse{}, nil, err
	}

	reserved, err := s.storageRepo.ReserveStorage(ctx, tx, userID, sizeBytes)
	if err != nil {
		_ = tx.Rollback(ctx)
		return models.SingleStartResponse{}, nil, err
	}
	if !reserved {
		_ = tx.Rollback(ctx)
		validationErrors.Add("size", ErrQuotaExceeded.Error())
		return models.SingleStartResponse{}, validationErrors, nil
	}

	file, err := s.fileRepo.CreateFile(ctx, tx, models.File{
		UserID:      userID,
		Bucket:      s.bucket,
		ObjectKey:   objectKey,
		Filename:    filename,
		ContentType: contentType,
		SizeBytes:   sizeBytes,
		Status:      FileStatusUploading,
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		return models.SingleStartResponse{}, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.SingleStartResponse{}, nil, err
	}

	return models.SingleStartResponse{
		FileID:    file.ID,
		ObjectKey: objectKey,
		UploadURL: uploadURL,
		ExpiresAt: time.Now().Add(s.uploadExpires),
	}, nil, nil
}

func (s *Service) PresignPart(ctx context.Context, userID, multipartID string, partNumber int32) (string, error) {
	userID = strings.TrimSpace(userID)
	multipartID = strings.TrimSpace(multipartID)
	if userID == "" {
		return "", ErrUnauthorized
	}
	if multipartID == "" || partNumber <= 0 {
		return "", ErrInvalidInput
	}

	upload, err := s.uploadRepo.GetMultipartForUser(ctx, s.db, multipartID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	if upload.Status == MultipartStatusCompleted || upload.Status == MultipartStatusAborted {
		return "", ErrInvalidInput
	}
	if partNumber > int32(upload.TotalParts) {
		return "", ErrInvalidInput
	}

	if err := s.uploadRepo.TouchMultipart(ctx, s.db, multipartID, MultipartStatusUploading); err != nil {
		return "", err
	}

	return s.r2.PresignUploadPart(ctx, upload.ObjectKey, upload.UploadID, partNumber, s.uploadExpires)
}

func (s *Service) CompleteMultipart(ctx context.Context, userID, multipartID string, parts []models.CompletedPartInput) error {
	userID = strings.TrimSpace(userID)
	multipartID = strings.TrimSpace(multipartID)
	if userID == "" {
		return ErrUnauthorized
	}
	if multipartID == "" {
		return ErrInvalidInput
	}
	if len(parts) == 0 {
		return ErrPartsRequired
	}

	upload, err := s.uploadRepo.GetMultipartForUser(ctx, s.db, multipartID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if upload.Status == MultipartStatusCompleted || upload.Status == MultipartStatusAborted {
		return ErrInvalidInput
	}
	if len(parts) != upload.TotalParts {
		return ErrInvalidInput
	}

	type dedupeKey struct {
		PartNumber int32
	}
	seen := map[dedupeKey]bool{}
	completed := make([]types.CompletedPart, 0, len(parts))
	stored := make([]models.StoredPart, 0, len(parts))
	for _, part := range parts {
		if part.PartNumber <= 0 || strings.TrimSpace(part.ETag) == "" {
			return ErrInvalidInput
		}
		if part.PartNumber > int32(upload.TotalParts) {
			return ErrInvalidInput
		}
		key := dedupeKey{PartNumber: part.PartNumber}
		if seen[key] {
			return ErrInvalidInput
		}
		seen[key] = true
		etag := strings.TrimSpace(part.ETag)
		completed = append(completed, types.CompletedPart{
			PartNumber: &part.PartNumber,
			ETag:       &etag,
		})
		stored = append(stored, models.StoredPart{
			PartNumber: part.PartNumber,
			ETag:       etag,
			Size:       part.Size,
		})
	}

	sort.Slice(completed, func(i, j int) bool {
		if completed[i].PartNumber == nil || completed[j].PartNumber == nil {
			return false
		}
		return *completed[i].PartNumber < *completed[j].PartNumber
	})

	if err := s.r2.CompleteMultipartUpload(ctx, upload.ObjectKey, upload.UploadID, completed); err != nil {
		_ = s.uploadRepo.TouchMultipart(ctx, s.db, multipartID, MultipartStatusFailed)
		_ = s.fileRepo.UpdateFileStatus(ctx, s.db, upload.FileID, FileStatusFailed)
		file, fileErr := s.fileRepo.GetFileByID(ctx, s.db, upload.FileID)
		if fileErr == nil {
			_, _ = s.storageRepo.ReleaseReservedStorage(ctx, s.db, userID, file.SizeBytes)
		}
		return err
	}

	storedJSON, err := json.Marshal(stored)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	if err := s.uploadRepo.UpdateMultipart(ctx, tx, multipartID, MultipartStatusCompleted, storedJSON); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := s.fileRepo.UpdateFileStatus(ctx, tx, upload.FileID, FileStatusComplete); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	file, err := s.fileRepo.GetFileByID(ctx, tx, upload.FileID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	committed, err := s.storageRepo.CommitStorage(ctx, tx, userID, file.SizeBytes)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if !committed {
		_ = tx.Rollback(ctx)
		return ErrUploadFailed
	}

	return tx.Commit(ctx)
}

func (s *Service) AbortMultipart(ctx context.Context, userID, multipartID string) error {
	userID = strings.TrimSpace(userID)
	multipartID = strings.TrimSpace(multipartID)
	if userID == "" {
		return ErrUnauthorized
	}
	if multipartID == "" {
		return ErrInvalidInput
	}

	upload, err := s.uploadRepo.GetMultipartForUser(ctx, s.db, multipartID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if err := s.r2.AbortMultipartUpload(ctx, upload.ObjectKey, upload.UploadID); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if err := s.uploadRepo.UpdateMultipart(ctx, tx, multipartID, MultipartStatusAborted, upload.UploadedParts); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := s.fileRepo.UpdateFileStatus(ctx, tx, upload.FileID, FileStatusAborted); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	file, err := s.fileRepo.GetFileByID(ctx, tx, upload.FileID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
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

	return tx.Commit(ctx)
}

func (s *Service) CompleteSingleUpload(ctx context.Context, userID, fileID string) error {
	userID = strings.TrimSpace(userID)
	fileID = strings.TrimSpace(fileID)
	if userID == "" {
		return ErrUnauthorized
	}
	if fileID == "" {
		return ErrInvalidInput
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
	if file.Status == FileStatusComplete || file.Status == FileStatusAborted {
		_ = tx.Rollback(ctx)
		return ErrInvalidInput
	}

	if err := s.fileRepo.UpdateFileStatus(ctx, tx, fileID, FileStatusComplete); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	committed, err := s.storageRepo.CommitStorage(ctx, tx, userID, file.SizeBytes)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if !committed {
		_ = tx.Rollback(ctx)
		return ErrUploadFailed
	}

	return tx.Commit(ctx)
}

func (s *Service) AbortSingleUpload(ctx context.Context, userID, fileID string) error {
	userID = strings.TrimSpace(userID)
	fileID = strings.TrimSpace(fileID)
	if userID == "" {
		return ErrUnauthorized
	}
	if fileID == "" {
		return ErrInvalidInput
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

	if err := s.fileRepo.UpdateFileStatus(ctx, tx, fileID, FileStatusAborted); err != nil {
		_ = tx.Rollback(ctx)
		return err
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

	return tx.Commit(ctx)
}

func (s *Service) AbortUploadByFile(ctx context.Context, userID, fileID string) error {
	userID = strings.TrimSpace(userID)
	fileID = strings.TrimSpace(fileID)
	if userID == "" {
		return ErrUnauthorized
	}
	if fileID == "" {
		return ErrInvalidInput
	}

	_, err := s.fileRepo.GetFileForUser(ctx, s.db, fileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	upload, err := s.uploadRepo.GetMultipartForFile(ctx, s.db, fileID, userID)
	if err == nil {
		return s.AbortMultipart(ctx, userID, upload.ID)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return s.AbortSingleUpload(ctx, userID, fileID)
	}
	return err
}

func (s *Service) PresignDownload(ctx context.Context, userID, fileID string) (string, error) {
	userID = strings.TrimSpace(userID)
	fileID = strings.TrimSpace(fileID)
	if userID == "" {
		return "", ErrUnauthorized
	}
	if fileID == "" {
		return "", ErrInvalidInput
	}

	file, err := s.fileRepo.GetFileForUser(ctx, s.db, fileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}

	return s.r2.PresignDownload(ctx, file.ObjectKey, s.downloadExpire)
}

func (s *Service) ListPendingUploads(ctx context.Context, userID string) ([]models.File, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrUnauthorized
	}
	return s.fileRepo.ListPendingForUser(ctx, s.db, userID)
}
