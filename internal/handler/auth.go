package handler

import (
	"context"
	"fmt"

	"github.com/sall-lah/store_notification/internal/domain"
	"github.com/sall-lah/store_notification/internal/mailer"
	"github.com/sall-lah/store_notification/internal/template"
)

// AuthHandler processes authentication lifecycle events and dispatches security emails.
type AuthHandler struct {
	renderer *template.Renderer
	mailer   mailer.Mailer
	appName  string
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(renderer *template.Renderer, mailer mailer.Mailer, appName string) *AuthHandler {
	if appName == "" {
		appName = "Store Platform"
	}
	return &AuthHandler{
		renderer: renderer,
		mailer:   mailer,
		appName:  appName,
	}
}

// HandleRegistrationOTP renders and sends the 6-digit registration verification OTP email.
func (h *AuthHandler) HandleRegistrationOTP(ctx context.Context, data *domain.AuthOtpEventData) error {
	if err := data.Validate(); err != nil {
		return fmt.Errorf("invalid registration OTP payload: %w", err)
	}

	emailCtx := domain.AuthEmailContext{
		Name:       data.Name,
		Email:      data.Email,
		OTPCode:    data.Code,
		ExpiryMins: 15,
		AppName:    h.appName,
	}

	htmlBody, textBody, err := h.renderer.RenderAuthRegistrationOTP(emailCtx)
	if err != nil {
		return fmt.Errorf("failed to render registration OTP template: %w", err)
	}

	msg := &domain.EmailMessage{
		To:       []string{data.Email},
		Subject:  fmt.Sprintf("[%s] Verify your email address", h.appName),
		HTMLBody: htmlBody,
		TextBody: textBody,
	}

	return h.mailer.Send(ctx, msg)
}

// HandlePasswordResetOTP renders and sends the password reset recovery OTP email.
func (h *AuthHandler) HandlePasswordResetOTP(ctx context.Context, data *domain.AuthOtpEventData) error {
	if err := data.Validate(); err != nil {
		return fmt.Errorf("invalid password reset OTP payload: %w", err)
	}

	emailCtx := domain.AuthEmailContext{
		Name:       data.Name,
		Email:      data.Email,
		OTPCode:    data.Code,
		ExpiryMins: 15,
		AppName:    h.appName,
	}

	htmlBody, textBody, err := h.renderer.RenderAuthPasswordResetOTP(emailCtx)
	if err != nil {
		return fmt.Errorf("failed to render password reset OTP template: %w", err)
	}

	msg := &domain.EmailMessage{
		To:       []string{data.Email},
		Subject:  fmt.Sprintf("[%s] Reset your password", h.appName),
		HTMLBody: htmlBody,
		TextBody: textBody,
	}

	return h.mailer.Send(ctx, msg)
}
