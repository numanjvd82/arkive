package syncsvc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"arkive/core/database"
	"arkive/core/models"
	syncrepo "arkive/core/repositories/sync"
	"arkive/pkg/validation"
)

type Service struct {
	db   database.PgPool
	repo *syncrepo.Repository
}

type ListEntriesInput struct {
	UserID         string
	FolderID       *string
	IncludeDeleted bool
	Limit          int
	Cursor         *models.SyncEntriesCursor
}

type ListEntriesResult struct {
	Entries    []models.SyncEntry
	NextCursor *string
	HasMore    bool
}

func NewService(db database.PgPool, repo *syncrepo.Repository) *Service {
	return &Service{
		db:   db,
		repo: repo,
	}
}

func (s *Service) ListEntries(ctx context.Context, input models.ListEntriesPageInput) (models.SyncEntriesResponse, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return models.SyncEntriesResponse{}, ErrInvalidInput
	}
	if input.Limit <= 0 {
		input.Limit = 100
	}
	if input.Limit > 500 {
		input.Limit = 500
	}
	if input.Cursor != nil {
		if _, err := cursorTypeOrder(input.Cursor.Type); err != nil {
			return models.SyncEntriesResponse{}, ErrInvalidInput
		}
		if _, err := validateUUIDValue(input.Cursor.ID); err != nil {
			return models.SyncEntriesResponse{}, err
		}
	}

	if input.FolderID != nil {
		folderID, err := validateUUIDValue(*input.FolderID)
		if err != nil {
			return models.SyncEntriesResponse{}, err
		}
		exists, err := s.repo.FolderExistsForUser(ctx, s.db, userID, folderID)
		if err != nil {
			return models.SyncEntriesResponse{}, err
		}
		if !exists {
			return models.SyncEntriesResponse{}, ErrNotFound
		}
		input.FolderID = &folderID
	}

	rows, err := s.repo.ListEntriesPage(ctx, s.db, models.ListEntriesPageInput{
		UserID:         userID,
		FolderID:       input.FolderID,
		IncludeDeleted: input.IncludeDeleted,
		Cursor:         input.Cursor,
		Limit:          input.Limit + 1,
	})
	if err != nil {
		return models.SyncEntriesResponse{}, err
	}

	hasMore := len(rows) > input.Limit
	if hasMore {
		rows = rows[:input.Limit]
	}

	entries := make([]models.SyncEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, models.SyncEntry{
			Type:              row.Type,
			ID:                row.ID,
			FolderID:          row.FolderID,
			ParentFolderID:    row.ParentFolderID,
			EncryptedMetadata: encodeSyncBytes([]byte(row.EncryptedMetadata)),
			EncryptedFileKey:  encodeSyncBytes([]byte(row.EncryptedFileKey)),
			EncryptedManifest: encodeSyncBytes([]byte(row.EncryptedManifest)),
			EncryptedName:     encodeSyncBytes([]byte(row.EncryptedName)),
			UpdatedAt:         row.UpdatedAt,
			DeletedAt:         row.DeletedAt,
			PurgedAt:          row.PurgedAt,
		})
	}

	var nextCursor *string
	if hasMore && len(rows) > 0 {
		encoded, err := EncodeEntryCursor(models.SyncEntriesCursor{
			UpdatedAt: rows[len(rows)-1].UpdatedAt,
			Type:      rows[len(rows)-1].Type,
			ID:        rows[len(rows)-1].ID,
		})
		if err != nil {
			return models.SyncEntriesResponse{}, err
		}
		nextCursor = &encoded
	}

	return models.SyncEntriesResponse{
		Entries:    entries,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func validateUUIDValue(value string) (string, error) {
	normalized, ok := validation.NormalizeUUID(value)
	if !ok {
		return "", ErrInvalidInput
	}
	return normalized, nil
}

func EncodeEntryCursor(cursor models.SyncEntriesCursor) (string, error) {
	payload, err := json.Marshal(struct {
		UpdatedAt string `json:"updated_at"`
		Type      string `json:"type"`
		ID        string `json:"id"`
	}{
		UpdatedAt: cursor.UpdatedAt.UTC().Format(time.RFC3339Nano),
		Type:      cursor.Type,
		ID:        cursor.ID,
	})
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func encodeSyncBytes(value []byte) string {
	if len(value) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(value)
}

func cursorTypeOrder(entryType string) (int, error) {
	switch entryType {
	case "folder":
		return 0, nil
	case "file":
		return 1, nil
	default:
		return 0, ErrInvalidInput
	}
}
