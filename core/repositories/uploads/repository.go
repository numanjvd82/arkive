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

func (r *Repository) CreateMultipart(ctx context.Context, db database.PgExecutor, upload models.MultipartUpload) (models.MultipartUpload, error) {
	var created models.MultipartUpload
	query := `INSERT INTO multipart_uploads (file_id, upload_id, bucket, object_key, chunk_size, total_parts, uploaded_parts, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, file_id, upload_id, bucket, object_key, chunk_size, total_parts, uploaded_parts, status, created_at, updated_at`
	if err := db.QueryRow(ctx, query,
		upload.FileID,
		upload.UploadID,
		upload.Bucket,
		upload.ObjectKey,
		upload.ChunkSize,
		upload.TotalParts,
		upload.UploadedParts,
		upload.Status,
	).Scan(
		&created.ID,
		&created.FileID,
		&created.UploadID,
		&created.Bucket,
		&created.ObjectKey,
		&created.ChunkSize,
		&created.TotalParts,
		&created.UploadedParts,
		&created.Status,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return models.MultipartUpload{}, err
	}
	return created, nil
}

func (r *Repository) GetMultipartForUser(ctx context.Context, db database.PgExecutor, multipartID, userID string) (models.MultipartUpload, error) {
	var upload models.MultipartUpload
	query := `SELECT m.id, m.file_id, m.upload_id, m.bucket, m.object_key, m.chunk_size, m.total_parts, m.uploaded_parts, m.status, m.created_at, m.updated_at
		FROM multipart_uploads m
		JOIN files f ON f.id = m.file_id
		WHERE m.id = $1 AND f.user_id = $2`
	if err := db.QueryRow(ctx, query, multipartID, userID).Scan(
		&upload.ID,
		&upload.FileID,
		&upload.UploadID,
		&upload.Bucket,
		&upload.ObjectKey,
		&upload.ChunkSize,
		&upload.TotalParts,
		&upload.UploadedParts,
		&upload.Status,
		&upload.CreatedAt,
		&upload.UpdatedAt,
	); err != nil {
		return models.MultipartUpload{}, err
	}
	return upload, nil
}

func (r *Repository) GetMultipartForFile(ctx context.Context, db database.PgExecutor, fileID, userID string) (models.MultipartUpload, error) {
	var upload models.MultipartUpload
	query := `SELECT m.id, m.file_id, m.upload_id, m.bucket, m.object_key, m.chunk_size, m.total_parts, m.uploaded_parts, m.status, m.created_at, m.updated_at
		FROM multipart_uploads m
		JOIN files f ON f.id = m.file_id
		WHERE m.file_id = $1 AND f.user_id = $2`
	if err := db.QueryRow(ctx, query, fileID, userID).Scan(
		&upload.ID,
		&upload.FileID,
		&upload.UploadID,
		&upload.Bucket,
		&upload.ObjectKey,
		&upload.ChunkSize,
		&upload.TotalParts,
		&upload.UploadedParts,
		&upload.Status,
		&upload.CreatedAt,
		&upload.UpdatedAt,
	); err != nil {
		return models.MultipartUpload{}, err
	}
	return upload, nil
}

func (r *Repository) UpdateMultipart(ctx context.Context, db database.PgExecutor, multipartID, status string, uploadedParts []byte) error {
	query := `UPDATE multipart_uploads
		SET status = $2, uploaded_parts = $3, updated_at = now()
		WHERE id = $1`
	_, err := db.Exec(ctx, query, multipartID, status, uploadedParts)
	return err
}

func (r *Repository) TouchMultipart(ctx context.Context, db database.PgExecutor, multipartID, status string) error {
	query := `UPDATE multipart_uploads
		SET status = $2, updated_at = now()
		WHERE id = $1`
	_, err := db.Exec(ctx, query, multipartID, status)
	return err
}
