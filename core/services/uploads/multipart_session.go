package uploads

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"arkive/core/models"
	"arkive/pkg/storage"
	"arkive/pkg/validation"
)

const (
	UploadStatusActive    = "active"
	UploadStatusCompleted = "completed"
	UploadStatusAborted   = "aborted"
	UploadStatusExpired   = "expired"
	UploadStatusFailed    = "failed"

	FileUploadPending   = "pending"
	FileUploadUploading = "uploading"
	FileUploadComplete  = "complete"
	FileUploadAborted   = "aborted"
	FileUploadFailed    = "failed"

	UploadPartPending  = "pending"
	UploadPartComplete = "complete"

	S3MultipartMinPartSizeBytes int64 = 5 * 1024 * 1024
	MaxPresignBatchParts              = 32
)

type MultipartUploadStartInput struct {
	OriginalSize      int64
	FileChunkSize     int64
	TotalChunks       int
	UploadPartSize    int64
	UploadPartCount   int
	EncryptionVersion int16
	FolderID          *string
}

type MultipartUploadCompleteInput struct {
	EncryptedMetadata string
	EncryptedFileKey  string
	EncryptedManifest string
	EncryptedHash     string
	HasThumbnail      bool
	ThumbnailMime     string
	ThumbnailWidth    int
	ThumbnailHeight   int
}

type UploadPartRecordInput struct {
	PartNumber    int
	EncryptedHash string
	ETag          string
}

type ThumbnailUploadInput struct {
	EncryptedSize int64
	Mime          string
	Width         int
	Height        int
}

func (s *Service) StartMultipartUploadSession(ctx context.Context, userID string, input MultipartUploadStartInput) (models.UploadStartResponse, validation.Errors, error) {
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
	settings, err := s.settingsRepo.GetUploadSettings(ctx, s.db)
	if err != nil {
		settings = models.UploadSettings{MaxQueueItems: 300}
	}
	if inFlight >= int64(settings.MaxQueueItems) {
		validationErrors.Add("queue", ErrQueueLimitReached.Error())
	}
	if input.OriginalSize <= 0 {
		validationErrors.Add("size", ErrFileSizeRequired.Error())
	}
	if input.FileChunkSize <= 0 {
		validationErrors.Add("fileChunkSize", "file chunk size is required")
	}
	if input.TotalChunks <= 0 {
		validationErrors.Add("totalChunks", "total chunks is required")
	}
	if input.UploadPartSize <= 0 {
		validationErrors.Add("uploadPartSize", "upload part size is required")
	}
	if input.UploadPartCount <= 0 {
		validationErrors.Add("uploadPartCount", "upload part count is required")
	}
	if input.UploadPartCount > 1 {
		provider, providerErr := s.storage.ActiveProvider(ctx)
		if providerErr != nil {
			return models.UploadStartResponse{}, nil, providerErr
		}
		if provider == "s3" && input.UploadPartSize < S3MultipartMinPartSizeBytes {
			validationErrors.Add("uploadPartSize", "upload part size must be at least 5 MiB for multipart S3 uploads")
		}
	}
	if validationErrors.HasAny() {
		return models.UploadStartResponse{}, validationErrors, nil
	}

	expectedChunks := int((input.OriginalSize + input.FileChunkSize - 1) / input.FileChunkSize)
	if expectedChunks != input.TotalChunks {
		validationErrors.Add("totalChunks", "total chunks does not match original size")
		return models.UploadStartResponse{}, validationErrors, nil
	}
	chunksPerUploadPart := int(input.UploadPartSize / input.FileChunkSize)
	if chunksPerUploadPart <= 0 {
		validationErrors.Add("uploadPartSize", "upload part size must be greater than or equal to file chunk size")
		return models.UploadStartResponse{}, validationErrors, nil
	}
	expectedUploadParts := (input.TotalChunks + chunksPerUploadPart - 1) / chunksPerUploadPart
	if expectedUploadParts != input.UploadPartCount {
		validationErrors.Add("uploadPartCount", "upload part count does not match original size")
		return models.UploadStartResponse{}, validationErrors, nil
	}
	if input.UploadPartSize%input.FileChunkSize != 0 {
		validationErrors.Add("uploadPartSize", "upload part size must align to file chunk size")
		return models.UploadStartResponse{}, validationErrors, nil
	}
	declaredEncryptedSize := reservedUploadSize(input.OriginalSize, input.TotalChunks)
	storageSettings, err := s.settingsRepo.GetStorageSettings(ctx, s.db)
	if err != nil {
		return models.UploadStartResponse{}, nil, err
	}
	input.FolderID, err = validateOptionalFolderID(input.FolderID)
	if err != nil {
		return models.UploadStartResponse{}, nil, err
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.UploadStartResponse{}, nil, err
	}

	if input.FolderID != nil {
		if _, err := s.folderRepo.GetForUser(ctx, tx, userID, *input.FolderID); err != nil {
			_ = tx.Rollback(ctx)
			if errors.Is(err, pgx.ErrNoRows) {
				return models.UploadStartResponse{}, nil, ErrNotFound
			}
			return models.UploadStartResponse{}, nil, err
		}
	}

	file, err := s.fileRepo.CreateEncryptedFile(ctx, tx, models.File{
		UserID:              userID,
		FolderID:            input.FolderID,
		EncryptedMetadata:   []byte{},
		EncryptedFileKey:    []byte{},
		EncryptedManifest:   []byte{},
		EncryptionVersion:   input.EncryptionVersion,
		ChunkSize:           input.FileChunkSize,
		ChunkCount:          input.TotalChunks,
		PlaintextSize:       input.OriginalSize,
		ActualEncryptedSize: 0,
		UploadStatus:        FileUploadUploading,
		ExpiresAt:           expiresAtPtr(time.Now().Add(s.uploadExpires)),
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		return models.UploadStartResponse{}, nil, err
	}
	objectKey, err := storage.BuildObjectKey(userID, file.ID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return models.UploadStartResponse{}, nil, err
	}
	providerUploadID, err := s.storage.CreateMultipartUpload(ctx, objectKey, "application/octet-stream")
	if err != nil {
		_ = tx.Rollback(ctx)
		return models.UploadStartResponse{}, nil, err
	}

	reserved, err := s.storageRepo.ReserveStorage(ctx, tx, userID, declaredEncryptedSize, storageSettings.MaxStorageBytes)
	if err != nil {
		_ = tx.Rollback(ctx)
		return models.UploadStartResponse{}, nil, err
	}
	if !reserved {
		_ = tx.Rollback(ctx)
		usedBytes, reservedBytes, usageErr := s.userRepo.GetStorageUsage(ctx, s.db, userID)
		if usageErr != nil {
			return models.UploadStartResponse{}, nil, ErrStorageLimitExceeded
		}
		return models.UploadStartResponse{}, nil, &StorageLimitExceededError{
			MaxBytes:       storageSettings.MaxStorageBytes,
			UsedBytes:      usedBytes + reservedBytes,
			RequestedBytes: declaredEncryptedSize,
		}
	}

	session, err := s.uploadRepo.CreateUploadSession(ctx, tx, models.UploadSession{
		FileID:           file.ID,
		ProviderUploadID: providerUploadID,
		UploadPartCount:  input.UploadPartCount,
		Status:           UploadStatusActive,
		ExpiresAt:        time.Now().Add(s.uploadExpires),
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		return models.UploadStartResponse{}, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		_ = s.storage.AbortMultipartUpload(ctx, objectKey, providerUploadID)
		return models.UploadStartResponse{}, nil, err
	}

	return models.UploadStartResponse{
		FileID:           file.ID,
		VaultID:          file.UserID,
		UploadSessionID:  session.ID,
		ProviderUploadID: providerUploadID,
		FileChunkSize:    input.FileChunkSize,
		TotalChunks:      input.TotalChunks,
		UploadPartSize:   input.UploadPartSize,
		UploadPartCount:  input.UploadPartCount,
	}, nil, nil
}

func (s *Service) PresignMultipartUploadPart(ctx context.Context, userID, uploadSessionID string, partNumber int) (string, error) {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return "", err
	}
	uploadSessionID, err = validateUploadID(uploadSessionID)
	if err != nil {
		return "", err
	}
	if partNumber <= 0 {
		return "", ErrInvalidInput
	}

	uploadSession, err := s.uploadRepo.GetUploadSessionForUser(ctx, s.db, uploadSessionID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	if uploadSession.Status != UploadStatusActive || isExpired(&uploadSession.ExpiresAt) {
		return "", ErrUploadCancelled
	}
	if uploadSession.UploadPartCount > 0 && partNumber > uploadSession.UploadPartCount {
		return "", ErrInvalidInput
	}

	if err := s.uploadRepo.UpdateUploadSessionStatus(ctx, s.db, uploadSessionID, UploadStatusActive); err != nil {
		return "", err
	}
	objectKey, err := storage.BuildObjectKey(userID, uploadSession.FileID)
	if err != nil {
		return "", err
	}
	return s.storage.PresignUploadPart(ctx, objectKey, uploadSession.ProviderUploadID, int32(partNumber), s.uploadExpires)
}

func (s *Service) PresignMultipartUploadParts(ctx context.Context, userID, uploadSessionID string, partNumbers []int) (map[string]string, error) {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return nil, err
	}
	uploadSessionID, err = validateUploadID(uploadSessionID)
	if err != nil {
		return nil, err
	}
	if len(partNumbers) == 0 {
		return nil, ErrInvalidInput
	}
	if len(partNumbers) > MaxPresignBatchParts {
		return nil, ErrInvalidInput
	}

	uploadSession, err := s.uploadRepo.GetUploadSessionForUser(ctx, s.db, uploadSessionID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if uploadSession.Status != UploadStatusActive || isExpired(&uploadSession.ExpiresAt) {
		return nil, ErrUploadCancelled
	}

	if err := s.uploadRepo.UpdateUploadSessionStatus(ctx, s.db, uploadSessionID, UploadStatusActive); err != nil {
		return nil, err
	}
	objectKey, err := storage.BuildObjectKey(userID, uploadSession.FileID)
	if err != nil {
		return nil, err
	}

	urls := make(map[string]string, len(partNumbers))
	seen := make(map[int]struct{}, len(partNumbers))
	for _, partNumber := range partNumbers {
		if partNumber <= 0 {
			return nil, ErrInvalidInput
		}
		if uploadSession.UploadPartCount > 0 && partNumber > uploadSession.UploadPartCount {
			return nil, ErrInvalidInput
		}
		if _, exists := seen[partNumber]; exists {
			continue
		}
		seen[partNumber] = struct{}{}

		url, presignErr := s.storage.PresignUploadPart(ctx, objectKey, uploadSession.ProviderUploadID, int32(partNumber), s.uploadExpires)
		if presignErr != nil {
			return nil, presignErr
		}
		urls[strconv.Itoa(partNumber)] = url
	}

	return urls, nil
}

func (s *Service) RecordMultipartUploadPart(ctx context.Context, userID, uploadSessionID string, input UploadPartRecordInput) error {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return err
	}
	uploadSessionID, err = validateUploadID(uploadSessionID)
	if err != nil {
		return err
	}
	if input.PartNumber <= 0 || strings.TrimSpace(input.EncryptedHash) == "" {
		return ErrInvalidInput
	}

	uploadSession, err := s.uploadRepo.GetUploadSessionForUser(ctx, s.db, uploadSessionID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if uploadSession.Status != UploadStatusActive || isExpired(&uploadSession.ExpiresAt) {
		return ErrUploadCancelled
	}

	encryptedHash, err := base64.StdEncoding.DecodeString(strings.TrimSpace(input.EncryptedHash))
	if err != nil {
		return ErrInvalidInput
	}
	_, err = s.uploadRepo.UpsertUploadPart(ctx, s.db, models.UploadPart{
		UploadSessionID: uploadSessionID,
		PartNumber:      input.PartNumber,
		ETag:            strings.TrimSpace(input.ETag),
		EncryptedHash:   encryptedHash,
	})
	return err
}

func (s *Service) CompleteMultipartUploadSession(ctx context.Context, userID, uploadSessionID string, input MultipartUploadCompleteInput) error {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return err
	}
	uploadSessionID, err = validateUploadID(uploadSessionID)
	if err != nil {
		return err
	}

	uploadSession, err := s.uploadRepo.GetUploadSessionForUser(ctx, s.db, uploadSessionID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if uploadSession.Status != UploadStatusActive {
		return ErrUploadCancelled
	}
	if strings.TrimSpace(input.EncryptedMetadata) == "" || strings.TrimSpace(input.EncryptedFileKey) == "" || strings.TrimSpace(input.EncryptedManifest) == "" {
		return ErrInvalidInput
	}
	if input.HasThumbnail {
		if strings.TrimSpace(input.ThumbnailMime) != thumbnailMimeWebP || input.ThumbnailWidth <= 0 || input.ThumbnailHeight <= 0 {
			return ErrInvalidInput
		}
	}

	file, err := s.fileRepo.GetEncryptedFileForUser(ctx, s.db, uploadSession.FileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	storageSettings, err := s.settingsRepo.GetStorageSettings(ctx, s.db)
	if err != nil {
		return err
	}

	parts, err := s.uploadRepo.ListUploadParts(ctx, s.db, uploadSessionID)
	if err != nil {
		return err
	}
	if uploadSession.UploadPartCount <= 0 || len(parts) != uploadSession.UploadPartCount {
		return ErrPartsRequired
	}
	encryptedMetadata, err := base64.StdEncoding.DecodeString(strings.TrimSpace(input.EncryptedMetadata))
	if err != nil {
		return ErrInvalidInput
	}
	encryptedFileKey, err := base64.StdEncoding.DecodeString(strings.TrimSpace(input.EncryptedFileKey))
	if err != nil {
		return ErrInvalidInput
	}
	encryptedManifest, err := base64.StdEncoding.DecodeString(strings.TrimSpace(input.EncryptedManifest))
	if err != nil {
		return ErrInvalidInput
	}
	var encryptedHash []byte
	if strings.TrimSpace(input.EncryptedHash) != "" {
		encryptedHash, err = base64.StdEncoding.DecodeString(strings.TrimSpace(input.EncryptedHash))
		if err != nil {
			return ErrInvalidInput
		}
	}
	thumbnailObjectKey := ""
	if input.HasThumbnail {
		thumbnailObjectKey, err = storage.BuildThumbnailObjectKey(userID, uploadSession.FileID)
		if err != nil {
			return err
		}
	}

	completedParts := make([]storage.CompletedPart, 0, len(parts))
	for idx, part := range parts {
		if part.PartNumber != idx+1 {
			return ErrPartsRequired
		}
		completedParts = append(completedParts, storage.CompletedPart{
			PartNumber: int32(part.PartNumber),
			ETag:       part.ETag,
		})
	}

	objectKey, err := storage.BuildObjectKey(userID, uploadSession.FileID)
	if err != nil {
		return err
	}
	if err := s.storage.CompleteMultipartUpload(ctx, objectKey, uploadSession.ProviderUploadID, completedParts); err != nil {
		if thumbnailObjectKey != "" {
			_ = s.storage.DeleteObject(ctx, thumbnailObjectKey)
		}
		_ = s.uploadRepo.UpdateUploadSessionStatus(ctx, s.db, uploadSessionID, UploadStatusFailed)
		_, _ = s.fileRepo.UpdateEncryptedFileStatusIf(ctx, s.db, file.ID, FileUploadFailed, []string{FileUploadUploading, FileUploadPending})
		return err
	}
	actualEncryptedSize, err := objectSizeWithRetry(ctx, func(measureCtx context.Context) (int64, error) {
		return s.storage.ObjectSize(measureCtx, objectKey)
	})
	if err != nil {
		if deleteErr := s.storage.DeleteObject(ctx, objectKey); deleteErr != nil {
			log.Printf("uploads: failed to delete object after size lookup failure file=%s key=%s: %v", file.ID, objectKey, deleteErr)
		}
		if thumbnailObjectKey != "" {
			_ = s.storage.DeleteObject(ctx, thumbnailObjectKey)
		}
		_ = s.uploadRepo.UpdateUploadSessionStatus(ctx, s.db, uploadSessionID, UploadStatusFailed)
		_, _ = s.fileRepo.UpdateEncryptedFileStatusIf(ctx, s.db, file.ID, FileUploadFailed, []string{FileUploadUploading, FileUploadPending})
		tx, txErr := s.db.BeginTx(ctx, pgx.TxOptions{})
		if txErr == nil {
			_, _ = s.storageRepo.ReleaseReservedStorage(ctx, tx, userID, reservedUploadSize(file.PlaintextSize, file.ChunkCount))
			_ = tx.Commit(ctx)
		}
		return err
	}
	thumbnailSizeBytes := int64(0)
	hasStoredThumbnail := false
	if thumbnailObjectKey != "" {
		thumbnailSizeBytes, err = objectSizeWithRetry(ctx, func(measureCtx context.Context) (int64, error) {
			return s.storage.ObjectSize(measureCtx, thumbnailObjectKey)
		})
		if err != nil {
			if deleteErr := s.storage.DeleteObject(ctx, thumbnailObjectKey); deleteErr != nil {
				log.Printf("uploads: failed to delete thumbnail after size lookup failure file=%s key=%s: %v", file.ID, thumbnailObjectKey, deleteErr)
			}
			log.Printf("uploads: continuing without thumbnail after size lookup failure file=%s key=%s: %v", file.ID, thumbnailObjectKey, err)
			thumbnailObjectKey = ""
			thumbnailSizeBytes = 0
		} else if thumbnailSizeBytes <= 0 || thumbnailSizeBytes > thumbnailMaxEncryptedBytes {
			if deleteErr := s.storage.DeleteObject(ctx, thumbnailObjectKey); deleteErr != nil {
				log.Printf("uploads: failed to delete invalid thumbnail file=%s key=%s: %v", file.ID, thumbnailObjectKey, deleteErr)
			}
			log.Printf("uploads: continuing without thumbnail after invalid size file=%s key=%s size=%d", file.ID, thumbnailObjectKey, thumbnailSizeBytes)
			thumbnailObjectKey = ""
			thumbnailSizeBytes = 0
		} else {
			hasStoredThumbnail = true
		}
	}
	reservedSize := reservedUploadSize(file.PlaintextSize, file.ChunkCount)
	actualStoredSize := actualEncryptedSize + thumbnailSizeBytes

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	finalized, err := s.storageRepo.FinalizeReservedStorage(ctx, tx, userID, reservedSize, actualStoredSize, storageSettings.MaxStorageBytes)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if !finalized {
		_ = tx.Rollback(ctx)
		if deleteErr := s.storage.DeleteObject(ctx, objectKey); deleteErr != nil {
			log.Printf("uploads: failed to delete object after quota finalize failure file=%s key=%s: %v", file.ID, objectKey, deleteErr)
		}
		if thumbnailObjectKey != "" {
			if deleteErr := s.storage.DeleteObject(ctx, thumbnailObjectKey); deleteErr != nil {
				log.Printf("uploads: failed to delete thumbnail after quota finalize failure file=%s key=%s: %v", file.ID, thumbnailObjectKey, deleteErr)
			}
		}
		failTx, failErr := s.db.BeginTx(ctx, pgx.TxOptions{})
		if failErr != nil {
			return ErrStorageLimitExceeded
		}
		if err := s.uploadRepo.UpdateUploadSessionStatus(ctx, failTx, uploadSessionID, UploadStatusFailed); err != nil {
			_ = failTx.Rollback(ctx)
			return err
		}
		if _, err := s.fileRepo.UpdateEncryptedFileStatusIf(ctx, failTx, file.ID, FileUploadFailed, []string{FileUploadUploading, FileUploadPending}); err != nil {
			_ = failTx.Rollback(ctx)
			return err
		}
		if _, err := s.storageRepo.ReleaseReservedStorage(ctx, failTx, userID, reservedSize); err != nil {
			_ = failTx.Rollback(ctx)
			return err
		}
		if err := failTx.Commit(ctx); err != nil {
			return err
		}
		usedBytes, reservedBytes, usageErr := s.userRepo.GetStorageUsage(ctx, s.db, userID)
		if usageErr != nil {
			return ErrStorageLimitExceeded
		}
		return &StorageLimitExceededError{
			MaxBytes:       storageSettings.MaxStorageBytes,
			UsedBytes:      usedBytes + reservedBytes,
			RequestedBytes: actualStoredSize,
		}
	}
	if err := s.uploadRepo.UpdateUploadSessionStatus(ctx, tx, uploadSessionID, UploadStatusCompleted); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := s.fileRepo.UpdateEncryptedFileEnvelope(ctx, tx, file.ID, encryptedMetadata, encryptedFileKey, encryptedManifest, encryptedHash); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	thumbnailStatus := "none"
	thumbnailMime := ""
	thumbnailWidth := 0
	thumbnailHeight := 0
	if input.HasThumbnail && hasStoredThumbnail {
		thumbnailStatus = "complete"
		thumbnailMime = strings.TrimSpace(input.ThumbnailMime)
		thumbnailWidth = input.ThumbnailWidth
		thumbnailHeight = input.ThumbnailHeight
	} else if input.HasThumbnail {
		thumbnailStatus = "failed"
	}
	if err := s.fileRepo.UpdateThumbnailInfo(ctx, tx, file.ID, thumbnailStatus, thumbnailSizeBytes, thumbnailMime, thumbnailWidth, thumbnailHeight); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := s.fileRepo.MarkEncryptedFileComplete(ctx, tx, file.ID, actualEncryptedSize, encryptedHash); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) AbortMultipartUploadSession(ctx context.Context, userID, uploadSessionID string) error {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return err
	}
	uploadSessionID, err = validateUploadID(uploadSessionID)
	if err != nil {
		return err
	}

	uploadSession, err := s.uploadRepo.GetUploadSessionForUser(ctx, s.db, uploadSessionID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	file, err := s.fileRepo.GetEncryptedFileForUser(ctx, s.db, uploadSession.FileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	objectKey, keyErr := storage.BuildObjectKey(userID, uploadSession.FileID)
	if keyErr == nil {
		_ = s.storage.AbortMultipartUpload(ctx, objectKey, uploadSession.ProviderUploadID)
	}
	if thumbnailKey, thumbErr := storage.BuildThumbnailObjectKey(userID, uploadSession.FileID); thumbErr == nil {
		_ = s.storage.DeleteObject(ctx, thumbnailKey)
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if err := s.uploadRepo.UpdateUploadSessionStatus(ctx, tx, uploadSessionID, UploadStatusAborted); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	updated, err := s.fileRepo.UpdateEncryptedFileStatusIf(ctx, tx, file.ID, FileUploadAborted, []string{FileUploadPending, FileUploadUploading})
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if updated {
		released, err := s.storageRepo.ReleaseReservedStorage(ctx, tx, userID, reservedUploadSize(file.PlaintextSize, file.ChunkCount))
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if !released {
			_ = tx.Rollback(ctx)
			return ErrUploadFailed
		}
	}
	return tx.Commit(ctx)
}

func (s *Service) PresignThumbnailUpload(ctx context.Context, userID, uploadSessionID string, input ThumbnailUploadInput) (string, error) {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return "", err
	}
	uploadSessionID, err = validateUploadID(uploadSessionID)
	if err != nil {
		return "", err
	}
	if input.EncryptedSize <= 0 || input.EncryptedSize > thumbnailMaxEncryptedBytes {
		return "", ErrFileTooLarge
	}
	if strings.TrimSpace(input.Mime) != thumbnailMimeWebP || input.Width <= 0 || input.Height <= 0 {
		return "", ErrInvalidInput
	}

	uploadSession, err := s.uploadRepo.GetUploadSessionForUser(ctx, s.db, uploadSessionID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	if uploadSession.Status != UploadStatusActive || isExpired(&uploadSession.ExpiresAt) {
		return "", ErrUploadCancelled
	}
	if err := s.uploadRepo.UpdateUploadSessionStatus(ctx, s.db, uploadSessionID, UploadStatusActive); err != nil {
		return "", err
	}

	file, err := s.fileRepo.GetEncryptedFileForUser(ctx, s.db, uploadSession.FileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	if file.UploadStatus != FileUploadUploading && file.UploadStatus != FileUploadPending {
		if file.UploadStatus == FileUploadFailed || file.UploadStatus == FileUploadAborted {
			return "", ErrUploadCancelled
		}
		return "", ErrInvalidInput
	}

	objectKey, err := storage.BuildThumbnailObjectKey(userID, file.ID)
	if err != nil {
		return "", err
	}
	return s.storage.PresignUpload(ctx, objectKey, "application/octet-stream", s.uploadExpires)
}
