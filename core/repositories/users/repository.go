package usersrepo

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

func (r *Repository) CountUsers(ctx context.Context, db database.PgExecutor) (int64, error) {
	query := `SELECT COUNT(*) FROM users`
	var count int64
	if err := db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) UpdateLoginActivity(ctx context.Context, db database.PgExecutor, userID string, loginAt time.Time, lastIP string) error {
	query := `UPDATE
		users
	SET
		last_login_at = $2,
		last_active_at = $2,
		last_ip = $3,
		updated_at = now()
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, userID, loginAt, lastIP)
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

func (r *Repository) ListInactiveUsers(ctx context.Context, db database.PgExecutor, cutoff time.Time) ([]models.User, error) {
	query := `SELECT
		id, last_active_at
	FROM
		users
	WHERE
		last_active_at <= $1`
	rows, err := db.Query(ctx, query, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.LastActiveAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return users, nil
}
