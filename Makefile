# Load environment variables from .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Binary name
BINARY_NAME=api
MAIN_PATH=cmd/api/main.go
GOOSE=$(shell which goose 2> /dev/null || echo $(shell go env GOPATH)/bin/goose)

.PHONY: all build run test clean lint migrate-up migrate-down migrate-status

all: build

## Build:
build:
	@echo "Building binary..."
	go build -o $(BINARY_NAME) $(MAIN_PATH)

## Run:
run: build
	@echo "Running application..."
	./$(BINARY_NAME)

## Dev (using air for hot reload if installed):
dev:
	@if command -v air > /dev/null; then \
		air -c .air.toml; \
	else \
		echo "Air is not installed. Running normally..."; \
		go run $(MAIN_PATH); \
	fi


## Clean:
clean:
	@echo "Cleaning up..."
	go clean
	rm -f $(BINARY_NAME)

## Lint:
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint is not installed. Please install it from https://golangci-lint.run/"; \
		exit 1; \
	fi

## Test:
test:
	@echo "Running tests..."
	go test -v ./...

## Migrations:
migrate-up:
	@echo "Running migrations up..."
	$(GOOSE) -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	@echo "Running migrations down..."
	$(GOOSE) -dir migrations postgres "$(DATABASE_URL)" down

migrate-status:
	@echo "Checking migration status..."
	$(GOOSE) -dir migrations postgres "$(DATABASE_URL)" status

migrate-create:
	@read -p "Enter migration name: " name; \
	$(GOOSE) -dir migrations create $$name sql
