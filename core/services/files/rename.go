package files

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type RenameFileInput struct {
	UserID            string
	FileID            string
	EncryptedMetadata []byte
}

func (s *Service) RenameFile(ctx context.Context, input RenameFileInput) error {
	userID, err := validateUserID(input.UserID)
	if err != nil {
		return err
	}
	fileID, err := validateUploadID(input.FileID)
	if err != nil {
		return err
	}
	if len(input.EncryptedMetadata) == 0 {
		return ErrInvalidInput
	}

	file, err := s.fileRepo.GetFileForUser(ctx, s.db, fileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if file.UploadStatus != "complete" {
		if file.UploadStatus == "failed" || file.UploadStatus == "aborted" {
			return ErrUploadCancelled
		}
		return ErrNotFound
	}
	if isExpired(file.ExpiresAt) {
		return ErrNotFound
	}

	renamed, err := s.fileRepo.RenameFileForUser(ctx, s.db, fileID, userID, input.EncryptedMetadata)
	if err != nil {
		return err
	}
	if !renamed {
		return ErrNotFound
	}
	return nil
}
