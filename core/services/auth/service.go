package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"arkive/core/database"
	"arkive/core/models"
	authrepo "arkive/core/repositories/auth"
)

type Service struct {
	db         database.PgPool
	authRepo   *authrepo.Repository
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

func NewService(db database.PgPool, authRepo *authrepo.Repository, cfg Config) *Service {
	return &Service{
		db:         db,
		authRepo:   authRepo,
		jwtSecret:  []byte(cfg.JWTSecret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
		sessionTTL: cfg.SessionTTL,
	}
}

func (s *Service) WebLogin(ctx context.Context, email, password string) (string, time.Time, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	if email == "" || password == "" {
		return "", time.Time{}, ErrInvalidInput
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", time.Time{}, err
	}

	user, err := s.authenticateUser(ctx, email, password)
	if err != nil {
		_ = tx.Rollback(ctx)
		return "", time.Time{}, err
	}

	sessionID, expiresAt, err := s.createSession(ctx, user.ID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return "", time.Time{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", time.Time{}, err
	}

	return sessionID, expiresAt, nil
}

func (s *Service) WebSignup(ctx context.Context, brandName, email, password string) (string, time.Time, error) {
	brandName = strings.TrimSpace(brandName)
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	if brandName == "" || email == "" || password == "" {
		return "", time.Time{}, ErrInvalidInput
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", time.Time{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		_ = tx.Rollback(ctx)
		return "", time.Time{}, err
	}

	user, err := s.authRepo.CreateUser(ctx, brandName, email, string(hash))
	if err != nil {
		_ = tx.Rollback(ctx)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_email_key":
				return "", time.Time{}, ErrEmailExists
			case "users_brand_name_key":
				return "", time.Time{}, ErrBrandNameExists
			}
		}
		return "", time.Time{}, err
	}

	sessionID, expiresAt, err := s.createSession(ctx, user.ID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return "", time.Time{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", time.Time{}, err
	}

	return sessionID, expiresAt, nil
}

func (s *Service) LogoutSession(ctx context.Context, sessionID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	if err := s.authRepo.DeleteSession(ctx, sessionID); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) LoginTokens(ctx context.Context, email, password string) (string, time.Time, string, time.Time, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	if email == "" || password == "" {
		return "", time.Time{}, "", time.Time{}, ErrInvalidInput
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}

	user, err := s.authenticateUser(ctx, email, password)
	if err != nil {
		_ = tx.Rollback(ctx)
		return "", time.Time{}, "", time.Time{}, err
	}

	refreshToken, refreshExpires, err := s.createRefreshToken(ctx, user.ID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return "", time.Time{}, "", time.Time{}, err
	}

	accessToken, accessExpires, err := s.CreateAccessToken(user.ID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return "", time.Time{}, "", time.Time{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}

	return accessToken, accessExpires, refreshToken, refreshExpires, nil
}

func (s *Service) RefreshTokens(ctx context.Context, token string) (string, time.Time, string, time.Time, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", time.Time{}, "", time.Time{}, ErrInvalidInput
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}

	hash := sha256.Sum256([]byte(token))
	id, userID, expiresAt, revokedAt, err := s.authRepo.GetRefreshToken(ctx, hash[:])
	if err != nil {
		_ = tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return "", time.Time{}, "", time.Time{}, ErrRefreshTokenInvalid
		}
		return "", time.Time{}, "", time.Time{}, err
	}
	if revokedAt != nil || time.Now().After(expiresAt) {
		_ = tx.Rollback(ctx)
		return "", time.Time{}, "", time.Time{}, ErrRefreshTokenInvalid
	}

	if err := s.authRepo.RevokeRefreshToken(ctx, id); err != nil {
		_ = tx.Rollback(ctx)
		return "", time.Time{}, "", time.Time{}, err
	}

	refreshToken, refreshExpires, err := s.createRefreshToken(ctx, userID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return "", time.Time{}, "", time.Time{}, err
	}

	accessToken, accessExpires, err := s.CreateAccessToken(userID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return "", time.Time{}, "", time.Time{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}

	return accessToken, accessExpires, refreshToken, refreshExpires, nil
}

func (s *Service) RevokeRefreshToken(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(token))
	revoked, err := s.authRepo.RevokeRefreshTokenByHash(ctx, hash[:])
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if !revoked {
		_ = tx.Rollback(ctx)
		return ErrRefreshTokenInvalid
	}

	return tx.Commit(ctx)
}

func (s *Service) GetUserByID(ctx context.Context, userID string) (models.User, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return models.User{}, ErrInvalidInput
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.User{}, err
	}

	user, err := s.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return models.User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.User{}, err
	}

	return user, nil
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

func (s *Service) authenticateUser(ctx context.Context, email, password string) (models.User, error) {
	user, hash, err := s.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrInvalidCredentials
		}
		return models.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return models.User{}, ErrInvalidCredentials
	}

	return user, nil
}

func (s *Service) createSession(ctx context.Context, userID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.sessionTTL)
	sessionID, err := s.authRepo.CreateSession(ctx, userID, expiresAt)
	if err != nil {
		return "", time.Time{}, err
	}
	return sessionID, expiresAt, nil
}

func (s *Service) createRefreshToken(ctx context.Context, userID string) (string, time.Time, error) {
	token, hash, err := generateToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(s.refreshTTL)
	if err := s.authRepo.CreateRefreshToken(ctx, userID, hash, expiresAt); err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
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
