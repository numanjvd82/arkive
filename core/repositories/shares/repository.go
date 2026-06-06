package sharerepo

import (
	"context"
	"time"

	"arkive/core/database"
	"arkive/core/models"
)

type Repository struct{}

type CreateShareLinkInput struct {
	OwnerUserID       string
	Token             string
	EncryptedShareKey []byte
	CryptoVersion     int16
	PasswordHash      *string
	ExpiresAt         *time.Time
	BurnAfterRead     bool
	MaxAccessCount    *int
	Status            string
}

type CreateShareItemInput struct {
	ShareLinkID  string
	ItemType     string
	FileID       string
	DisplayOrder int
}

type CreateShareSnapshotFileInput struct {
	ShareItemID              string
	FileID                   string
	EncryptedRelativePath    []byte
	EncryptedFileKeyForShare []byte
	DisplayOrder             int
}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) CreateShareLink(ctx context.Context, db database.PgExecutor, input CreateShareLinkInput) (models.ShareLink, error) {
	var created models.ShareLink
	query := `INSERT INTO share_links
		(owner_user_id, token, encrypted_share_key, crypto_version, password_hash, expires_at, burn_after_read, max_access_count, status)
	VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING
		id, owner_user_id, token, slug, status, title_encrypted, description_encrypted,
		encrypted_share_key, crypto_version, password_hash, password_mode, expires_at,
		revoked_at, allow_preview, allow_download, comments_enabled, reactions_enabled,
		burn_after_read, access_count, max_access_count, consumed_at, show_exif, show_location,
		strip_exif_download, created_at, updated_at`
	if err := db.QueryRow(ctx, query,
		input.OwnerUserID,
		input.Token,
		input.EncryptedShareKey,
		input.CryptoVersion,
		input.PasswordHash,
		input.ExpiresAt,
		input.BurnAfterRead,
		input.MaxAccessCount,
		input.Status,
	).Scan(
		&created.ID,
		&created.OwnerUserID,
		&created.Token,
		&created.Slug,
		&created.Status,
		&created.TitleEncrypted,
		&created.DescriptionEncrypted,
		&created.EncryptedShareKey,
		&created.CryptoVersion,
		&created.PasswordHash,
		&created.PasswordMode,
		&created.ExpiresAt,
		&created.RevokedAt,
		&created.AllowPreview,
		&created.AllowDownload,
		&created.CommentsEnabled,
		&created.ReactionsEnabled,
		&created.BurnAfterRead,
		&created.AccessCount,
		&created.MaxAccessCount,
		&created.ConsumedAt,
		&created.ShowEXIF,
		&created.ShowLocation,
		&created.StripEXIFDownload,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return models.ShareLink{}, err
	}
	return created, nil
}

func (r *Repository) CreateShareItem(ctx context.Context, db database.PgExecutor, input CreateShareItemInput) (models.ShareItem, error) {
	var created models.ShareItem
	query := `INSERT INTO share_items
		(share_link_id, item_type, file_id, display_order)
	VALUES
		($1, $2, $3, $4)
	RETURNING
		id, share_link_id, item_type, file_id, display_order, created_at`
	if err := db.QueryRow(ctx, query,
		input.ShareLinkID,
		input.ItemType,
		input.FileID,
		input.DisplayOrder,
	).Scan(
		&created.ID,
		&created.ShareLinkID,
		&created.ItemType,
		&created.FileID,
		&created.DisplayOrder,
		&created.CreatedAt,
	); err != nil {
		return models.ShareItem{}, err
	}
	return created, nil
}

func (r *Repository) CreateShareSnapshotFile(ctx context.Context, db database.PgExecutor, input CreateShareSnapshotFileInput) (models.ShareSnapshotFile, error) {
	var created models.ShareSnapshotFile
	query := `INSERT INTO share_snapshot_files
		(share_item_id, file_id, encrypted_relative_path, encrypted_file_key_for_share, display_order)
	VALUES
		($1, $2, $3, $4, $5)
	RETURNING
		id, share_item_id, file_id, encrypted_relative_path, encrypted_file_key_for_share, display_order, created_at`
	if err := db.QueryRow(ctx, query,
		input.ShareItemID,
		input.FileID,
		input.EncryptedRelativePath,
		input.EncryptedFileKeyForShare,
		input.DisplayOrder,
	).Scan(
		&created.ID,
		&created.ShareItemID,
		&created.FileID,
		&created.EncryptedRelativePath,
		&created.EncryptedFileKeyForShare,
		&created.DisplayOrder,
		&created.CreatedAt,
	); err != nil {
		return models.ShareSnapshotFile{}, err
	}
	return created, nil
}

func (r *Repository) GetShareByToken(ctx context.Context, db database.PgExecutor, token string) (models.Share, error) {
	var share models.Share
	query := `SELECT
		sl.id, ssf.file_id, sl.owner_user_id, sl.token, sl.encrypted_share_key, sl.allow_preview, sl.allow_download, sl.burn_after_read,
		sl.access_count, sl.max_access_count, sl.password_hash, sl.expires_at, sl.status, sl.revoked_at, sl.consumed_at, sl.created_at, sl.updated_at
	FROM
		share_links sl
	JOIN
		share_items si ON si.share_link_id = sl.id
	JOIN
		share_snapshot_files ssf ON ssf.share_item_id = si.id
	JOIN
		files f ON f.id = ssf.file_id
	WHERE
		sl.token = $1
		AND f.deleted_at IS NULL
	ORDER BY
		ssf.display_order ASC
	LIMIT 1`
	if err := db.QueryRow(ctx, query, token).Scan(
		&share.ID,
		&share.FileID,
		&share.OwnerUserID,
		&share.Token,
		&share.EncryptedShareKey,
		&share.AllowPreview,
		&share.AllowDownload,
		&share.BurnAfterRead,
		&share.AccessCount,
		&share.MaxAccessCount,
		&share.PasswordHash,
		&share.ExpiresAt,
		&share.Status,
		&share.RevokedAt,
		&share.ConsumedAt,
		&share.CreatedAt,
		&share.UpdatedAt,
	); err != nil {
		return models.Share{}, err
	}
	return share, nil
}

func (r *Repository) GetShareForFile(ctx context.Context, db database.PgExecutor, fileID string) (models.Share, error) {
	var share models.Share
	query := `SELECT
		sl.id, ssf.file_id, sl.owner_user_id, sl.token, sl.encrypted_share_key, sl.allow_preview, sl.allow_download, sl.burn_after_read,
		sl.access_count, sl.max_access_count, sl.password_hash, sl.expires_at, sl.status, sl.revoked_at, sl.consumed_at, sl.created_at, sl.updated_at
	FROM
		share_links sl
	JOIN
		share_items si ON si.share_link_id = sl.id
	JOIN
		share_snapshot_files ssf ON ssf.share_item_id = si.id
	JOIN
		files f ON f.id = ssf.file_id
	WHERE
		ssf.file_id = $1
		AND f.deleted_at IS NULL
	ORDER BY
		ssf.display_order ASC
	LIMIT 1`
	if err := db.QueryRow(ctx, query, fileID).Scan(
		&share.ID,
		&share.FileID,
		&share.OwnerUserID,
		&share.Token,
		&share.EncryptedShareKey,
		&share.AllowPreview,
		&share.AllowDownload,
		&share.BurnAfterRead,
		&share.AccessCount,
		&share.MaxAccessCount,
		&share.PasswordHash,
		&share.ExpiresAt,
		&share.Status,
		&share.RevokedAt,
		&share.ConsumedAt,
		&share.CreatedAt,
		&share.UpdatedAt,
	); err != nil {
		return models.Share{}, err
	}
	return share, nil
}

func (r *Repository) GetShareForFileForUser(ctx context.Context, db database.PgExecutor, fileID, ownerUserID string) (models.Share, error) {
	var share models.Share
	query := `SELECT
		sl.id, ssf.file_id, sl.owner_user_id, sl.token, sl.encrypted_share_key, sl.allow_preview, sl.allow_download, sl.burn_after_read,
		sl.access_count, sl.max_access_count, sl.password_hash, sl.expires_at, sl.status, sl.revoked_at, sl.consumed_at, sl.created_at, sl.updated_at
	FROM
		share_links sl
	JOIN
		share_items si ON si.share_link_id = sl.id
	JOIN
		share_snapshot_files ssf ON ssf.share_item_id = si.id
	JOIN
		files f ON f.id = ssf.file_id
	WHERE
		ssf.file_id = $1 AND sl.owner_user_id = $2
		AND f.deleted_at IS NULL
	ORDER BY
		ssf.display_order ASC
	LIMIT 1`
	if err := db.QueryRow(ctx, query, fileID, ownerUserID).Scan(
		&share.ID,
		&share.FileID,
		&share.OwnerUserID,
		&share.Token,
		&share.EncryptedShareKey,
		&share.AllowPreview,
		&share.AllowDownload,
		&share.BurnAfterRead,
		&share.AccessCount,
		&share.MaxAccessCount,
		&share.PasswordHash,
		&share.ExpiresAt,
		&share.Status,
		&share.RevokedAt,
		&share.ConsumedAt,
		&share.CreatedAt,
		&share.UpdatedAt,
	); err != nil {
		return models.Share{}, err
	}
	return share, nil
}

func (r *Repository) GetShareForUser(ctx context.Context, db database.PgExecutor, shareID, ownerUserID string) (models.Share, error) {
	var share models.Share
	query := `SELECT
		sl.id, ssf.file_id, sl.owner_user_id, sl.token, sl.encrypted_share_key, sl.allow_preview, sl.allow_download, sl.burn_after_read,
		sl.access_count, sl.max_access_count, sl.password_hash, sl.expires_at, sl.status, sl.revoked_at, sl.consumed_at, sl.created_at, sl.updated_at
	FROM
		share_links sl
	JOIN
		share_items si ON si.share_link_id = sl.id
	JOIN
		share_snapshot_files ssf ON ssf.share_item_id = si.id
	JOIN
		files f ON f.id = ssf.file_id
	WHERE
		sl.id = $1 AND sl.owner_user_id = $2
		AND f.deleted_at IS NULL
	ORDER BY
		ssf.display_order ASC
	LIMIT 1`
	if err := db.QueryRow(ctx, query, shareID, ownerUserID).Scan(
		&share.ID,
		&share.FileID,
		&share.OwnerUserID,
		&share.Token,
		&share.EncryptedShareKey,
		&share.AllowPreview,
		&share.AllowDownload,
		&share.BurnAfterRead,
		&share.AccessCount,
		&share.MaxAccessCount,
		&share.PasswordHash,
		&share.ExpiresAt,
		&share.Status,
		&share.RevokedAt,
		&share.ConsumedAt,
		&share.CreatedAt,
		&share.UpdatedAt,
	); err != nil {
		return models.Share{}, err
	}
	return share, nil
}

func (r *Repository) UpdateShareForUser(ctx context.Context, db database.PgExecutor, shareID, ownerUserID string, passwordHash *string, expiresAt *time.Time, burnAfterRead bool, maxAccessCount *int) (models.Share, error) {
	var share models.Share
	query := `UPDATE
		share_links
	SET
		password_hash = $3,
		expires_at = $4,
		burn_after_read = $5,
		max_access_count = $6,
		updated_at = now()
	WHERE
		id = $1 AND owner_user_id = $2 AND status = 'active'
	RETURNING
		id, owner_user_id, token, encrypted_share_key, allow_preview, allow_download, burn_after_read,
		access_count, max_access_count, password_hash, expires_at, status, revoked_at, consumed_at, created_at, updated_at`
	var shareIDValue string
	if err := db.QueryRow(ctx, query, shareID, ownerUserID, passwordHash, expiresAt, burnAfterRead, maxAccessCount).Scan(
		&shareIDValue,
		&share.OwnerUserID,
		&share.Token,
		&share.EncryptedShareKey,
		&share.AllowPreview,
		&share.AllowDownload,
		&share.BurnAfterRead,
		&share.AccessCount,
		&share.MaxAccessCount,
		&share.PasswordHash,
		&share.ExpiresAt,
		&share.Status,
		&share.RevokedAt,
		&share.ConsumedAt,
		&share.CreatedAt,
		&share.UpdatedAt,
	); err != nil {
		return models.Share{}, err
	}
	share.ID = shareIDValue
	linked, err := r.GetShareForUser(ctx, db, shareIDValue, ownerUserID)
	if err != nil {
		return models.Share{}, err
	}
	share.FileID = linked.FileID
	return share, nil
}

func (r *Repository) DeleteShareForUser(ctx context.Context, db database.PgExecutor, shareID, ownerUserID string) (bool, error) {
	query := `DELETE FROM
		share_links
	WHERE
		id = $1 AND owner_user_id = $2`
	tag, err := db.Exec(ctx, query, shareID, ownerUserID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *Repository) RevokeShareForUser(ctx context.Context, db database.PgExecutor, shareID, ownerUserID string) (models.Share, error) {
	var share models.Share
	query := `UPDATE
		share_links
	SET
		status = 'revoked',
		revoked_at = now(),
		updated_at = now()
	WHERE
		id = $1 AND owner_user_id = $2 AND status = 'active'
	RETURNING
		id, owner_user_id, token, encrypted_share_key, allow_preview, allow_download, burn_after_read,
		access_count, max_access_count, password_hash, expires_at, status, revoked_at, consumed_at, created_at, updated_at`
	var shareIDValue string
	if err := db.QueryRow(ctx, query, shareID, ownerUserID).Scan(
		&shareIDValue,
		&share.OwnerUserID,
		&share.Token,
		&share.EncryptedShareKey,
		&share.AllowPreview,
		&share.AllowDownload,
		&share.BurnAfterRead,
		&share.AccessCount,
		&share.MaxAccessCount,
		&share.PasswordHash,
		&share.ExpiresAt,
		&share.Status,
		&share.RevokedAt,
		&share.ConsumedAt,
		&share.CreatedAt,
		&share.UpdatedAt,
	); err != nil {
		return models.Share{}, err
	}
	share.ID = shareIDValue
	linked, err := r.GetShareForUser(ctx, db, shareIDValue, ownerUserID)
	if err != nil {
		return models.Share{}, err
	}
	share.FileID = linked.FileID
	return share, nil
}

func (r *Repository) ActivateShareForUser(ctx context.Context, db database.PgExecutor, shareID, ownerUserID string) (models.Share, error) {
	var share models.Share
	query := `UPDATE
		share_links
	SET
		status = 'active',
		revoked_at = NULL,
		updated_at = now()
	WHERE
		id = $1 AND owner_user_id = $2 AND status = 'revoked'
	RETURNING
		id, owner_user_id, token, encrypted_share_key, allow_preview, allow_download, burn_after_read,
		access_count, max_access_count, password_hash, expires_at, status, revoked_at, consumed_at, created_at, updated_at`
	var shareIDValue string
	if err := db.QueryRow(ctx, query, shareID, ownerUserID).Scan(
		&shareIDValue,
		&share.OwnerUserID,
		&share.Token,
		&share.EncryptedShareKey,
		&share.AllowPreview,
		&share.AllowDownload,
		&share.BurnAfterRead,
		&share.AccessCount,
		&share.MaxAccessCount,
		&share.PasswordHash,
		&share.ExpiresAt,
		&share.Status,
		&share.RevokedAt,
		&share.ConsumedAt,
		&share.CreatedAt,
		&share.UpdatedAt,
	); err != nil {
		return models.Share{}, err
	}
	share.ID = shareIDValue
	linked, err := r.GetShareForUser(ctx, db, shareIDValue, ownerUserID)
	if err != nil {
		return models.Share{}, err
	}
	share.FileID = linked.FileID
	return share, nil
}

func (r *Repository) ConsumeShareByToken(ctx context.Context, db database.PgExecutor, token string) (models.Share, bool, error) {
	var share models.Share
	query := `UPDATE share_links
	SET
	  access_count = access_count + 1,
	  consumed_at = CASE
	    WHEN access_count + 1 >= max_access_count THEN now()
	    ELSE consumed_at
	  END,
	  status = CASE
	    WHEN access_count + 1 >= max_access_count THEN 'burned'
	    ELSE status
	  END,
	  updated_at = now()
	WHERE token = $1
	  AND status = 'active'
	  AND (expires_at IS NULL OR expires_at > now())
	  AND burn_after_read = true
	  AND max_access_count IS NOT NULL
	  AND access_count < max_access_count
	  AND EXISTS (
		SELECT 1
		FROM share_items si
		JOIN share_snapshot_files ssf ON ssf.share_item_id = si.id
		JOIN files f ON f.id = ssf.file_id
		WHERE si.share_link_id = share_links.id
		  AND f.deleted_at IS NULL
	  )
	RETURNING
	  id, owner_user_id, token, encrypted_share_key, allow_preview, allow_download, burn_after_read,
	  access_count, max_access_count, password_hash, expires_at, status, revoked_at, consumed_at, created_at, updated_at`
	err := db.QueryRow(ctx, query, token).Scan(
		&share.ID,
		&share.OwnerUserID,
		&share.Token,
		&share.EncryptedShareKey,
		&share.AllowPreview,
		&share.AllowDownload,
		&share.BurnAfterRead,
		&share.AccessCount,
		&share.MaxAccessCount,
		&share.PasswordHash,
		&share.ExpiresAt,
		&share.Status,
		&share.RevokedAt,
		&share.ConsumedAt,
		&share.CreatedAt,
		&share.UpdatedAt,
	)
	if err != nil {
		return models.Share{}, false, err
	}
	return share, true, nil
}

func (r *Repository) ListSharesForUser(ctx context.Context, db database.PgExecutor, ownerUserID string) ([]models.ShareWithFile, error) {
	rows, err := db.Query(ctx, `SELECT
		sl.id,
		ssf.file_id,
		sl.owner_user_id,
		sl.token,
		sl.encrypted_share_key,
		sl.allow_preview,
		sl.allow_download,
		sl.burn_after_read,
		sl.access_count,
		sl.max_access_count,
		sl.password_hash,
		sl.expires_at,
		sl.status,
		sl.revoked_at,
		sl.consumed_at,
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
		AND f.deleted_at IS NULL
	ORDER BY
		sl.created_at DESC`, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	shares := []models.ShareWithFile{}
	for rows.Next() {
		var share models.ShareWithFile
		if err := rows.Scan(
			&share.ID,
			&share.FileID,
			&share.OwnerUserID,
			&share.Token,
			&share.EncryptedShareKey,
			&share.AllowPreview,
			&share.AllowDownload,
			&share.BurnAfterRead,
			&share.AccessCount,
			&share.MaxAccessCount,
			&share.PasswordHash,
			&share.ExpiresAt,
			&share.Status,
			&share.RevokedAt,
			&share.ConsumedAt,
			&share.CreatedAt,
			&share.UpdatedAt,
			&share.FileName,
			&share.FileContentType,
			&share.FileSizeBytes,
			&share.FileUpdatedAt,
		); err != nil {
			return nil, err
		}
		shares = append(shares, share)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return shares, nil
}

func (r *Repository) GetPublicShareRecord(ctx context.Context, db database.PgExecutor, token string) (models.PublicShareRecord, error) {
	var record models.PublicShareRecord
	query := `SELECT
		sl.id,
		sl.token,
		ssf.file_id,
		f.user_id,
		ssf.encrypted_file_key_for_share,
		f.encrypted_metadata,
		f.encrypted_manifest,
		f.encrypted_hash,
		f.encryption_version,
		f.chunk_size,
		f.chunk_count,
		f.plaintext_size
	FROM
		share_links sl
	JOIN
		share_items si ON si.share_link_id = sl.id
	JOIN
		share_snapshot_files ssf ON ssf.share_item_id = si.id
	JOIN
		files f ON f.id = ssf.file_id
	WHERE
		sl.token = $1
		AND f.deleted_at IS NULL
	ORDER BY
		ssf.display_order ASC
	LIMIT 1`
	if err := db.QueryRow(ctx, query, token).Scan(
		&record.ShareID,
		&record.Token,
		&record.FileID,
		&record.VaultID,
		&record.EncryptedFileKeyForShare,
		&record.EncryptedMetadata,
		&record.EncryptedManifest,
		&record.EncryptedHash,
		&record.EncryptionVersion,
		&record.ChunkSize,
		&record.TotalChunks,
		&record.PlaintextSize,
	); err != nil {
		return models.PublicShareRecord{}, err
	}
	return record, nil
}
