package mailer

import "errors"

type Message struct {
	To      string
	Subject string
	HTML    string
}

type Mailer interface {
	Send(msg Message) error
}

type Config struct {
	Provider string
	From     string
	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
}

func NewMailerFromConfig(cfg Config) (Mailer, error) {
	switch cfg.Provider {
	case "smtp":
		return NewSMTPMailer(SMTPConfig{
			Host: cfg.SMTPHost,
			Port: cfg.SMTPPort,
			User: cfg.SMTPUser,
			Pass: cfg.SMTPPass,
			From: cfg.From,
		})
	case "noop":
		return NewNoopMailer(), nil
	default:
		if cfg.Provider == "" {
			return NewNoopMailer(), nil
		}
		return nil, errors.New("unknown email provider: " + cfg.Provider)
	}
}
