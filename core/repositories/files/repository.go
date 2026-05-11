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
	query := `INSERT INTO files
		(user_id, encrypted_metadata, encrypted_file_key, encrypted_manifest, encryption_version, chunk_size, chunk_count, plaintext_size, encrypted_size, encrypted_hash, upload_status, expires_at)
	VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	RETURNING
		id, user_id, encrypted_metadata, encrypted_file_key, encrypted_manifest, encryption_version, chunk_size, chunk_count,
		plaintext_size, encrypted_size, encrypted_hash, upload_status, completed_at, created_at, updated_at, expires_at`
	if err := db.QueryRow(ctx, query,
		file.UserID,
		file.EncryptedMetadata,
		file.EncryptedFileKey,
		file.EncryptedManifest,
		file.EncryptionVersion,
		file.ChunkSize,
		file.ChunkCount,
		file.PlaintextSize,
		file.EncryptedSize,
		file.EncryptedHash,
		file.UploadStatus,
		file.ExpiresAt,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.EncryptedMetadata,
		&created.EncryptedFileKey,
		&created.EncryptedManifest,
		&created.EncryptionVersion,
		&created.ChunkSize,
		&created.ChunkCount,
		&created.PlaintextSize,
		&created.EncryptedSize,
		&created.EncryptedHash,
		&created.UploadStatus,
		&created.CompletedAt,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.ExpiresAt,
	); err != nil {
		return models.File{}, err
	}
	return created, nil
}

func (r *Repository) GetEncryptedFileForUser(ctx context.Context, db database.PgExecutor, fileID, userID string) (models.File, error) {
	var file models.File
	query := `SELECT
		id, user_id, encrypted_metadata, encrypted_file_key, encrypted_manifest, encryption_version, chunk_size, chunk_count,
		plaintext_size, encrypted_size, encrypted_hash, upload_status, completed_at, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		id = $1 AND user_id = $2`
	if err := db.QueryRow(ctx, query, fileID, userID).Scan(
		&file.ID,
		&file.UserID,
		&file.EncryptedMetadata,
		&file.EncryptedFileKey,
		&file.EncryptedManifest,
		&file.EncryptionVersion,
		&file.ChunkSize,
		&file.ChunkCount,
		&file.PlaintextSize,
		&file.EncryptedSize,
		&file.EncryptedHash,
		&file.UploadStatus,
		&file.CompletedAt,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.ExpiresAt,
	); err != nil {
		return models.File{}, err
	}
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
	query := `SELECT
		id, user_id, encrypted_metadata, encrypted_file_key, encrypted_manifest, encryption_version,
		chunk_size, chunk_count, plaintext_size, encrypted_size, encrypted_hash, upload_status
	FROM
		files
	WHERE
		id = $1 AND user_id = $2`
	if err := db.QueryRow(ctx, query, fileID, userID).Scan(
		&file.ID,
		&file.UserID,
		&file.EncryptedMetadata,
		&file.EncryptedFileKey,
		&file.EncryptedManifest,
		&file.EncryptionVersion,
		&file.ChunkSize,
		&file.ChunkCount,
		&file.PlaintextSize,
		&file.EncryptedSize,
		&file.EncryptedHash,
		&file.UploadStatus,
	); err != nil {
		return models.File{}, err
	}
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
	query := `SELECT
		id, user_id, encrypted_metadata, encrypted_file_key, encrypted_manifest, encryption_version, chunk_size, chunk_count,
		plaintext_size, encrypted_size, encrypted_hash, upload_status, completed_at, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		id = $1`
	if err := db.QueryRow(ctx, query, fileID).Scan(
		&file.ID,
		&file.UserID,
		&file.EncryptedMetadata,
		&file.EncryptedFileKey,
		&file.EncryptedManifest,
		&file.EncryptionVersion,
		&file.ChunkSize,
		&file.ChunkCount,
		&file.PlaintextSize,
		&file.EncryptedSize,
		&file.EncryptedHash,
		&file.UploadStatus,
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
		id, user_id, encrypted_metadata, encrypted_file_key, encrypted_manifest, encryption_version, chunk_size, chunk_count,
		plaintext_size, encrypted_size, encrypted_hash, upload_status, completed_at, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		id = $1 AND user_id = $2`
	if err := db.QueryRow(ctx, query, fileID, userID).Scan(
		&file.ID,
		&file.UserID,
		&file.EncryptedMetadata,
		&file.EncryptedFileKey,
		&file.EncryptedManifest,
		&file.EncryptionVersion,
		&file.ChunkSize,
		&file.ChunkCount,
		&file.PlaintextSize,
		&file.EncryptedSize,
		&file.EncryptedHash,
		&file.UploadStatus,
		&file.CompletedAt,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.ExpiresAt,
	); err != nil {
		return models.File{}, err
	}
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
		id, user_id, encrypted_metadata, encrypted_file_key, encrypted_manifest, encryption_version, chunk_size, chunk_count,
		plaintext_size, encrypted_size, encrypted_hash, upload_status, completed_at, created_at, updated_at, expires_at
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
		if err := rows.Scan(
			&file.ID,
			&file.UserID,
			&file.EncryptedMetadata,
			&file.EncryptedFileKey,
			&file.EncryptedManifest,
			&file.EncryptionVersion,
			&file.ChunkSize,
			&file.ChunkCount,
			&file.PlaintextSize,
			&file.EncryptedSize,
			&file.EncryptedHash,
			&file.UploadStatus,
			&file.CompletedAt,
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
