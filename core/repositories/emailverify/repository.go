package emailverify

import (
	"context"
	"time"

	"arkive/core/database"
)

type Repository struct{}

func New() *Repository { return &Repository{} }

func (r *Repository) CreateToken(ctx context.Context, db database.PgExecutor, userID string, tokenHash []byte, expiresAt time.Time) error {
	query := `INSERT INTO email_verification_tokens
		(user_id, token_hash, expires_at)
	VALUES
		($1, $2, $3)`
	_, err := db.Exec(ctx, query, userID, tokenHash, expiresAt)
	return err
}

func (r *Repository) ConsumeToken(ctx context.Context, db database.PgExecutor, tokenHash []byte, now time.Time) (string, error) {
	var userID string
	query := `UPDATE
		email_verification_tokens
	SET
		used_at = $2
	WHERE
		token_hash = $1
		AND used_at IS NULL
		AND expires_at > $2
	RETURNING
		user_id`
	if err := db.QueryRow(ctx, query, tokenHash, now).Scan(&userID); err != nil {
		return "", err
	}
	return userID, nil
}

func (r *Repository) DeleteByHash(ctx context.Context, db database.PgExecutor, tokenHash []byte) error {
	query := `DELETE FROM
		email_verification_tokens
	WHERE
		token_hash = $1`
	_, err := db.Exec(ctx, query, tokenHash)
	return err
}

func (r *Repository) DeleteIfExpiredOrUsed(ctx context.Context, db database.PgExecutor, tokenHash []byte, now time.Time) error {
	query := `DELETE FROM
		email_verification_tokens
	WHERE
		token_hash = $1
		AND (used_at IS NOT NULL OR expires_at <= $2)`
	_, err := db.Exec(ctx, query, tokenHash, now)
	return err
}

func (r *Repository) MarkEmailVerified(ctx context.Context, db database.PgExecutor, userID string) error {
	query := `UPDATE
		users
	SET
		is_email_verified = true,
		updated_at = now()
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, userID)
	return err
}
