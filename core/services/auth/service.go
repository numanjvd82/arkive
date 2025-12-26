package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"arkive/core/database"
	"arkive/core/models"
	authrepo "arkive/core/repositories/auth"
	sessionrepo "arkive/core/repositories/session"
	jwtservice "arkive/core/services/jwt"
	"arkive/pkg/tokens"
	"arkive/pkg/validation"
)

type Service struct {
	db          database.PgPool
	authRepo    *authrepo.Repository
	sessionRepo *sessionrepo.Repository
	jwtService  *jwtservice.Service
	accessTTL   time.Duration
	refreshTTL  time.Duration
	sessionTTL  time.Duration
}

type Config struct {
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	SessionTTL time.Duration
}

func NewService(
	db database.PgPool,
	authRepo *authrepo.Repository,
	sessionRepo *sessionrepo.Repository,
	jwtService *jwtservice.Service,
	cfg Config,
) *Service {
	return &Service{
		db:          db,
		authRepo:    authRepo,
		sessionRepo: sessionRepo,
		jwtService:  jwtService,
		accessTTL:   cfg.AccessTTL,
		refreshTTL:  cfg.RefreshTTL,
		sessionTTL:  cfg.SessionTTL,
	}
}

func (s *Service) WebLogin(ctx context.Context, email, password string) (string, time.Time, validation.Errors, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	if email == "" || password == "" {
		validationErrors := validation.New()
		if email == "" {
			validationErrors.Add("email", ErrEmailRequired.Error())
		}
		if password == "" {
			validationErrors.Add("password", ErrPasswordRequired.Error())
		}
		return "", time.Time{}, validationErrors, nil
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", time.Time{}, nil, err
	}

	user, err := s.authenticateUser(ctx, email, password)
	if err != nil {
		_ = tx.Rollback(ctx)
		if errors.Is(err, ErrInvalidCredentials) {
			validationErrors := validation.New()
			validationErrors.Add(validation.GeneralKey, ErrLoginInvalid.Error())
			return "", time.Time{}, validationErrors, nil
		}
		return "", time.Time{}, nil, err
	}

	sessionID, expiresAt, err := s.createSession(ctx, tx, user.ID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return "", time.Time{}, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", time.Time{}, nil, err
	}

	return sessionID, expiresAt, nil, nil
}

func (s *Service) WebSignup(ctx context.Context, brandName, email, password, confirmPassword string) (validation.Errors, error) {
	brandName = strings.TrimSpace(brandName)
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	confirmPassword = strings.TrimSpace(confirmPassword)
	validationErrors := validation.New()
	if brandName == "" {
		validationErrors.Add("brand_name", ErrBrandNameRequired.Error())
	}
	if email == "" {
		validationErrors.Add("email", ErrEmailRequired.Error())
	}
	if password == "" {
		validationErrors.Add("password", ErrPasswordRequired.Error())
	}
	if confirmPassword == "" {
		validationErrors.Add("confirm_password", ErrConfirmPasswordRequired.Error())
	}
	if password != "" {
		switch validation.PasswordIssueFor(password) {
		case validation.PasswordTooShort:
			validationErrors.Add("password", ErrPasswordTooShort.Error())
		case validation.PasswordMissingLower:
			validationErrors.Add("password", ErrPasswordMissingLower.Error())
		case validation.PasswordMissingUpper:
			validationErrors.Add("password", ErrPasswordMissingUpper.Error())
		case validation.PasswordMissingSymbol:
			validationErrors.Add("password", ErrPasswordMissingSymbol.Error())
		}
	}
	if password != "" && confirmPassword != "" && password != confirmPassword {
		validationErrors.Add("confirm_password", ErrPasswordMismatch.Error())
	}
	if validationErrors.HasAny() {
		return validationErrors, nil
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	if _, _, err := s.authRepo.GetUserByEmail(ctx, tx, email); err == nil {
		validationErrors.Add("email", ErrEmailExists.Error())
	} else if !errors.Is(err, pgx.ErrNoRows) {
		_ = tx.Rollback(ctx)
		return nil, err
	}

	if _, err := s.authRepo.GetUserByBrandName(ctx, tx, brandName); err == nil {
		validationErrors.Add("brand_name", ErrBrandNameExists.Error())
	} else if !errors.Is(err, pgx.ErrNoRows) {
		_ = tx.Rollback(ctx)
		return nil, err
	}

	if validationErrors.HasAny() {
		_ = tx.Rollback(ctx)
		return validationErrors, nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}

	_, err = s.authRepo.CreateUser(ctx, tx, brandName, email, string(hash))
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return nil, nil
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

	if err := s.sessionRepo.DeleteSession(ctx, tx, sessionID); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) ValidateSession(ctx context.Context, sessionID string) (string, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return "", ErrSessionNotFound
	}

	userID, _, err := s.sessionRepo.GetSessionByID(ctx, s.db, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrSessionNotFound
		}
		return "", err
	}
	return userID, nil
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

	refreshToken, refreshExpires, err := s.createRefreshToken(ctx, tx, user.ID)
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
	id, userID, expiresAt, revokedAt, err := s.sessionRepo.GetRefreshToken(ctx, tx, hash[:])
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

	if err := s.sessionRepo.RevokeRefreshToken(ctx, tx, id); err != nil {
		_ = tx.Rollback(ctx)
		return "", time.Time{}, "", time.Time{}, err
	}

	refreshToken, refreshExpires, err := s.createRefreshToken(ctx, tx, userID)
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
	revoked, err := s.sessionRepo.RevokeRefreshTokenByHash(ctx, tx, hash[:])
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

	user, err := s.authRepo.GetUserByID(ctx, tx, userID)
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
	return s.jwtService.CreateAccessToken(userID, s.accessTTL)
}

func (s *Service) ParseAccessToken(tokenString string) (string, error) {
	userID, err := s.jwtService.ParseAccessToken(tokenString)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	return userID, nil
}

func (s *Service) authenticateUser(ctx context.Context, email, password string) (models.User, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.User{}, err
	}
	user, hash, err := s.authRepo.GetUserByEmail(ctx, tx, email)
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

func (s *Service) createSession(ctx context.Context, db database.PgExecutor, userID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.sessionTTL)
	sessionID, err := s.sessionRepo.CreateSession(ctx, db, userID, expiresAt)
	if err != nil {
		return "", time.Time{}, err
	}
	return sessionID, expiresAt, nil
}

func (s *Service) createRefreshToken(ctx context.Context, db database.PgExecutor, userID string) (string, time.Time, error) {
	token, hash, err := tokens.Generate()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(s.refreshTTL)
	if err := s.sessionRepo.CreateRefreshToken(ctx, db, userID, hash, expiresAt); err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}
