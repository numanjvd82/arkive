package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"arkive/pkg/tokens"
)

var (
	ErrVerifyTokenInvalid = errors.New("verification link is invalid or expired")
)

type EmailSender interface {
	SendVerifyEmail(to, verifyURL string) error
}

type VerifyConfig struct {
	PublicBaseURL string
}

func (s *Service) ConfigureEmailVerification(sender EmailSender, cfg VerifyConfig) {
	s.emailSender = sender
	s.publicBaseURL = strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/")
}

func (s *Service) sendEmailVerification(ctx context.Context, tx pgx.Tx, userID, email string) error {
	if s.emailSender == nil {
		return nil
	}
	if s.publicBaseURL == "" {
		return nil
	}

	token, hash, err := tokens.Generate()
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.emailVerifyRepo.CreateToken(ctx, tx, userID, hash, expiresAt); err != nil {
		return err
	}

	verifyURL := s.publicBaseURL + "/verify-email?token=" + url.QueryEscape(token)
	return s.emailSender.SendVerifyEmail(email, verifyURL)
}

func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return ErrVerifyTokenInvalid
	}
	hash := sha256TokenHash(token)
	now := time.Now()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	userID, err := s.emailVerifyRepo.ConsumeToken(ctx, tx, hash, now)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Best-effort cleanup: if the link is expired/used, delete it.
			_ = s.emailVerifyRepo.DeleteIfExpiredOrUsed(ctx, tx, hash, now)
			return ErrVerifyTokenInvalid
		}
		return err
	}

	if err := s.emailVerifyRepo.MarkEmailVerified(ctx, tx, userID); err != nil {
		return err
	}
	// Clean up token row after successful verification.
	if err := s.emailVerifyRepo.DeleteByHash(ctx, tx, hash); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func sha256TokenHash(token string) []byte {
	// Keep token hashing consistent with pkg/tokens.Generate.
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}
