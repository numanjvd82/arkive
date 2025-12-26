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

func (r *Repository) CreateFile(ctx context.Context, db database.PgExecutor, file models.File) (models.File, error) {
	var created models.File
	query := `INSERT INTO files (user_id, bucket, object_key, filename, content_type, size_bytes, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, bucket, object_key, filename, content_type, size_bytes, status, created_at, updated_at`
	if err := db.QueryRow(ctx, query,
		file.UserID,
		file.Bucket,
		file.ObjectKey,
		file.Filename,
		file.ContentType,
		file.SizeBytes,
		file.Status,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.Bucket,
		&created.ObjectKey,
		&created.Filename,
		&created.ContentType,
		&created.SizeBytes,
		&created.Status,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return models.File{}, err
	}
	return created, nil
}

func (r *Repository) UpdateFileStatus(ctx context.Context, db database.PgExecutor, fileID, status string) error {
	query := `UPDATE files
		SET status = $2, updated_at = now()
		WHERE id = $1`
	_, err := db.Exec(ctx, query, fileID, status)
	return err
}

func (r *Repository) GetFileByID(ctx context.Context, db database.PgExecutor, fileID string) (models.File, error) {
	var file models.File
	query := `SELECT id, user_id, bucket, object_key, filename, content_type, size_bytes, status, created_at, updated_at
		FROM files
		WHERE id = $1`
	if err := db.QueryRow(ctx, query, fileID).Scan(
		&file.ID,
		&file.UserID,
		&file.Bucket,
		&file.ObjectKey,
		&file.Filename,
		&file.ContentType,
		&file.SizeBytes,
		&file.Status,
		&file.CreatedAt,
		&file.UpdatedAt,
	); err != nil {
		return models.File{}, err
	}
	return file, nil
}

func (r *Repository) GetFileForUser(ctx context.Context, db database.PgExecutor, fileID, userID string) (models.File, error) {
	var file models.File
	query := `SELECT id, user_id, bucket, object_key, filename, content_type, size_bytes, status, created_at, updated_at
		FROM files
		WHERE id = $1 AND user_id = $2`
	if err := db.QueryRow(ctx, query, fileID, userID).Scan(
		&file.ID,
		&file.UserID,
		&file.Bucket,
		&file.ObjectKey,
		&file.Filename,
		&file.ContentType,
		&file.SizeBytes,
		&file.Status,
		&file.CreatedAt,
		&file.UpdatedAt,
	); err != nil {
		return models.File{}, err
	}
	return file, nil
}

func (r *Repository) CreateShare(ctx context.Context, db database.PgExecutor, share models.FileShare) (models.FileShare, error) {
	var created models.FileShare
	query := `INSERT INTO file_shares (file_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, file_id, token_hash, expires_at, created_at, updated_at`
	if err := db.QueryRow(ctx, query,
		share.FileID,
		share.TokenHash,
		share.ExpiresAt,
	).Scan(
		&created.ID,
		&created.FileID,
		&created.TokenHash,
		&created.ExpiresAt,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return models.FileShare{}, err
	}
	return created, nil
}

func (r *Repository) GetShareByTokenHash(ctx context.Context, db database.PgExecutor, tokenHash []byte) (models.FileShare, error) {
	var share models.FileShare
	query := `SELECT id, file_id, token_hash, expires_at, created_at, updated_at
		FROM file_shares
		WHERE token_hash = $1`
	if err := db.QueryRow(ctx, query, tokenHash).Scan(
		&share.ID,
		&share.FileID,
		&share.TokenHash,
		&share.ExpiresAt,
		&share.CreatedAt,
		&share.UpdatedAt,
	); err != nil {
		return models.FileShare{}, err
	}
	return share, nil
}
