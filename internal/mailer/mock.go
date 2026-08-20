package mailer

import (
	"context"
	"sync"

	"github.com/sall-lah/store_notification/internal/domain"
)

// MockMailer records sent messages in memory for unit and integration testing.
type MockMailer struct {
	mu           sync.Mutex
	SentMessages []*domain.EmailMessage
	ShouldFail   bool
	FailErr      error
}

// NewMockMailer creates a thread-safe mock mailer.
func NewMockMailer() *MockMailer {
	return &MockMailer{
		SentMessages: make([]*domain.EmailMessage, 0),
	}
}

// Send records the email into SentMessages slice.
func (m *MockMailer) Send(ctx context.Context, msg *domain.EmailMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ShouldFail {
		if m.FailErr != nil {
			return m.FailErr
		}
		return context.DeadlineExceeded
	}

	m.SentMessages = append(m.SentMessages, msg)
	return nil
}

// Ping verifies mock mailer health.
func (m *MockMailer) Ping(ctx context.Context) error {
	if m.ShouldFail {
		return context.DeadlineExceeded
	}
	return nil
}

// Close resets the mock mailer.
func (m *MockMailer) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SentMessages = nil
	return nil
}

// GetSentCount returns the number of messages sent through this mock.
func (m *MockMailer) GetSentCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.SentMessages)
}

// GetLastMessage returns the most recent message sent, or nil if none.
func (m *MockMailer) GetLastMessage() *domain.EmailMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.SentMessages) == 0 {
		return nil
	}
	return m.SentMessages[len(m.SentMessages)-1]
}
