package filerepo

import (
	"context"
	"strings"

	"arkive/core/database"
	"arkive/core/models"
)

func (r *Repository) SearchCompletedForUser(ctx context.Context, db database.PgExecutor, userID, query string, limit int) ([]models.File, error) {
	query = strings.TrimSpace(query)
	if limit <= 0 {
		limit = 5
	}
	pattern := "%" + query + "%"
	rows, err := db.Query(ctx, `SELECT
		id, user_id, encrypted_metadata, encrypted_file_key, encryption_version, chunk_size, chunk_count,
		plaintext_size, encrypted_size, encrypted_hash, upload_status, storage_backend, completed_at, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		user_id = $1
		AND upload_status = 'complete'
		AND expires_at IS NULL
		AND (
			id::text ILIKE $2
			OR upload_status ILIKE $2
		)
	ORDER BY
		updated_at DESC
	LIMIT $3`, userID, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := []models.File{}
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
		files = append(files, file)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return files, nil
}
