package uploads

import (
	"context"
	"path"
	"strings"
	"time"

	"arkive/core/database"
	"arkive/core/models"
)

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

func (s *Service) cleanupFolderPath(ctx context.Context, db database.PgExecutor, userID, folderPath string) error {
	folderPath = normalizeFolderPath(folderPath)
	if folderPath == "" {
		return nil
	}

	parts := strings.Split(folderPath, "/")
	paths := make([]string, 0, len(parts))
	current := ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		if current == "" {
			current = part
		} else {
			current = current + "/" + part
		}
		paths = append(paths, current)
	}

	for i := len(paths) - 1; i >= 0; i-- {
		path := paths[i]
		count, err := s.fileRepo.CountFilesInFolderTree(ctx, db, userID, path)
		if err != nil {
			return err
		}
		if count > 0 {
			break
		}
		if err := s.folderRepo.DeleteByPath(ctx, db, userID, path); err != nil {
			return err
		}
	}

	return nil
}

func expiresAtPtr(t time.Time) *time.Time {
	return &t
}
