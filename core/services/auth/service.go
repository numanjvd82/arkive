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
	emailverifyrepo "arkive/core/repositories/emailverify"
	sessionrepo "arkive/core/repositories/session"
	usersrepo "arkive/core/repositories/users"
	"arkive/pkg/validation"
)

type Service struct {
	db              database.PgPool
	authRepo        *authrepo.Repository
	sessionRepo     *sessionrepo.Repository
	userRepo        *usersrepo.Repository
	emailVerifyRepo *emailverifyrepo.Repository

	emailSender   EmailSender
	publicBaseURL string
	sessionTTL    time.Duration
}

type Config struct {
	SessionTTL time.Duration
}

func NewService(
	db database.PgPool,
	authRepo *authrepo.Repository,
	sessionRepo *sessionrepo.Repository,
	userRepo *usersrepo.Repository,
	cfg Config,
) *Service {
	return &Service{
		db:              db,
		authRepo:        authRepo,
		sessionRepo:     sessionRepo,
		userRepo:        userRepo,
		emailVerifyRepo: emailverifyrepo.New(),
		sessionTTL:      cfg.SessionTTL,
	}
}

func (s *Service) WebLogin(ctx context.Context, email, password, lastIP string) (string, time.Time, validation.Errors, error) {
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
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	user, err := s.authenticateUser(ctx, tx, email, password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			validationErrors := validation.New()
			validationErrors.Add(validation.GeneralKey, ErrLoginInvalid.Error())
			return "", time.Time{}, validationErrors, nil
		}
		return "", time.Time{}, nil, err
	}

	if s.EmailVerificationEnabled() && !user.IsEmailVerified {
		validationErrors := validation.New()
		validationErrors.Add(validation.GeneralKey, ErrEmailNotVerified.Error())
		return "", time.Time{}, validationErrors, nil
	}

	sessionID, expiresAt, err := s.createSession(ctx, tx, user.ID, lastIP)
	if err != nil {
		return "", time.Time{}, nil, err
	}

	loginAt := time.Now()
	if err := s.userRepo.UpdateLoginActivity(ctx, tx, user.ID, loginAt, lastIP); err != nil {
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

	user, err := s.authRepo.CreateUser(ctx, tx, brandName, email, string(hash))
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}

	// Best-effort email verification; if email provider isn't configured (dev), this is a no-op.
	if err := s.sendEmailVerification(ctx, tx, user.ID, email); err != nil {
		_ = tx.Rollback(ctx)
		// Treat email send as part of signup: no email means no account.
		validationErrors.Add(validation.GeneralKey, ErrEmailSendFailed.Error())
		return validationErrors, nil
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

func (s *Service) EmailVerificationEnabled() bool {
	return s.emailSender != nil && strings.TrimSpace(s.publicBaseURL) != ""
}
