package uploads

import (
	"context"
	"errors"
	"time"
)

func (s *Service) CleanupExpiredUploads(ctx context.Context) (int, error) {
	files, err := s.fileRepo.ListExpiredUploads(ctx, s.db, time.Now())
	if err != nil {
		return 0, err
	}

	cleaned := 0
	var lastErr error
	for _, file := range files {
		if err := s.AbortSingleUpload(ctx, file.UserID, file.ID); err != nil {
			if errors.Is(err, ErrNotFound) || errors.Is(err, ErrUploadCancelled) {
				continue
			}
			lastErr = err
			continue
		}
		cleaned++
	}

	return cleaned, lastErr
}
