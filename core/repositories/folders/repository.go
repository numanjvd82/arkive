package foldersrepo

import (
	"context"

	"arkive/core/database"
	"arkive/core/models"
)

type Repository struct{}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) Create(ctx context.Context, db database.PgExecutor, folder models.Folder) (models.Folder, error) {
	var created models.Folder
	var parentFolderID *string
	query := `INSERT INTO folders
		(user_id, vault_id, parent_folder_id, encrypted_name, encrypted_metadata)
	VALUES
		($1, $2, $3, $4, $5)
	RETURNING
		id, user_id, vault_id, parent_folder_id, encrypted_name, encrypted_metadata, created_at, updated_at, deleted_at`
	if err := db.QueryRow(
		ctx,
		query,
		folder.UserID,
		folder.VaultID,
		folder.ParentFolderID,
		folder.EncryptedName,
		folder.EncryptedMetadata,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.VaultID,
		&parentFolderID,
		&created.EncryptedName,
		&created.EncryptedMetadata,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.DeletedAt,
	); err != nil {
		return models.Folder{}, err
	}
	created.ParentFolderID = parentFolderID
	return created, nil
}

func (r *Repository) GetForUser(ctx context.Context, db database.PgExecutor, userID, folderID string) (models.Folder, error) {
	var folder models.Folder
	var parentFolderID *string
	query := `SELECT
		id, user_id, vault_id, parent_folder_id, encrypted_name, encrypted_metadata, created_at, updated_at, deleted_at
	FROM
		folders
	WHERE
		id = $1
		AND user_id = $2
		AND deleted_at IS NULL`
	if err := db.QueryRow(ctx, query, folderID, userID).Scan(
		&folder.ID,
		&folder.UserID,
		&folder.VaultID,
		&parentFolderID,
		&folder.EncryptedName,
		&folder.EncryptedMetadata,
		&folder.CreatedAt,
		&folder.UpdatedAt,
		&folder.DeletedAt,
	); err != nil {
		return models.Folder{}, err
	}
	folder.ParentFolderID = parentFolderID
	return folder, nil
}

func (r *Repository) CountChildFolders(ctx context.Context, db database.PgExecutor, userID string, parentID *string) (int, error) {
	var count int
	query := `SELECT
		COUNT(1)
	FROM
		folders
	WHERE
		user_id = $1
		AND deleted_at IS NULL
		AND (
			($2::uuid IS NULL AND parent_folder_id IS NULL)
			OR parent_folder_id = $2
		)`
	if err := db.QueryRow(ctx, query, userID, parentID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) ListChildFolders(ctx context.Context, db database.PgExecutor, userID string, parentID *string, limit, offset int) ([]models.Folder, error) {
	query := `SELECT
		id, user_id, vault_id, parent_folder_id, encrypted_name, encrypted_metadata, created_at, updated_at, deleted_at
	FROM
		folders
	WHERE
		user_id = $1
		AND deleted_at IS NULL
		AND (
			($2::uuid IS NULL AND parent_folder_id IS NULL)
			OR parent_folder_id = $2
		)
	ORDER BY created_at DESC
	LIMIT $3 OFFSET $4`
	rows, err := db.Query(ctx, query, userID, parentID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	folders := []models.Folder{}
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
		); err != nil {
			return nil, err
		}
		folder.ParentFolderID = parentFolderID
		folders = append(folders, folder)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return folders, nil
}

func (r *Repository) MoveFolders(ctx context.Context, db database.PgExecutor, userID string, folderIDs []string, targetFolderID *string) (int64, error) {
	query := `UPDATE folders
	SET
		parent_folder_id = $3,
		updated_at = now()
	WHERE
		user_id = $1
		AND id = ANY($2::uuid[])
		AND deleted_at IS NULL`
	tag, err := db.Exec(ctx, query, userID, folderIDs, targetFolderID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *Repository) TargetIsDescendant(ctx context.Context, db database.PgExecutor, userID string, movingFolderIDs []string, targetFolderID string) (bool, error) {
	var exists bool
	query := `WITH RECURSIVE descendants AS (
		SELECT id
		FROM folders
		WHERE user_id = $1
			AND deleted_at IS NULL
			AND id = ANY($2::uuid[])
		UNION ALL
		SELECT f.id
		FROM folders f
		INNER JOIN descendants d ON f.parent_folder_id = d.id
		WHERE f.user_id = $1
			AND f.deleted_at IS NULL
	)
	SELECT EXISTS(
		SELECT 1
		FROM descendants
		WHERE id = $3::uuid
	)`
	if err := db.QueryRow(ctx, query, userID, movingFolderIDs, targetFolderID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
