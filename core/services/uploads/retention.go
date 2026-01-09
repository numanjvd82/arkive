package uploads

import (
	"context"
	"time"
)

const (
	InactivityWindow       = 30 * 24 * time.Hour
	InactivityDeleteGrace  = 7 * 24 * time.Hour
)

func (s *Service) ApplyInactivityRetention(ctx context.Context) (int64, int, error) {
	if s.userRepo == nil {
		return 0, 0, nil
	}
	now := time.Now()
	cutoff := now.Add(-InactivityWindow)
	inactiveUsers, err := s.userRepo.ListInactiveUsers(ctx, s.db, cutoff)
	if err != nil {
		return 0, 0, err
	}

	marked := int64(0)
	expiresAt := now.Add(InactivityDeleteGrace)
	for _, user := range inactiveUsers {
		updated, markErr := s.fileRepo.MarkInactiveFilesForUser(ctx, s.db, user.ID, expiresAt)
		if markErr != nil {
			return marked, 0, markErr
		}
		marked += updated
	}

	expiredFiles, err := s.fileRepo.ListExpiredCompleteFiles(ctx, s.db, now)
	if err != nil {
		return marked, 0, err
	}

	deleted := 0
	for _, file := range expiredFiles {
		if err := s.DeleteFile(ctx, file.UserID, file.ID); err != nil {
			return marked, deleted, err
		}
		deleted++
	}

	return marked, deleted, nil
}
