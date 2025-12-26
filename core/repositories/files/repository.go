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

func (r *Repository) UpdateFileSize(ctx context.Context, db database.PgExecutor, fileID string, sizeBytes int64) error {
	query := `UPDATE files
		SET size_bytes = $2, updated_at = now()
		WHERE id = $1`
	_, err := db.Exec(ctx, query, fileID, sizeBytes)
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

func (r *Repository) ListPendingForUser(ctx context.Context, db database.PgExecutor, userID string) ([]models.File, error) {
	query := `SELECT id, user_id, bucket, object_key, filename, content_type, size_bytes, status, created_at, updated_at
		FROM files
		WHERE user_id = $1 AND status IN ('pending', 'uploading')
		ORDER BY created_at DESC`
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
			&file.Filename,
			&file.ContentType,
			&file.SizeBytes,
			&file.Status,
			&file.CreatedAt,
			&file.UpdatedAt,
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
