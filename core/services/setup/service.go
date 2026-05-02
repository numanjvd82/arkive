package setup

import (
	"context"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"arkive/core/database"
	"arkive/core/models"
	authrepo "arkive/core/repositories/auth"
	settingsrepo "arkive/core/repositories/settings"
	usersrepo "arkive/core/repositories/users"
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
	BrandName       string
	Email           string
	Password        string
	ConfirmPassword string
	Storage         models.StorageSettings
	LocalStorageGB  string
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
	input.Storage = normalizeStorageSettings(input.Storage)

	validationErrors := validateAdminInput(input.BrandName, input.Email, input.Password, input.ConfirmPassword)
	ValidateStorageSettings(input.Storage, validationErrors)
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

	if input.Storage.Provider == "local" {
		if err := os.MkdirAll(input.Storage.LocalPath, 0o700); err != nil {
			validationErrors.Add("local_path", "local storage path is not writable")
			return models.User{}, validationErrors, nil
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, nil, err
	}

	user, err := s.authRepo.CreateVerifiedUser(ctx, tx, input.BrandName, input.Email, string(hash))
	if err != nil {
		return models.User{}, nil, err
	}
	quotaBytes := input.Storage.MaxStorageBytes
	if quotaBytes == 0 {
		quotaBytes = math.MaxInt64
	}
	if err := s.userRepo.UpdateQuota(ctx, tx, user.ID, quotaBytes); err != nil {
		return models.User{}, nil, err
	}
	if err := s.settingsRepo.SaveStorageSettings(ctx, tx, input.Storage); err != nil {
		return models.User{}, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.User{}, nil, err
	}
	return user, nil, nil
}

func BuildStorageSettings(provider, localPath, storageGB, accessKeyID, secretAccessKey, sessionToken, bucket, endpoint, region, usePathStyle string) (models.StorageSettings, validation.Errors) {
	validationErrors := validation.New()
	storageGB = strings.TrimSpace(storageGB)
	maxBytes := int64(0)
	if storageGB != "" {
		gb, err := strconv.ParseInt(storageGB, 10, 64)
		if err != nil || gb < 0 {
			validationErrors.Add("storage_gb", "storage limit must be 0 or a positive number")
		} else if gb > 0 {
			maxBytes = gb * 1024 * 1024 * 1024
		}
	}
	return models.StorageSettings{
		Provider:          strings.ToLower(strings.TrimSpace(provider)),
		LocalPath:         strings.TrimSpace(localPath),
		MaxStorageBytes:   maxBytes,
		S3AccessKeyID:     strings.TrimSpace(accessKeyID),
		S3SecretAccessKey: strings.TrimSpace(secretAccessKey),
		S3SessionToken:    strings.TrimSpace(sessionToken),
		S3Bucket:          strings.TrimSpace(bucket),
		S3Endpoint:        strings.TrimSpace(endpoint),
		S3Region:          strings.TrimSpace(region),
		S3UsePathStyle:    strings.TrimSpace(usePathStyle) == "on" || strings.TrimSpace(usePathStyle) == "true",
	}, validationErrors
}

func normalizeStorageSettings(settings models.StorageSettings) models.StorageSettings {
	settings.Provider = strings.ToLower(strings.TrimSpace(settings.Provider))
	if settings.Provider == "" {
		settings.Provider = "local"
	}
	settings.LocalPath = strings.TrimSpace(settings.LocalPath)
	if settings.Provider == "local" && settings.LocalPath != "" {
		if abs, err := filepath.Abs(settings.LocalPath); err == nil {
			settings.LocalPath = abs
		}
	}
	settings.S3Region = strings.TrimSpace(settings.S3Region)
	if settings.Provider == "s3" && settings.S3Region == "" {
		settings.S3Region = "auto"
	}
	return settings
}

func ValidateStorageSettings(settings models.StorageSettings, validationErrors validation.Errors) {
	switch settings.Provider {
	case "local":
		if settings.LocalPath == "" {
			validationErrors.Add("local_path", "local storage path is required")
		}
	case "s3":
		if settings.S3AccessKeyID == "" {
			validationErrors.Add("s3_access_key_id", "access key is required")
		}
		if settings.S3SecretAccessKey == "" {
			validationErrors.Add("s3_secret_access_key", "secret key is required")
		}
		if settings.S3Bucket == "" {
			validationErrors.Add("s3_bucket", "bucket is required")
		}
		if settings.S3Endpoint == "" {
			validationErrors.Add("s3_endpoint", "endpoint is required")
		}
	default:
		validationErrors.Add("storage_provider", "choose local or S3-compatible storage")
	}
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
