package uploads

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	FreeRestoreDailyBytes int64 = 2 * 1024 * 1024 * 1024
	RestoreWindow               = 24 * time.Hour
)

func (s *Service) TouchUserActivity(ctx context.Context, userID string, isPremium bool) error {
	return s.touchUserActivity(ctx, userID, isPremium)
}

func (s *Service) touchUserActivity(ctx context.Context, userID string, isPremium bool) error {
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
	if isPremium {
		if _, err := s.fileRepo.ClearExpiryForUserCompletedFiles(ctx, tx, userID); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		return tx.Commit(ctx)
	}
	if s.restoreRepo == nil {
		return tx.Commit(ctx)
	}

	usedBytes, err := s.restoreRepo.SumUsageSince(ctx, tx, userID, activeAt.Add(-RestoreWindow))
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	remaining := FreeRestoreDailyBytes - usedBytes
	if remaining <= 0 {
		return tx.Commit(ctx)
	}

	archivedFiles, err := s.fileRepo.ListArchivedFilesForUser(ctx, tx, userID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	for _, file := range archivedFiles {
		if file.SizeBytes <= 0 || file.SizeBytes > remaining {
			continue
		}
		if err := s.fileRepo.ClearFileExpiry(ctx, tx, file.ID); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if err := s.restoreRepo.AddUsage(ctx, tx, userID, file.SizeBytes); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		remaining -= file.SizeBytes
		if remaining <= 0 {
			break
		}
	}

	return tx.Commit(ctx)
}
