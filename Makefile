.PHONY: all build test run clean docker-build

APP_NAME := store_notification
BINARY_DIR := bin

all: test build

build:
	@echo "==> Building binary..."
	go build -o $(BINARY_DIR)/$(APP_NAME) ./cmd/server

test:
	@echo "==> Running all unit tests..."
	go test -v -race ./...

run:
	@echo "==> Starting notification service..."
	go run ./cmd/server

clean:
	@echo "==> Cleaning build artifacts..."
	rm -rf $(BINARY_DIR)

docker-build:
	@echo "==> Building Docker image..."
	docker build -t $(APP_NAME):latest .
