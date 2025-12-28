package uploads

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"arkive/core/models"
)

func (s *Service) GetFileForDisplay(ctx context.Context, userID, fileID string) (models.File, error) {
	userID = strings.TrimSpace(userID)
	fileID = strings.TrimSpace(fileID)
	if userID == "" {
		return models.File{}, ErrUnauthorized
	}
	if fileID == "" {
		return models.File{}, ErrInvalidInput
	}

	file, err := s.fileRepo.GetFileForUser(ctx, s.db, fileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.File{}, ErrNotFound
		}
		return models.File{}, err
	}

	if file.Status != FileStatusComplete {
		if file.Status == FileStatusFailed || file.Status == FileStatusAborted {
			return models.File{}, ErrUploadCancelled
		}
		return models.File{}, ErrNotFound
	}

	return file, nil
}

func (s *Service) PresignView(ctx context.Context, userID, fileID string) (string, error) {
	file, err := s.GetFileForDisplay(ctx, userID, fileID)
	if err != nil {
		return "", err
	}
	if !isViewableContentType(file.ContentType) {
		return "", ErrInvalidInput
	}
	return s.r2.PresignDownload(ctx, file.ObjectKey, file.Filename, "inline", s.downloadExpire)
}

func isViewableContentType(contentType string) bool {
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	return strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "video/")
}
