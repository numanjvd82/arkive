package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/jackc/pgx/v5"

	"arkive/core/services/auth/templates"
	"arkive/pkg/mailer"
	"arkive/pkg/tokens"
)

var (
	ErrVerifyTokenInvalid = errors.New("verification link is invalid or expired")
)

func (s *Service) sendEmailVerification(ctx context.Context, tx pgx.Tx, userID, email string) error {
	if s.mailer == nil {
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

	tmplBytes, err := templates.FS.ReadFile("verify_email.html")
	if err != nil {
		return err
	}
	tmpl, err := template.New("verify_email").Parse(string(tmplBytes))
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct{ VerifyURL string }{VerifyURL: verifyURL}); err != nil {
		return err
	}

	return s.mailer.Send(mailer.Message{
		To:      email,
		Subject: "Confirm your Arkive email",
		HTML:    buf.String(),
	})
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
