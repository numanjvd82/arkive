package mailer

import (
	"log"
	"strings"
)

type NoopMailer struct{}

func NewNoopMailer() *NoopMailer {
	return &NoopMailer{}
}

func (m *NoopMailer) Send(msg Message) error {
	to := strings.TrimSpace(msg.To)
	if to == "" {
		to = "(no recipient)"
	}
	subject := strings.TrimSpace(msg.Subject)
	if subject == "" {
		subject = "(no subject)"
	}
	log.Printf("[mailer] TO: %s | SUBJECT: %s | HTML: %d bytes", to, subject, len(msg.HTML))
	return nil
}
