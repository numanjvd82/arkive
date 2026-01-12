package folderrepo

import (
	"context"

	"arkive/core/database"
	"arkive/core/models"
)

type Repository struct{}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) CreateFolder(ctx context.Context, db database.PgExecutor, folder models.Folder) error {
	query := `INSERT INTO folders
		(user_id, path, name, parent_path)
	VALUES
		($1, $2, $3, $4)
	ON CONFLICT (user_id, path) DO NOTHING`
	_, err := db.Exec(ctx, query,
		folder.UserID,
		folder.Path,
		folder.Name,
		folder.ParentPath,
	)
	return err
}

func (r *Repository) ListByParent(ctx context.Context, db database.PgExecutor, userID, parentPath string) ([]models.Folder, error) {
	query := `SELECT
		id, user_id, path, name, parent_path, created_at, updated_at
	FROM
		folders
	WHERE
		user_id = $1
		AND parent_path = $2
	ORDER BY
		name ASC`
	rows, err := db.Query(ctx, query, userID, parentPath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []models.Folder
	for rows.Next() {
		var folder models.Folder
		if err := rows.Scan(
			&folder.ID,
			&folder.UserID,
			&folder.Path,
			&folder.Name,
			&folder.ParentPath,
			&folder.CreatedAt,
			&folder.UpdatedAt,
		); err != nil {
			return nil, err
		}
		folders = append(folders, folder)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return folders, nil
}

func (r *Repository) DeleteByPath(ctx context.Context, db database.PgExecutor, userID, path string) error {
	query := `DELETE FROM
		folders
	WHERE
		user_id = $1
		AND path = $2`
	_, err := db.Exec(ctx, query, userID, path)
	return err
}
