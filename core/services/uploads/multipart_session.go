package uploads

import (
	"context"
	"encoding/base64"
	"errors"
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
	EncryptedSize int64
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
	if inFlight >= int64(s.maxQueueItems) {
		validationErrors.Add("queue", ErrQueueLimitReached.Error())
	}
	if inFlight >= int64(s.maxUploadConcurrency) {
		validationErrors.Add("queue", ErrConcurrentLimit.Error())
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

	provider, err := s.storage.ActiveProvider(ctx)
	if err != nil {
		return models.UploadStartResponse{}, nil, err
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.UploadStartResponse{}, nil, err
	}

	file, err := s.fileRepo.CreateEncryptedFile(ctx, tx, models.File{
		UserID:            userID,
		EncryptedMetadata: []byte{},
		EncryptedFileKey:  []byte{},
		EncryptedManifest: []byte{},
		EncryptionVersion: input.EncryptionVersion,
		ChunkSize:         input.PartSize,
		ChunkCount:        input.TotalParts,
		PlaintextSize:     input.OriginalSize,
		UploadStatus:      FileUploadUploading,
		StorageBackend:    provider,
		ExpiresAt:         expiresAtPtr(time.Now().Add(s.uploadExpires)),
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

	reserved, err := s.storageRepo.ReserveStorage(ctx, tx, userID, input.OriginalSize)
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
		OwnerID:          userID,
		Provider:         provider,
		StorageKey:       objectKey,
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
	return s.storage.PresignUploadPart(ctx, uploadSession.StorageKey, uploadSession.ProviderUploadID, int32(partNumber), s.uploadExpires)
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
	if input.PartNumber <= 0 || input.EncryptedSize <= 0 || strings.TrimSpace(input.EncryptedHash) == "" {
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
	now := time.Now()
	_, err = s.uploadRepo.UpsertUploadPart(ctx, s.db, models.UploadPart{
		UploadSessionID: uploadSessionID,
		PartNumber:      input.PartNumber,
		ETag:            strings.TrimSpace(input.ETag),
		EncryptedSize:   input.EncryptedSize,
		EncryptedHash:   encryptedHash,
		UploadStatus:    UploadPartComplete,
		UploadedAt:      &now,
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
	fileChunks := make([]models.FileChunk, 0, len(parts))
	var encryptedSize int64
	for idx, part := range parts {
		if part.PartNumber != idx+1 {
			return ErrPartsRequired
		}
		completedParts = append(completedParts, storage.CompletedPart{
			PartNumber: int32(part.PartNumber),
			ETag:       part.ETag,
		})
		plaintextSize := file.ChunkSize
		if idx == len(parts)-1 {
			remainder := file.PlaintextSize - int64(idx)*file.ChunkSize
			if remainder > 0 {
				plaintextSize = remainder
			}
		}
		fileChunks = append(fileChunks, models.FileChunk{
			FileID:        file.ID,
			ChunkIndex:    idx,
			StorageKey:    uploadSession.StorageKey,
			PlaintextSize: plaintextSize,
			EncryptedSize: part.EncryptedSize,
			EncryptedHash: part.EncryptedHash,
			UploadStatus:  UploadPartComplete,
			UploadedAt:    part.UploadedAt,
		})
		encryptedSize += part.EncryptedSize
	}

	if err := s.storage.CompleteMultipartUpload(ctx, uploadSession.StorageKey, uploadSession.ProviderUploadID, completedParts); err != nil {
		_ = s.uploadRepo.UpdateUploadSessionStatus(ctx, s.db, uploadSessionID, UploadStatusFailed)
		_, _ = s.fileRepo.UpdateEncryptedFileStatusIf(ctx, s.db, file.ID, FileUploadFailed, []string{FileUploadUploading, FileUploadPending})
		return err
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if err := s.uploadRepo.UpdateUploadSessionStatus(ctx, tx, uploadSessionID, UploadStatusCompleted); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := s.fileRepo.UpdateEncryptedFileEnvelope(ctx, tx, file.ID, encryptedMetadata, encryptedFileKey, encryptedManifest, encryptedHash); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := s.uploadRepo.ReplaceFileChunks(ctx, tx, file.ID, fileChunks); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := s.fileRepo.MarkEncryptedFileComplete(ctx, tx, file.ID, encryptedSize, encryptedHash); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	committed, err := s.storageRepo.CommitStorage(ctx, tx, userID, file.PlaintextSize)
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

	_ = s.storage.AbortMultipartUpload(ctx, uploadSession.StorageKey, uploadSession.ProviderUploadID)

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
		released, err := s.storageRepo.ReleaseReservedStorage(ctx, tx, userID, file.PlaintextSize)
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
