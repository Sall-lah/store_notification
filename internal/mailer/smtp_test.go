package mailer

import (
	"context"
	"testing"

	"github.com/sall-lah/store_notification/internal/domain"
)

func TestMockMailerSend(t *testing.T) {
	mock := NewMockMailer()
	ctx := context.Background()

	msg := &domain.EmailMessage{
		To:       []string{"test@example.com"},
		Subject:  "Welcome!",
		HTMLBody: "<h1>Hello</h1>",
		TextBody: "Hello",
	}

	if err := mock.Send(ctx, msg); err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}

	if mock.GetSentCount() != 1 {
		t.Fatalf("expected 1 sent message, got %d", mock.GetSentCount())
	}

	last := mock.GetLastMessage()
	if last.Subject != "Welcome!" || len(last.To) != 1 || last.To[0] != "test@example.com" {
		t.Errorf("unexpected message recorded: %+v", last)
	}
}
