package uploadrepo

import (
	"context"

	"arkive/core/database"
	"arkive/core/models"
)

type Repository struct{}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) CreateUploadSession(ctx context.Context, db database.PgExecutor, upload models.UploadSession) (models.UploadSession, error) {
	var created models.UploadSession
	query := `INSERT INTO upload_sessions
		(file_id, storage_key, provider_upload_id, status, expires_at)
	VALUES
		($1, $2, $3, $4, $5)
	RETURNING
		id, file_id, storage_key, provider_upload_id, status, expires_at, created_at, updated_at`
	if err := db.QueryRow(ctx, query,
		upload.FileID,
		upload.StorageKey,
		upload.ProviderUploadID,
		upload.Status,
		upload.ExpiresAt,
	).Scan(
		&created.ID,
		&created.FileID,
		&created.StorageKey,
		&created.ProviderUploadID,
		&created.Status,
		&created.ExpiresAt,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return models.UploadSession{}, err
	}
	return created, nil
}

func (r *Repository) GetUploadSessionForUser(ctx context.Context, db database.PgExecutor, uploadSessionID, ownerID string) (models.UploadSession, error) {
	var upload models.UploadSession
	query := `SELECT
		upload_sessions.id, upload_sessions.file_id, upload_sessions.storage_key, upload_sessions.provider_upload_id,
		upload_sessions.status, upload_sessions.expires_at, upload_sessions.created_at, upload_sessions.updated_at
	FROM
		upload_sessions
	INNER JOIN
		files ON files.id = upload_sessions.file_id
	WHERE
		upload_sessions.id = $1 AND files.user_id = $2`
	if err := db.QueryRow(ctx, query, uploadSessionID, ownerID).Scan(
		&upload.ID,
		&upload.FileID,
		&upload.StorageKey,
		&upload.ProviderUploadID,
		&upload.Status,
		&upload.ExpiresAt,
		&upload.CreatedAt,
		&upload.UpdatedAt,
	); err != nil {
		return models.UploadSession{}, err
	}
	return upload, nil
}

func (r *Repository) UpdateUploadSessionStatus(ctx context.Context, db database.PgExecutor, uploadSessionID, status string) error {
	query := `UPDATE
		upload_sessions
	SET
		status = $2,
		updated_at = now()
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, uploadSessionID, status)
	return err
}

func (r *Repository) UpsertUploadPart(ctx context.Context, db database.PgExecutor, part models.UploadPart) (models.UploadPart, error) {
	var stored models.UploadPart
	query := `INSERT INTO upload_parts
		(upload_session_id, part_number, etag, encrypted_hash)
	VALUES
		($1, $2, $3, $4)
	ON CONFLICT (upload_session_id, part_number)
	DO UPDATE SET
		etag = EXCLUDED.etag,
		encrypted_hash = EXCLUDED.encrypted_hash
	RETURNING
		id, upload_session_id, part_number, COALESCE(etag, ''), encrypted_hash, created_at`
	if err := db.QueryRow(ctx, query,
		part.UploadSessionID,
		part.PartNumber,
		part.ETag,
		part.EncryptedHash,
	).Scan(
		&stored.ID,
		&stored.UploadSessionID,
		&stored.PartNumber,
		&stored.ETag,
		&stored.EncryptedHash,
		&stored.CreatedAt,
	); err != nil {
		return models.UploadPart{}, err
	}
	return stored, nil
}

func (r *Repository) ListUploadParts(ctx context.Context, db database.PgExecutor, uploadSessionID string) ([]models.UploadPart, error) {
	rows, err := db.Query(ctx, `SELECT
		id, upload_session_id, part_number, COALESCE(etag, ''), encrypted_hash, created_at
	FROM
		upload_parts
	WHERE
		upload_session_id = $1
	ORDER BY
		part_number ASC`, uploadSessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []models.UploadPart
	for rows.Next() {
		var part models.UploadPart
		if err := rows.Scan(
			&part.ID,
			&part.UploadSessionID,
			&part.PartNumber,
			&part.ETag,
			&part.EncryptedHash,
			&part.CreatedAt,
		); err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return parts, nil
}

func (r *Repository) ReplaceFileChunks(ctx context.Context, db database.PgExecutor, fileID string, chunks []models.FileChunk) error {
	if _, err := db.Exec(ctx, `DELETE FROM file_chunks WHERE file_id = $1`, fileID); err != nil {
		return err
	}
	for _, chunk := range chunks {
		query := `INSERT INTO file_chunks
			(file_id, chunk_index, storage_key, plaintext_size, encrypted_size, encrypted_hash)
		VALUES
			($1, $2, $3, $4, $5, $6)`
		if _, err := db.Exec(ctx, query,
			chunk.FileID,
			chunk.ChunkIndex,
			chunk.StorageKey,
			chunk.PlaintextSize,
			chunk.EncryptedSize,
			chunk.EncryptedHash,
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ListFileChunksByFile(ctx context.Context, db database.PgExecutor, fileID string) ([]models.FileChunk, error) {
	rows, err := db.Query(ctx, `SELECT
		id, file_id, chunk_index, storage_key, plaintext_size, encrypted_size, encrypted_hash, created_at
	FROM
		file_chunks
	WHERE
		file_id = $1
	ORDER BY
		chunk_index ASC`, fileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chunks []models.FileChunk
	for rows.Next() {
		var chunk models.FileChunk
		if err := rows.Scan(
			&chunk.ID,
			&chunk.FileID,
			&chunk.ChunkIndex,
			&chunk.StorageKey,
			&chunk.PlaintextSize,
			&chunk.EncryptedSize,
			&chunk.EncryptedHash,
			&chunk.CreatedAt,
		); err != nil {
			return nil, err
		}
		chunks = append(chunks, chunk)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return chunks, nil
}
