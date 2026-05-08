package shares

import (
	"context"
	"errors"
	"strings"
	"time"

	"arkive/core/database"
	"arkive/core/models"
	filerepo "arkive/core/repositories/files"
	sharerepo "arkive/core/repositories/shares"
	"arkive/pkg/tokens"
	"arkive/pkg/validation"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	ShareStatusActive  = "active"
	ShareStatusRevoked = "revoked"
)

type Service struct {
	db        database.PgPool
	fileRepo  *filerepo.Repository
	shareRepo *sharerepo.Repository
}

type CreateInput struct {
	FileID      string
	OwnerUserID string
	Token       string
	Password    string
	ExpiresAt   *time.Time
}

func NewService(db database.PgPool, fileRepo *filerepo.Repository, shareRepo *sharerepo.Repository) *Service {
	return &Service{
		db:        db,
		fileRepo:  fileRepo,
		shareRepo: shareRepo,
	}
}

func (s *Service) CreateShare(ctx context.Context, input CreateInput) (models.Share, validation.Errors, error) {
	input.FileID = strings.TrimSpace(input.FileID)
	input.OwnerUserID = strings.TrimSpace(input.OwnerUserID)
	input.Token = strings.TrimSpace(input.Token)
	input.Password = strings.TrimSpace(input.Password)

	validationErrors := validation.New()
	if input.OwnerUserID == "" {
		return models.Share{}, nil, ErrUnauthorized
	}
	if input.FileID == "" {
		validationErrors.Add("fileId", ErrFileIDRequired.Error())
	}
	if input.Token != "" && !isTokenValid(input.Token) {
		validationErrors.Add("token", ErrTokenInvalid.Error())
	}
	if input.ExpiresAt != nil && !input.ExpiresAt.After(time.Now()) {
		validationErrors.Add("expiresAt", ErrExpiryInvalid.Error())
	}
	if input.Password != "" {
		if message := sharePasswordValidationMessage(input.Password); message != "" {
			validationErrors.Add("password", message)
		}
	}
	if validationErrors.HasAny() {
		return models.Share{}, validationErrors, nil
	}

	file, err := s.fileRepo.GetFileForUser(ctx, s.db, input.FileID, input.OwnerUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Share{}, nil, ErrNotFound
		}
		return models.Share{}, nil, err
	}
	if file.UploadStatus != FileStatusComplete {
		return models.Share{}, nil, ErrInvalidInput
	}

	if _, err := s.shareRepo.GetShareForFile(ctx, s.db, input.FileID); err == nil {
		return models.Share{}, nil, ErrShareExists
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return models.Share{}, nil, err
	}

	token := input.Token
	if token == "" {
		generated, _, err := tokens.Generate()
		if err != nil {
			return models.Share{}, nil, err
		}
		token = generated
	}
	if !isTokenValid(token) {
		validationErrors.Add("token", ErrTokenInvalid.Error())
		return models.Share{}, validationErrors, nil
	}

	var passwordHash *string
	if input.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return models.Share{}, nil, ErrPasswordHashFailed
		}
		hashStr := string(hashed)
		passwordHash = &hashStr
	}

	share := models.Share{
		FileID:       input.FileID,
		OwnerUserID:  input.OwnerUserID,
		Token:        token,
		PasswordHash: passwordHash,
		ExpiresAt:    input.ExpiresAt,
		Status:       ShareStatusActive,
	}
	created, err := s.shareRepo.CreateShare(ctx, s.db, share)
	if err != nil {
		return models.Share{}, nil, err
	}
	return created, nil, nil
}

func (s *Service) GetShareByToken(ctx context.Context, token string) (models.Share, error) {
	return s.shareRepo.GetShareByToken(ctx, s.db, token)
}

func (s *Service) GetShareForFile(ctx context.Context, fileID string) (models.Share, error) {
	return s.shareRepo.GetShareForFile(ctx, s.db, fileID)
}

func (s *Service) GetShareForFileForUser(ctx context.Context, fileID, ownerUserID string) (models.Share, error) {
	fileID = strings.TrimSpace(fileID)
	ownerUserID = strings.TrimSpace(ownerUserID)
	if ownerUserID == "" {
		return models.Share{}, ErrUnauthorized
	}
	if fileID == "" {
		return models.Share{}, ErrInvalidInput
	}
	return s.shareRepo.GetShareForFileForUser(ctx, s.db, fileID, ownerUserID)
}

func (s *Service) UpdateShareForUser(ctx context.Context, shareID, ownerUserID string, expiresAt *time.Time, password string, requirePassword bool) (models.Share, validation.Errors, error) {
	shareID = strings.TrimSpace(shareID)
	ownerUserID = strings.TrimSpace(ownerUserID)
	password = strings.TrimSpace(password)

	validationErrors := validation.New()
	if ownerUserID == "" {
		return models.Share{}, nil, ErrUnauthorized
	}
	if shareID == "" {
		return models.Share{}, nil, ErrInvalidInput
	}
	if expiresAt != nil && !expiresAt.After(time.Now()) {
		validationErrors.Add("expiresAt", ErrExpiryInvalid.Error())
	}
	existing, err := s.shareRepo.GetShareForUser(ctx, s.db, shareID, ownerUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Share{}, nil, ErrNotFound
		}
		return models.Share{}, nil, err
	}

	var passwordHash *string
	if requirePassword {
		if password == "" && existing.PasswordHash == nil {
			validationErrors.Add("password", ErrPasswordRequired.Error())
		}
		if password != "" {
			if message := sharePasswordValidationMessage(password); message != "" {
				validationErrors.Add("password", message)
			}
		}
	}
	if validationErrors.HasAny() {
		return models.Share{}, validationErrors, nil
	}

	if requirePassword {
		if password == "" && existing.PasswordHash != nil {
			passwordHash = existing.PasswordHash
		} else {
			hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return models.Share{}, nil, ErrPasswordHashFailed
			}
			hashStr := string(hashed)
			passwordHash = &hashStr
		}
	}

	updated, err := s.shareRepo.UpdateShareForUser(ctx, s.db, shareID, ownerUserID, passwordHash, expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Share{}, nil, ErrNotFound
		}
		return models.Share{}, nil, err
	}
	return updated, nil, nil
}

func sharePasswordValidationMessage(password string) string {
	switch validation.PasswordIssueFor(password) {
	case validation.PasswordTooShort:
		return ErrPasswordTooShort.Error()
	case validation.PasswordMissingLower:
		return ErrPasswordNeedLower.Error()
	case validation.PasswordMissingUpper:
		return ErrPasswordNeedUpper.Error()
	case validation.PasswordMissingSymbol:
		return ErrPasswordNeedSymbol.Error()
	default:
		return ""
	}
}

func (s *Service) DeleteShareForUser(ctx context.Context, shareID, ownerUserID string) (bool, error) {
	shareID = strings.TrimSpace(shareID)
	ownerUserID = strings.TrimSpace(ownerUserID)
	if ownerUserID == "" {
		return false, ErrUnauthorized
	}
	if shareID == "" {
		return false, ErrInvalidInput
	}
	return s.shareRepo.DeleteShareForUser(ctx, s.db, shareID, ownerUserID)
}

func (s *Service) ListSharesForUser(ctx context.Context, ownerUserID string) ([]models.ShareWithFile, error) {
	ownerUserID = strings.TrimSpace(ownerUserID)
	if ownerUserID == "" {
		return nil, ErrUnauthorized
	}
	return s.shareRepo.ListSharesForUser(ctx, s.db, ownerUserID)
}

func (s *Service) SearchSharesForUser(ctx context.Context, ownerUserID, query string, limit int) ([]models.ShareWithFile, error) {
	ownerUserID = strings.TrimSpace(ownerUserID)
	query = strings.TrimSpace(query)
	if ownerUserID == "" {
		return nil, ErrUnauthorized
	}
	if query == "" {
		return []models.ShareWithFile{}, nil
	}
	return s.shareRepo.SearchSharesForUser(ctx, s.db, ownerUserID, query, limit)
}

func isTokenValid(token string) bool {
	if len(token) < TokenMinLength || len(token) > TokenMaxLength {
		return false
	}
	for _, r := range token {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return false
		}
	}
	return true
}
