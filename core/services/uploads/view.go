package uploads

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"arkive/core/models"
	"arkive/pkg/storage"
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

	if file.UploadStatus != FileUploadComplete {
		if file.UploadStatus == FileUploadFailed || file.UploadStatus == FileUploadAborted {
			return models.File{}, ErrUploadCancelled
		}
		return models.File{}, ErrNotFound
	}
	if isExpired(file.ExpiresAt) {
		return models.File{}, ErrNotFound
	}

	return file, nil
}

func (s *Service) PresignView(ctx context.Context, userID, fileID string) (string, error) {
	file, err := s.GetFileForDisplay(ctx, userID, fileID)
	if err != nil {
		return "", err
	}
	if file.UploadStatus != FileUploadComplete {
		return "", ErrInvalidInput
	}
	objectKey, err := storage.BuildObjectKey(userID, file.ID)
	if err != nil {
		return "", err
	}
	return s.storage.PresignDownload(ctx, objectKey, file.ID, "inline", s.downloadExpire)
}

func (s *Service) GetFileForShare(ctx context.Context, fileID string) (models.File, error) {
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return models.File{}, ErrInvalidInput
	}

	file, err := s.fileRepo.GetFileByID(ctx, s.db, fileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.File{}, ErrNotFound
		}
		return models.File{}, err
	}

	if file.UploadStatus != FileUploadComplete {
		if file.UploadStatus == FileUploadFailed || file.UploadStatus == FileUploadAborted {
			return models.File{}, ErrUploadCancelled
		}
		return models.File{}, ErrNotFound
	}
	if isExpired(file.ExpiresAt) {
		return models.File{}, ErrNotFound
	}

	return file, nil
}

func (s *Service) PresignShare(ctx context.Context, fileID string) (string, error) {
	file, err := s.GetFileForShare(ctx, fileID)
	if err != nil {
		return "", err
	}

	disposition := "attachment"
	if isViewableContentType(file.UploadStatus) {
		disposition = "inline"
	}

	objectKey, err := storage.BuildObjectKey(file.UserID, file.ID)
	if err != nil {
		return "", err
	}
	return s.storage.PresignDownload(ctx, objectKey, file.ID, disposition, s.downloadExpire)
}

func (s *Service) PresignShareView(ctx context.Context, fileID string) (string, error) {
	file, err := s.GetFileForShare(ctx, fileID)
	if err != nil {
		return "", err
	}
	return s.PresignShareViewForFile(ctx, file)
}

func (s *Service) PresignShareDownload(ctx context.Context, fileID string) (string, error) {
	file, err := s.GetFileForShare(ctx, fileID)
	if err != nil {
		return "", err
	}
	return s.PresignShareDownloadForFile(ctx, file)
}

func (s *Service) PresignShareViewForFile(ctx context.Context, file models.File) (string, error) {
	if file.UploadStatus != FileUploadComplete {
		return "", ErrInvalidInput
	}
	objectKey, err := storage.BuildObjectKey(file.UserID, file.ID)
	if err != nil {
		return "", err
	}
	return s.storage.PresignDownload(ctx, objectKey, file.ID, "inline", s.downloadExpire)
}

func (s *Service) PresignShareDownloadForFile(ctx context.Context, file models.File) (string, error) {
	expiry := s.shareDownloadExpire
	if expiry <= 0 {
		expiry = s.downloadExpire
	}
	objectKey, err := storage.BuildObjectKey(file.UserID, file.ID)
	if err != nil {
		return "", err
	}
	return s.storage.PresignDownload(ctx, objectKey, file.ID, "attachment", expiry)
}

func isViewableContentType(contentType string) bool {
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	switch contentType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return true
	case "video/mp4", "video/webm", "video/ogg":
		return true
	default:
		return false
	}
}
