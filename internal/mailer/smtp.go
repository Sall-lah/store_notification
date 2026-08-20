package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/sall-lah/store_notification/internal/config"
	"github.com/sall-lah/store_notification/internal/domain"
	"github.com/wneessen/go-mail"
)

// SMTPMailer implements the Mailer interface using standard SMTP transport.
type SMTPMailer struct {
	client    *mail.Client
	fromEmail string
	fromName  string
}

// NewSMTPMailer constructs a configured SMTP mailer client.
func NewSMTPMailer(cfg *config.Config) (*SMTPMailer, error) {
	clientOpts := []mail.Option{
		mail.WithPort(cfg.SMTPPort),
		mail.WithTimeout(15 * time.Second),
	}

	if cfg.SMTPUser != "" {
		clientOpts = append(clientOpts,
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(cfg.SMTPUser),
			mail.WithPassword(cfg.SMTPPassword),
		)
	}

	tlsConfig := &tls.Config{
		ServerName:         cfg.SMTPHost,
		InsecureSkipVerify: cfg.SMTPInsecureSkipVerify,
	}

	if cfg.SMTPPort == 465 {
		clientOpts = append(clientOpts,
			mail.WithSSLPort(true),
			mail.WithTLSPolicy(mail.TLSMandatory),
			mail.WithTLSConfig(tlsConfig),
		)
	} else if cfg.SMTPPort == 587 || cfg.SMTPRequireTLS {
		clientOpts = append(clientOpts,
			mail.WithTLSPolicy(mail.TLSMandatory),
			mail.WithTLSConfig(tlsConfig),
		)
	} else {
		clientOpts = append(clientOpts,
			mail.WithTLSPolicy(mail.TLSOpportunistic),
			mail.WithTLSConfig(tlsConfig),
		)
	}

	client, err := mail.NewClient(cfg.SMTPHost, clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create SMTP client: %w", err)
	}

	return &SMTPMailer{
		client:    client,
		fromEmail: cfg.SMTPFromEmail,
		fromName:  cfg.SMTPFromName,
	}, nil
}

// Send sends an email with retry handling for transient connectivity errors.
func (s *SMTPMailer) Send(ctx context.Context, msg *domain.EmailMessage) error {
	m := mail.NewMsg()

	fromFormatted := fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	if s.fromName == "" {
		fromFormatted = s.fromEmail
	}

	if err := m.From(fromFormatted); err != nil {
		return fmt.Errorf("invalid from address: %w", err)
	}

	if err := m.To(msg.To...); err != nil {
		return fmt.Errorf("invalid to recipient(s): %w", err)
	}

	if len(msg.Cc) > 0 {
		_ = m.Cc(msg.Cc...)
	}
	if len(msg.Bcc) > 0 {
		_ = m.Bcc(msg.Bcc...)
	}

	m.Subject(msg.Subject)

	if msg.HTMLBody != "" {
		m.SetBodyString(mail.TypeTextHTML, msg.HTMLBody)
		if msg.TextBody != "" {
			m.AddAlternativeString(mail.TypeTextPlain, msg.TextBody)
		}
	} else if msg.TextBody != "" {
		m.SetBodyString(mail.TypeTextPlain, msg.TextBody)
	}

	// Dispatch with retry on transient network errors
	var lastErr error
	maxRetries := 2
	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := s.client.DialAndSendWithContext(ctx, m); err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt*500) * time.Millisecond)
			continue
		}
		return nil
	}

	return fmt.Errorf("failed to deliver email: %w", lastErr)
}

// Ping tests connection capability to the SMTP host.
func (s *SMTPMailer) Ping(ctx context.Context) error {
	return s.client.DialWithContext(ctx)
}

// Close closes the SMTP mailer client.
func (s *SMTPMailer) Close() error {
	return s.client.Close()
}
