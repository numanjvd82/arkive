package setup

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"arkive/core/database"
	"arkive/core/models"
	authrepo "arkive/core/repositories/auth"
	settingsrepo "arkive/core/repositories/settings"
	usersrepo "arkive/core/repositories/users"
	settingssvc "arkive/core/services/settings"
	"arkive/pkg/tokens"
	"arkive/pkg/validation"
)

var ErrAlreadyInitialized = errors.New("instance already initialized")

type Service struct {
	db           database.PgPool
	authRepo     *authrepo.Repository
	userRepo     *usersrepo.Repository
	settingsRepo *settingsrepo.Repository
}

type InitialAdminInput struct {
	BrandName          string
	Email              string
	Password           string
	ConfirmPassword    string
	VaultSalt          []byte
	EncryptedMasterKey []byte
	Storage            models.StorageSettings
	LocalStorageGB     string
}

func NewService(db database.PgPool, authRepo *authrepo.Repository, userRepo *usersrepo.Repository, settingsRepo *settingsrepo.Repository) *Service {
	return &Service{
		db:           db,
		authRepo:     authRepo,
		userRepo:     userRepo,
		settingsRepo: settingsRepo,
	}
}

func (s *Service) IsInitialized(ctx context.Context) (bool, error) {
	count, err := s.userRepo.CountUsers(ctx, s.db)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Service) CreateInitialAdmin(ctx context.Context, input InitialAdminInput) (models.User, validation.Errors, error) {
	input.BrandName = strings.TrimSpace(input.BrandName)
	input.Email = strings.TrimSpace(input.Email)
	input.Password = strings.TrimSpace(input.Password)
	input.ConfirmPassword = strings.TrimSpace(input.ConfirmPassword)
	input.Storage = settingssvc.NormalizeStorageSettings(input.Storage)

	validationErrors := validateAdminInput(input.BrandName, input.Email, input.Password, input.ConfirmPassword)
	if len(input.VaultSalt) != 16 {
		validationErrors.Add(validation.GeneralKey, "vault salt is missing or invalid")
	}
	if len(input.EncryptedMasterKey) == 0 {
		validationErrors.Add(validation.GeneralKey, "encrypted master key is missing")
	}
	settingssvc.ValidateStorageSettings(input.Storage, validationErrors)
	if validationErrors.HasAny() {
		return models.User{}, validationErrors, nil
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.User{}, nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	count, err := s.userRepo.CountUsers(ctx, tx)
	if err != nil {
		return models.User{}, nil, err
	}
	if count > 0 {
		return models.User{}, nil, ErrAlreadyInitialized
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, nil, err
	}

	user, err := s.authRepo.CreateVerifiedUser(ctx, tx, input.BrandName, input.Email, string(hash), input.VaultSalt, input.EncryptedMasterKey)
	if err != nil {
		return models.User{}, nil, err
	}
	if err := settingssvc.SaveStorageSettingsTx(ctx, tx, s.settingsRepo, s.userRepo, user.ID, input.Storage); err != nil {
		return models.User{}, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.User{}, nil, err
	}
	return user, nil, nil
}

func (s *Service) IssueRecoverySetupToken(ctx context.Context, userID string, ttl time.Duration) (string, error) {
	token, _, err := tokens.Generate()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().Add(ttl)
	if err := s.userRepo.SetRecoverySetupToken(ctx, s.db, userID, token, expiresAt); err != nil {
		return "", err
	}
	return token, nil
}

func (s *Service) HasValidRecoverySetupToken(ctx context.Context, token string) (bool, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return false, nil
	}
	return s.userRepo.HasValidRecoverySetupToken(ctx, s.db, token, time.Now())
}

func (s *Service) ClearRecoverySetupToken(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}
	return s.userRepo.ClearRecoverySetupToken(ctx, s.db, token)
}

func validateAdminInput(brandName, email, password, confirmPassword string) validation.Errors {
	validationErrors := validation.New()
	if brandName == "" {
		validationErrors.Add("brand_name", "brand name is required")
	}
	if email == "" {
		validationErrors.Add("email", "email is required")
	}
	if password == "" {
		validationErrors.Add("password", "password is required")
	}
	if confirmPassword == "" {
		validationErrors.Add("confirm_password", "confirm password is required")
	}
	if password != "" {
		switch validation.PasswordIssueFor(password) {
		case validation.PasswordTooShort:
			validationErrors.Add("password", "password must be at least 8 characters")
		case validation.PasswordMissingLower:
			validationErrors.Add("password", "password must include a lowercase letter")
		case validation.PasswordMissingUpper:
			validationErrors.Add("password", "password must include an uppercase letter")
		case validation.PasswordMissingSymbol:
			validationErrors.Add("password", "password must include a symbol")
		}
	}
	if password != "" && confirmPassword != "" && password != confirmPassword {
		validationErrors.Add("confirm_password", "passwords do not match")
	}
	return validationErrors
}
