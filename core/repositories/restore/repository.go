package restoreusage

import (
	"context"
	"time"

	"arkive/core/database"
)

type Repository struct{}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) AddUsage(ctx context.Context, db database.PgExecutor, userID string, sizeBytes int64) error {
	query := `INSERT INTO restore_usage
		(user_id, size_bytes)
	VALUES
		($1, $2)`
	_, err := db.Exec(ctx, query, userID, sizeBytes)
	return err
}

func (r *Repository) SumUsageSince(ctx context.Context, db database.PgExecutor, userID string, since time.Time) (int64, error) {
	var total int64
	query := `SELECT
		COALESCE(SUM(size_bytes), 0)
	FROM
		restore_usage
	WHERE
		user_id = $1 AND created_at >= $2`
	if err := db.QueryRow(ctx, query, userID, since).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}
