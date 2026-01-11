package filerepo

import (
	"context"
	"time"

	"arkive/core/database"
	"arkive/core/models"
)

type Repository struct{}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) CreateFile(ctx context.Context, db database.PgExecutor, file models.File) (models.File, error) {
	var created models.File
	query := `INSERT INTO files
		(user_id, bucket, object_key, folder_path, filename, content_type, size_bytes, status, expires_at)
	VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING
		id, user_id, bucket, object_key, folder_path, filename, content_type, size_bytes,
		video_width, video_height, video_duration_seconds,
		status, created_at, updated_at, expires_at`
	if err := db.QueryRow(ctx, query,
		file.UserID,
		file.Bucket,
		file.ObjectKey,
		file.FolderPath,
		file.Filename,
		file.ContentType,
		file.SizeBytes,
		file.Status,
		file.ExpiresAt,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.Bucket,
		&created.ObjectKey,
		&created.FolderPath,
		&created.Filename,
		&created.ContentType,
		&created.SizeBytes,
		&created.VideoWidth,
		&created.VideoHeight,
		&created.VideoDurationSeconds,
		&created.Status,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.ExpiresAt,
	); err != nil {
		return models.File{}, err
	}
	return created, nil
}

func (r *Repository) UpdateFileStatus(ctx context.Context, db database.PgExecutor, fileID, status string) error {
	query := `UPDATE
		files
	SET
		status = $2, updated_at = now()
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, fileID, status)
	return err
}

func (r *Repository) UpdateFileStatusIf(ctx context.Context, db database.PgExecutor, fileID, status string, allowed []string) (bool, error) {
	query := `UPDATE
		files
	SET
		status = $2, updated_at = now()
	WHERE
		id = $1 AND status = ANY($3)`
	tag, err := db.Exec(ctx, query, fileID, status, allowed)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *Repository) UpdateFileSize(ctx context.Context, db database.PgExecutor, fileID string, sizeBytes int64) error {
	query := `UPDATE
		files
	SET
		size_bytes = $2, updated_at = now()
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, fileID, sizeBytes)
	return err
}

func (r *Repository) UpdateFileContentType(ctx context.Context, db database.PgExecutor, fileID, contentType string) error {
	query := `UPDATE
		files
	SET
		content_type = $2, updated_at = now()
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, fileID, contentType)
	return err
}

func (r *Repository) UpdateFileExpiry(ctx context.Context, db database.PgExecutor, fileID string, expiresAt time.Time) error {
	query := `UPDATE
		files
	SET
		expires_at = $2, updated_at = now()
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, fileID, expiresAt)
	return err
}

func (r *Repository) ClearFileExpiry(ctx context.Context, db database.PgExecutor, fileID string) error {
	query := `UPDATE
		files
	SET
		expires_at = NULL, updated_at = now()
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, fileID)
	return err
}

func (r *Repository) MarkInactiveFilesForUser(ctx context.Context, db database.PgExecutor, userID string, expiresAt time.Time) (int64, error) {
	query := `UPDATE
		files
	SET
		expires_at = $2, updated_at = now()
	WHERE
		user_id = $1
		AND status = 'complete'
		AND expires_at IS NULL`
	tag, err := db.Exec(ctx, query, userID, expiresAt)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *Repository) ClearExpiryForUserCompletedFiles(ctx context.Context, db database.PgExecutor, userID string) (int64, error) {
	query := `UPDATE
		files
	SET
		expires_at = NULL, updated_at = now()
	WHERE
		user_id = $1
		AND status = 'complete'
		AND expires_at IS NOT NULL`
	tag, err := db.Exec(ctx, query, userID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *Repository) ListArchivedFilesForUser(ctx context.Context, db database.PgExecutor, userID string) ([]models.File, error) {
	query := `SELECT
		id, user_id, bucket, object_key, folder_path, filename, content_type, size_bytes,
		video_width, video_height, video_duration_seconds,
		status, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		user_id = $1
		AND status = 'complete'
		AND expires_at IS NOT NULL
	ORDER BY
		expires_at ASC`
	rows, err := db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var file models.File
		if err := rows.Scan(
			&file.ID,
			&file.UserID,
			&file.Bucket,
			&file.ObjectKey,
			&file.FolderPath,
			&file.Filename,
			&file.ContentType,
			&file.SizeBytes,
			&file.VideoWidth,
			&file.VideoHeight,
			&file.VideoDurationSeconds,
			&file.Status,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.ExpiresAt,
		); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return files, nil
}

func (r *Repository) CountActiveFilesForUser(ctx context.Context, db database.PgExecutor, userID string) (int64, error) {
	var total int64
	query := `SELECT
		COUNT(*)
	FROM
		files
	WHERE
		user_id = $1
		AND status IN ('pending', 'uploading', 'complete')`
	if err := db.QueryRow(ctx, query, userID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) CountArchivedFilesForUser(ctx context.Context, db database.PgExecutor, userID string) (int64, error) {
	var total int64
	query := `SELECT
		COUNT(*)
	FROM
		files
	WHERE
		user_id = $1
		AND status = 'complete'
		AND expires_at IS NOT NULL`
	if err := db.QueryRow(ctx, query, userID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) UpdateVideoMetadata(ctx context.Context, db database.PgExecutor, fileID string, width, height int, durationSeconds int64) error {
	query := `UPDATE
		files
	SET
		video_width = $2,
		video_height = $3,
		video_duration_seconds = $4,
		updated_at = now()
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, fileID, width, height, durationSeconds)
	return err
}

func (r *Repository) GetFileByID(ctx context.Context, db database.PgExecutor, fileID string) (models.File, error) {
	var file models.File
	query := `SELECT
		id, user_id, bucket, object_key, folder_path, filename, content_type, size_bytes,
		video_width, video_height, video_duration_seconds,
		status, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		id = $1`
	if err := db.QueryRow(ctx, query, fileID).Scan(
		&file.ID,
		&file.UserID,
		&file.Bucket,
		&file.ObjectKey,
		&file.FolderPath,
		&file.Filename,
		&file.ContentType,
		&file.SizeBytes,
		&file.VideoWidth,
		&file.VideoHeight,
		&file.VideoDurationSeconds,
		&file.Status,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.ExpiresAt,
	); err != nil {
		return models.File{}, err
	}
	return file, nil
}

func (r *Repository) GetFileForUser(ctx context.Context, db database.PgExecutor, fileID, userID string) (models.File, error) {
	var file models.File
	query := `SELECT
		id, user_id, bucket, object_key, folder_path, filename, content_type, size_bytes,
		video_width, video_height, video_duration_seconds,
		status, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		id = $1 AND user_id = $2`
	if err := db.QueryRow(ctx, query, fileID, userID).Scan(
		&file.ID,
		&file.UserID,
		&file.Bucket,
		&file.ObjectKey,
		&file.FolderPath,
		&file.Filename,
		&file.ContentType,
		&file.SizeBytes,
		&file.VideoWidth,
		&file.VideoHeight,
		&file.VideoDurationSeconds,
		&file.Status,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.ExpiresAt,
	); err != nil {
		return models.File{}, err
	}
	return file, nil
}

func (r *Repository) ListPendingForUser(ctx context.Context, db database.PgExecutor, userID string) ([]models.File, error) {
	query := `SELECT
		id, user_id, bucket, object_key, folder_path, filename, content_type, size_bytes,
		video_width, video_height, video_duration_seconds,
		status, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		user_id = $1 AND status IN ('pending', 'uploading')
	ORDER BY
		created_at DESC`
	rows, err := db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var file models.File
		if err := rows.Scan(
			&file.ID,
			&file.UserID,
			&file.Bucket,
			&file.ObjectKey,
			&file.FolderPath,
			&file.Filename,
			&file.ContentType,
			&file.SizeBytes,
			&file.VideoWidth,
			&file.VideoHeight,
			&file.VideoDurationSeconds,
			&file.Status,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.ExpiresAt,
		); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return files, nil
}

func (r *Repository) ListCompletedForUser(ctx context.Context, db database.PgExecutor, userID string) ([]models.File, error) {
	query := `SELECT
		id, user_id, bucket, object_key, folder_path, filename, content_type, size_bytes,
		video_width, video_height, video_duration_seconds,
		status, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		user_id = $1
		AND status = 'complete'
		AND expires_at IS NULL
	ORDER BY
		created_at DESC`
	rows, err := db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var file models.File
		if err := rows.Scan(
			&file.ID,
			&file.UserID,
			&file.Bucket,
			&file.ObjectKey,
			&file.FolderPath,
			&file.Filename,
			&file.ContentType,
			&file.SizeBytes,
			&file.VideoWidth,
			&file.VideoHeight,
			&file.VideoDurationSeconds,
			&file.Status,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.ExpiresAt,
		); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return files, nil
}

func (r *Repository) ListCompletedForUserInFolder(ctx context.Context, db database.PgExecutor, userID, folderPath string) ([]models.File, error) {
	query := `SELECT
		id, user_id, bucket, object_key, folder_path, filename, content_type, size_bytes,
		video_width, video_height, video_duration_seconds,
		status, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		user_id = $1
		AND folder_path = $2
		AND status = 'complete'
		AND expires_at IS NULL
	ORDER BY
		created_at DESC`
	rows, err := db.Query(ctx, query, userID, folderPath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var file models.File
		if err := rows.Scan(
			&file.ID,
			&file.UserID,
			&file.Bucket,
			&file.ObjectKey,
			&file.FolderPath,
			&file.Filename,
			&file.ContentType,
			&file.SizeBytes,
			&file.VideoWidth,
			&file.VideoHeight,
			&file.VideoDurationSeconds,
			&file.Status,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.ExpiresAt,
		); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return files, nil
}

func (r *Repository) DeleteFileForUser(ctx context.Context, db database.PgExecutor, fileID, userID string) (bool, error) {
	query := `DELETE FROM
		files
	WHERE
		id = $1 AND user_id = $2`
	tag, err := db.Exec(ctx, query, fileID, userID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *Repository) ListExpiredCompleteFiles(ctx context.Context, db database.PgExecutor, cutoff time.Time) ([]models.File, error) {
	query := `SELECT
		id, user_id, bucket, object_key, folder_path, filename, content_type, size_bytes,
		video_width, video_height, video_duration_seconds,
		status, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		status = 'complete'
		AND expires_at IS NOT NULL
		AND expires_at <= $1
	ORDER BY
		expires_at ASC`
	rows, err := db.Query(ctx, query, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var file models.File
		if err := rows.Scan(
			&file.ID,
			&file.UserID,
			&file.Bucket,
			&file.ObjectKey,
			&file.FolderPath,
			&file.Filename,
			&file.ContentType,
			&file.SizeBytes,
			&file.VideoWidth,
			&file.VideoHeight,
			&file.VideoDurationSeconds,
			&file.Status,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.ExpiresAt,
		); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return files, nil
}

func (r *Repository) ListExpiredUploads(ctx context.Context, db database.PgExecutor, cutoff time.Time) ([]models.File, error) {
	query := `SELECT
		id, user_id, bucket, object_key, folder_path, filename, content_type, size_bytes,
		video_width, video_height, video_duration_seconds,
		status, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		status IN ('pending', 'uploading')
		AND expires_at IS NOT NULL
		AND expires_at <= $1
	ORDER BY
		expires_at ASC`
	rows, err := db.Query(ctx, query, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var file models.File
		if err := rows.Scan(
			&file.ID,
			&file.UserID,
			&file.Bucket,
			&file.ObjectKey,
			&file.FolderPath,
			&file.Filename,
			&file.ContentType,
			&file.SizeBytes,
			&file.VideoWidth,
			&file.VideoHeight,
			&file.VideoDurationSeconds,
			&file.Status,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.ExpiresAt,
		); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return files, nil
}
