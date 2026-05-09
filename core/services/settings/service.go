package settings

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"

	"arkive/core/database"
	"arkive/core/models"
	settingsrepo "arkive/core/repositories/settings"
	usersrepo "arkive/core/repositories/users"
	"arkive/pkg/storage/localclient"
	"arkive/pkg/validation"
)

type Service struct {
	db           database.PgPool
	settingsRepo *settingsrepo.Repository
	userRepo     *usersrepo.Repository
}

type StorageInput struct {
	Provider          string
	LocalPath         string
	StorageGB         string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3SessionToken    string
	S3Bucket          string
	S3Endpoint        string
	S3Region          string
	S3UsePathStyle    string
}

type EmailInput struct {
	Provider      string
	From          string
	PublicBaseURL string
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPass      string
}

type UploadInput struct {
	MaxUploadConcurrency string
	MaxQueueItems        string
}

func NewService(db database.PgPool, settingsRepo *settingsrepo.Repository, userRepo *usersrepo.Repository) *Service {
	return &Service{
		db:           db,
		settingsRepo: settingsRepo,
		userRepo:     userRepo,
	}
}

func NewLocalStorage(db database.PgPool, settingsRepo *settingsrepo.Repository) *localclient.Client {
	return localclient.New(func(ctx context.Context) (string, error) {
		settings, err := settingsRepo.GetStorageSettings(ctx, db)
		if err != nil {
			return "", err
		}
		return settings.LocalPath, nil
	})
}

func (s *Service) StorageSettings(ctx context.Context) (models.StorageSettings, error) {
	return s.settingsRepo.GetStorageSettings(ctx, s.db)
}

func (s *Service) EmailSettings(ctx context.Context) (models.EmailSettings, error) {
	return s.settingsRepo.GetEmailSettings(ctx, s.db)
}

func (s *Service) UploadSettings(ctx context.Context) (models.UploadSettings, error) {
	return s.settingsRepo.GetUploadSettings(ctx, s.db)
}

func (s *Service) UpdateStorageSettings(ctx context.Context, userID string, input StorageInput) (models.StorageSettings, validation.Errors, error) {
	current, _ := s.settingsRepo.GetStorageSettings(ctx, s.db)
	settings, validationErrors := BuildStorageSettings(input)
	if settings.Provider == "s3" && strings.TrimSpace(settings.S3SecretAccessKey) == "" {
		settings.S3SecretAccessKey = current.S3SecretAccessKey
	}
	ValidateStorageSettings(settings, validationErrors)
	if validationErrors.HasAny() {
		return settings, validationErrors, nil
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.StorageSettings{}, nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := SaveStorageSettingsTx(ctx, tx, s.settingsRepo, s.userRepo, userID, settings); err != nil {
		return models.StorageSettings{}, nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return models.StorageSettings{}, nil, err
	}
	return settings, nil, nil
}

func (s *Service) UpdateEmailSettings(ctx context.Context, userID string, input EmailInput) (models.EmailSettings, validation.Errors, error) {
	current, _ := s.settingsRepo.GetEmailSettings(ctx, s.db)
	settings, validationErrors := BuildEmailSettings(input)
	if settings.Provider == "smtp" && strings.TrimSpace(settings.From) == "" {
		settings.From = current.From
	}
	ValidateEmailSettings(settings, validationErrors)
	if validationErrors.HasAny() {
		return settings, validationErrors, nil
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.EmailSettings{}, nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.settingsRepo.SaveEmailSettings(ctx, tx, settings); err != nil {
		return models.EmailSettings{}, nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return models.EmailSettings{}, nil, err
	}
	return settings, nil, nil
}

func (s *Service) UpdateUploadSettings(ctx context.Context, userID string, input UploadInput) (models.UploadSettings, validation.Errors, error) {
	settings, validationErrors := BuildUploadSettings(input)
	ValidateUploadSettings(settings, validationErrors)
	if validationErrors.HasAny() {
		return settings, validationErrors, nil
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.UploadSettings{}, nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.settingsRepo.SaveUploadSettings(ctx, tx, settings); err != nil {
		return models.UploadSettings{}, nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return models.UploadSettings{}, nil, err
	}
	return settings, nil, nil
}

func SaveStorageSettingsTx(ctx context.Context, tx pgx.Tx, settingsRepo *settingsrepo.Repository, userRepo *usersrepo.Repository, userID string, settings models.StorageSettings) error {
	settings = NormalizeStorageSettings(settings)
	if settings.Provider == "local" {
		if err := os.MkdirAll(settings.LocalPath, 0o700); err != nil {
			return err
		}
	}
	if err := userRepo.UpdateQuota(ctx, tx, userID, quotaBytes(settings.MaxStorageBytes)); err != nil {
		return err
	}
	return settingsRepo.SaveStorageSettings(ctx, tx, settings)
}

func BuildStorageSettings(input StorageInput) (models.StorageSettings, validation.Errors) {
	validationErrors := validation.New()
	storageGB := strings.TrimSpace(input.StorageGB)
	maxBytes := int64(0)
	if storageGB != "" {
		gb, err := strconv.ParseInt(storageGB, 10, 64)
		if err != nil || gb < 0 {
			validationErrors.Add("storage_gb", "storage limit must be 0 or a positive number")
		} else if gb > 0 {
			maxBytes = gb * 1024 * 1024 * 1024
		}
	}
	return NormalizeStorageSettings(models.StorageSettings{
		Provider:          strings.ToLower(strings.TrimSpace(input.Provider)),
		LocalPath:         strings.TrimSpace(input.LocalPath),
		MaxStorageBytes:   maxBytes,
		S3AccessKeyID:     strings.TrimSpace(input.S3AccessKeyID),
		S3SecretAccessKey: strings.TrimSpace(input.S3SecretAccessKey),
		S3SessionToken:    strings.TrimSpace(input.S3SessionToken),
		S3Bucket:          strings.TrimSpace(input.S3Bucket),
		S3Endpoint:        strings.TrimSpace(input.S3Endpoint),
		S3Region:          strings.TrimSpace(input.S3Region),
		S3UsePathStyle:    strings.TrimSpace(input.S3UsePathStyle) == "on" || strings.TrimSpace(input.S3UsePathStyle) == "true",
	}), validationErrors
}

func NormalizeStorageSettings(settings models.StorageSettings) models.StorageSettings {
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

func BuildEmailSettings(input EmailInput) (models.EmailSettings, validation.Errors) {
	validationErrors := validation.New()
	smtpPort := 0
	if strings.TrimSpace(input.SMTPPort) != "" {
		port, err := strconv.Atoi(strings.TrimSpace(input.SMTPPort))
		if err != nil || port <= 0 {
			validationErrors.Add("smtp_port", "smtp port must be a positive number")
		} else {
			smtpPort = port
		}
	}
	if smtpPort == 0 {
		smtpPort = 587
	}
	return models.EmailSettings{
		Provider:      strings.ToLower(strings.TrimSpace(input.Provider)),
		From:          strings.TrimSpace(input.From),
		PublicBaseURL: strings.TrimSpace(input.PublicBaseURL),
		SMTPHost:      strings.TrimSpace(input.SMTPHost),
		SMTPPort:      smtpPort,
		SMTPUser:      strings.TrimSpace(input.SMTPUser),
		SMTPPass:      strings.TrimSpace(input.SMTPPass),
	}, validationErrors
}

func ValidateEmailSettings(settings models.EmailSettings, validationErrors validation.Errors) {
	switch settings.Provider {
	case "noop":
		return
	case "smtp":
		if settings.From == "" {
			validationErrors.Add("email_from", "from address is required")
		}
		if settings.SMTPHost == "" {
			validationErrors.Add("smtp_host", "smtp host is required")
		}
	default:
		validationErrors.Add("email_provider", "choose noop or smtp")
	}
}

func BuildUploadSettings(input UploadInput) (models.UploadSettings, validation.Errors) {
	validationErrors := validation.New()
	maxUploadConcurrency := 0
	if strings.TrimSpace(input.MaxUploadConcurrency) != "" {
		n, err := strconv.Atoi(strings.TrimSpace(input.MaxUploadConcurrency))
		if err != nil || n <= 0 {
			validationErrors.Add("max_upload_concurrency", "must be a positive number")
		} else {
			maxUploadConcurrency = n
		}
	}
	if maxUploadConcurrency == 0 {
		maxUploadConcurrency = 4
	}
	maxQueueItems := 0
	if strings.TrimSpace(input.MaxQueueItems) != "" {
		n, err := strconv.Atoi(strings.TrimSpace(input.MaxQueueItems))
		if err != nil || n <= 0 {
			validationErrors.Add("max_queue_items", "must be a positive number")
		} else {
			maxQueueItems = n
		}
	}
	if maxQueueItems == 0 {
		maxQueueItems = 300
	}
	return models.UploadSettings{MaxUploadConcurrency: maxUploadConcurrency, MaxQueueItems: maxQueueItems}, validationErrors
}

func ValidateUploadSettings(settings models.UploadSettings, validationErrors validation.Errors) {
	if settings.MaxUploadConcurrency <= 0 {
		validationErrors.Add("max_upload_concurrency", "must be a positive number")
	}
	if settings.MaxQueueItems <= 0 {
		validationErrors.Add("max_queue_items", "must be a positive number")
	}
}

func quotaBytes(maxStorageBytes int64) int64 {
	if maxStorageBytes <= 0 {
		return math.MaxInt64
	}
	return maxStorageBytes
}
