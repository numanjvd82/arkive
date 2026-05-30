package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"arkive/core/database"
	"arkive/core/models"
	authrepo "arkive/core/repositories/auth"
	sessionrepo "arkive/core/repositories/session"
	usersrepo "arkive/core/repositories/users"
	"arkive/pkg/tokens"
	"arkive/pkg/validation"
)

type Service struct {
	db          database.PgPool
	authRepo    *authrepo.Repository
	sessionRepo *sessionrepo.Repository
	userRepo    *usersrepo.Repository

	sessionTTL time.Duration
	resetTTL   time.Duration
	resetURL   string
}

type Config struct {
	SessionTTL            time.Duration
	PasswordResetTokenTTL time.Duration
	BaseURL               string
}

type LoginUnlockResult struct {
	SessionID          string
	ExpiresAt          time.Time
	VaultSalt          []byte
	EncryptedMasterKey []byte
}

type VaultUnlockResult struct {
	VaultSalt          []byte
	EncryptedMasterKey []byte
}

type PasswordResetRequestResult struct {
	ResetToken string
	ResetURL   string
	ExpiresAt  time.Time
}

type PasswordRecoveryVault struct {
	UserID                     string
	VaultSalt                  []byte
	EncryptedMasterKeyRecovery []byte
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
		resetTTL:    cfg.PasswordResetTokenTTL,
		resetURL:    strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
	}
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
	if err := s.userRepo.UpdateLoginActivity(ctx, tx, user.ID, loginAt); err != nil {
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

func (s *Service) UnlockVaultWithSession(ctx context.Context, userID, password string) (VaultUnlockResult, validation.Errors, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" || password == "" {
		validationErrors := validation.New()
		if password == "" {
			validationErrors.Add("password", ErrPasswordRequired.Error())
		}
		if userID == "" {
			return VaultUnlockResult{}, nil, ErrUnauthorized
		}
		return VaultUnlockResult{}, validationErrors, nil
	}

	user, err := s.authRepo.GetUserByID(ctx, s.db, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return VaultUnlockResult{}, nil, ErrUnauthorized
		}
		return VaultUnlockResult{}, nil, err
	}
	_, authErr := s.authenticateUser(ctx, s.db, user.Email, password)
	if authErr != nil {
		if errors.Is(authErr, ErrInvalidCredentials) {
			validationErrors := validation.New()
			validationErrors.Add(validation.GeneralKey, ErrLoginInvalid.Error())
			return VaultUnlockResult{}, validationErrors, nil
		}
		return VaultUnlockResult{}, nil, authErr
	}
	if len(user.VaultSalt) == 0 || len(user.EncryptedMasterKey) == 0 {
		return VaultUnlockResult{}, nil, ErrVaultNotConfigured
	}
	return VaultUnlockResult{
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

func (s *Service) RequestPasswordReset(ctx context.Context, email string) (PasswordResetRequestResult, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return PasswordResetRequestResult{}, nil
	}

	user, err := s.authRepo.GetUserByEmail(ctx, s.db, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PasswordResetRequestResult{}, nil
		}
		return PasswordResetRequestResult{}, err
	}
	if len(user.EncryptedMasterKeyRecovery) == 0 {
		return PasswordResetRequestResult{}, nil
	}

	token, rawHash, err := tokens.Generate()
	if err != nil {
		return PasswordResetRequestResult{}, err
	}
	expiresAt := time.Now().Add(s.resetTTL)
	if err := s.userRepo.SetPasswordResetToken(ctx, s.db, email, hex.EncodeToString(rawHash), expiresAt); err != nil {
		return PasswordResetRequestResult{}, err
	}

	return PasswordResetRequestResult{
		ResetToken: token,
		ResetURL:   s.buildPasswordResetURL(token),
		ExpiresAt:  expiresAt,
	}, nil
}

func (s *Service) LoadPasswordRecoveryVault(ctx context.Context, resetToken string) (PasswordRecoveryVault, error) {
	user, err := s.lookupPasswordResetUser(ctx, s.db, resetToken)
	if err != nil {
		return PasswordRecoveryVault{}, err
	}
	if len(user.VaultSalt) == 0 || len(user.EncryptedMasterKeyRecovery) == 0 {
		return PasswordRecoveryVault{}, ErrVaultNotConfigured
	}
	return PasswordRecoveryVault{
		UserID:                     user.ID,
		VaultSalt:                  user.VaultSalt,
		EncryptedMasterKeyRecovery: user.EncryptedMasterKeyRecovery,
	}, nil
}

func (s *Service) CompletePasswordRecovery(
	ctx context.Context,
	resetToken string,
	newPassword string,
	newVaultSalt []byte,
	newEncryptedMasterKey []byte,
) (validation.Errors, error) {
	resetToken = strings.TrimSpace(resetToken)
	newPassword = strings.TrimSpace(newPassword)
	tokenHash := sha256.Sum256([]byte(resetToken))
	tokenHashHex := hex.EncodeToString(tokenHash[:])

	validationErrors := validation.New()
	if resetToken == "" {
		validationErrors.Add("token", ErrPasswordResetToken.Error())
	}
	if newPassword == "" {
		validationErrors.Add("newPassword", ErrPasswordRequired.Error())
	} else {
		switch validation.PasswordIssueFor(newPassword) {
		case validation.PasswordTooShort:
			validationErrors.Add("newPassword", "password must be at least 8 characters")
		case validation.PasswordMissingLower:
			validationErrors.Add("newPassword", "password must include a lowercase letter")
		case validation.PasswordMissingUpper:
			validationErrors.Add("newPassword", "password must include an uppercase letter")
		case validation.PasswordMissingSymbol:
			validationErrors.Add("newPassword", "password must include a symbol")
		}
	}
	if len(newVaultSalt) != 16 {
		validationErrors.Add("vaultSalt", "vault salt is missing or invalid")
	}
	if len(newEncryptedMasterKey) == 0 {
		validationErrors.Add("encryptedMasterKey", "encrypted master key is missing")
	}
	if validationErrors.HasAny() {
		return validationErrors, nil
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	user, err := s.lookupPasswordResetUser(ctx, tx, resetToken)
	if err != nil {
		if errors.Is(err, ErrPasswordResetToken) {
			validationErrors := validation.New()
			validationErrors.Add("token", ErrPasswordResetToken.Error())
			return validationErrors, nil
		}
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	if err := s.userRepo.CompletePasswordRecovery(ctx, tx, user.ID, tokenHashHex, string(hash), newVaultSalt, newEncryptedMasterKey); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			validationErrors := validation.New()
			validationErrors.Add("token", ErrPasswordResetToken.Error())
			return validationErrors, nil
		}
		return nil, err
	}
	if err := s.sessionRepo.DeleteSessionsByUserID(ctx, tx, user.ID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *Service) authenticateUser(ctx context.Context, db database.PgExecutor, email, password string) (models.User, error) {
	user, err := s.authRepo.GetUserByEmail(ctx, db, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrInvalidCredentials
		}
		return models.User{}, err
	}
	hash, err := s.authRepo.GetPasswordHashByEmail(ctx, db, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrInvalidCredentials
		}
		return models.User{}, err
	}
	if hash == nil {
		return models.User{}, ErrInvalidCredentials
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

func (s *Service) buildPasswordResetURL(token string) string {
	base := s.resetURL
	return fmt.Sprintf("%s/reset-password?token=%s", base, url.QueryEscape(token))
}

func (s *Service) lookupPasswordResetUser(ctx context.Context, db database.PgExecutor, resetToken string) (*models.User, error) {
	resetToken = strings.TrimSpace(resetToken)
	if resetToken == "" {
		return nil, ErrPasswordResetToken
	}
	tokenHash := sha256.Sum256([]byte(resetToken))
	user, err := s.userRepo.FindByPasswordResetTokenHash(ctx, db, hex.EncodeToString(tokenHash[:]))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPasswordResetToken
		}
		return nil, err
	}
	return user, nil
}
