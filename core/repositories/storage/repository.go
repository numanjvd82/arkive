package storagerepo

import (
	"context"

	"arkive/core/database"
)

type Repository struct{}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) ReserveStorage(ctx context.Context, db database.PgExecutor, userID string, sizeBytes int64) (bool, error) {
	query := `UPDATE
		users
	SET
		reserved_bytes = reserved_bytes + $2
	WHERE
		id = $1 AND used_bytes + reserved_bytes + $2 <= quota_bytes`
	tag, err := db.Exec(ctx, query, userID, sizeBytes)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *Repository) CommitStorage(ctx context.Context, db database.PgExecutor, userID string, sizeBytes int64) (bool, error) {
	query := `UPDATE
		users
	SET
		used_bytes = used_bytes + $2,
		reserved_bytes = reserved_bytes - $2
	WHERE
		id = $1 AND reserved_bytes >= $2`
	tag, err := db.Exec(ctx, query, userID, sizeBytes)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *Repository) ReleaseReservedStorage(ctx context.Context, db database.PgExecutor, userID string, sizeBytes int64) (bool, error) {
	query := `UPDATE
		users
	SET
		reserved_bytes = reserved_bytes - $2
	WHERE
		id = $1 AND reserved_bytes >= $2`
	tag, err := db.Exec(ctx, query, userID, sizeBytes)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *Repository) DecreaseUsedStorage(ctx context.Context, db database.PgExecutor, userID string, sizeBytes int64) error {
	query := `UPDATE
		users
	SET
		used_bytes = GREATEST(used_bytes - $2, 0)
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, userID, sizeBytes)
	return err
}
