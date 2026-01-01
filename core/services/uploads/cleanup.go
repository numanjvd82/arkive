package uploads

import (
	"context"
	"errors"
	"time"
)

func (s *Service) CleanupExpiredUploads(ctx context.Context) (int, error) {
	expiredMultiparts, err := s.uploadRepo.ListExpiredMultiparts(ctx, s.db, time.Now())
	if err != nil {
		return 0, err
	}

	cleaned := 0
	var lastErr error
	for _, upload := range expiredMultiparts {
		if err := s.AbortMultipart(ctx, upload.UserID, upload.MultipartID); err != nil {
			if errors.Is(err, ErrNotFound) || errors.Is(err, ErrUploadCancelled) {
				continue
			}
			lastErr = err
			continue
		}
		cleaned++
	}

	files, err := s.fileRepo.ListExpiredUploads(ctx, s.db, time.Now())
	if err != nil {
		return 0, err
	}

	for _, file := range files {
		if err := s.AbortUploadByFile(ctx, file.UserID, file.ID); err != nil {
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
