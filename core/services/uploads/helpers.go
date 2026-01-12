package uploads

import (
	"context"
	"errors"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"arkive/core/database"
	"arkive/core/models"
	"arkive/pkg/storage/r2"
)

func (s *Service) loadMultipartForUserOrFile(ctx context.Context, userID, uploadID string) (models.MultipartUpload, models.File, bool, error) {
	upload, err := s.uploadRepo.GetMultipartForUser(ctx, s.db, uploadID, userID)
	if err == nil {
		file, fileErr := s.fileRepo.GetFileForUser(ctx, s.db, upload.FileID, userID)
		if fileErr != nil {
			return models.MultipartUpload{}, models.File{}, false, fileErr
		}
		return upload, file, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return models.MultipartUpload{}, models.File{}, false, err
	}

	upload, err = s.uploadRepo.GetMultipartForFile(ctx, s.db, uploadID, userID)
	if err == nil {
		file, fileErr := s.fileRepo.GetFileForUser(ctx, s.db, upload.FileID, userID)
		if fileErr != nil {
			return models.MultipartUpload{}, models.File{}, false, fileErr
		}
		return upload, file, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return models.MultipartUpload{}, models.File{}, false, err
	}

	return models.MultipartUpload{}, models.File{}, false, nil
}

func (s *Service) ensureNotExpired(ctx context.Context, userID string, file models.File, multipart *models.MultipartUpload) error {
	if isExpired(file.ExpiresAt) {
		if multipart != nil {
			_ = s.AbortMultipart(ctx, userID, multipart.ID)
		} else {
			_ = s.AbortSingleUpload(ctx, userID, file.ID)
		}
		return ErrUploadCancelled
	}
	if multipart != nil && isExpired(multipart.ExpiresAt) {
		_ = s.AbortMultipart(ctx, userID, multipart.ID)
		return ErrUploadCancelled
	}
	return nil
}

func (s *Service) refreshExpiry(ctx context.Context, fileID, multipartID string) error {
	expiresAt := time.Now().Add(s.uploadExpires)
	if fileID != "" {
		if err := s.fileRepo.UpdateFileExpiry(ctx, s.db, fileID, expiresAt); err != nil {
			return err
		}
	}
	if multipartID != "" {
		if err := s.uploadRepo.UpdateMultipartExpiry(ctx, s.db, multipartID, expiresAt); err != nil {
			return err
		}
	}
	return nil
}

func normalizeDetectedContentType(contentType string) string {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	if contentType == "application/octet-stream" || contentType == "" {
		return ""
	}
	return contentType
}

func detectObjectContentType(ctx context.Context, client *r2.Client, key string) (string, error) {
	data, err := client.ReadObjectRange(ctx, key, 0, 511)
	if err != nil {
		return "", err
	}
	return normalizeDetectedContentType(http.DetectContentType(data)), nil
}

func (s *Service) resolveAndPersistContentType(ctx context.Context, file models.File, key string, tx pgx.Tx) (int64, error) {
	actualSize, actualContentType, err := s.r2.HeadObjectDetails(ctx, key)
	if err != nil {
		return 0, err
	}

	detectedContentType, err := detectObjectContentType(ctx, s.r2, key)
	if err != nil {
		return 0, err
	}

	resolvedContentType := normalizeDetectedContentType(actualContentType)
	if detectedContentType != "" {
		resolvedContentType = detectedContentType
	}
	if resolvedContentType != "" && resolvedContentType != file.ContentType {
		if err := s.fileRepo.UpdateFileContentType(ctx, tx, file.ID, resolvedContentType); err != nil {
			return 0, err
		}
	}

	return actualSize, nil
}

func validateUserID(userID string) (string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", ErrUnauthorized
	}
	return userID, nil
}

func validateUploadID(uploadID string) (string, error) {
	uploadID = strings.TrimSpace(uploadID)
	if uploadID == "" {
		return "", ErrInvalidInput
	}
	return uploadID, nil
}

func normalizeFolderPath(folderPath string) string {
	folderPath = strings.TrimSpace(strings.ReplaceAll(folderPath, "\\", "/"))
	if folderPath == "" {
		return ""
	}
	cleaned := path.Clean(folderPath)
	cleaned = strings.TrimPrefix(cleaned, "/")
	cleaned = strings.TrimSuffix(cleaned, "/")
	if cleaned == "." || cleaned == "" || strings.HasPrefix(cleaned, "..") {
		return ""
	}
	return cleaned
}

func NormalizeFolderPath(folderPath string) string {
	return normalizeFolderPath(folderPath)
}

func (s *Service) ensureFolderPath(ctx context.Context, db database.PgExecutor, userID, folderPath string) error {
	folderPath = normalizeFolderPath(folderPath)
	if folderPath == "" {
		return nil
	}

	parts := strings.Split(folderPath, "/")
	parentPath := ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		path := part
		if parentPath != "" {
			path = parentPath + "/" + part
		}
		err := s.folderRepo.CreateFolder(ctx, db, models.Folder{
			UserID:     userID,
			Path:       path,
			Name:       part,
			ParentPath: parentPath,
		})
		if err != nil {
			return err
		}
		parentPath = path
	}
	return nil
}

func validatePartNumber(partNumber int32, totalParts int) error {
	if partNumber <= 0 || partNumber > int32(totalParts) {
		return ErrInvalidInput
	}
	return nil
}
