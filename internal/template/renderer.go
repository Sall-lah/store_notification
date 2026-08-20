package template

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	textTemplate "text/template"

	"github.com/sall-lah/store_notification/internal/domain"
	"github.com/sall-lah/store_notification/templates"
)

// Renderer provides thread-safe template rendering for all transactional email types.
type Renderer struct{}

// NewRenderer instantiates the template rendering engine.
func NewRenderer() *Renderer {
	return &Renderer{}
}

// RenderHTML loads the shared layout wrapper along with the specified content template
// and renders them with the provided data context.
func (r *Renderer) RenderHTML(contentFile string, data any) (string, error) {
	tmpl, err := htmlTemplate.New("base.html").ParseFS(templates.FS, "layout/base.html", contentFile)
	if err != nil {
		return "", fmt.Errorf("failed to parse html templates (%s): %w", contentFile, err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base.html", data); err != nil {
		return "", fmt.Errorf("failed to execute html template (%s): %w", contentFile, err)
	}

	return buf.String(), nil
}

// RenderText parses and executes a plain-text template.
func (r *Renderer) RenderText(contentFile string, data any) (string, error) {
	tmpl, err := textTemplate.ParseFS(templates.FS, contentFile)
	if err != nil {
		return "", fmt.Errorf("failed to parse text template (%s): %w", contentFile, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute text template (%s): %w", contentFile, err)
	}

	return buf.String(), nil
}

// RenderAuthRegistrationOTP renders registration OTP emails in HTML and text formats.
func (r *Renderer) RenderAuthRegistrationOTP(ctx domain.AuthEmailContext) (string, string, error) {
	htmlBody, err := r.RenderHTML("auth/registration_otp.html", ctx)
	if err != nil {
		return "", "", err
	}
	textBody, err := r.RenderText("auth/registration_otp.txt", ctx)
	if err != nil {
		return htmlBody, "", nil // Graceful fallback if text template missing
	}
	return htmlBody, textBody, nil
}

// RenderAuthPasswordResetOTP renders password reset OTP emails in HTML and text formats.
func (r *Renderer) RenderAuthPasswordResetOTP(ctx domain.AuthEmailContext) (string, string, error) {
	htmlBody, err := r.RenderHTML("auth/password_reset_otp.html", ctx)
	if err != nil {
		return "", "", err
	}
	textBody, err := r.RenderText("auth/password_reset_otp.txt", ctx)
	if err != nil {
		return htmlBody, "", nil
	}
	return htmlBody, textBody, nil
}

// RenderOrderCreated renders order created customer email.
func (r *Renderer) RenderOrderCreated(ctx domain.OrderEmailContext) (string, string, error) {
	htmlBody, err := r.RenderHTML("order/order_created.html", ctx)
	if err != nil {
		return "", "", err
	}
	return htmlBody, "", nil
}

// RenderOrderPaid renders order paid customer receipt.
func (r *Renderer) RenderOrderPaid(ctx domain.OrderEmailContext) (string, string, error) {
	htmlBody, err := r.RenderHTML("order/order_paid.html", ctx)
	if err != nil {
		return "", "", err
	}
	return htmlBody, "", nil
}

// RenderOrderCancelled renders order cancellation notice.
func (r *Renderer) RenderOrderCancelled(ctx domain.OrderEmailContext) (string, string, error) {
	htmlBody, err := r.RenderHTML("order/order_cancelled.html", ctx)
	if err != nil {
		return "", "", err
	}
	return htmlBody, "", nil
}

// RenderOrderFulfilled renders shipping & fulfillment confirmation notice.
func (r *Renderer) RenderOrderFulfilled(ctx domain.OrderEmailContext) (string, string, error) {
	htmlBody, err := r.RenderHTML("order/order_fulfilled.html", ctx)
	if err != nil {
		return "", "", err
	}
	return htmlBody, "", nil
}

// RenderAdminOrderAlert renders store admin alert for paid orders.
func (r *Renderer) RenderAdminOrderAlert(ctx domain.AdminOrderAlertContext) (string, string, error) {
	htmlBody, err := r.RenderHTML("admin/admin_order_alert.html", ctx)
	if err != nil {
		return "", "", err
	}
	return htmlBody, "", nil
}
