package files

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"arkive/core/models"
)

type RenameFileInput struct {
	UserID            string
	FileID            string
	EncryptedMetadata []byte
	SearchTokens      []models.FileSearchToken
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
	searchTokens, err := NormalizeSearchTokens(input.SearchTokens, MaxSearchTokensPerFile)
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
	if file.UploadStatus != "complete" {
		if file.UploadStatus == "failed" || file.UploadStatus == "aborted" {
			return ErrUploadCancelled
		}
		return ErrNotFound
	}
	if isExpired(file.ExpiresAt) {
		return ErrNotFound
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	renamed, err := s.fileRepo.RenameFileForUser(ctx, tx, fileID, userID, input.EncryptedMetadata)
	if err != nil {
		return err
	}
	if !renamed {
		return ErrNotFound
	}
	if err := s.fileRepo.ReplaceSearchTokensForFile(ctx, tx, userID, file.UserID, fileID, searchTokens); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}
