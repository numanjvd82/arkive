package uploads

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"

	"arkive/pkg/storage"
	"github.com/jackc/pgx/v5"
)

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
	if file.UploadStatus != FileUploadComplete {
		if file.UploadStatus == FileUploadFailed || file.UploadStatus == FileUploadAborted {
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
