# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build dependencies & ca-certificates
RUN apk add --no-cache git ca-certificates

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o store_notification ./cmd/server

# Production runtime stage
FROM alpine:3.21

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/store_notification .

# Expose HTTP health and documentation port
EXPOSE 8070

# Run service
ENTRYPOINT ["/app/store_notification"]
