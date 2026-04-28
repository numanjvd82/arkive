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
	query := `INSERT INTO users
		(brand_name, email, password_hash)
	VALUES
		($1, $2, $3)
	RETURNING
		id, brand_name, email`
	if err := db.QueryRow(ctx, query, brandName, email, passwordHash).Scan(&user.ID, &user.BrandName, &user.Email); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, db database.PgExecutor, email string) (models.User, *string, error) {
	var user models.User
	var hash *string
	query := `SELECT
		id, brand_name, email, password_hash
	FROM
		users
	WHERE
		email = $1`
	if err := db.QueryRow(ctx, query, email).Scan(&user.ID, &user.BrandName, &user.Email, &hash); err != nil {
		return models.User{}, nil, err
	}
	return user, hash, nil
}

func (r *Repository) GetUserByID(ctx context.Context, db database.PgExecutor, userID string) (models.User, error) {
	var user models.User
	query := `SELECT
		id, brand_name, email, quota_bytes, used_bytes, reserved_bytes, is_premium,
		is_email_verified, is_banned, ban_reason, last_login_at, last_active_at, last_ip::text, created_at, updated_at
	FROM
		users
	WHERE
		id = $1`
	if err := db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.BrandName,
		&user.Email,
		&user.QuotaBytes,
		&user.UsedBytes,
		&user.ReservedBytes,
		&user.IsPremium,
		&user.IsEmailVerified,
		&user.IsBanned,
		&user.BanReason,
		&user.LastLoginAt,
		&user.LastActiveAt,
		&user.LastIP,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *Repository) GetUserByBrandName(ctx context.Context, db database.PgExecutor, brandName string) (models.User, error) {
	var user models.User
	query := `SELECT
		id, brand_name, email
	FROM
		users
	WHERE
		brand_name = $1`
	if err := db.QueryRow(ctx, query, brandName).Scan(&user.ID, &user.BrandName, &user.Email); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *Repository) GetUserByGoogleSub(ctx context.Context, db database.PgExecutor, sub string) (models.User, error) {
	var user models.User
	query := `SELECT
		id, brand_name, email
	FROM
		users
	WHERE
		google_sub = $1`
	if err := db.QueryRow(ctx, query, sub).Scan(&user.ID, &user.BrandName, &user.Email); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *Repository) UpdateLastLogin(ctx context.Context, db database.PgExecutor, userID, lastIP string) error {
	query := `UPDATE
		users
	SET
		last_login_at = now(),
		last_ip = NULLIF($2, '')::inet
	WHERE
		id = $1`
	_, err := db.Exec(ctx, query, userID, lastIP)
	return err
}

func (r *Repository) CreateUserWithGoogleProfile(ctx context.Context, db database.PgExecutor, brandName, email, sub, givenName, familyName string, emailVerified bool, pictureURL string) (models.User, error) {
	var user models.User
	query := `INSERT INTO users
		(brand_name, email, password_hash, google_sub, google_given_name, google_family_name, google_email_verified, google_picture_url, is_email_verified)
	VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING
		id, brand_name, email`
	if err := db.QueryRow(ctx, query, brandName, email, nil, sub, givenName, familyName, emailVerified, pictureURL, emailVerified).Scan(&user.ID, &user.BrandName, &user.Email); err != nil {
		return models.User{}, err
	}
	return user, nil
}
