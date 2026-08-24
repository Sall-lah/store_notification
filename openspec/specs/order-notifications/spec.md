# order-notifications

## Purpose
Order lifecycle email notifications for customer receipts, cancellations, fulfillment updates, and store admin new order alerts.

## Requirements

### Requirement: Order Created Email Notification
The system SHALL dispatch an order summary email with payment instructions to the customer upon consuming an `order.created` event.

#### Scenario: Dispatching order created receipt to customer
- **WHEN** an `order.created` event is consumed containing `orderNumber`, `userEmail`, `totalAmount`, and `snapRedirectUrl`
- **THEN** the system renders the `order_created` template and emails the customer with order breakdown and payment link.

### Requirement: Order Paid Confirmation and Store Admin Alert
The system SHALL dispatch a payment receipt email to the customer AND dispatch a new paid order alert email to store administrators (`STORE_ADMIN_EMAILS`) upon consuming an `order.paid` event.

#### Scenario: Dispatching order confirmation to customer
- **WHEN** an `order.paid` event is consumed
- **THEN** the system renders the `order_paid` template and sends a payment receipt to `userEmail`.

#### Scenario: Dispatching order alert to store administrators
- **WHEN** an `order.paid` event is consumed and `STORE_ADMIN_EMAILS` are configured
- **THEN** the system renders the `admin_order_alert` template and emails each configured store administrator address with order items and total amount.

### Requirement: Order Cancelled Email Notification
The system SHALL dispatch an order cancellation notice to the customer upon consuming an `order.cancelled` event.

#### Scenario: Customer receives cancellation notice
- **WHEN** an `order.cancelled` event is consumed
- **THEN** the system renders the `order_cancelled` template and emails the customer confirming the cancellation and release of held items/reservations.

### Requirement: Order Expired Email Notification
The system SHALL dispatch a payment window expiration notice to the customer upon consuming an `order.expired` event.

#### Scenario: Customer receives payment window expired notice
- **WHEN** an `order.expired` event is consumed
- **THEN** the system renders the `order_expired` template and emails the customer notifying them that the payment window elapsed and items have been released back to stock.

### Requirement: Order Fulfilled Email Notification
The system SHALL dispatch a delivery and shipping confirmation email to the customer upon consuming an `order.fulfilled` event.

#### Scenario: Customer receives fulfillment and shipping notice
- **WHEN** an `order.fulfilled` event is consumed
- **THEN** the system renders the `order_fulfilled` template with tracking and delivery details and emails the customer.
