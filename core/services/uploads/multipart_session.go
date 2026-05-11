package uploads

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
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
)

type MultipartUploadStartInput struct {
	OriginalSize      int64
	PartSize          int64
	TotalParts        int
	EncryptionVersion int16
}

type MultipartUploadCompleteInput struct {
	EncryptedMetadata string
	EncryptedFileKey  string
	EncryptedManifest string
	EncryptedHash     string
}

type UploadPartRecordInput struct {
	PartNumber    int
	EncryptedHash string
	ETag          string
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
	if input.PartSize <= 0 {
		validationErrors.Add("partSize", "part size is required")
	}
	if input.TotalParts <= 0 {
		validationErrors.Add("totalParts", "total parts is required")
	}
	if validationErrors.HasAny() {
		return models.UploadStartResponse{}, validationErrors, nil
	}

	expectedParts := int((input.OriginalSize + input.PartSize - 1) / input.PartSize)
	if expectedParts != input.TotalParts {
		validationErrors.Add("totalParts", "total parts does not match original size")
		return models.UploadStartResponse{}, validationErrors, nil
	}
	declaredEncryptedSize := encryptedFileSize(input.OriginalSize, input.TotalParts)

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.UploadStartResponse{}, nil, err
	}

	file, err := s.fileRepo.CreateEncryptedFile(ctx, tx, models.File{
		UserID:              userID,
		EncryptedMetadata:   []byte{},
		EncryptedFileKey:    []byte{},
		EncryptedManifest:   []byte{},
		EncryptionVersion:   input.EncryptionVersion,
		ChunkSize:           input.PartSize,
		ChunkCount:          input.TotalParts,
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

	reserved, err := s.storageRepo.ReserveStorage(ctx, tx, userID, declaredEncryptedSize)
	if err != nil {
		_ = tx.Rollback(ctx)
		return models.UploadStartResponse{}, nil, err
	}
	if !reserved {
		_ = tx.Rollback(ctx)
		return models.UploadStartResponse{}, nil, ErrQuotaExceeded
	}

	session, err := s.uploadRepo.CreateUploadSession(ctx, tx, models.UploadSession{
		FileID:           file.ID,
		ProviderUploadID: providerUploadID,
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
		PartSize:         input.PartSize,
		TotalParts:       input.TotalParts,
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

	if err := s.uploadRepo.UpdateUploadSessionStatus(ctx, s.db, uploadSessionID, UploadStatusActive); err != nil {
		return "", err
	}
	objectKey, err := storage.BuildObjectKey(userID, uploadSession.FileID)
	if err != nil {
		return "", err
	}
	return s.storage.PresignUploadPart(ctx, objectKey, uploadSession.ProviderUploadID, int32(partNumber), s.uploadExpires)
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

	file, err := s.fileRepo.GetEncryptedFileForUser(ctx, s.db, uploadSession.FileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	parts, err := s.uploadRepo.ListUploadParts(ctx, s.db, uploadSessionID)
	if err != nil {
		return err
	}
	if len(parts) != file.ChunkCount {
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
		_ = s.uploadRepo.UpdateUploadSessionStatus(ctx, s.db, uploadSessionID, UploadStatusFailed)
		_, _ = s.fileRepo.UpdateEncryptedFileStatusIf(ctx, s.db, file.ID, FileUploadFailed, []string{FileUploadUploading, FileUploadPending})
		tx, txErr := s.db.BeginTx(ctx, pgx.TxOptions{})
		if txErr == nil {
			_, _ = s.storageRepo.ReleaseReservedStorage(ctx, tx, userID, encryptedFileSize(file.PlaintextSize, file.ChunkCount))
			_ = tx.Commit(ctx)
		}
		return err
	}
	reservedSize := encryptedFileSize(file.PlaintextSize, file.ChunkCount)

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	finalized, err := s.storageRepo.FinalizeReservedStorage(ctx, tx, userID, reservedSize, actualEncryptedSize)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if !finalized {
		_ = tx.Rollback(ctx)
		if deleteErr := s.storage.DeleteObject(ctx, objectKey); deleteErr != nil {
			log.Printf("uploads: failed to delete object after quota finalize failure file=%s key=%s: %v", file.ID, objectKey, deleteErr)
		}
		failTx, failErr := s.db.BeginTx(ctx, pgx.TxOptions{})
		if failErr != nil {
			return ErrQuotaExceeded
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
		return ErrQuotaExceeded
	}
	if err := s.uploadRepo.UpdateUploadSessionStatus(ctx, tx, uploadSessionID, UploadStatusCompleted); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := s.fileRepo.UpdateEncryptedFileEnvelope(ctx, tx, file.ID, encryptedMetadata, encryptedFileKey, encryptedManifest, encryptedHash); err != nil {
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
		released, err := s.storageRepo.ReleaseReservedStorage(ctx, tx, userID, encryptedFileSize(file.PlaintextSize, file.ChunkCount))
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
