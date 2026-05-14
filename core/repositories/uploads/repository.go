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
		(file_id, provider_upload_id, upload_part_count, status, expires_at)
	VALUES
		($1, $2, $3, $4, $5)
	RETURNING
		id, file_id, provider_upload_id, upload_part_count, status, expires_at, created_at, updated_at`
	if err := db.QueryRow(ctx, query,
		upload.FileID,
		upload.ProviderUploadID,
		upload.UploadPartCount,
		upload.Status,
		upload.ExpiresAt,
	).Scan(
		&created.ID,
		&created.FileID,
		&created.ProviderUploadID,
		&created.UploadPartCount,
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
		upload_sessions.id, upload_sessions.file_id, upload_sessions.provider_upload_id,
		upload_sessions.upload_part_count,
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
		&upload.ProviderUploadID,
		&upload.UploadPartCount,
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
