package authrepo

import (
	"context"

	"arkive/core/database"
	"arkive/core/models"
)

type Repository struct {
}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) CreateUser(ctx context.Context, db database.PgExecutor, brandName, email, passwordHash string) (models.User, error) {
	var user models.User
	query := `INSERT INTO users (brand_name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, brand_name, email`
	if err := db.QueryRow(ctx, query, brandName, email, passwordHash).Scan(&user.ID, &user.BrandName, &user.Email); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, db database.PgExecutor, email string) (models.User, string, error) {
	var user models.User
	var hash string
	query := `SELECT id, brand_name, email, password_hash
		FROM users
		WHERE email = $1`
	if err := db.QueryRow(ctx, query, email).Scan(&user.ID, &user.BrandName, &user.Email, &hash); err != nil {
		return models.User{}, "", err
	}
	return user, hash, nil
}

func (r *Repository) GetUserByID(ctx context.Context, db database.PgExecutor, userID string) (models.User, error) {
	var user models.User
	query := `SELECT id, brand_name, email
		FROM users
		WHERE id = $1`
	if err := db.QueryRow(ctx, query, userID).Scan(&user.ID, &user.BrandName, &user.Email); err != nil {
		return models.User{}, err
	}
	return user, nil
}
