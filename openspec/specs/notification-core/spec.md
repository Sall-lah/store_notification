# notification-core

## Purpose
Core ingestion, parsing, routing, and Redis-backed idempotency deduplication for incoming Kafka domain events.

## Requirements

### Requirement: Standard Event Envelope Parsing
The system SHALL ingest and deserialize JSON event envelopes from Kafka topics into a strongly typed `EventEnvelope` structure containing `event_id`, `event_type`, `timestamp`, `producer`, and `data`.

#### Scenario: Successfully deserializing a valid event envelope
- **WHEN** a valid JSON event envelope message is consumed from a Kafka topic
- **THEN** the system parses the envelope, extracts the metadata and payload, and forwards it to the event router.

#### Scenario: Handling malformed JSON message
- **WHEN** an unparseable or schema-invalid JSON message is consumed from Kafka
- **THEN** the system logs an error with the raw payload and message offset, skips execution, and acknowledges the message to prevent partition blocking.

### Requirement: Redis-backed Idempotency Deduplication
The system SHALL ensure exactly-once processing for each unique `event_id` using an atomic Redis `SETNX` lock with a 24-hour TTL (`notif:evt:{event_id}`).

#### Scenario: Processing a new event for the first time
- **WHEN** an event with a new `event_id` is received
- **THEN** the Redis key `notif:evt:{event_id}` is set with a 24-hour expiry and execution continues to handler dispatch.

#### Scenario: Dropping duplicate events upon redelivery
- **WHEN** an event with an existing `event_id` is re-delivered due to a Kafka consumer rebalance or retry
- **THEN** the Redis key check returns already exists, the duplicate event is safely acknowledged, and no duplicate email is dispatched.

### Requirement: Event Routing and Dispatching
The system SHALL route incoming valid events to their registered domain handlers based on `event_type`.

#### Scenario: Routing known event types
- **WHEN** an event with `event_type` matching `auth.registration_otp` or `order.created` is processed
- **THEN** the system invokes the respective registered event handler.

#### Scenario: Handling unmapped event types
- **WHEN** an event with an unrecognized `event_type` is consumed
- **THEN** the system logs an unhandled event warning and acknowledges the message without error.
