package filerepo

import (
	"context"

	"arkive/core/database"
	"arkive/core/models"
)

type Repository struct{}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) CreateEncryptedFile(ctx context.Context, db database.PgExecutor, file models.File) (models.File, error) {
	var created models.File
	var folderID *string
	query := `INSERT INTO files
		(user_id, folder_id, encrypted_metadata, encrypted_file_key, encrypted_manifest, encryption_version, chunk_size, chunk_count, plaintext_size, actual_encrypted_size, encrypted_hash, upload_status, expires_at)
	VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	RETURNING
		id, user_id, folder_id, encrypted_metadata, encrypted_file_key, encrypted_manifest, encryption_version, chunk_size, chunk_count,
		plaintext_size, actual_encrypted_size, encrypted_hash, upload_status,
		thumbnail_status, thumbnail_size_bytes, thumbnail_mime, thumbnail_width, thumbnail_height,
		completed_at, created_at, updated_at, expires_at`
	if err := db.QueryRow(ctx, query,
		file.UserID,
		file.FolderID,
		file.EncryptedMetadata,
		file.EncryptedFileKey,
		file.EncryptedManifest,
		file.EncryptionVersion,
		file.ChunkSize,
		file.ChunkCount,
		file.PlaintextSize,
		file.ActualEncryptedSize,
		file.EncryptedHash,
		file.UploadStatus,
		file.ExpiresAt,
	).Scan(
		&created.ID,
		&created.UserID,
		&folderID,
		&created.EncryptedMetadata,
		&created.EncryptedFileKey,
		&created.EncryptedManifest,
		&created.EncryptionVersion,
		&created.ChunkSize,
		&created.ChunkCount,
		&created.PlaintextSize,
		&created.ActualEncryptedSize,
		&created.EncryptedHash,
		&created.UploadStatus,
		&created.ThumbnailStatus,
		&created.ThumbnailSizeBytes,
		&created.ThumbnailMime,
		&created.ThumbnailWidth,
		&created.ThumbnailHeight,
		&created.CompletedAt,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.ExpiresAt,
	); err != nil {
		return models.File{}, err
	}
	created.FolderID = folderID
	return created, nil
}

func (r *Repository) GetEncryptedFileForUser(ctx context.Context, db database.PgExecutor, fileID, userID string) (models.File, error) {
	var file models.File
	var folderID *string
	query := `SELECT
		id, user_id, folder_id, encrypted_metadata, encrypted_file_key, encrypted_manifest, encryption_version, chunk_size, chunk_count,
		plaintext_size, actual_encrypted_size, encrypted_hash, upload_status,
		thumbnail_status, thumbnail_size_bytes, thumbnail_mime, thumbnail_width, thumbnail_height,
		completed_at, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		id = $1 AND user_id = $2`
	if err := db.QueryRow(ctx, query, fileID, userID).Scan(
		&file.ID,
		&file.UserID,
		&folderID,
		&file.EncryptedMetadata,
		&file.EncryptedFileKey,
		&file.EncryptedManifest,
		&file.EncryptionVersion,
		&file.ChunkSize,
		&file.ChunkCount,
		&file.PlaintextSize,
		&file.ActualEncryptedSize,
		&file.EncryptedHash,
		&file.UploadStatus,
		&file.ThumbnailStatus,
		&file.ThumbnailSizeBytes,
		&file.ThumbnailMime,
		&file.ThumbnailWidth,
		&file.ThumbnailHeight,
		&file.CompletedAt,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.ExpiresAt,
	); err != nil {
		return models.File{}, err
	}
	file.FolderID = folderID
	return file, nil
}

func (r *Repository) UpdateEncryptedFileEnvelope(ctx context.Context, db database.PgExecutor, fileID string, encryptedMetadata, encryptedFileKey, encryptedManifest, encryptedHash []byte) error {
	query := `UPDATE
		files
	SET
		encrypted_metadata = $2,
		encrypted_file_key = $3,
		encrypted_manifest = $4,
		encrypted_hash = $5,
		updated_at = now()
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, fileID, encryptedMetadata, encryptedFileKey, encryptedManifest, encryptedHash)
	return err
}

func (r *Repository) GetEncryptedFileRecordForUser(ctx context.Context, db database.PgExecutor, fileID, userID string) (models.File, error) {
	var file models.File
	var folderID *string
	query := `SELECT
		id, user_id, folder_id, encrypted_metadata, encrypted_file_key, encrypted_manifest, encryption_version,
		chunk_size, chunk_count, plaintext_size, actual_encrypted_size, encrypted_hash, upload_status,
		thumbnail_status, thumbnail_size_bytes, thumbnail_mime, thumbnail_width, thumbnail_height
	FROM
		files
	WHERE
		id = $1 AND user_id = $2`
	if err := db.QueryRow(ctx, query, fileID, userID).Scan(
		&file.ID,
		&file.UserID,
		&folderID,
		&file.EncryptedMetadata,
		&file.EncryptedFileKey,
		&file.EncryptedManifest,
		&file.EncryptionVersion,
		&file.ChunkSize,
		&file.ChunkCount,
		&file.PlaintextSize,
		&file.ActualEncryptedSize,
		&file.EncryptedHash,
		&file.UploadStatus,
		&file.ThumbnailStatus,
		&file.ThumbnailSizeBytes,
		&file.ThumbnailMime,
		&file.ThumbnailWidth,
		&file.ThumbnailHeight,
	); err != nil {
		return models.File{}, err
	}
	file.FolderID = folderID
	return file, nil
}

func (r *Repository) MarkEncryptedFileComplete(ctx context.Context, db database.PgExecutor, fileID string, actualEncryptedSize int64, encryptedHash []byte) error {
	query := `UPDATE
		files
	SET
		actual_encrypted_size = $2,
		encrypted_hash = $3,
		upload_status = 'complete',
		completed_at = now(),
		expires_at = NULL,
		updated_at = now()
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, fileID, actualEncryptedSize, encryptedHash)
	return err
}

func (r *Repository) UpdateThumbnailInfo(ctx context.Context, db database.PgExecutor, fileID, status string, sizeBytes int64, mime string, width, height int) error {
	query := `UPDATE
		files
	SET
		thumbnail_status = $2,
		thumbnail_size_bytes = $3,
		thumbnail_mime = $4,
		thumbnail_width = $5,
		thumbnail_height = $6,
		updated_at = now()
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, fileID, status, sizeBytes, mime, width, height)
	return err
}

func (r *Repository) UpdateEncryptedFileStatusIf(ctx context.Context, db database.PgExecutor, fileID, status string, allowed []string) (bool, error) {
	query := `UPDATE
		files
	SET
		upload_status = $2,
		updated_at = now()
	WHERE
		id = $1 AND upload_status = ANY($3)`
	tag, err := db.Exec(ctx, query, fileID, status, allowed)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *Repository) CountInFlightForUser(ctx context.Context, db database.PgExecutor, userID string) (int64, error) {
	var total int64
	query := `SELECT
		COUNT(DISTINCT files.id)
	FROM
		files
	INNER JOIN
		upload_sessions ON upload_sessions.file_id = files.id
	WHERE
		files.user_id = $1
		AND files.upload_status IN ('pending', 'uploading')
		AND upload_sessions.status = 'active'
		AND upload_sessions.expires_at > now()`
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
		AND upload_status = 'complete'
		AND expires_at IS NOT NULL`
	if err := db.QueryRow(ctx, query, userID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) GetFileByID(ctx context.Context, db database.PgExecutor, fileID string) (models.File, error) {
	var file models.File
	var folderID *string
	query := `SELECT
		id, user_id, folder_id, encrypted_metadata, encrypted_file_key, encrypted_manifest, encryption_version, chunk_size, chunk_count,
		plaintext_size, actual_encrypted_size, encrypted_hash, upload_status,
		thumbnail_status, thumbnail_size_bytes, thumbnail_mime, thumbnail_width, thumbnail_height,
		completed_at, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		id = $1`
	if err := db.QueryRow(ctx, query, fileID).Scan(
		&file.ID,
		&file.UserID,
		&folderID,
		&file.EncryptedMetadata,
		&file.EncryptedFileKey,
		&file.EncryptedManifest,
		&file.EncryptionVersion,
		&file.ChunkSize,
		&file.ChunkCount,
		&file.PlaintextSize,
		&file.ActualEncryptedSize,
		&file.EncryptedHash,
		&file.UploadStatus,
		&file.ThumbnailStatus,
		&file.ThumbnailSizeBytes,
		&file.ThumbnailMime,
		&file.ThumbnailWidth,
		&file.ThumbnailHeight,
		&file.CompletedAt,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.ExpiresAt,
	); err != nil {
		return models.File{}, err
	}
	file.FolderID = folderID
	return file, nil
}

func (r *Repository) GetFileForUser(ctx context.Context, db database.PgExecutor, fileID, userID string) (models.File, error) {
	var file models.File
	var folderID *string
	query := `SELECT
		id, user_id, folder_id, encrypted_metadata, encrypted_file_key, encrypted_manifest, encryption_version, chunk_size, chunk_count,
		plaintext_size, actual_encrypted_size, encrypted_hash, upload_status,
		thumbnail_status, thumbnail_size_bytes, thumbnail_mime, thumbnail_width, thumbnail_height,
		completed_at, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		id = $1 AND user_id = $2`
	if err := db.QueryRow(ctx, query, fileID, userID).Scan(
		&file.ID,
		&file.UserID,
		&folderID,
		&file.EncryptedMetadata,
		&file.EncryptedFileKey,
		&file.EncryptedManifest,
		&file.EncryptionVersion,
		&file.ChunkSize,
		&file.ChunkCount,
		&file.PlaintextSize,
		&file.ActualEncryptedSize,
		&file.EncryptedHash,
		&file.UploadStatus,
		&file.ThumbnailStatus,
		&file.ThumbnailSizeBytes,
		&file.ThumbnailMime,
		&file.ThumbnailWidth,
		&file.ThumbnailHeight,
		&file.CompletedAt,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.ExpiresAt,
	); err != nil {
		return models.File{}, err
	}
	file.FolderID = folderID
	return file, nil
}

func (r *Repository) ListCompletedForUser(ctx context.Context, db database.PgExecutor, userID string, page, pageSize int) ([]models.File, error) {
	if pageSize <= 0 {
		pageSize = 25
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize
	query := `SELECT
		id, user_id, folder_id, encrypted_metadata, encrypted_file_key, encrypted_manifest, encryption_version, chunk_size, chunk_count,
		plaintext_size, actual_encrypted_size, encrypted_hash, upload_status, completed_at, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		user_id = $1
		AND upload_status = 'complete'
		AND expires_at IS NULL
	ORDER BY updated_at DESC
	LIMIT $2 OFFSET $3`
	rows, err := db.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var file models.File
		var folderID *string
		if err := rows.Scan(
			&file.ID,
			&file.UserID,
			&folderID,
			&file.EncryptedMetadata,
			&file.EncryptedFileKey,
			&file.EncryptedManifest,
			&file.EncryptionVersion,
			&file.ChunkSize,
			&file.ChunkCount,
			&file.PlaintextSize,
			&file.ActualEncryptedSize,
			&file.EncryptedHash,
			&file.UploadStatus,
			&file.CompletedAt,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.ExpiresAt,
		); err != nil {
			return nil, err
		}
		file.FolderID = folderID
		files = append(files, file)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return files, nil
}

func (r *Repository) CountCompletedForUser(ctx context.Context, db database.PgExecutor, userID string) (int, error) {
	query := `SELECT
		COUNT(1)
	FROM
		files
	WHERE
		user_id = $1
		AND upload_status = 'complete'
		AND expires_at IS NULL`
	var count int
	if err := db.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) CountByFolder(ctx context.Context, db database.PgExecutor, userID string, folderID *string) (int, error) {
	query := `SELECT
		COUNT(1)
	FROM
		files
	WHERE
		user_id = $1
		AND upload_status = 'complete'
		AND expires_at IS NULL
		AND (
			($2::uuid IS NULL AND folder_id IS NULL)
			OR folder_id = $2
		)`
	var count int
	if err := db.QueryRow(ctx, query, userID, folderID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) ListByFolder(ctx context.Context, db database.PgExecutor, userID string, folderID *string, limit, offset int) ([]models.File, error) {
	query := `SELECT
		id, user_id, folder_id, encrypted_metadata, encrypted_file_key, encrypted_manifest, encryption_version, chunk_size, chunk_count,
		plaintext_size, actual_encrypted_size, encrypted_hash, upload_status, completed_at, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		user_id = $1
		AND upload_status = 'complete'
		AND expires_at IS NULL
		AND (
			($2::uuid IS NULL AND folder_id IS NULL)
			OR folder_id = $2
		)
	ORDER BY created_at DESC
	LIMIT $3 OFFSET $4`
	rows, err := db.Query(ctx, query, userID, folderID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := []models.File{}
	for rows.Next() {
		var file models.File
		var scannedFolderID *string
		if err := rows.Scan(
			&file.ID,
			&file.UserID,
			&scannedFolderID,
			&file.EncryptedMetadata,
			&file.EncryptedFileKey,
			&file.EncryptedManifest,
			&file.EncryptionVersion,
			&file.ChunkSize,
			&file.ChunkCount,
			&file.PlaintextSize,
			&file.ActualEncryptedSize,
			&file.EncryptedHash,
			&file.UploadStatus,
			&file.CompletedAt,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.ExpiresAt,
		); err != nil {
			return nil, err
		}
		file.FolderID = scannedFolderID
		files = append(files, file)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return files, nil
}

func (r *Repository) MoveFilesToFolder(ctx context.Context, db database.PgExecutor, userID string, fileIDs []string, targetFolderID *string) (int64, error) {
	query := `UPDATE files
	SET
		folder_id = $3,
		updated_at = now()
	WHERE
		user_id = $1
		AND id = ANY($2::uuid[])
		AND upload_status = 'complete'
		AND expires_at IS NULL`
	tag, err := db.Exec(ctx, query, userID, fileIDs, targetFolderID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *Repository) RenameFileForUser(ctx context.Context, db database.PgExecutor, fileID, userID string, encryptedMetadata []byte) (bool, error) {
	query := `UPDATE files
	SET
		encrypted_metadata = $3,
		updated_at = now()
	WHERE
		id = $1
		AND user_id = $2
		AND upload_status = 'complete'
		AND expires_at IS NULL`
	tag, err := db.Exec(ctx, query, fileID, userID, encryptedMetadata)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
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
