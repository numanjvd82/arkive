package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"arkive/core/database"
	"arkive/core/models"
	authrepo "arkive/core/repositories/auth"
	sessionrepo "arkive/core/repositories/session"
	"arkive/pkg/tokens"
	"arkive/pkg/validation"
)

type Service struct {
	db             database.PgPool
	authRepo       *authrepo.Repository
	sessionRepo    *sessionrepo.Repository
	sessionTTL     time.Duration
	googleClientID string
}

type Config struct {
	SessionTTL     time.Duration
	GoogleClientID string
}

func NewService(
	db database.PgPool,
	authRepo *authrepo.Repository,
	sessionRepo *sessionrepo.Repository,
	cfg Config,
) *Service {
	return &Service{
		db:             db,
		authRepo:       authRepo,
		sessionRepo:    sessionRepo,
		sessionTTL:     cfg.SessionTTL,
		googleClientID: cfg.GoogleClientID,
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
		if errors.Is(err, ErrLoginUseGoogle) {
			validationErrors := validation.New()
			validationErrors.Add(validation.GeneralKey, ErrLoginUseGoogle.Error())
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

	if hash == nil || *hash == "" {
		return models.User{}, ErrLoginUseGoogle
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*hash), []byte(password)); err != nil {
		return models.User{}, ErrInvalidCredentials
	}

	return user, nil
}

func (s *Service) WebGoogleLogin(ctx context.Context, credential string) (string, time.Time, error) {
	credential = strings.TrimSpace(credential)
	if credential == "" {
		return "", time.Time{}, ErrInvalidInput
	}
	if s.googleClientID == "" {
		return "", time.Time{}, ErrGoogleClientNotConfigured
	}

	payload, err := s.fetchGoogleTokenInfo(ctx, credential)
	if err != nil {
		return "", time.Time{}, ErrGoogleTokenInvalid
	}

	email := strings.TrimSpace(payload.Email)
	sub := strings.TrimSpace(payload.Sub)
	givenName := strings.TrimSpace(payload.GivenName)
	familyName := strings.TrimSpace(payload.FamilyName)
	pictureURL := strings.TrimSpace(payload.PictureURL)
	emailVerified := payload.EmailVerified

	if email == "" || sub == "" {
		return "", time.Time{}, ErrGoogleTokenInvalid
	}
	if !emailVerified {
		return "", time.Time{}, ErrGoogleEmailNotVerified
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", time.Time{}, err
	}

	user, err := s.authRepo.GetUserByGoogleSub(ctx, tx, sub)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		_ = tx.Rollback(ctx)
		return "", time.Time{}, err
	}

	if errors.Is(err, pgx.ErrNoRows) {
		_, _, err := s.authRepo.GetUserByEmail(ctx, tx, email)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			_ = tx.Rollback(ctx)
			return "", time.Time{}, err
		}

		if !errors.Is(err, pgx.ErrNoRows) {
			_ = tx.Rollback(ctx)
			return "", time.Time{}, ErrGoogleEmailHasPassword
		}

		displayName := strings.TrimSpace(strings.Join([]string{givenName, familyName}, " "))
		brandName, err := s.uniqueBrandName(ctx, tx, displayName, email)
		if err != nil {
			_ = tx.Rollback(ctx)
			return "", time.Time{}, err
		}

		user, err = s.authRepo.CreateUserWithGoogleProfile(ctx, tx, brandName, email, sub, givenName, familyName, emailVerified, pictureURL)
		if err != nil {
			_ = tx.Rollback(ctx)
			return "", time.Time{}, err
		}
	}

	sessionID, expiresAt, err := s.createSession(ctx, tx, user.ID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return "", time.Time{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", time.Time{}, err
	}

	return sessionID, expiresAt, nil
}

type googleTokenInfo struct {
	Aud           string `json:"aud"`
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	PictureURL    string `json:"picture"`
	EmailVerified bool   `json:"-"`
}

func (s *Service) fetchGoogleTokenInfo(ctx context.Context, credential string) (googleTokenInfo, error) {
	values := url.Values{}
	values.Set("id_token", credential)
	endpoint := "https://oauth2.googleapis.com/tokeninfo?" + values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return googleTokenInfo{}, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return googleTokenInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return googleTokenInfo{}, ErrGoogleTokenInvalid
	}

	var raw struct {
		Aud           string      `json:"aud"`
		Sub           string      `json:"sub"`
		Email         string      `json:"email"`
		GivenName     string      `json:"given_name"`
		FamilyName    string      `json:"family_name"`
		PictureURL    string      `json:"picture"`
		EmailVerified interface{} `json:"email_verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return googleTokenInfo{}, err
	}

	if raw.Aud != s.googleClientID {
		return googleTokenInfo{}, ErrGoogleTokenInvalid
	}

	verified := false
	switch value := raw.EmailVerified.(type) {
	case bool:
		verified = value
	case string:
		verified = strings.EqualFold(value, "true")
	}

	return googleTokenInfo{
		Aud:           raw.Aud,
		Sub:           raw.Sub,
		Email:         raw.Email,
		GivenName:     raw.GivenName,
		FamilyName:    raw.FamilyName,
		PictureURL:    raw.PictureURL,
		EmailVerified: verified,
	}, nil
}

func (s *Service) uniqueBrandName(ctx context.Context, db database.PgExecutor, name, email string) (string, error) {
	base := strings.TrimSpace(name)
	if base == "" {
		base = strings.TrimSpace(strings.Split(email, "@")[0])
	}
	if base == "" {
		base = "Arkive"
	}

	candidate := base
	for i := 0; i < 5; i++ {
		if _, err := s.authRepo.GetUserByBrandName(ctx, db, candidate); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return candidate, nil
			}
			return "", err
		}
		candidate = fmt.Sprintf("%s-%d", base, i+2)
	}

	token, _, err := tokens.Generate()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", base, token[:6]), nil
}

func (s *Service) createSession(ctx context.Context, db database.PgExecutor, userID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.sessionTTL)
	sessionID, err := s.sessionRepo.CreateSession(ctx, db, userID, expiresAt)
	if err != nil {
		return "", time.Time{}, err
	}
	return sessionID, expiresAt, nil
}

func (s *Service) GoogleClientID() string {
	return s.googleClientID
}
