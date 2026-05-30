package files

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"arkive/core/models"
	"arkive/pkg/storage"
)

type FileList struct {
	Files      []models.File
	TotalFiles int
}

type EncryptedFileRecord struct {
	FileID            string
	VaultID           string
	EncryptionVersion int16
	ChunkSize         int64
	TotalChunks       int
	PlaintextSize     int64
	EncryptedHash     string
	EncryptedMetadata string
	EncryptedFileKey  string
	EncryptedManifest string
	SourceURL         string
}

func (s *Service) CountArchivedFiles(ctx context.Context, userID string) (int64, error) {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return 0, err
	}
	return s.fileRepo.CountArchivedFilesForUser(ctx, s.db, userID)
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

func (s *Service) SearchCompletedUploadsByTokens(ctx context.Context, userID, vaultID string, tokenHashes [][]byte, limit int) ([]models.File, error) {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return nil, err
	}
	vaultID, err = validateUserID(vaultID)
	if err != nil {
		return nil, err
	}
	if len(tokenHashes) == 0 {
		return []models.File{}, nil
	}
	if len(tokenHashes) > MaxSearchQueryTokens {
		return nil, ErrInvalidInput
	}
	return s.fileRepo.SearchCompletedForTokens(ctx, s.db, userID, vaultID, tokenHashes, limit)
}

func (s *Service) GetFileForDisplay(ctx context.Context, userID, fileID string) (models.File, error) {
	userID = strings.TrimSpace(userID)
	fileID = strings.TrimSpace(fileID)
	if userID == "" {
		return models.File{}, ErrUnauthorized
	}
	if fileID == "" {
		return models.File{}, ErrInvalidInput
	}

	file, err := s.fileRepo.GetFileForUser(ctx, s.db, fileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.File{}, ErrNotFound
		}
		return models.File{}, err
	}

	if file.UploadStatus != "complete" {
		if file.UploadStatus == "failed" || file.UploadStatus == "aborted" {
			return models.File{}, ErrUploadCancelled
		}
		return models.File{}, ErrNotFound
	}
	if isExpired(file.ExpiresAt) {
		return models.File{}, ErrNotFound
	}

	return file, nil
}

func (s *Service) PresignDownload(ctx context.Context, userID, fileID string) (string, error) {
	file, err := s.GetFileForDisplay(ctx, userID, fileID)
	if err != nil {
		return "", err
	}
	objectKey, err := storage.BuildObjectKey(userID, file.ID)
	if err != nil {
		return "", err
	}
	return s.storage.PresignDownload(ctx, objectKey, file.ID, "attachment", s.downloadExpire)
}

func (s *Service) PresignView(ctx context.Context, userID, fileID string) (string, error) {
	file, err := s.GetFileForDisplay(ctx, userID, fileID)
	if err != nil {
		return "", err
	}
	objectKey, err := storage.BuildObjectKey(userID, file.ID)
	if err != nil {
		return "", err
	}
	return s.storage.PresignDownload(ctx, objectKey, file.ID, "inline", s.downloadExpire)
}

func (s *Service) PresignThumbnailDownload(ctx context.Context, userID, fileID string) (string, error) {
	file, err := s.GetFileForDisplay(ctx, userID, fileID)
	if err != nil {
		return "", err
	}
	if file.ThumbnailStatus != "complete" || file.ThumbnailSizeBytes <= 0 {
		return "", ErrNotFound
	}
	objectKey, err := storage.BuildThumbnailObjectKey(userID, file.ID)
	if err != nil {
		return "", err
	}
	size, sizeErr := s.storage.ObjectSize(ctx, objectKey)
	if sizeErr != nil || size <= 0 {
		return "", ErrNotFound
	}
	return s.storage.PresignDownload(ctx, objectKey, "thumbnail.enc", "inline", s.downloadExpire)
}

func (s *Service) GetFileForShare(ctx context.Context, fileID string) (models.File, error) {
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return models.File{}, ErrInvalidInput
	}

	file, err := s.fileRepo.GetFileByID(ctx, s.db, fileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.File{}, ErrNotFound
		}
		return models.File{}, err
	}

	if file.UploadStatus != "complete" {
		if file.UploadStatus == "failed" || file.UploadStatus == "aborted" {
			return models.File{}, ErrUploadCancelled
		}
		return models.File{}, ErrNotFound
	}
	if isExpired(file.ExpiresAt) {
		return models.File{}, ErrNotFound
	}

	return file, nil
}

func (s *Service) PresignShareDownloadForFile(ctx context.Context, file models.File) (string, error) {
	expiry := s.shareDownloadExpire
	if expiry <= 0 {
		expiry = s.downloadExpire
	}
	objectKey, err := storage.BuildObjectKey(file.UserID, file.ID)
	if err != nil {
		return "", err
	}
	return s.storage.PresignDownload(ctx, objectKey, file.ID, "attachment", expiry)
}

func (s *Service) PresignShareSourceForFile(ctx context.Context, file models.File) (string, error) {
	expiry := s.shareDownloadExpire
	if expiry <= 0 {
		expiry = s.downloadExpire
	}
	objectKey, err := storage.BuildObjectKey(file.UserID, file.ID)
	if err != nil {
		return "", err
	}
	return s.storage.PresignDownload(ctx, objectKey, "ciphertext.bin", "inline", expiry)
}

func (s *Service) GetEncryptedFileRecord(ctx context.Context, userID, fileID string) (EncryptedFileRecord, error) {
	userID = strings.TrimSpace(userID)
	fileID = strings.TrimSpace(fileID)
	if userID == "" {
		return EncryptedFileRecord{}, ErrUnauthorized
	}
	if fileID == "" {
		return EncryptedFileRecord{}, ErrInvalidInput
	}

	file, err := s.fileRepo.GetEncryptedFileRecordForUser(ctx, s.db, fileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EncryptedFileRecord{}, ErrNotFound
		}
		return EncryptedFileRecord{}, err
	}
	if file.UploadStatus != "complete" {
		if file.UploadStatus == "failed" || file.UploadStatus == "aborted" {
			return EncryptedFileRecord{}, ErrUploadCancelled
		}
		return EncryptedFileRecord{}, ErrNotFound
	}

	objectKey, err := storage.BuildObjectKey(file.UserID, file.ID)
	if err != nil {
		return EncryptedFileRecord{}, err
	}
	sourceURL, err := s.storage.PresignDownload(ctx, objectKey, "ciphertext.bin", "inline", s.downloadExpire)
	if err != nil {
		return EncryptedFileRecord{}, err
	}

	return EncryptedFileRecord{
		FileID:            file.ID,
		VaultID:           file.UserID,
		EncryptionVersion: file.EncryptionVersion,
		ChunkSize:         file.ChunkSize,
		TotalChunks:       file.ChunkCount,
		PlaintextSize:     file.PlaintextSize,
		EncryptedHash:     base64.StdEncoding.EncodeToString(file.EncryptedHash),
		EncryptedMetadata: base64.StdEncoding.EncodeToString(file.EncryptedMetadata),
		EncryptedFileKey:  base64.StdEncoding.EncodeToString(file.EncryptedFileKey),
		EncryptedManifest: base64.StdEncoding.EncodeToString(file.EncryptedManifest),
		SourceURL:         sourceURL,
	}, nil
}

func (s *Service) TouchUserActivity(ctx context.Context, userID string) error {
	if s.userRepo == nil {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	activeAt := time.Now()
	if err := s.userRepo.TouchUserActivity(ctx, tx, userID, activeAt); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
