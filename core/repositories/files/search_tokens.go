package filerepo

import (
	"context"

	"arkive/core/database"
	"arkive/core/models"
)

func (r *Repository) InsertSearchTokensForFile(ctx context.Context, db database.PgExecutor, userID, vaultID, fileID string, tokens []models.FileSearchToken) error {
	if len(tokens) == 0 {
		return nil
	}
	query := `INSERT INTO file_search_tokens
		(user_id, vault_id, file_id, token_hash, field, weight)
	VALUES
		($1, $2, $3, $4, $5, $6)
	ON CONFLICT DO NOTHING`
	for _, token := range tokens {
		if _, err := db.Exec(ctx, query, userID, vaultID, fileID, token.TokenHash, token.Field, token.Weight); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ReplaceSearchTokensForFile(ctx context.Context, db database.PgExecutor, userID, vaultID, fileID string, tokens []models.FileSearchToken) error {
	if _, err := db.Exec(ctx, `DELETE FROM file_search_tokens WHERE user_id = $1 AND vault_id = $2 AND file_id = $3`, userID, vaultID, fileID); err != nil {
		return err
	}
	return r.InsertSearchTokensForFile(ctx, db, userID, vaultID, fileID, tokens)
}

func (r *Repository) SearchCompletedForTokens(ctx context.Context, db database.PgExecutor, userID, vaultID string, tokenHashes [][]byte, limit int) ([]models.File, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := db.Query(ctx, `SELECT
		f.id,
		f.user_id,
		f.folder_id,
		f.encrypted_metadata,
		f.encrypted_file_key,
		f.encrypted_manifest,
		f.encryption_version,
		f.chunk_size,
		f.chunk_count,
		f.plaintext_size,
		f.actual_encrypted_size,
		f.encrypted_hash,
		f.upload_status,
		f.thumbnail_status,
		f.thumbnail_size_bytes,
		f.thumbnail_mime,
		f.thumbnail_width,
		f.thumbnail_height,
		f.completed_at,
		f.created_at,
		f.updated_at,
		f.expires_at,
		SUM(fst.weight) AS score
	FROM
		file_search_tokens fst
	INNER JOIN files f ON f.id = fst.file_id
	WHERE
		fst.user_id = $1
		AND fst.vault_id = $2
		AND fst.token_hash = ANY($3::bytea[])
		AND f.upload_status = 'complete'
		AND f.expires_at IS NULL
	GROUP BY
		f.id
	ORDER BY
		score DESC,
		f.updated_at DESC
	LIMIT $4`, userID, vaultID, tokenHashes, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make([]models.File, 0, limit)
	for rows.Next() {
		var file models.File
		var folderID *string
		var score int64
		if err := rows.Scan(
			&file.ID,
			&file.UserID,
			&folderID,
			&file.EncryptedMetadata,
			&file.EncryptedFileKey,
			&file.EncryptedManifest,
			&file.EncryptionVersion,
			&file.ChunkSize,
			&file.ChunkCount,
			&file.PlaintextSize,
			&file.ActualEncryptedSize,
			&file.EncryptedHash,
			&file.UploadStatus,
			&file.ThumbnailStatus,
			&file.ThumbnailSizeBytes,
			&file.ThumbnailMime,
			&file.ThumbnailWidth,
			&file.ThumbnailHeight,
			&file.CompletedAt,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.ExpiresAt,
			&score,
		); err != nil {
			return nil, err
		}
		file.FolderID = folderID
		files = append(files, file)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return files, nil
}
