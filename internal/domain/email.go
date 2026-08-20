package domain

// RecipientType denotes whether an email is destined for an end-customer or store administrator.
type RecipientType string

const (
	RecipientUser       RecipientType = "USER"
	RecipientStoreAdmin RecipientType = "STORE_ADMIN"
)

// EmailRecipient represents an email address with an optional display name.
type EmailRecipient struct {
	Name  string
	Email string
}

// EmailMessage contains the fully prepared message ready for transport via SMTP.
type EmailMessage struct {
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	HTMLBody    string
	TextBody    string
	Attachments []EmailAttachment
}

// EmailAttachment represents an optional binary attachment sent with an email.
type EmailAttachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// AuthEmailContext represents dynamic template variables for authentication OTP emails.
type AuthEmailContext struct {
	Name        string
	Email       string
	OTPCode     string
	ExpiryMins  int
	SupportLink string
	AppName     string
}

// OrderEmailContext represents dynamic template variables for customer order notifications.
type OrderEmailContext struct {
	OrderNumber     string
	CustomerName    string
	CustomerEmail   string
	Status          string
	TotalAmount     float64
	ShippingFee     float64
	ShippingAddress string
	PaymentType     string
	SnapRedirectURL string
	PaidAtFormatted string
	Items           []OrderItemData
	Reason          string
	SupportLink     string
	AppName         string
}

// AdminOrderAlertContext represents dynamic template variables for store administrator alerts.
type AdminOrderAlertContext struct {
	OrderNumber     string
	CustomerEmail   string
	TotalAmount     float64
	ItemCount       int
	Items           []OrderItemData
	PaidAtFormatted string
	AdminPortalURL  string
	AppName         string
}
