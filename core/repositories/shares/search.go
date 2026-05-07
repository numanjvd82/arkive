package sharerepo

import (
	"context"
	"strings"

	"arkive/core/database"
	"arkive/core/models"
)

func (r *Repository) SearchSharesForUser(ctx context.Context, db database.PgExecutor, ownerUserID, query string, limit int) ([]models.ShareWithFile, error) {
	query = strings.TrimSpace(query)
	if limit <= 0 {
		limit = 5
	}
	pattern := "%" + query + "%"
	rows, err := db.Query(ctx, `SELECT
		s.id,
		s.file_id,
		s.owner_user_id,
		s.token,
		s.password_hash,
		s.expires_at,
		s.status,
		s.revoked_at,
		s.created_at,
		s.updated_at,
		concat('file-', left(f.id::text, 8)),
		'application/octet-stream',
		f.plaintext_size,
		f.updated_at
	FROM
		shares s
	JOIN
		files f ON f.id = s.file_id
	WHERE
		s.owner_user_id = $1
		AND (
			f.id::text ILIKE $2
			OR s.token ILIKE $2
		)
	ORDER BY
		s.updated_at DESC
	LIMIT $3`, ownerUserID, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.ShareWithFile{}
	for rows.Next() {
		var share models.ShareWithFile
		if err := rows.Scan(
			&share.ID,
			&share.FileID,
			&share.OwnerUserID,
			&share.Token,
			&share.PasswordHash,
			&share.ExpiresAt,
			&share.Status,
			&share.RevokedAt,
			&share.CreatedAt,
			&share.UpdatedAt,
			&share.FileName,
			&share.FileContentType,
			&share.FileSizeBytes,
			&share.FileUpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, share)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
