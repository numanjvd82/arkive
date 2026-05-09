package settingsrepo

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"

	"arkive/core/database"
	"arkive/core/models"
)

var ErrStorageSettingsNotFound = errors.New("storage settings not found")

type Repository struct{}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) GetStorageSettings(ctx context.Context, db database.PgExecutor) (models.StorageSettings, error) {
	rows, err := db.Query(ctx, `SELECT key, value FROM instance_settings WHERE key LIKE 'storage.%'`)
	if err != nil {
		return models.StorageSettings{}, err
	}
	defer rows.Close()

	values := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return models.StorageSettings{}, err
		}
		values[key] = value
	}
	if rows.Err() != nil {
		return models.StorageSettings{}, rows.Err()
	}
	if len(values) == 0 {
		return models.StorageSettings{}, ErrStorageSettingsNotFound
	}

	maxStorageBytes, _ := strconv.ParseInt(values["storage.max_bytes"], 10, 64)
	usePathStyle, _ := strconv.ParseBool(values["storage.s3_use_path_style"])
	settings := models.StorageSettings{
		Provider:          strings.TrimSpace(values["storage.provider"]),
		LocalPath:         strings.TrimSpace(values["storage.local_path"]),
		MaxStorageBytes:   maxStorageBytes,
		S3AccessKeyID:     strings.TrimSpace(values["storage.s3_access_key_id"]),
		S3SecretAccessKey: values["storage.s3_secret_access_key"],
		S3SessionToken:    values["storage.s3_session_token"],
		S3Bucket:          strings.TrimSpace(values["storage.s3_bucket"]),
		S3Endpoint:        strings.TrimSpace(values["storage.s3_endpoint"]),
		S3Region:          strings.TrimSpace(values["storage.s3_region"]),
		S3UsePathStyle:    usePathStyle,
	}
	return settings, nil
}

func (r *Repository) GetEmailSettings(ctx context.Context, db database.PgExecutor) (models.EmailSettings, error) {
	rows, err := db.Query(ctx, `SELECT key, value FROM instance_settings WHERE key LIKE 'email.%'`)
	if err != nil {
		return models.EmailSettings{}, err
	}
	defer rows.Close()

	values := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return models.EmailSettings{}, err
		}
		values[key] = value
	}
	if rows.Err() != nil {
		return models.EmailSettings{}, rows.Err()
	}
	if len(values) == 0 {
		return models.EmailSettings{}, ErrStorageSettingsNotFound
	}

	smtpPort, _ := strconv.Atoi(values["email.smtp_port"])
	if smtpPort <= 0 {
		smtpPort = 587
	}
	settings := models.EmailSettings{
		Provider:      strings.TrimSpace(values["email.provider"]),
		From:          strings.TrimSpace(values["email.from"]),
		PublicBaseURL: strings.TrimSpace(values["email.public_base_url"]),
		SMTPHost:      strings.TrimSpace(values["email.smtp_host"]),
		SMTPPort:      smtpPort,
		SMTPUser:      strings.TrimSpace(values["email.smtp_user"]),
		SMTPPass:      values["email.smtp_pass"],
	}
	return settings, nil
}

func (r *Repository) SaveEmailSettings(ctx context.Context, db database.PgExecutor, settings models.EmailSettings) error {
	values := map[string]string{
		"email.provider":        settings.Provider,
		"email.from":            settings.From,
		"email.public_base_url": settings.PublicBaseURL,
		"email.smtp_host":       settings.SMTPHost,
		"email.smtp_port":       strconv.Itoa(settings.SMTPPort),
		"email.smtp_user":       settings.SMTPUser,
		"email.smtp_pass":       settings.SMTPPass,
	}
	for key, value := range values {
		if _, err := db.Exec(ctx, `
			INSERT INTO instance_settings (key, value, updated_at)
			VALUES ($1, $2, now())
			ON CONFLICT (key)
			DO UPDATE SET value = EXCLUDED.value, updated_at = now()
		`, key, value); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) GetUploadSettings(ctx context.Context, db database.PgExecutor) (models.UploadSettings, error) {
	rows, err := db.Query(ctx, `SELECT key, value FROM instance_settings WHERE key LIKE 'upload.%'`)
	if err != nil {
		return models.UploadSettings{}, err
	}
	defer rows.Close()

	values := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return models.UploadSettings{}, err
		}
		values[key] = value
	}
	if rows.Err() != nil {
		return models.UploadSettings{}, rows.Err()
	}
	if len(values) == 0 {
		return models.UploadSettings{}, ErrStorageSettingsNotFound
	}

	maxQueueItems, _ := strconv.Atoi(values["upload.max_queue_items"])
	if maxQueueItems <= 0 {
		maxQueueItems = 300
	}
	return models.UploadSettings{
		MaxQueueItems: maxQueueItems,
	}, nil
}

func (r *Repository) SaveUploadSettings(ctx context.Context, db database.PgExecutor, settings models.UploadSettings) error {
	values := map[string]string{
		"upload.max_queue_items": strconv.Itoa(settings.MaxQueueItems),
	}
	for key, value := range values {
		if _, err := db.Exec(ctx, `
			INSERT INTO instance_settings (key, value, updated_at)
			VALUES ($1, $2, now())
			ON CONFLICT (key)
			DO UPDATE SET value = EXCLUDED.value, updated_at = now()
		`, key, value); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) SaveStorageSettings(ctx context.Context, db database.PgExecutor, settings models.StorageSettings) error {
	values := map[string]string{
		"storage.provider":             settings.Provider,
		"storage.local_path":           settings.LocalPath,
		"storage.max_bytes":            strconv.FormatInt(settings.MaxStorageBytes, 10),
		"storage.s3_access_key_id":     settings.S3AccessKeyID,
		"storage.s3_secret_access_key": settings.S3SecretAccessKey,
		"storage.s3_session_token":     settings.S3SessionToken,
		"storage.s3_bucket":            settings.S3Bucket,
		"storage.s3_endpoint":          settings.S3Endpoint,
		"storage.s3_region":            settings.S3Region,
		"storage.s3_use_path_style":    strconv.FormatBool(settings.S3UsePathStyle),
	}
	for key, value := range values {
		if _, err := db.Exec(ctx, `
			INSERT INTO instance_settings (key, value, updated_at)
			VALUES ($1, $2, now())
			ON CONFLICT (key)
			DO UPDATE SET value = EXCLUDED.value, updated_at = now()
		`, key, value); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) HasStorageSettings(ctx context.Context, db database.PgExecutor) (bool, error) {
	var exists bool
	err := db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM instance_settings WHERE key = 'storage.provider')`).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return exists, err
}
