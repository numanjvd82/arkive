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
		sl.id,
		ssf.file_id,
		sl.owner_user_id,
		sl.token,
		sl.encrypted_share_key,
		sl.password_hash,
		sl.expires_at,
		sl.status,
		sl.revoked_at,
		sl.created_at,
		sl.updated_at,
		concat('file-', left(f.id::text, 8)),
		'application/octet-stream',
		f.plaintext_size,
		f.updated_at
	FROM
		share_links sl
	JOIN
		share_items si ON si.share_link_id = sl.id
	JOIN
		share_snapshot_files ssf ON ssf.share_item_id = si.id
	JOIN
		files f ON f.id = ssf.file_id
	WHERE
		sl.owner_user_id = $1
		AND (
			f.id::text ILIKE $2
			OR sl.token ILIKE $2
		)
	ORDER BY
		sl.updated_at DESC
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
			&share.EncryptedShareKey,
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
