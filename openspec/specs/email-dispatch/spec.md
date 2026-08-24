# email-dispatch

## Purpose
HTML and plain-text template rendering engine and thread-safe SMTP transport client.

## Requirements

### Requirement: Email Template Rendering Engine
The system SHALL provide an HTML and plain-text template rendering engine utilizing Go `html/template` and `text/template` with embedded layout styling and template helpers.

#### Scenario: Rendering template with dynamic context
- **WHEN** a template name and data context map are provided to the renderer
- **THEN** the system generates sanitized, responsive HTML output and a plaintext alternative.

#### Scenario: Template not found
- **WHEN** an unregistered or missing template is requested
- **THEN** the renderer returns an explicit template error and prevents email dispatch.

### Requirement: SMTP Email Dispatcher
The system SHALL provide a thread-safe SMTP email transport client supporting TLS/STARTTLS, authentication, and sender identity (`From`, `To`, `Subject`, `Body`, `MIME-Version`).

#### Scenario: Successfully sending an email over SMTP
- **WHEN** an outbound email message is dispatched to the SMTP client
- **THEN** the client connects to the configured SMTP host, authenticates, and sends the message successfully.

#### Scenario: Handling temporary SMTP connection failures
- **WHEN** the SMTP server is temporarily unreachable or returns a 4xx error
- **THEN** the dispatcher retries with exponential backoff before reporting a terminal error.
