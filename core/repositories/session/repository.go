package sessionrepo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"arkive/core/database"
)

type Repository struct {
}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) CreateSession(ctx context.Context, db database.PgExecutor, userID string, expiresAt time.Time) (string, error) {
	var sessionID string
	query := `INSERT INTO sessions (user_id, expires_at)
		VALUES ($1, $2)
		RETURNING id`
	if err := db.QueryRow(ctx, query, userID, expiresAt).Scan(&sessionID); err != nil {
		return "", err
	}
	return sessionID, nil
}

func (r *Repository) DeleteSession(ctx context.Context, db database.PgExecutor, sessionID string) error {
	query := `DELETE FROM sessions
		WHERE id = $1`
	_, err := db.Exec(ctx, query, sessionID)
	return err
}

func (r *Repository) GetSessionByID(ctx context.Context, db database.PgExecutor, sessionID string) (string, time.Time, error) {
	var userID string
	var expiresAt time.Time
	query := `SELECT user_id, expires_at
		FROM sessions
		WHERE id = $1`
	if err := db.QueryRow(ctx, query, sessionID).Scan(&userID, &expiresAt); err != nil {
		return "", time.Time{}, err
	}
	if expiresAt.Before(time.Now()) {
		return "", time.Time{}, pgx.ErrNoRows
	}
	return userID, expiresAt, nil
}

func (r *Repository) CreateRefreshToken(ctx context.Context, db database.PgExecutor, userID string, hash []byte, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`
	_, err := db.Exec(ctx, query, userID, hash, expiresAt)
	return err
}

func (r *Repository) GetRefreshToken(ctx context.Context, db database.PgExecutor, hash []byte) (string, string, time.Time, *time.Time, error) {
	var id string
	var userID string
	var expiresAt time.Time
	var revokedAt *time.Time
	query := `SELECT id, user_id, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1`
	if err := db.QueryRow(ctx, query, hash).Scan(&id, &userID, &expiresAt, &revokedAt); err != nil {
		return "", "", time.Time{}, nil, err
	}
	return id, userID, expiresAt, revokedAt, nil
}

func (r *Repository) RevokeRefreshToken(ctx context.Context, db database.PgExecutor, id string) error {
	_, err := db.Exec(ctx, `UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE id = $1`, id)
	return err
}

func (r *Repository) RevokeRefreshTokenByHash(ctx context.Context, db database.PgExecutor, hash []byte) (bool, error) {
	tag, err := db.Exec(ctx, `UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE token_hash = $1 AND revoked_at IS NULL`, hash)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
