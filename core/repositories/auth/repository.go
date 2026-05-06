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

func (r *Repository) CreateVerifiedUser(ctx context.Context, db database.PgExecutor, brandName, email, passwordHash string, vaultSalt, encryptedMasterKey []byte) (models.User, error) {
	var user models.User
	query := `INSERT INTO users
		(brand_name, email, password_hash, vault_salt, encrypted_master_key)
	VALUES
		($1, $2, $3, $4, $5)
	RETURNING
		id, brand_name, email, vault_salt, encrypted_master_key`
	if err := db.QueryRow(ctx, query, brandName, email, passwordHash, vaultSalt, encryptedMasterKey).Scan(
		&user.ID,
		&user.BrandName,
		&user.Email,
		&user.VaultSalt,
		&user.EncryptedMasterKey,
	); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, db database.PgExecutor, email string) (models.User, *string, error) {
	var user models.User
	var hash *string
	query := `SELECT
		id, brand_name, email, password_hash, vault_salt, encrypted_master_key
	FROM
		users
	WHERE
		email = $1`
	if err := db.QueryRow(ctx, query, email).Scan(&user.ID, &user.BrandName, &user.Email, &hash, &user.VaultSalt, &user.EncryptedMasterKey); err != nil {
		return models.User{}, nil, err
	}
	return user, hash, nil
}

func (r *Repository) GetUserByID(ctx context.Context, db database.PgExecutor, userID string) (models.User, error) {
	var user models.User
	query := `SELECT
		id, brand_name, email, vault_salt, encrypted_master_key, quota_bytes, used_bytes, reserved_bytes,
		last_login_at, created_at
	FROM
		users
	WHERE
		id = $1`
	if err := db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.BrandName,
		&user.Email,
		&user.VaultSalt,
		&user.EncryptedMasterKey,
		&user.QuotaBytes,
		&user.UsedBytes,
		&user.ReservedBytes,
		&user.LastLoginAt,
		&user.CreatedAt,
	); err != nil {
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
