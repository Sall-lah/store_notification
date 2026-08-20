package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/sall-lah/store_notification/internal/domain"
	"github.com/sall-lah/store_notification/internal/mailer"
	"github.com/sall-lah/store_notification/internal/template"
)

// OrderHandler processes order lifecycle events and dispatches customer and administrator emails.
type OrderHandler struct {
	renderer    *template.Renderer
	mailer      mailer.Mailer
	adminEmails []string
	appName     string
}

// NewOrderHandler constructs an OrderHandler.
func NewOrderHandler(renderer *template.Renderer, mailer mailer.Mailer, adminEmails []string, appName string) *OrderHandler {
	if appName == "" {
		appName = "Store Platform"
	}
	return &OrderHandler{
		renderer:    renderer,
		mailer:      mailer,
		adminEmails: adminEmails,
		appName:     appName,
	}
}

// HandleOrderCreated dispatches payment instructions and order breakdown to the customer.
func (h *OrderHandler) HandleOrderCreated(ctx context.Context, data *domain.OrderEventData) error {
	if err := data.Validate(); err != nil {
		return fmt.Errorf("invalid order.created payload: %w", err)
	}

	emailCtx := domain.OrderEmailContext{
		OrderNumber:     data.OrderNumber,
		CustomerEmail:   data.UserEmail,
		Status:          data.Status,
		TotalAmount:     data.TotalAmount,
		ShippingFee:     data.ShippingFee,
		ShippingAddress: data.ShippingAddress,
		PaymentType:     data.PaymentType,
		SnapRedirectURL: data.SnapRedirectURL,
		Items:           data.Items,
		AppName:         h.appName,
	}

	htmlBody, textBody, err := h.renderer.RenderOrderCreated(emailCtx)
	if err != nil {
		return fmt.Errorf("failed to render order_created template: %w", err)
	}

	msg := &domain.EmailMessage{
		To:       []string{data.UserEmail},
		Subject:  fmt.Sprintf("[%s] Order #%s Confirmation & Payment", h.appName, data.OrderNumber),
		HTMLBody: htmlBody,
		TextBody: textBody,
	}

	return h.mailer.Send(ctx, msg)
}

// HandleOrderPaid sends a payment receipt to the customer AND a new order alert to store administrators.
func (h *OrderHandler) HandleOrderPaid(ctx context.Context, data *domain.OrderEventData) error {
	if err := data.Validate(); err != nil {
		return fmt.Errorf("invalid order.paid payload: %w", err)
	}

	paidTimeStr := time.Now().UTC().Format("2006-01-02 15:04:05 MST")
	if data.PaidAt != nil {
		paidTimeStr = data.PaidAt.UTC().Format("2006-01-02 15:04:05 MST")
	}

	// 1. Send receipt to customer
	customerCtx := domain.OrderEmailContext{
		OrderNumber:     data.OrderNumber,
		CustomerEmail:   data.UserEmail,
		Status:          "PAID",
		TotalAmount:     data.TotalAmount,
		ShippingFee:     data.ShippingFee,
		ShippingAddress: data.ShippingAddress,
		PaymentType:     data.PaymentType,
		PaidAtFormatted: paidTimeStr,
		Items:           data.Items,
		AppName:         h.appName,
	}

	htmlBody, textBody, err := h.renderer.RenderOrderPaid(customerCtx)
	if err != nil {
		return fmt.Errorf("failed to render order_paid template: %w", err)
	}

	customerMsg := &domain.EmailMessage{
		To:       []string{data.UserEmail},
		Subject:  fmt.Sprintf("[%s] Payment Confirmed: Order #%s", h.appName, data.OrderNumber),
		HTMLBody: htmlBody,
		TextBody: textBody,
	}

	if err := h.mailer.Send(ctx, customerMsg); err != nil {
		return fmt.Errorf("failed to send customer receipt: %w", err)
	}

	// 2. Send alert to store administrators if configured
	if len(h.adminEmails) > 0 {
		adminCtx := domain.AdminOrderAlertContext{
			OrderNumber:     data.OrderNumber,
			CustomerEmail:   data.UserEmail,
			TotalAmount:     data.TotalAmount,
			ItemCount:       len(data.Items),
			Items:           data.Items,
			PaidAtFormatted: paidTimeStr,
			AppName:         h.appName,
		}

		adminHTML, adminText, err := h.renderer.RenderAdminOrderAlert(adminCtx)
		if err == nil {
			adminMsg := &domain.EmailMessage{
				To:       h.adminEmails,
				Subject:  fmt.Sprintf("⚡ [Admin Alert] New Paid Order #%s ($%.2f)", data.OrderNumber, data.TotalAmount),
				HTMLBody: adminHTML,
				TextBody: adminText,
			}
			_ = h.mailer.Send(ctx, adminMsg)
		}
	}

	return nil
}

// HandleOrderCancelled sends a cancellation notification to the customer.
func (h *OrderHandler) HandleOrderCancelled(ctx context.Context, data *domain.OrderEventData) error {
	if err := data.Validate(); err != nil {
		return fmt.Errorf("invalid order.cancelled payload: %w", err)
	}

	emailCtx := domain.OrderEmailContext{
		OrderNumber:   data.OrderNumber,
		CustomerEmail: data.UserEmail,
		Status:        "CANCELLED",
		TotalAmount:   data.TotalAmount,
		Reason:        data.Reason,
		AppName:       h.appName,
	}

	htmlBody, textBody, err := h.renderer.RenderOrderCancelled(emailCtx)
	if err != nil {
		return fmt.Errorf("failed to render order_cancelled template: %w", err)
	}

	msg := &domain.EmailMessage{
		To:       []string{data.UserEmail},
		Subject:  fmt.Sprintf("[%s] Order #%s Cancelled", h.appName, data.OrderNumber),
		HTMLBody: htmlBody,
		TextBody: textBody,
	}

	return h.mailer.Send(ctx, msg)
}

// HandleOrderFulfilled sends delivery/shipping confirmation to the customer.
func (h *OrderHandler) HandleOrderFulfilled(ctx context.Context, data *domain.OrderEventData) error {
	if err := data.Validate(); err != nil {
		return fmt.Errorf("invalid order.fulfilled payload: %w", err)
	}

	emailCtx := domain.OrderEmailContext{
		OrderNumber:     data.OrderNumber,
		CustomerEmail:   data.UserEmail,
		Status:          "FULFILLED",
		TotalAmount:     data.TotalAmount,
		ShippingAddress: data.ShippingAddress,
		AppName:         h.appName,
	}

	htmlBody, textBody, err := h.renderer.RenderOrderFulfilled(emailCtx)
	if err != nil {
		return fmt.Errorf("failed to render order_fulfilled template: %w", err)
	}

	msg := &domain.EmailMessage{
		To:       []string{data.UserEmail},
		Subject:  fmt.Sprintf("[%s] Order #%s Dispatched & On The Way", h.appName, data.OrderNumber),
		HTMLBody: htmlBody,
		TextBody: textBody,
	}

	return h.mailer.Send(ctx, msg)
}
