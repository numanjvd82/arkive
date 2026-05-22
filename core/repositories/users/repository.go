package usersrepo

import (
	"context"
	"time"

	"arkive/core/database"
)

type Repository struct{}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) CountUsers(ctx context.Context, db database.PgExecutor) (int64, error) {
	query := `SELECT COUNT(*) FROM users`
	var count int64
	if err := db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) GetStorageUsage(ctx context.Context, db database.PgExecutor, userID string) (int64, int64, error) {
	query := `SELECT used_bytes, reserved_bytes FROM users WHERE id = $1`
	var usedBytes int64
	var reservedBytes int64
	if err := db.QueryRow(ctx, query, userID).Scan(&usedBytes, &reservedBytes); err != nil {
		return 0, 0, err
	}
	return usedBytes, reservedBytes, nil
}

func (r *Repository) UpdateLoginActivity(ctx context.Context, db database.PgExecutor, userID string, loginAt time.Time) error {
	query := `UPDATE
		users
	SET
		last_login_at = $2,
		last_active_at = $2,
		updated_at = now()
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, userID, loginAt)
	return err
}

func (r *Repository) TouchUserActivity(ctx context.Context, db database.PgExecutor, userID string, activeAt time.Time) error {
	query := `UPDATE
		users
	SET
		last_active_at = $2,
		updated_at = now()
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, userID, activeAt)
	return err
}

func (r *Repository) SetRecoverySetupToken(ctx context.Context, db database.PgExecutor, userID, token string, expiresAt time.Time) error {
	query := `UPDATE
		users
	SET
		recovery_setup_token = $2,
		recovery_setup_token_expires_at = $3,
		updated_at = now()
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, userID, token, expiresAt)
	return err
}

func (r *Repository) HasValidRecoverySetupToken(ctx context.Context, db database.PgExecutor, token string, now time.Time) (bool, error) {
	query := `SELECT EXISTS (
		SELECT 1
		FROM users
		WHERE recovery_setup_token = $1
			AND recovery_setup_token_expires_at IS NOT NULL
			AND recovery_setup_token_expires_at > $2
	)`
	var ok bool
	if err := db.QueryRow(ctx, query, token, now).Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}

func (r *Repository) ClearRecoverySetupToken(ctx context.Context, db database.PgExecutor, token string) error {
	query := `UPDATE
		users
	SET
		recovery_setup_token = NULL,
		recovery_setup_token_expires_at = NULL,
		updated_at = now()
	WHERE
		recovery_setup_token = $1`
	_, err := db.Exec(ctx, query, token)
	return err
}
