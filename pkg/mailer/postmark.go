package mailer

import (
	"errors"
	"strings"

	"github.com/keighl/postmark"
)

type PostmarkMailer struct {
	client *postmark.Client
	from   string
}

func NewPostmarkMailer(serverToken, from string) (*PostmarkMailer, error) {
	serverToken = strings.TrimSpace(serverToken)
	if serverToken == "" {
		return nil, errors.New("postmark server token is required")
	}
	from = strings.TrimSpace(from)
	if from == "" {
		return nil, errors.New("postmark from address is required")
	}
	return &PostmarkMailer{
		client: postmark.NewClient(serverToken, ""),
		from:   from,
	}, nil
}

func (m *PostmarkMailer) Send(msg Message) error {
	to := strings.TrimSpace(msg.To)
	if to == "" {
		return errors.New("recipient is required")
	}

	_, err := m.client.SendEmail(postmark.Email{
		From:     m.from,
		To:       to,
		Subject:  strings.TrimSpace(msg.Subject),
		HtmlBody: msg.HTML,
	})
	return err
}
