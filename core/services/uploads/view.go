package uploads

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"arkive/core/models"
)

func (s *Service) GetFileForView(ctx context.Context, userID, fileID string) (models.File, string, error) {
	userID = strings.TrimSpace(userID)
	fileID = strings.TrimSpace(fileID)
	if userID == "" {
		return models.File{}, "", ErrUnauthorized
	}
	if fileID == "" {
		return models.File{}, "", ErrInvalidInput
	}

	file, err := s.fileRepo.GetFileForUser(ctx, s.db, fileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.File{}, "", ErrNotFound
		}
		return models.File{}, "", err
	}

	if file.Status != FileStatusComplete {
		if file.Status == FileStatusFailed || file.Status == FileStatusAborted {
			return models.File{}, "", ErrUploadCancelled
		}
		return models.File{}, "", ErrNotFound
	}

	viewURL := ""
	if isViewableContentType(file.ContentType) {
		viewURL, err = s.r2.PresignDownload(ctx, file.ObjectKey, file.Filename, "inline", s.downloadExpire)
		if err != nil {
			return models.File{}, "", err
		}
	}

	return file, viewURL, nil
}

func isViewableContentType(contentType string) bool {
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	return strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "video/")
}
