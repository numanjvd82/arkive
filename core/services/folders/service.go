package folders

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"arkive/core/database"
	"arkive/core/models"
	filerepo "arkive/core/repositories/files"
	foldersrepo "arkive/core/repositories/folders"
	filessvc "arkive/core/services/files"
	"arkive/pkg/validation"
)

const (
	defaultPageSize = 50
	maxPageSize     = 100
	maxMoveEntries  = 1000
	maxFolderDepth  = 100
)

type Service struct {
	db         database.PgPool
	folderRepo *foldersrepo.Repository
	fileRepo   *filerepo.Repository
	filesSvc   *filessvc.Service
}

func NewService(db database.PgPool, folderRepo *foldersrepo.Repository, fileRepo *filerepo.Repository, filesSvc *filessvc.Service) *Service {
	return &Service{
		db:         db,
		folderRepo: folderRepo,
		fileRepo:   fileRepo,
		filesSvc:   filesSvc,
	}
}

func (s *Service) CreateFolder(ctx context.Context, input CreateFolderInput) (models.Folder, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" || len(input.EncryptedName) == 0 {
		return models.Folder{}, ErrInvalidInput
	}
	vaultID := strings.TrimSpace(input.VaultID)
	if vaultID == "" {
		vaultID = userID
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.Folder{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if input.ParentFolderID != nil {
		parentID, err := validateUUIDValue(*input.ParentFolderID)
		if err != nil {
			return models.Folder{}, err
		}
		parent, err := s.folderRepo.GetForUser(ctx, tx, userID, parentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return models.Folder{}, ErrNotFound
			}
			return models.Folder{}, err
		}
		vaultID = parent.VaultID
		input.ParentFolderID = &parentID
	}

	folder, err := s.folderRepo.Create(ctx, tx, models.Folder{
		UserID:            userID,
		VaultID:           vaultID,
		ParentFolderID:    input.ParentFolderID,
		EncryptedName:     input.EncryptedName,
		EncryptedMetadata: input.EncryptedMetadata,
	})
	if err != nil {
		return models.Folder{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return models.Folder{}, err
	}
	return folder, nil
}

func (s *Service) ListEntries(ctx context.Context, input ListEntriesInput) (ListEntriesResult, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return ListEntriesResult{}, ErrInvalidInput
	}
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = defaultPageSize
	}
	if input.PageSize > maxPageSize {
		input.PageSize = maxPageSize
	}
	if input.FolderID != nil {
		folderID, err := validateUUIDValue(*input.FolderID)
		if err != nil {
			return ListEntriesResult{}, err
		}
		if _, err := s.folderRepo.GetForUser(ctx, s.db, userID, folderID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ListEntriesResult{}, ErrNotFound
			}
			return ListEntriesResult{}, err
		}
		input.FolderID = &folderID
	}

	folderCount, err := s.folderRepo.CountChildFolders(ctx, s.db, userID, input.FolderID)
	if err != nil {
		return ListEntriesResult{}, err
	}
	fileCount, err := s.fileRepo.CountByFolder(ctx, s.db, userID, input.FolderID)
	if err != nil {
		return ListEntriesResult{}, err
	}

	totalEntries := folderCount + fileCount
	totalPages := 0
	if totalEntries > 0 {
		totalPages = (totalEntries + input.PageSize - 1) / input.PageSize
	}
	offset := (input.Page - 1) * input.PageSize

	result := ListEntriesResult{
		FolderID:     input.FolderID,
		Page:         input.Page,
		PageSize:     input.PageSize,
		TotalEntries: totalEntries,
		TotalPages:   totalPages,
		Path:         []models.Folder{},
		Folders:      []models.Folder{},
		Files:        []models.File{},
	}
	if input.FolderID != nil {
		path, err := s.folderPath(ctx, s.db, userID, *input.FolderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ListEntriesResult{}, ErrNotFound
			}
			return ListEntriesResult{}, err
		}
		result.Path = path
	}
	if totalEntries == 0 || offset >= totalEntries {
		return result, nil
	}

	if offset < folderCount {
		folderLimit := input.PageSize
		if remainingFolders := folderCount - offset; remainingFolders < folderLimit {
			folderLimit = remainingFolders
		}
		folders, err := s.folderRepo.ListChildFolders(ctx, s.db, userID, input.FolderID, folderLimit, offset)
		if err != nil {
			return ListEntriesResult{}, err
		}
		result.Folders = folders

		remaining := input.PageSize - len(folders)
		if remaining > 0 {
			files, err := s.fileRepo.ListByFolder(ctx, s.db, userID, input.FolderID, remaining, 0)
			if err != nil {
				return ListEntriesResult{}, err
			}
			result.Files = files
		}
		return result, nil
	}

	fileOffset := offset - folderCount
	files, err := s.fileRepo.ListByFolder(ctx, s.db, userID, input.FolderID, input.PageSize, fileOffset)
	if err != nil {
		return ListEntriesResult{}, err
	}
	result.Files = files
	return result, nil
}

func (s *Service) ValidateFolderAccess(ctx context.Context, userID, folderID string) (models.Folder, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return models.Folder{}, ErrInvalidInput
	}
	var err error
	folderID, err = validateUUIDValue(folderID)
	if err != nil {
		return models.Folder{}, err
	}

	folder, err := s.folderRepo.GetForUser(ctx, s.db, userID, folderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Folder{}, ErrNotFound
		}
		return models.Folder{}, err
	}
	return folder, nil
}

func (s *Service) RenameFolder(ctx context.Context, input RenameFolderInput) error {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" || len(input.EncryptedName) == 0 || len(input.EncryptedMetadata) == 0 {
		return ErrInvalidInput
	}

	folderID, err := validateUUIDValue(input.FolderID)
	if err != nil {
		return err
	}

	renamed, err := s.folderRepo.RenameFolderForUser(ctx, s.db, userID, folderID, input.EncryptedName, input.EncryptedMetadata)
	if err != nil {
		return err
	}
	if !renamed {
		return ErrNotFound
	}
	return nil
}

func (s *Service) folderPath(ctx context.Context, db database.PgExecutor, userID, folderID string) ([]models.Folder, error) {
	path := []models.Folder{}
	currentID := folderID
	for depth := 0; strings.TrimSpace(currentID) != "" && depth < maxFolderDepth; depth++ {
		folder, err := s.folderRepo.GetForUser(ctx, db, userID, currentID)
		if err != nil {
			return nil, err
		}
		path = append(path, folder)
		if folder.ParentFolderID == nil {
			break
		}
		currentID = strings.TrimSpace(*folder.ParentFolderID)
	}
	if len(path) >= maxFolderDepth && currentID != "" {
		return nil, ErrInvalidInput
	}

	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path, nil
}

func (s *Service) MoveEntries(ctx context.Context, input MoveEntriesInput) error {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return ErrInvalidInput
	}
	fileIDs := uniqueNonEmpty(input.FileIDs)
	folderIDs := uniqueNonEmpty(input.FolderIDs)
	if len(fileIDs) == 0 && len(folderIDs) == 0 {
		return ErrInvalidInput
	}
	if len(fileIDs)+len(folderIDs) > maxMoveEntries {
		return ErrInvalidInput
	}
	var err error
	fileIDs, err = validateUUIDList(fileIDs)
	if err != nil {
		return err
	}
	folderIDs, err = validateUUIDList(folderIDs)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var targetFolderID *string
	if input.TargetFolderID != nil {
		targetID, err := validateUUIDValue(*input.TargetFolderID)
		if err != nil {
			return err
		}
		if _, err := s.folderRepo.GetForUser(ctx, tx, userID, targetID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}
		targetFolderID = &targetID
	}

	if len(folderIDs) > 0 && targetFolderID != nil {
		for _, folderID := range folderIDs {
			if folderID == *targetFolderID {
				return ErrInvalidMove
			}
		}
		isDescendant, err := s.folderRepo.TargetIsDescendant(ctx, tx, userID, folderIDs, *targetFolderID)
		if err != nil {
			return err
		}
		if isDescendant {
			return ErrInvalidMove
		}
	}

	if len(folderIDs) > 0 {
		moved, err := s.folderRepo.MoveFolders(ctx, tx, userID, folderIDs, targetFolderID)
		if err != nil {
			return err
		}
		if moved != int64(len(folderIDs)) {
			return ErrNotFound
		}
	}

	if len(fileIDs) > 0 {
		moved, err := s.fileRepo.MoveFilesToFolder(ctx, tx, userID, fileIDs, targetFolderID)
		if err != nil {
			return err
		}
		if moved != int64(len(fileIDs)) {
			return ErrNotFound
		}
	}

	return tx.Commit(ctx)
}

func (s *Service) ResolveDeleteScope(ctx context.Context, input ResolveDeleteScopeInput) (DeleteScope, error) {
	return s.resolveDeleteScope(ctx, s.db, input)
}

func (s *Service) resolveDeleteScope(ctx context.Context, db database.PgExecutor, input ResolveDeleteScopeInput) (DeleteScope, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return DeleteScope{}, ErrInvalidInput
	}

	fileIDs := uniqueNonEmpty(input.FileIDs)
	folderIDs := uniqueNonEmpty(input.FolderIDs)
	if len(fileIDs) == 0 && len(folderIDs) == 0 {
		return DeleteScope{}, ErrInvalidInput
	}

	var err error
	fileIDs, err = validateUUIDList(fileIDs)
	if err != nil {
		return DeleteScope{}, err
	}
	folderIDs, err = validateUUIDList(folderIDs)
	if err != nil {
		return DeleteScope{}, err
	}

	explicitFiles := make([]string, 0, len(fileIDs))
	for _, fileID := range fileIDs {
		file, getErr := s.fileRepo.GetFileForUser(ctx, db, fileID, userID)
		if getErr != nil {
			if errors.Is(getErr, pgx.ErrNoRows) {
				return DeleteScope{}, ErrNotFound
			}
			return DeleteScope{}, getErr
		}
		if file.UploadStatus != "complete" || file.ExpiresAt != nil {
			return DeleteScope{}, ErrNotFound
		}
		explicitFiles = append(explicitFiles, file.ID)
	}

	resolvedFolderIDs := []string{}
	folderFileIDs := []string{}
	if len(folderIDs) > 0 {
		resolvedFolderIDs, err = s.folderRepo.DescendantFolderIDs(ctx, db, userID, folderIDs)
		if err != nil {
			return DeleteScope{}, err
		}
		if len(resolvedFolderIDs) == 0 {
			return DeleteScope{}, ErrNotFound
		}

		resolvedSet := make(map[string]struct{}, len(resolvedFolderIDs))
		for _, folderID := range resolvedFolderIDs {
			resolvedSet[folderID] = struct{}{}
		}
		for _, folderID := range folderIDs {
			if _, ok := resolvedSet[folderID]; !ok {
				return DeleteScope{}, ErrNotFound
			}
		}

		folderFileIDs, err = s.folderRepo.FileIDsInFolders(ctx, db, userID, resolvedFolderIDs)
		if err != nil {
			return DeleteScope{}, err
		}
	}

	return DeleteScope{
		FolderIDs: uniqueNonEmpty(resolvedFolderIDs),
		FileIDs:   uniqueNonEmpty(append(explicitFiles, folderFileIDs...)),
	}, nil
}

func (s *Service) DeleteEntries(ctx context.Context, input DeleteEntriesInput) (DeleteEntriesResult, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return DeleteEntriesResult{}, ErrInvalidInput
	}
	if s.filesSvc == nil {
		return DeleteEntriesResult{}, ErrInvalidInput
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return DeleteEntriesResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	scope, err := s.resolveDeleteScope(ctx, tx, ResolveDeleteScopeInput{
		UserID:    userID,
		FileIDs:   input.FileIDs,
		FolderIDs: input.FolderIDs,
	})
	if err != nil {
		return DeleteEntriesResult{}, err
	}

	deletedFiles := make([]models.File, 0)
	if len(scope.FileIDs) > 0 {
		deletedFiles, err = s.filesSvc.DeleteFilesWithinTx(ctx, tx, userID, scope.FileIDs)
		if err != nil {
			return DeleteEntriesResult{}, err
		}
	}

	deletedFolders := 0
	if len(scope.FolderIDs) > 0 {
		affected, err := s.folderRepo.SoftDeleteFolders(ctx, tx, userID, scope.FolderIDs)
		if err != nil {
			return DeleteEntriesResult{}, err
		}
		if affected != int64(len(scope.FolderIDs)) {
			return DeleteEntriesResult{}, ErrNotFound
		}
		deletedFolders = int(affected)
	}

	if err := tx.Commit(ctx); err != nil {
		return DeleteEntriesResult{}, err
	}

	if len(deletedFiles) > 0 {
		s.filesSvc.CleanupDeletedFiles(ctx, userID, deletedFiles)
	}

	return DeleteEntriesResult{
		DeletedFiles:   len(deletedFiles),
		DeletedFolders: deletedFolders,
	}, nil
}

func uniqueNonEmpty(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func validateUUIDList(values []string) ([]string, error) {
	out := make([]string, 0, len(values))
	for _, value := range values {
		validated, err := validateUUIDValue(value)
		if err != nil {
			return nil, err
		}
		out = append(out, validated)
	}
	return out, nil
}

func validateUUIDValue(value string) (string, error) {
	normalized, ok := validation.NormalizeUUID(value)
	if !ok {
		return "", ErrInvalidInput
	}
	return normalized, nil
}
