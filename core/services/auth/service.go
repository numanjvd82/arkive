package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"arkive/core/database"
)

type Service struct {
	db         database.PgExecutor
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	sessionTTL time.Duration
}

type Config struct {
	JWTSecret  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	SessionTTL time.Duration
}

type User struct {
	ID        string
	BrandName string
	Email     string
}

func NewService(db database.PgExecutor, cfg Config) *Service {
	return &Service{
		db:         db,
		jwtSecret:  []byte(cfg.JWTSecret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
		sessionTTL: cfg.SessionTTL,
	}
}

func (s *Service) CreateUser(ctx context.Context, brandName, email, password string) (User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	var user User
	query := `insert into users (brand_name, email, password_hash) values ($1, $2, $3) returning id, brand_name, email`
	if err := s.db.QueryRow(ctx, query, brandName, email, string(hash)).Scan(&user.ID, &user.BrandName, &user.Email); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_email_key":
				return User{}, ErrEmailExists
			case "users_brand_name_key":
				return User{}, ErrBrandNameExists
			}
		}
		return User{}, err
	}

	return user, nil
}

func (s *Service) Authenticate(ctx context.Context, email, password string) (User, error) {
	var user User
	var hash string
	query := `select id, brand_name, email, password_hash from users where email = $1`
	if err := s.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.BrandName, &user.Email, &hash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrInvalidCredentials
		}
		return User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return User{}, ErrInvalidCredentials
	}

	return user, nil
}

func (s *Service) CreateSession(ctx context.Context, userID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.sessionTTL)
	var sessionID string
	query := `insert into sessions (user_id, expires_at) values ($1, $2) returning id`
	if err := s.db.QueryRow(ctx, query, userID, expiresAt).Scan(&sessionID); err != nil {
		return "", time.Time{}, err
	}
	return sessionID, expiresAt, nil
}

func (s *Service) DeleteSession(ctx context.Context, sessionID string) error {
	query := `delete from sessions where id = $1`
	_, err := s.db.Exec(ctx, query, sessionID)
	return err
}

func (s *Service) CreateAccessToken(userID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.accessTTL)
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func (s *Service) ParseAccessToken(tokenString string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	parsed, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil || !parsed.Valid {
		return "", ErrInvalidCredentials
	}
	if claims.Subject == "" {
		return "", ErrInvalidCredentials
	}
	return claims.Subject, nil
}

func (s *Service) CreateRefreshToken(ctx context.Context, userID string) (string, time.Time, error) {
	token, hash, err := generateToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(s.refreshTTL)
	query := `insert into refresh_tokens (user_id, token_hash, expires_at) values ($1, $2, $3)`
	if _, err := s.db.Exec(ctx, query, userID, hash, expiresAt); err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

func (s *Service) RotateRefreshToken(ctx context.Context, token string) (string, time.Time, string, error) {
	hash := sha256.Sum256([]byte(token))
	var id string
	var userID string
	var expiresAt time.Time
	var revokedAt *time.Time
	query := `select id, user_id, expires_at, revoked_at from refresh_tokens where token_hash = $1`
	if err := s.db.QueryRow(ctx, query, hash[:]).Scan(&id, &userID, &expiresAt, &revokedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", time.Time{}, "", ErrRefreshTokenInvalid
		}
		return "", time.Time{}, "", err
	}
	if revokedAt != nil || time.Now().After(expiresAt) {
		return "", time.Time{}, "", ErrRefreshTokenInvalid
	}

	if _, err := s.db.Exec(ctx, `update refresh_tokens set revoked_at = now() where id = $1`, id); err != nil {
		return "", time.Time{}, "", err
	}

	newToken, newExpiresAt, err := s.CreateRefreshToken(ctx, userID)
	if err != nil {
		return "", time.Time{}, "", err
	}

	return newToken, newExpiresAt, userID, nil
}

func (s *Service) RevokeRefreshToken(ctx context.Context, token string) error {
	hash := sha256.Sum256([]byte(token))
	tag, err := s.db.Exec(ctx, `update refresh_tokens set revoked_at = now() where token_hash = $1 and revoked_at is null`, hash[:])
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrRefreshTokenInvalid
	}
	return nil
}

func (s *Service) GetUserByID(ctx context.Context, userID string) (User, error) {
	var user User
	query := `select id, brand_name, email from users where id = $1`
	if err := s.db.QueryRow(ctx, query, userID).Scan(&user.ID, &user.BrandName, &user.Email); err != nil {
		return User{}, err
	}
	return user, nil
}

func generateToken() (string, []byte, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	hash := sha256.Sum256([]byte(token))
	return token, hash[:], nil
}
