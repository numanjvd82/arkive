package files

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"

	"arkive/core/models"
	"arkive/pkg/storage"
)

const deletedFilePurgeBatchSize = 100

func (s *Service) PurgeDeletedFiles(ctx context.Context) (int, error) {
	files, err := s.fileRepo.ListDeletedPendingPurge(ctx, s.db, deletedFilePurgeBatchSize)
	if err != nil {
		return 0, err
	}
	if len(files) == 0 {
		return 0, nil
	}

	purgedIDs := make([]string, 0, len(files))
	for _, file := range files {
		if err := s.purgeDeletedFileObjects(ctx, file); err != nil {
			log.Printf("files: purge skipped file=%s user=%s: %v", file.ID, file.UserID, err)
			continue
		}
		purgedIDs = append(purgedIDs, file.ID)
	}
	if len(purgedIDs) == 0 {
		return 0, nil
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	affected, err := s.fileRepo.MarkFilesPurged(ctx, tx, purgedIDs)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return int(affected), nil
}

func (s *Service) purgeDeletedFileObjects(ctx context.Context, file models.File) error {
	objectKey, err := storage.BuildObjectKey(file.UserID, file.ID)
	if err != nil {
		return err
	}
	if err := s.storage.DeleteObject(ctx, objectKey); err != nil {
		return err
	}
	thumbnailKey, err := storage.BuildThumbnailObjectKey(file.UserID, file.ID)
	if err != nil {
		return err
	}
	if err := s.storage.DeleteObject(ctx, thumbnailKey); err != nil {
		return err
	}
	return nil
}
