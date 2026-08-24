## ADDED Requirements

### Requirement: Registration OTP Email Dispatch
The system SHALL dispatch a 6-digit account verification OTP email to the registered user upon consuming an `auth.registration_otp` event.

#### Scenario: Sending registration verification OTP email
- **WHEN** an `auth.registration_otp` event is consumed with user `email`, `code`, and `name`
- **THEN** the system renders the registration OTP template and sends an email to `email` with subject "Verify your account".

#### Scenario: Missing required email or OTP code in auth payload
- **WHEN** an `auth.registration_otp` event payload lacks an `email` or `code`
- **THEN** the handler logs a validation error and aborts sending without throwing an uncaught panic.

### Requirement: Password Reset OTP Email Dispatch
The system SHALL dispatch a 6-digit password reset OTP email to the user upon consuming an `auth.password_reset_otp` event.

#### Scenario: Sending password reset OTP email
- **WHEN** an `auth.password_reset_otp` event is consumed with user `email`, `code`, and `name`
- **THEN** the system renders the password reset template and sends an email to `email` with subject "Reset your password".
