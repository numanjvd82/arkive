package sharerepo

import (
	"context"
	"time"

	"arkive/core/database"
	"arkive/core/models"
)

type Repository struct{}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) CreateShare(ctx context.Context, db database.PgExecutor, share models.Share) (models.Share, error) {
	var created models.Share
	query := `INSERT INTO shares
		(file_id, owner_user_id, token, password_hash, expires_at, status)
	VALUES
		($1, $2, $3, $4, $5, $6)
	RETURNING
		id, file_id, owner_user_id, token, password_hash, expires_at,
		status, revoked_at, created_at, updated_at`
	if err := db.QueryRow(ctx, query,
		share.FileID,
		share.OwnerUserID,
		share.Token,
		share.PasswordHash,
		share.ExpiresAt,
		share.Status,
	).Scan(
		&created.ID,
		&created.FileID,
		&created.OwnerUserID,
		&created.Token,
		&created.PasswordHash,
		&created.ExpiresAt,
		&created.Status,
		&created.RevokedAt,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return models.Share{}, err
	}
	return created, nil
}

func (r *Repository) GetShareByToken(ctx context.Context, db database.PgExecutor, token string) (models.Share, error) {
	var share models.Share
	query := `SELECT
		id, file_id, owner_user_id, token, password_hash, expires_at,
		status, revoked_at, created_at, updated_at
	FROM
		shares
	WHERE
		token = $1`
	if err := db.QueryRow(ctx, query, token).Scan(
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
	); err != nil {
		return models.Share{}, err
	}
	return share, nil
}

func (r *Repository) GetShareForFile(ctx context.Context, db database.PgExecutor, fileID string) (models.Share, error) {
	var share models.Share
	query := `SELECT
		id, file_id, owner_user_id, token, password_hash, expires_at,
		status, revoked_at, created_at, updated_at
	FROM
		shares
	WHERE
		file_id = $1`
	if err := db.QueryRow(ctx, query, fileID).Scan(
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
	); err != nil {
		return models.Share{}, err
	}
	return share, nil
}

func (r *Repository) GetShareForFileForUser(ctx context.Context, db database.PgExecutor, fileID, ownerUserID string) (models.Share, error) {
	var share models.Share
	query := `SELECT
		id, file_id, owner_user_id, token, password_hash, expires_at,
		status, revoked_at, created_at, updated_at
	FROM
		shares
	WHERE
		file_id = $1 AND owner_user_id = $2`
	if err := db.QueryRow(ctx, query, fileID, ownerUserID).Scan(
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
	); err != nil {
		return models.Share{}, err
	}
	return share, nil
}

func (r *Repository) RevokeShareForUser(ctx context.Context, db database.PgExecutor, shareID, ownerUserID string) (bool, error) {
	query := `UPDATE
		shares
	SET
		status = 'revoked',
		revoked_at = now(),
		updated_at = now()
	WHERE
		id = $1 AND owner_user_id = $2 AND status = 'active'`
	tag, err := db.Exec(ctx, query, shareID, ownerUserID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *Repository) UpdateShareForUser(ctx context.Context, db database.PgExecutor, shareID, ownerUserID string, passwordHash *string, expiresAt *time.Time) (models.Share, error) {
	var share models.Share
	query := `UPDATE
		shares
	SET
		password_hash = $3,
		expires_at = $4,
		updated_at = now()
	WHERE
		id = $1 AND owner_user_id = $2 AND status = 'active'
	RETURNING
		id, file_id, owner_user_id, token, password_hash, expires_at,
		status, revoked_at, created_at, updated_at`
	if err := db.QueryRow(ctx, query, shareID, ownerUserID, passwordHash, expiresAt).Scan(
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
	); err != nil {
		return models.Share{}, err
	}
	return share, nil
}

func (r *Repository) DeleteShareForUser(ctx context.Context, db database.PgExecutor, shareID, ownerUserID string) (bool, error) {
	query := `DELETE FROM
		shares
	WHERE
		id = $1 AND owner_user_id = $2`
	tag, err := db.Exec(ctx, query, shareID, ownerUserID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
