package files

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"arkive/core/models"
	"arkive/pkg/storage"
)

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
	defer func() { _ = tx.Rollback(ctx) }()

	files, err := s.DeleteFilesWithinTx(ctx, tx, userID, []string{fileID})
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	s.CleanupDeletedFiles(ctx, userID, files)
	return nil
}

func (s *Service) DeleteFiles(ctx context.Context, userID string, fileIDs []string) (int, error) {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return 0, err
	}
	uniqueIDs, err := normalizeDeleteFileIDs(fileIDs)
	if err != nil {
		return 0, err
	}
	if len(uniqueIDs) == 0 {
		return 0, ErrInvalidInput
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	files, err := s.DeleteFilesWithinTx(ctx, tx, userID, uniqueIDs)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	s.CleanupDeletedFiles(ctx, userID, files)
	return len(files), nil
}

func (s *Service) DeleteFilesWithinTx(ctx context.Context, tx pgx.Tx, userID string, fileIDs []string) ([]models.File, error) {
	var err error
	userID, err = validateUserID(userID)
	if err != nil {
		return nil, err
	}
	uniqueIDs, err := normalizeDeleteFileIDs(fileIDs)
	if err != nil {
		return nil, err
	}
	if len(uniqueIDs) == 0 {
		return nil, ErrInvalidInput
	}
	return s.deleteFilesWithinTx(ctx, tx, userID, uniqueIDs)
}

func (s *Service) deleteFilesWithinTx(ctx context.Context, tx pgx.Tx, userID string, fileIDs []string) ([]models.File, error) {
	files := make([]models.File, 0, len(fileIDs))
	for _, fileID := range fileIDs {
		file, getErr := s.fileRepo.GetFileForUser(ctx, tx, fileID, userID)
		if getErr != nil {
			if errors.Is(getErr, pgx.ErrNoRows) {
				return nil, ErrNotFound
			}
			return nil, getErr
		}

		switch file.UploadStatus {
		case "complete":
			if err := s.storageRepo.DecreaseUsedStorage(ctx, tx, userID, totalStoredSize(file)); err != nil {
				return nil, err
			}
		case "pending", "uploading":
			return nil, ErrUploadCancelled
		default:
			return nil, ErrNotFound
		}

		deleted, deleteErr := s.fileRepo.DeleteFileForUser(ctx, tx, fileID, userID)
		if deleteErr != nil {
			return nil, deleteErr
		}
		if !deleted {
			return nil, ErrNotFound
		}
		files = append(files, file)
	}
	return files, nil
}

func (s *Service) CleanupDeletedFiles(ctx context.Context, userID string, files []models.File) {
	for _, file := range files {
		if objectKey, keyErr := storage.BuildObjectKey(userID, file.ID); keyErr == nil {
			_ = s.storage.DeleteObject(ctx, objectKey)
		}
		if thumbnailKey, keyErr := storage.BuildThumbnailObjectKey(userID, file.ID); keyErr == nil {
			_ = s.storage.DeleteObject(ctx, thumbnailKey)
		}
	}
}

func normalizeDeleteFileIDs(fileIDs []string) ([]string, error) {
	uniqueIDs := make([]string, 0, len(fileIDs))
	seen := make(map[string]struct{}, len(fileIDs))
	for _, rawID := range fileIDs {
		fileID, validateErr := validateUploadID(rawID)
		if validateErr != nil {
			return nil, validateErr
		}
		if _, exists := seen[fileID]; exists {
			continue
		}
		seen[fileID] = struct{}{}
		uniqueIDs = append(uniqueIDs, fileID)
	}
	return uniqueIDs, nil
}
