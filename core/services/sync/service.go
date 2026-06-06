package syncsvc

import (
	"context"
	"encoding/base64"
	"sort"
	"strings"

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
}

type ListEntriesResult struct {
	Entries []models.SyncEntry
}

func NewService(db database.PgPool, repo *syncrepo.Repository) *Service {
	return &Service{
		db:   db,
		repo: repo,
	}
}

func (s *Service) ListEntries(ctx context.Context, input ListEntriesInput) (ListEntriesResult, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return ListEntriesResult{}, ErrInvalidInput
	}

	if input.FolderID != nil {
		folderID, err := validateUUIDValue(*input.FolderID)
		if err != nil {
			return ListEntriesResult{}, err
		}
		exists, err := s.repo.FolderExistsForUser(ctx, s.db, userID, folderID)
		if err != nil {
			return ListEntriesResult{}, err
		}
		if !exists {
			return ListEntriesResult{}, ErrNotFound
		}
		input.FolderID = &folderID
	}

	files, err := s.repo.ListFilesByFolder(ctx, s.db, userID, input.FolderID, input.IncludeDeleted)
	if err != nil {
		return ListEntriesResult{}, err
	}
	folders, err := s.repo.ListFoldersByParent(ctx, s.db, userID, input.FolderID, input.IncludeDeleted)
	if err != nil {
		return ListEntriesResult{}, err
	}

	entries := make([]models.SyncEntry, 0, len(files)+len(folders))
	for _, file := range files {
		entries = append(entries, models.SyncEntry{
			Type:              "file",
			ID:                file.ID,
			FolderID:          file.FolderID,
			EncryptedMetadata: base64.StdEncoding.EncodeToString(file.EncryptedMetadata),
			EncryptedFileKey:  base64.StdEncoding.EncodeToString(file.EncryptedFileKey),
			EncryptedManifest: base64.StdEncoding.EncodeToString(file.EncryptedManifest),
			UpdatedAt:         file.UpdatedAt,
			DeletedAt:         file.DeletedAt,
			PurgedAt:          file.PurgedAt,
		})
	}
	for _, folder := range folders {
		entries = append(entries, models.SyncEntry{
			Type:              "folder",
			ID:                folder.ID,
			ParentFolderID:    folder.ParentFolderID,
			EncryptedName:     base64.StdEncoding.EncodeToString(folder.EncryptedName),
			EncryptedMetadata: base64.StdEncoding.EncodeToString(folder.EncryptedMetadata),
			UpdatedAt:         folder.UpdatedAt,
			DeletedAt:         folder.DeletedAt,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].UpdatedAt.Equal(entries[j].UpdatedAt) {
			if entries[i].ID == entries[j].ID {
				return entries[i].Type < entries[j].Type
			}
			return entries[i].ID < entries[j].ID
		}
		return entries[i].UpdatedAt.After(entries[j].UpdatedAt)
	})

	return ListEntriesResult{Entries: entries}, nil
}

func validateUUIDValue(value string) (string, error) {
	normalized, ok := validation.NormalizeUUID(value)
	if !ok {
		return "", ErrInvalidInput
	}
	return normalized, nil
}
