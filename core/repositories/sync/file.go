package syncrepo

import (
	"context"

	"arkive/core/database"
	"arkive/core/models"
)

type Repository struct{}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) FolderExistsForUser(ctx context.Context, db database.PgExecutor, userID, folderID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(
		SELECT 1
		FROM folders
		WHERE id = $1
			AND user_id = $2
	)`
	if err := db.QueryRow(ctx, query, folderID, userID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) ListEntriesPage(ctx context.Context, db database.PgExecutor, input models.ListEntriesPageInput) ([]models.SyncEntry, error) {
	cursorUpdatedAt := any(nil)
	cursorTypeOrder := any(nil)
	cursorID := any(nil)

	if input.Cursor != nil {
		cursorUpdatedAt = input.Cursor.UpdatedAt
		cursorID = input.Cursor.ID
		switch input.Cursor.Type {
		case "folder":
			cursorTypeOrder = 0
		case "file":
			cursorTypeOrder = 1
		default:
			cursorTypeOrder = nil
		}
	}

	query := `WITH entries AS (
		SELECT
			0 AS entry_type_order,
			'folder' AS entry_type,
			id,
			NULL::uuid AS folder_id,
			parent_folder_id,
			encrypted_metadata,
			NULL::bytea AS encrypted_file_key,
			NULL::bytea AS encrypted_manifest,
			encrypted_name,
			updated_at,
			deleted_at,
			NULL::timestamptz AS purged_at
		FROM folders
		WHERE user_id = $1
			AND (
				($2::uuid IS NULL AND parent_folder_id IS NULL)
				OR parent_folder_id = $2
			)
			AND ($3::boolean OR deleted_at IS NULL)

		UNION ALL

		SELECT
			1 AS entry_type_order,
			'file' AS entry_type,
			id,
			folder_id,
			NULL::uuid AS parent_folder_id,
			encrypted_metadata,
			encrypted_file_key,
			encrypted_manifest,
			NULL::bytea AS encrypted_name,
			updated_at,
			deleted_at,
			purged_at
		FROM files
		WHERE user_id = $1
			AND upload_status = 'complete'
			AND expires_at IS NULL
			AND (
				($2::uuid IS NULL AND folder_id IS NULL)
				OR folder_id = $2
			)
			AND ($3::boolean OR deleted_at IS NULL)
	)
	SELECT
		entry_type,
		id::text,
		folder_id::text,
		parent_folder_id::text,
		encrypted_metadata,
		encrypted_file_key,
		encrypted_manifest,
		encrypted_name,
		updated_at,
		deleted_at,
		purged_at
	FROM entries
	WHERE
		$4::timestamptz IS NULL
		OR updated_at < $4
		OR (updated_at = $4 AND entry_type_order > $5)
		OR (updated_at = $4 AND entry_type_order = $5 AND id > $6::uuid)
	ORDER BY updated_at DESC, entry_type_order ASC, id ASC
	LIMIT $7`

	rows, err := db.Query(ctx, query,
		input.UserID,
		input.FolderID,
		input.IncludeDeleted,
		cursorUpdatedAt,
		cursorTypeOrder,
		cursorID,
		input.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]models.SyncEntry, 0)
	for rows.Next() {
		var entry models.SyncEntry
		if err := rows.Scan(
			&entry.Type,
			&entry.ID,
			&entry.FolderID,
			&entry.ParentFolderID,
			&entry.EncryptedMetadata,
			&entry.EncryptedFileKey,
			&entry.EncryptedManifest,
			&entry.EncryptedName,
			&entry.UpdatedAt,
			&entry.DeletedAt,
			&entry.PurgedAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return entries, nil
}

func (r *Repository) ListFilesByFolder(ctx context.Context, db database.PgExecutor, userID string, folderID *string, includeDeleted bool) ([]models.File, error) {
	query := `SELECT
		id,
		user_id,
		folder_id,
		encrypted_metadata,
		encrypted_file_key,
		encrypted_manifest,
		updated_at,
		deleted_at,
		purged_at
	FROM
		files
	WHERE
		user_id = $1
		AND upload_status = 'complete'
		AND expires_at IS NULL
		AND (
			($2::uuid IS NULL AND folder_id IS NULL)
			OR folder_id = $2
		)`
	if !includeDeleted {
		query += `
		AND deleted_at IS NULL`
	}
	query += `
	ORDER BY updated_at DESC, id ASC`

	rows, err := db.Query(ctx, query, userID, folderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make([]models.File, 0)
	for rows.Next() {
		var file models.File
		var scannedFolderID *string
		if err := rows.Scan(
			&file.ID,
			&file.UserID,
			&scannedFolderID,
			&file.EncryptedMetadata,
			&file.EncryptedFileKey,
			&file.EncryptedManifest,
			&file.UpdatedAt,
			&file.DeletedAt,
			&file.PurgedAt,
		); err != nil {
			return nil, err
		}
		file.FolderID = scannedFolderID
		files = append(files, file)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return files, nil
}

func (r *Repository) ListFoldersByParent(ctx context.Context, db database.PgExecutor, userID string, parentFolderID *string, includeDeleted bool) ([]models.Folder, error) {
	query := `SELECT
		id,
		user_id,
		vault_id,
		parent_folder_id,
		encrypted_name,
		encrypted_metadata,
		updated_at,
		deleted_at
	FROM
		folders
	WHERE
		user_id = $1
		AND (
			($2::uuid IS NULL AND parent_folder_id IS NULL)
			OR parent_folder_id = $2
		)`
	if !includeDeleted {
		query += `
		AND deleted_at IS NULL`
	}
	query += `
	ORDER BY updated_at DESC, id ASC`

	rows, err := db.Query(ctx, query, userID, parentFolderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	folders := make([]models.Folder, 0)
	for rows.Next() {
		var folder models.Folder
		var scannedParentFolderID *string
		if err := rows.Scan(
			&folder.ID,
			&folder.UserID,
			&folder.VaultID,
			&scannedParentFolderID,
			&folder.EncryptedName,
			&folder.EncryptedMetadata,
			&folder.UpdatedAt,
			&folder.DeletedAt,
		); err != nil {
			return nil, err
		}
		folder.ParentFolderID = scannedParentFolderID
		folders = append(folders, folder)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return folders, nil
}
