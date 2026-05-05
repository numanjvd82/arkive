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
		id, user_id, bucket, object_key, filename, content_type, size_bytes,
		video_width, video_height, video_duration_seconds,
		status, created_at, updated_at, expires_at
	FROM
		files
	WHERE
		user_id = $1
		AND status = 'complete'
		AND expires_at IS NULL
		AND (
			filename ILIKE $2
			OR content_type ILIKE $2
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
			&file.Bucket,
			&file.ObjectKey,
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return files, nil
}
