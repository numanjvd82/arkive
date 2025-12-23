package authrepo

import (
	"context"
	"time"

	"arkive/core/database"
	"arkive/core/models"
)

type Repository struct {
	db database.PgExecutor
}

func New(db database.PgExecutor) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, brandName, email, passwordHash string) (models.User, error) {
	var user models.User
	query := `INSERT INTO users (brand_name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, brand_name, email`
	if err := r.db.QueryRow(ctx, query, brandName, email, passwordHash).Scan(&user.ID, &user.BrandName, &user.Email); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (models.User, string, error) {
	var user models.User
	var hash string
	query := `SELECT id, brand_name, email, password_hash
		FROM users
		WHERE email = $1`
	if err := r.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.BrandName, &user.Email, &hash); err != nil {
		return models.User{}, "", err
	}
	return user, hash, nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID string) (models.User, error) {
	var user models.User
	query := `SELECT id, brand_name, email
		FROM users
		WHERE id = $1`
	if err := r.db.QueryRow(ctx, query, userID).Scan(&user.ID, &user.BrandName, &user.Email); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *Repository) CreateSession(ctx context.Context, userID string, expiresAt time.Time) (string, error) {
	var sessionID string
	query := `INSERT INTO sessions (user_id, expires_at)
		VALUES ($1, $2)
		RETURNING id`
	if err := r.db.QueryRow(ctx, query, userID, expiresAt).Scan(&sessionID); err != nil {
		return "", err
	}
	return sessionID, nil
}

func (r *Repository) DeleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM sessions
		WHERE id = $1`
	_, err := r.db.Exec(ctx, query, sessionID)
	return err
}

func (r *Repository) CreateRefreshToken(ctx context.Context, userID string, hash []byte, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, userID, hash, expiresAt)
	return err
}

func (r *Repository) GetRefreshToken(ctx context.Context, hash []byte) (string, string, time.Time, *time.Time, error) {
	var id string
	var userID string
	var expiresAt time.Time
	var revokedAt *time.Time
	query := `SELECT id, user_id, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1`
	if err := r.db.QueryRow(ctx, query, hash).Scan(&id, &userID, &expiresAt, &revokedAt); err != nil {
		return "", "", time.Time{}, nil, err
	}
	return id, userID, expiresAt, revokedAt, nil
}

func (r *Repository) RevokeRefreshToken(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE id = $1`, id)
	return err
}

func (r *Repository) RevokeRefreshTokenByHash(ctx context.Context, hash []byte) (bool, error) {
	tag, err := r.db.Exec(ctx, `UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE token_hash = $1 AND revoked_at IS NULL`, hash)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
