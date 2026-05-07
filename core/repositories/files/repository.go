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

func hydrateLegacyFile(file *models.File) {
	if file == nil {
		return
	}
	if file.Filename == "" {
		file.Filename = "file-" + shortID(file.ID)
	}
	if file.ContentType == "" {
		file.ContentType = "application/octet-stream"
	}
	if file.SizeBytes == 0 {
		file.SizeBytes = file.PlaintextSize
	}
	if file.Status == "" {
		file.Status = file.UploadStatus
	}
	if file.Bucket == "" {
		file.Bucket = file.StorageBackend
	}
}

func shortID(value string) string {
	if len(value) >= 8 {
		return value[:8]
	}
	return value
}

func (r *Repository) CreateFile(ctx context.Context, db database.PgExecutor, file models.File) (models.File, error) {
	var created models.File
	query := `INSERT INTO files
		(user_id, encrypted_metadata, encrypted_file_key, encryption_version, chunk_size, chunk_count, plaintext_size, encrypted_size, encrypted_hash, upload_status, storage_backend, expires_at)
	VALUES
		($1, $2, $3, 1, $4, 1, $5, $5, NULL, $6, $7, $8)
	RETURNING
		id, user_id, encrypted_metadata, encrypted_file_key, encryption_version, chunk_size, chunk_count,
		plaintext_size, encrypted_size, encrypted_hash, upload_status, storage_backend, completed_at, created_at, updated_at, expires_at`
	if err := db.QueryRow(ctx, query,
		file.UserID,
		[]byte{},
		[]byte{},
		file.SizeBytes,
		file.SizeBytes,
		"pending",
		"local",
		file.ExpiresAt,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.EncryptedMetadata,
		&created.EncryptedFileKey,
		&created.EncryptionVersion,
		&created.ChunkSize,
		&created.ChunkCount,
		&created.PlaintextSize,
		&created.EncryptedSize,
		&created.EncryptedHash,
		&created.UploadStatus,
		&created.StorageBackend,
		&created.CompletedAt,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.ExpiresAt,
	); err != nil {
		return models.File{}, err
	}
	hydrateLegacyFile(&created)
	return created, nil
}

func (r *Repository) CreateEncryptedFile(ctx context.Context, db database.PgExecutor, file models.File) (models.File, error) {
	var created models.File
	query := `INSERT INTO files
		(user_id, encrypted_metadata, encrypted_file_key, encryption_version, chunk_size, chunk_count, plaintext_size, encrypted_size, encrypted_hash, upload_status, storage_backend, expires_at)
	VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	RETURNING
		id, user_id, encrypted_metadata, encrypted_file_key, encryption_version, chunk_size, chunk_count,
		plaintext_size, encrypted_size, encrypted_hash, upload_status, storage_backend, completed_at, created_at, updated_at, expires_at`
	if err := db.QueryRow(ctx, query,
		file.UserID,
		file.EncryptedMetadata,
		file.EncryptedFileKey,
		file.EncryptionVersion,
		file.ChunkSize,
		file.ChunkCount,
		file.PlaintextSize,
		file.EncryptedSize,
		file.EncryptedHash,
		file.UploadStatus,
		file.StorageBackend,
		file.ExpiresAt,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.EncryptedMetadata,
		&created.EncryptedFileKey,
		&created.EncryptionVersion,
		&created.ChunkSize,
		&created.ChunkCount,
		&created.PlaintextSize,
		&created.EncryptedSize,
		&created.EncryptedHash,
		&created.UploadStatus,
		&created.StorageBackend,
		&created.CompletedAt,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.ExpiresAt,
	); err != nil {
		return models.File{}, err
	}
	hydrateLegacyFile(&created)
	return created, nil
}

func (r *Repository) GetEncryptedFileForUser(ctx context.Context, db database.PgExecutor, fileID, userID string) (models.File, error) {
	var file models.File
	query := `SELECT
		id, user_id, encrypted_metadata, encrypted_file_key, encryption_version, chunk_size, chunk_count,
		plaintext_size, encrypted_size, encrypted_hash, upload_status, storage_backend, completed_at, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		id = $1 AND user_id = $2`
	if err := db.QueryRow(ctx, query, fileID, userID).Scan(
		&file.ID,
		&file.UserID,
		&file.EncryptedMetadata,
		&file.EncryptedFileKey,
		&file.EncryptionVersion,
		&file.ChunkSize,
		&file.ChunkCount,
		&file.PlaintextSize,
		&file.EncryptedSize,
		&file.EncryptedHash,
		&file.UploadStatus,
		&file.StorageBackend,
		&file.CompletedAt,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.ExpiresAt,
	); err != nil {
		return models.File{}, err
	}
	hydrateLegacyFile(&file)
	return file, nil
}

func (r *Repository) MarkEncryptedFileComplete(ctx context.Context, db database.PgExecutor, fileID string, encryptedSize int64, encryptedHash []byte) error {
	query := `UPDATE
		files
	SET
		encrypted_size = $2,
		encrypted_hash = $3,
		upload_status = 'complete',
		completed_at = now(),
		expires_at = NULL,
		updated_at = now()
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, fileID, encryptedSize, encryptedHash)
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

func (r *Repository) UpdateFileStatus(ctx context.Context, db database.PgExecutor, fileID, status string) error {
	query := `UPDATE
		files
	SET
		upload_status = $2, updated_at = now()
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, fileID, status)
	return err
}

func (r *Repository) UpdateFileStatusIf(ctx context.Context, db database.PgExecutor, fileID, status string, allowed []string) (bool, error) {
	query := `UPDATE
		files
	SET
		upload_status = $2, updated_at = now()
	WHERE
		id = $1 AND upload_status = ANY($3)`
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
		plaintext_size = $2, updated_at = now()
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, fileID, sizeBytes)
	return err
}

func (r *Repository) UpdateFileContentType(ctx context.Context, db database.PgExecutor, fileID, contentType string) error {
	return nil
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
		AND upload_status = 'complete'
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
		AND upload_status = 'complete'
		AND expires_at IS NOT NULL`
	tag, err := db.Exec(ctx, query, userID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *Repository) ListArchivedFilesForUser(ctx context.Context, db database.PgExecutor, userID string) ([]models.File, error) {
	query := `SELECT
		id, user_id, encrypted_metadata, encrypted_file_key, encryption_version, chunk_size, chunk_count,
		plaintext_size, encrypted_size, encrypted_hash, upload_status, storage_backend, completed_at, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		user_id = $1
		AND upload_status = 'complete'
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
			&file.EncryptedMetadata,
			&file.EncryptedFileKey,
			&file.EncryptionVersion,
			&file.ChunkSize,
			&file.ChunkCount,
			&file.PlaintextSize,
			&file.EncryptedSize,
			&file.EncryptedHash,
			&file.UploadStatus,
			&file.StorageBackend,
			&file.CompletedAt,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.ExpiresAt,
		); err != nil {
			return nil, err
		}
		hydrateLegacyFile(&file)
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
		AND upload_status IN ('pending', 'uploading', 'complete')`
	if err := db.QueryRow(ctx, query, userID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) CountInFlightForUser(ctx context.Context, db database.PgExecutor, userID string) (int64, error) {
	var total int64
	query := `SELECT
		COUNT(*)
	FROM
		files
	WHERE
		user_id = $1
		AND upload_status IN ('pending', 'uploading')`
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

func (r *Repository) UpdateVideoMetadata(ctx context.Context, db database.PgExecutor, fileID string, width, height int, durationSeconds int64) error {
	return nil
}

func (r *Repository) GetFileByID(ctx context.Context, db database.PgExecutor, fileID string) (models.File, error) {
	var file models.File
	query := `SELECT
		id, user_id, encrypted_metadata, encrypted_file_key, encryption_version, chunk_size, chunk_count,
		plaintext_size, encrypted_size, encrypted_hash, upload_status, storage_backend, completed_at, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		id = $1`
	if err := db.QueryRow(ctx, query, fileID).Scan(
		&file.ID,
		&file.UserID,
		&file.EncryptedMetadata,
		&file.EncryptedFileKey,
		&file.EncryptionVersion,
		&file.ChunkSize,
		&file.ChunkCount,
		&file.PlaintextSize,
		&file.EncryptedSize,
		&file.EncryptedHash,
		&file.UploadStatus,
		&file.StorageBackend,
		&file.CompletedAt,
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
		id, user_id, encrypted_metadata, encrypted_file_key, encryption_version, chunk_size, chunk_count,
		plaintext_size, encrypted_size, encrypted_hash, upload_status, storage_backend, completed_at, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		id = $1 AND user_id = $2`
	if err := db.QueryRow(ctx, query, fileID, userID).Scan(
		&file.ID,
		&file.UserID,
		&file.EncryptedMetadata,
		&file.EncryptedFileKey,
		&file.EncryptionVersion,
		&file.ChunkSize,
		&file.ChunkCount,
		&file.PlaintextSize,
		&file.EncryptedSize,
		&file.EncryptedHash,
		&file.UploadStatus,
		&file.StorageBackend,
		&file.CompletedAt,
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
		id, user_id, encrypted_metadata, encrypted_file_key, encryption_version, chunk_size, chunk_count,
		plaintext_size, encrypted_size, encrypted_hash, upload_status, storage_backend, completed_at, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		user_id = $1 AND upload_status IN ('pending', 'uploading')
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
			&file.EncryptedMetadata,
			&file.EncryptedFileKey,
			&file.EncryptionVersion,
			&file.ChunkSize,
			&file.ChunkCount,
			&file.PlaintextSize,
			&file.EncryptedSize,
			&file.EncryptedHash,
			&file.UploadStatus,
			&file.StorageBackend,
			&file.CompletedAt,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.ExpiresAt,
		); err != nil {
			return nil, err
		}
		hydrateLegacyFile(&file)
		files = append(files, file)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return files, nil
}

func (r *Repository) ListCompletedForUser(ctx context.Context, db database.PgExecutor, userID, sort string, page, pageSize int) ([]models.File, error) {
	if pageSize <= 0 {
		pageSize = 25
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize
	orderBy := resolveFilesOrderBy(sort)
	query := `SELECT
		id, user_id, encrypted_metadata, encrypted_file_key, encryption_version, chunk_size, chunk_count,
		plaintext_size, encrypted_size, encrypted_hash, upload_status, storage_backend, completed_at, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		user_id = $1
		AND upload_status = 'complete'
		AND expires_at IS NULL
	ORDER BY ` + orderBy + `
	LIMIT $2 OFFSET $3`
	rows, err := db.Query(ctx, query, userID, pageSize, offset)
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
			&file.EncryptedMetadata,
			&file.EncryptedFileKey,
			&file.EncryptionVersion,
			&file.ChunkSize,
			&file.ChunkCount,
			&file.PlaintextSize,
			&file.EncryptedSize,
			&file.EncryptedHash,
			&file.UploadStatus,
			&file.StorageBackend,
			&file.CompletedAt,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.ExpiresAt,
		); err != nil {
			return nil, err
		}
		hydrateLegacyFile(&file)
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

func resolveFilesOrderBy(sort string) string {
	switch sort {
	case "name_asc":
		return "created_at ASC"
	case "name_desc":
		return "created_at DESC"
	case "size_asc":
		return "plaintext_size ASC"
	case "size_desc":
		return "plaintext_size DESC"
	case "updated_asc":
		return "updated_at ASC"
	case "updated_desc":
		return "updated_at DESC"
	default:
		return "updated_at DESC"
	}
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
		id, user_id, encrypted_metadata, encrypted_file_key, encryption_version, chunk_size, chunk_count,
		plaintext_size, encrypted_size, encrypted_hash, upload_status, storage_backend, completed_at, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		upload_status = 'complete'
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
			&file.EncryptedMetadata,
			&file.EncryptedFileKey,
			&file.EncryptionVersion,
			&file.ChunkSize,
			&file.ChunkCount,
			&file.PlaintextSize,
			&file.EncryptedSize,
			&file.EncryptedHash,
			&file.UploadStatus,
			&file.StorageBackend,
			&file.CompletedAt,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.ExpiresAt,
		); err != nil {
			return nil, err
		}
		hydrateLegacyFile(&file)
		files = append(files, file)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return files, nil
}

func (r *Repository) ListExpiredUploads(ctx context.Context, db database.PgExecutor, cutoff time.Time) ([]models.File, error) {
	query := `SELECT
		id, user_id, encrypted_metadata, encrypted_file_key, encryption_version, chunk_size, chunk_count,
		plaintext_size, encrypted_size, encrypted_hash, upload_status, storage_backend, completed_at, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		upload_status IN ('pending', 'uploading')
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
			&file.EncryptedMetadata,
			&file.EncryptedFileKey,
			&file.EncryptionVersion,
			&file.ChunkSize,
			&file.ChunkCount,
			&file.PlaintextSize,
			&file.EncryptedSize,
			&file.EncryptedHash,
			&file.UploadStatus,
			&file.StorageBackend,
			&file.CompletedAt,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.ExpiresAt,
		); err != nil {
			return nil, err
		}
		hydrateLegacyFile(&file)
		files = append(files, file)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return files, nil
}
