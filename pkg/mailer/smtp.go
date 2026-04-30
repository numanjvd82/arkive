package mailer

import (
	"fmt"
	"net/smtp"
	"strings"
)

type SMTPConfig struct {
	Host string
	Port int
	User string
	Pass string
	From string
}

type SMTPMailer struct {
	config SMTPConfig
}

func NewSMTPMailer(cfg SMTPConfig) (*SMTPMailer, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return nil, fmt.Errorf("smtp host is required")
	}
	if cfg.Port <= 0 {
		cfg.Port = 587
	}
	if strings.TrimSpace(cfg.From) == "" {
		return nil, fmt.Errorf("smtp from address is required")
	}
	return &SMTPMailer{config: cfg}, nil
}

func (m *SMTPMailer) Send(msg Message) error {
	to := strings.TrimSpace(msg.To)
	if to == "" {
		return fmt.Errorf("recipient is required")
	}

	subject := strings.TrimSpace(msg.Subject)
	if subject == "" {
		subject = "(no subject)"
	}

	html := strings.TrimSpace(msg.HTML)

	headers := make([]string, 0, 6)
	headers = append(headers, "From: "+m.config.From)
	headers = append(headers, "To: "+to)
	headers = append(headers, "Subject: "+subject)
	headers = append(headers, "MIME-Version: 1.0")
	headers = append(headers, "Content-Type: text/html; charset=\"utf-8\"")
	headers = append(headers, "")
	headers = append(headers, html)

	body := []byte(strings.Join(headers, "\r\n"))

	addr := fmt.Sprintf("%s:%d", m.config.Host, m.config.Port)

	var auth smtp.Auth
	if m.config.User != "" || m.config.Pass != "" {
		auth = smtp.PlainAuth("", m.config.User, m.config.Pass, m.config.Host)
	}

	return smtp.SendMail(addr, auth, m.config.From, []string{to}, body)
}
