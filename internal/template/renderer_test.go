package template

import (
	"strings"
	"testing"

	"github.com/sall-lah/store_notification/internal/domain"
)

func TestRenderAuthRegistrationOTP(t *testing.T) {
	renderer := NewRenderer()
	ctx := domain.AuthEmailContext{
		Name:       "Jane Doe",
		Email:      "jane@example.com",
		OTPCode:    "847291",
		ExpiryMins: 15,
		AppName:    "Acme Store",
	}

	html, text, err := renderer.RenderAuthRegistrationOTP(ctx)
	if err != nil {
		t.Fatalf("failed to render registration OTP email: %v", err)
	}

	if !strings.Contains(html, "847291") {
		t.Errorf("expected HTML to contain OTP code 847291, got: %s", html)
	}
	if !strings.Contains(html, "Acme Store") {
		t.Errorf("expected HTML to contain AppName Acme Store")
	}
	if !strings.Contains(text, "847291") {
		t.Errorf("expected text to contain OTP code 847291, got: %s", text)
	}
}

func TestRenderAdminOrderAlert(t *testing.T) {
	renderer := NewRenderer()
	ctx := domain.AdminOrderAlertContext{
		OrderNumber:     "ORD-2026-99",
		CustomerEmail:   "buyer@example.com",
		TotalAmount:     249.99,
		ItemCount:       1,
		PaidAtFormatted: "2026-08-20 12:00:00",
		Items: []domain.OrderItemData{
			{
				ProductName: "Mechanical Keyboard",
				Price:       249.99,
				Quantity:    1,
				Subtotal:    249.99,
			},
		},
		AppName: "Acme Store",
	}

	html, _, err := renderer.RenderAdminOrderAlert(ctx)
	if err != nil {
		t.Fatalf("failed to render admin order alert: %v", err)
	}

	if !strings.Contains(html, "ORD-2026-99") {
		t.Errorf("expected HTML to contain order number ORD-2026-99")
	}
	if !strings.Contains(html, "Mechanical Keyboard") {
		t.Errorf("expected HTML to contain product name Mechanical Keyboard")
	}
}
