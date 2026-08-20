package mailer

import (
	"context"

	"github.com/sall-lah/store_notification/internal/domain"
)

// Mailer defines the outbound email dispatch interface.
// It abstracts SMTP or third-party email providers (SES, Resend) behind a unified contract.
type Mailer interface {
	// Send delivers an email message synchronously with retry capabilities.
	Send(ctx context.Context, msg *domain.EmailMessage) error

	// Ping verifies connection capability to the email server.
	Ping(ctx context.Context) error

	// Close terminates any persistent connections or connection pools.
	Close() error
}
