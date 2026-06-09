package foldersrepo

import (
	"context"

	"arkive/core/database"
	"arkive/core/models"
)

func (r *Repository) InsertSearchTokensForFolder(ctx context.Context, db database.PgExecutor, userID, vaultID, folderID string, tokens []models.FileSearchToken) error {
	if len(tokens) == 0 {
		return nil
	}
	query := `INSERT INTO folder_search_tokens
		(user_id, vault_id, folder_id, token_hash, field, weight)
	VALUES
		($1, $2, $3, $4, $5, $6)
	ON CONFLICT DO NOTHING`
	for _, token := range tokens {
		if _, err := db.Exec(ctx, query, userID, vaultID, folderID, token.TokenHash, token.Field, token.Weight); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ReplaceSearchTokensForFolder(ctx context.Context, db database.PgExecutor, userID, vaultID, folderID string, tokens []models.FileSearchToken) error {
	if _, err := db.Exec(ctx, `DELETE FROM folder_search_tokens WHERE user_id = $1 AND vault_id = $2 AND folder_id = $3`, userID, vaultID, folderID); err != nil {
		return err
	}
	return r.InsertSearchTokensForFolder(ctx, db, userID, vaultID, folderID, tokens)
}

func (r *Repository) SearchFoldersForTokens(ctx context.Context, db database.PgExecutor, userID, vaultID string, tokenHashes [][]byte, limit int) ([]models.Folder, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := db.Query(ctx, `SELECT
		f.id,
		f.user_id,
		f.vault_id,
		f.parent_folder_id,
		f.encrypted_name,
		f.encrypted_metadata,
		f.created_at,
		f.updated_at,
		f.deleted_at,
		SUM(fst.weight) AS score
	FROM
		folder_search_tokens fst
	INNER JOIN folders f ON f.id = fst.folder_id
	WHERE
		fst.user_id = $1
		AND fst.vault_id = $2
		AND fst.token_hash = ANY($3::bytea[])
		AND f.deleted_at IS NULL
	GROUP BY
		f.id
	HAVING
		COUNT(DISTINCT fst.token_hash) = cardinality($3::bytea[])
	ORDER BY
		score DESC,
		f.updated_at DESC
	LIMIT $4`, userID, vaultID, tokenHashes, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	folders := make([]models.Folder, 0, limit)
	for rows.Next() {
		var folder models.Folder
		var parentFolderID *string
		if err := rows.Scan(
			&folder.ID,
			&folder.UserID,
			&folder.VaultID,
			&parentFolderID,
			&folder.EncryptedName,
			&folder.EncryptedMetadata,
			&folder.CreatedAt,
			&folder.UpdatedAt,
			&folder.DeletedAt,
			&folder.SearchScore,
		); err != nil {
			return nil, err
		}
		folder.ParentFolderID = parentFolderID
		folders = append(folders, folder)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return folders, nil
}
