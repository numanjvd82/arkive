package email

import (
	"errors"
	"strings"

	"github.com/keighl/postmark"
)

type Sender struct {
	client *postmark.Client
}

func NewPostmarkSender(serverToken string) (*Sender, error) {
	serverToken = strings.TrimSpace(serverToken)
	if serverToken == "" {
		return nil, errors.New("missing postmark server token")
	}
	return &Sender{client: postmark.NewClient(serverToken, "")}, nil
}

func (s *Sender) SendVerifyEmail(to, verifyURL string) error {
	to = strings.TrimSpace(to)
	verifyURL = strings.TrimSpace(verifyURL)
	if to == "" || verifyURL == "" {
		return errors.New("missing to or verify url")
	}

	htmlBody, err := RenderHTMLTemplate("verify_email.html", struct {
		Subject   string
		VerifyURL string
	}{
		Subject:   "Confirm your Arkive email",
		VerifyURL: verifyURL,
	})
	if err != nil {
		return err
	}

	_, err = s.client.SendEmail(postmark.Email{
		From:     "Arkive <no-reply@arkive.sh>",
		To:       to,
		Subject:  "Confirm your Arkive email",
		TextBody: "Confirm your email address by opening this link:\n\n" + verifyURL + "\n\nThis link expires in 24 hours.",
		HtmlBody: htmlBody,
		Tag:      "email_verification",
	})
	return err
}
