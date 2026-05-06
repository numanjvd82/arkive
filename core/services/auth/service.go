package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"arkive/core/database"
	"arkive/core/models"
	authrepo "arkive/core/repositories/auth"
	sessionrepo "arkive/core/repositories/session"
	usersrepo "arkive/core/repositories/users"
	"arkive/pkg/mailer"
	"arkive/pkg/validation"
)

type Service struct {
	db          database.PgPool
	authRepo    *authrepo.Repository
	sessionRepo *sessionrepo.Repository
	userRepo    *usersrepo.Repository

	mailer        mailer.Mailer
	publicBaseURL string
	sessionTTL    time.Duration
}

type Config struct {
	SessionTTL time.Duration
}

type LoginUnlockResult struct {
	SessionID          string
	ExpiresAt          time.Time
	VaultSalt          []byte
	EncryptedMasterKey []byte
}

func NewService(
	db database.PgPool,
	authRepo *authrepo.Repository,
	sessionRepo *sessionrepo.Repository,
	userRepo *usersrepo.Repository,
	cfg Config,
) *Service {
	return &Service{
		db:          db,
		authRepo:    authRepo,
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		sessionTTL:  cfg.SessionTTL,
	}
}

func (s *Service) SetMailer(m mailer.Mailer, publicBaseURL string) {
	s.mailer = m
	s.publicBaseURL = strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
}

func (s *Service) EmailVerificationEnabled() bool {
	return s.mailer != nil && strings.TrimSpace(s.publicBaseURL) != ""
}

func (s *Service) WebLogin(ctx context.Context, email, password, lastIP string) (string, time.Time, validation.Errors, error) {
	result, validationErrors, err := s.LoginAndLoadVault(ctx, email, password, lastIP)
	if err != nil || (validationErrors != nil && validationErrors.HasAny()) {
		return "", time.Time{}, validationErrors, err
	}
	return result.SessionID, result.ExpiresAt, nil, nil
}

func (s *Service) LoginAndLoadVault(ctx context.Context, email, password, lastIP string) (LoginUnlockResult, validation.Errors, error) {
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
		return LoginUnlockResult{}, validationErrors, nil
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return LoginUnlockResult{}, nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	user, err := s.authenticateUser(ctx, tx, email, password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			validationErrors := validation.New()
			validationErrors.Add(validation.GeneralKey, ErrLoginInvalid.Error())
			return LoginUnlockResult{}, validationErrors, nil
		}
		return LoginUnlockResult{}, nil, err
	}

	if len(user.VaultSalt) == 0 || len(user.EncryptedMasterKey) == 0 {
		return LoginUnlockResult{}, nil, ErrVaultNotConfigured
	}

	sessionID, expiresAt, err := s.createSession(ctx, tx, user.ID, lastIP)
	if err != nil {
		return LoginUnlockResult{}, nil, err
	}

	loginAt := time.Now()
	if err := s.userRepo.UpdateLoginActivity(ctx, tx, user.ID, loginAt, lastIP); err != nil {
		_ = tx.Rollback(ctx)
		return LoginUnlockResult{}, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return LoginUnlockResult{}, nil, err
	}

	return LoginUnlockResult{
		SessionID:          sessionID,
		ExpiresAt:          expiresAt,
		VaultSalt:          user.VaultSalt,
		EncryptedMasterKey: user.EncryptedMasterKey,
	}, nil, nil
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

func (s *Service) authenticateUser(ctx context.Context, db database.PgExecutor, email, password string) (models.User, error) {
	user, hash, err := s.authRepo.GetUserByEmail(ctx, db, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrInvalidCredentials
		}
		return models.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*hash), []byte(password)); err != nil {
		return models.User{}, ErrInvalidCredentials
	}

	return user, nil
}

func (s *Service) createSession(ctx context.Context, db database.PgExecutor, userID, lastIP string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.sessionTTL)
	if err := s.authRepo.UpdateLastLogin(ctx, db, userID, strings.TrimSpace(lastIP)); err != nil {
		return "", time.Time{}, err
	}
	sessionID, err := s.sessionRepo.CreateSession(ctx, db, userID, expiresAt)
	if err != nil {
		return "", time.Time{}, err
	}
	return sessionID, expiresAt, nil
}
