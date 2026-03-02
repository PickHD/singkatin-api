package infrastructure

import (
	"context"
	"fmt"
	"singkatin-api/auth/internal/config"
	"singkatin-api/auth/pkg/logger"

	"gopkg.in/gomail.v2"
)

type EmailProvider struct {
	dialer *gomail.Dialer
	from   string
}

func NewEmailProvider(cfg *config.Config) *EmailProvider {
	d := gomail.NewDialer(cfg.Mailer.Host, cfg.Mailer.Port, cfg.Mailer.Username, cfg.Mailer.Password)

	return &EmailProvider{dialer: d, from: cfg.Mailer.Sender}
}

func (e *EmailProvider) SendEmail(ctx context.Context, to, subject, body string) error {

	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("SINGKATIN System <%s>", e.from))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	if err := e.dialer.DialAndSend(m); err != nil {
		logger.Errorf("failed send email, error: %v", err)
		return err
	}
	return nil
}

func (e *EmailProvider) Close() error {
	return nil
}

func (e *EmailProvider) GetDialer() *gomail.Dialer {
	return e.dialer
}
