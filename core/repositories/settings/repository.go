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
	partConcurrency, _ := strconv.Atoi(values["upload.part_concurrency"])
	if partConcurrency <= 0 {
		partConcurrency = 3
	}
	staleUploadHours, _ := strconv.Atoi(values["upload.stale_hours"])
	if staleUploadHours <= 0 {
		staleUploadHours = 1
	}
	return models.UploadSettings{
		MaxQueueItems:    maxQueueItems,
		PartConcurrency:  partConcurrency,
		StaleUploadHours: staleUploadHours,
	}, nil
}

func (r *Repository) SaveUploadSettings(ctx context.Context, db database.PgExecutor, settings models.UploadSettings) error {
	values := map[string]string{
		"upload.max_queue_items":  strconv.Itoa(settings.MaxQueueItems),
		"upload.part_concurrency": strconv.Itoa(settings.PartConcurrency),
		"upload.stale_hours":      strconv.Itoa(settings.StaleUploadHours),
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

func (r *Repository) GetPreviewSettings(ctx context.Context, db database.PgExecutor) (models.PreviewSettings, error) {
	rows, err := db.Query(ctx, `SELECT key, value FROM instance_settings WHERE key LIKE 'preview.%'`)
	if err != nil {
		return models.PreviewSettings{}, err
	}
	defer rows.Close()

	values := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return models.PreviewSettings{}, err
		}
		values[key] = value
	}
	if rows.Err() != nil {
		return models.PreviewSettings{}, rows.Err()
	}
	if len(values) == 0 {
		return models.PreviewSettings{}, ErrStorageSettingsNotFound
	}

	imageMaxBytes, _ := strconv.ParseInt(values["preview.image_max_bytes"], 10, 64)
	if imageMaxBytes <= 0 {
		imageMaxBytes = 50 * 1024 * 1024
	}
	videoMaxBytes, _ := strconv.ParseInt(values["preview.video_max_bytes"], 10, 64)
	if videoMaxBytes <= 0 {
		videoMaxBytes = 128 * 1024 * 1024
	}
	textMaxBytes, _ := strconv.ParseInt(values["preview.text_max_bytes"], 10, 64)
	if textMaxBytes <= 0 {
		textMaxBytes = 2 * 1024 * 1024
	}

	return models.PreviewSettings{
		ImageMaxBytes: imageMaxBytes,
		VideoMaxBytes: videoMaxBytes,
		TextMaxBytes:  textMaxBytes,
	}, nil
}

func (r *Repository) SavePreviewSettings(ctx context.Context, db database.PgExecutor, settings models.PreviewSettings) error {
	values := map[string]string{
		"preview.image_max_bytes": strconv.FormatInt(settings.ImageMaxBytes, 10),
		"preview.video_max_bytes": strconv.FormatInt(settings.VideoMaxBytes, 10),
		"preview.text_max_bytes":  strconv.FormatInt(settings.TextMaxBytes, 10),
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
