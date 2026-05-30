package settings

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"

	"arkive/core/database"
	"arkive/core/models"
	settingsrepo "arkive/core/repositories/settings"
	"arkive/pkg/storage/localclient"
	"arkive/pkg/validation"
)

type Service struct {
	db           database.PgPool
	settingsRepo *settingsrepo.Repository
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

type UploadInput struct {
	MaxQueueItems    string
	PartConcurrency  string
	StaleUploadHours string
}

type PreviewInput struct {
	ImageMaxMB string
	VideoMaxMB string
	TextMaxMB  string
}

func NewService(db database.PgPool, settingsRepo *settingsrepo.Repository) *Service {
	return &Service{
		db:           db,
		settingsRepo: settingsRepo,
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

func (s *Service) UploadSettings(ctx context.Context) (models.UploadSettings, error) {
	return s.settingsRepo.GetUploadSettings(ctx, s.db)
}

func (s *Service) PreviewSettings(ctx context.Context) (models.PreviewSettings, error) {
	return s.settingsRepo.GetPreviewSettings(ctx, s.db)
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

	if err := SaveStorageSettingsTx(ctx, tx, s.settingsRepo, settings); err != nil {
		return models.StorageSettings{}, nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return models.StorageSettings{}, nil, err
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

func (s *Service) UpdatePreviewSettings(ctx context.Context, userID string, input PreviewInput) (models.PreviewSettings, validation.Errors, error) {
	settings, validationErrors := BuildPreviewSettings(input)
	ValidatePreviewSettings(settings, validationErrors)
	if validationErrors.HasAny() {
		return settings, validationErrors, nil
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.PreviewSettings{}, nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.settingsRepo.SavePreviewSettings(ctx, tx, settings); err != nil {
		return models.PreviewSettings{}, nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return models.PreviewSettings{}, nil, err
	}
	return settings, nil, nil
}

func SaveStorageSettingsTx(ctx context.Context, tx pgx.Tx, settingsRepo *settingsrepo.Repository, settings models.StorageSettings) error {
	settings = NormalizeStorageSettings(settings)
	if settings.Provider == "local" {
		if err := os.MkdirAll(settings.LocalPath, 0o700); err != nil {
			return err
		}
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

func DefaultUploadSettings() models.UploadSettings {
	return models.UploadSettings{
		MaxQueueItems:    300,
		PartConcurrency:  3,
		StaleUploadHours: 1,
	}
}

func DefaultPreviewSettings() models.PreviewSettings {
	return models.PreviewSettings{
		ImageMaxBytes: 50 * 1024 * 1024,
		VideoMaxBytes: 128 * 1024 * 1024,
		TextMaxBytes:  2 * 1024 * 1024,
	}
}

func BuildUploadSettings(input UploadInput) (models.UploadSettings, validation.Errors) {
	validationErrors := validation.New()
	defaults := DefaultUploadSettings()
	maxQueueItems := defaults.MaxQueueItems
	if strings.TrimSpace(input.MaxQueueItems) != "" {
		n, err := strconv.Atoi(strings.TrimSpace(input.MaxQueueItems))
		if err != nil || n <= 0 {
			validationErrors.Add("max_queue_items", "must be a positive number")
		} else {
			maxQueueItems = n
		}
	}
	partConcurrency := defaults.PartConcurrency
	if strings.TrimSpace(input.PartConcurrency) != "" {
		n, err := strconv.Atoi(strings.TrimSpace(input.PartConcurrency))
		if err != nil || n <= 0 {
			validationErrors.Add("part_concurrency", "must be a positive number")
		} else {
			partConcurrency = n
		}
	}
	staleUploadHours := defaults.StaleUploadHours
	if strings.TrimSpace(input.StaleUploadHours) != "" {
		n, err := strconv.Atoi(strings.TrimSpace(input.StaleUploadHours))
		if err != nil || n <= 0 {
			validationErrors.Add("stale_upload_hours", "must be a positive number")
		} else {
			staleUploadHours = n
		}
	}
	return models.UploadSettings{
		MaxQueueItems:    maxQueueItems,
		PartConcurrency:  partConcurrency,
		StaleUploadHours: staleUploadHours,
	}, validationErrors
}

func ValidateUploadSettings(settings models.UploadSettings, validationErrors validation.Errors) {
	if settings.MaxQueueItems <= 0 {
		validationErrors.Add("max_queue_items", "must be a positive number")
	}
	if settings.PartConcurrency <= 0 || settings.PartConcurrency > 8 {
		validationErrors.Add("part_concurrency", "must be between 1 and 8")
	}
	if settings.StaleUploadHours <= 0 || settings.StaleUploadHours > 168 {
		validationErrors.Add("stale_upload_hours", "must be between 1 and 168")
	}
}

func BuildPreviewSettings(input PreviewInput) (models.PreviewSettings, validation.Errors) {
	validationErrors := validation.New()
	defaults := DefaultPreviewSettings()
	imageMaxBytes := defaults.ImageMaxBytes
	videoMaxBytes := defaults.VideoMaxBytes
	textMaxBytes := defaults.TextMaxBytes

	parseMB := func(value string, field string, fallback int64) int64 {
		value = strings.TrimSpace(value)
		if value == "" {
			return fallback
		}
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil || n <= 0 {
			validationErrors.Add(field, "must be a positive number")
			return fallback
		}
		return n * 1024 * 1024
	}

	imageMaxBytes = parseMB(input.ImageMaxMB, "image_max_mb", imageMaxBytes)
	videoMaxBytes = parseMB(input.VideoMaxMB, "video_max_mb", videoMaxBytes)
	textMaxBytes = parseMB(input.TextMaxMB, "text_max_mb", textMaxBytes)

	return models.PreviewSettings{
		ImageMaxBytes: imageMaxBytes,
		VideoMaxBytes: videoMaxBytes,
		TextMaxBytes:  textMaxBytes,
	}, validationErrors
}

func ValidatePreviewSettings(settings models.PreviewSettings, validationErrors validation.Errors) {
	if settings.ImageMaxBytes < 1*1024*1024 || settings.ImageMaxBytes > 512*1024*1024 {
		validationErrors.Add("image_max_mb", "must be between 1 and 512")
	}
	if settings.VideoMaxBytes < 1*1024*1024 || settings.VideoMaxBytes > 2048*1024*1024 {
		validationErrors.Add("video_max_mb", "must be between 1 and 2048")
	}
	if settings.TextMaxBytes < 1*1024 || settings.TextMaxBytes > 32*1024*1024 {
		validationErrors.Add("text_max_mb", "must be between 1 and 32")
	}
}
