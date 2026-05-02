package setup

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"arkive/core/database"
	"arkive/core/models"
	authrepo "arkive/core/repositories/auth"
	usersrepo "arkive/core/repositories/users"
	"arkive/pkg/validation"
)

var ErrAlreadyInitialized = errors.New("instance already initialized")

type Service struct {
	db       database.PgPool
	authRepo *authrepo.Repository
	userRepo *usersrepo.Repository
}

func NewService(db database.PgPool, authRepo *authrepo.Repository, userRepo *usersrepo.Repository) *Service {
	return &Service{
		db:       db,
		authRepo: authRepo,
		userRepo: userRepo,
	}
}

func (s *Service) IsInitialized(ctx context.Context) (bool, error) {
	count, err := s.userRepo.CountUsers(ctx, s.db)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Service) CreateInitialAdmin(ctx context.Context, brandName, email, password, confirmPassword string) (models.User, validation.Errors, error) {
	brandName = strings.TrimSpace(brandName)
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	confirmPassword = strings.TrimSpace(confirmPassword)

	validationErrors := validateAdminInput(brandName, email, password, confirmPassword)
	if validationErrors.HasAny() {
		return models.User{}, validationErrors, nil
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.User{}, nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	count, err := s.userRepo.CountUsers(ctx, tx)
	if err != nil {
		return models.User{}, nil, err
	}
	if count > 0 {
		return models.User{}, nil, ErrAlreadyInitialized
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, nil, err
	}

	user, err := s.authRepo.CreateVerifiedUser(ctx, tx, brandName, email, string(hash))
	if err != nil {
		return models.User{}, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.User{}, nil, err
	}
	return user, nil, nil
}

func validateAdminInput(brandName, email, password, confirmPassword string) validation.Errors {
	validationErrors := validation.New()
	if brandName == "" {
		validationErrors.Add("brand_name", "brand name is required")
	}
	if email == "" {
		validationErrors.Add("email", "email is required")
	}
	if password == "" {
		validationErrors.Add("password", "password is required")
	}
	if confirmPassword == "" {
		validationErrors.Add("confirm_password", "confirm password is required")
	}
	if password != "" {
		switch validation.PasswordIssueFor(password) {
		case validation.PasswordTooShort:
			validationErrors.Add("password", "password must be at least 8 characters")
		case validation.PasswordMissingLower:
			validationErrors.Add("password", "password must include a lowercase letter")
		case validation.PasswordMissingUpper:
			validationErrors.Add("password", "password must include an uppercase letter")
		case validation.PasswordMissingSymbol:
			validationErrors.Add("password", "password must include a symbol")
		}
	}
	if password != "" && confirmPassword != "" && password != confirmPassword {
		validationErrors.Add("confirm_password", "passwords do not match")
	}
	return validationErrors
}
