package sessionrepo

import (
	"context"
	"time"

	"arkive/core/database"
)

type Repository struct {
}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) CreateSession(ctx context.Context, db database.PgExecutor, userID string, expiresAt time.Time) (string, error) {
	var sessionID string
	query := `INSERT INTO sessions
		(user_id, expires_at)
	VALUES
		($1, $2)
	RETURNING
		id`
	if err := db.QueryRow(ctx, query, userID, expiresAt).Scan(&sessionID); err != nil {
		return "", err
	}
	return sessionID, nil
}

func (r *Repository) DeleteSession(ctx context.Context, db database.PgExecutor, sessionID string) error {
	query := `DELETE FROM
		sessions
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, sessionID)
	return err
}

func (r *Repository) DeleteSessionsByUserID(ctx context.Context, db database.PgExecutor, userID string) error {
	query := `DELETE FROM
		sessions
	WHERE
		user_id = $1`
	_, err := db.Exec(ctx, query, userID)
	return err
}

func (r *Repository) GetSessionByID(ctx context.Context, db database.PgExecutor, sessionID string) (string, time.Time, error) {
	var userID string
	var expiresAt time.Time
	query := `SELECT
		user_id, expires_at
	FROM
		sessions
	WHERE
		id = $1 AND expires_at > now()`
	if err := db.QueryRow(ctx, query, sessionID).Scan(&userID, &expiresAt); err != nil {
		return "", time.Time{}, err
	}
	return userID, expiresAt, nil
}
