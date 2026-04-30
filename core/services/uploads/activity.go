package uploads

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

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
