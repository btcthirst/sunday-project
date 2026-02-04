.PHONY: all dev build test clean deps frontend-deps backend-deps help

# Default target
all: build

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## [-a-zA-Z0-9_]+:' Makefile | sed 's/## //' | awk -F ':' '{printf "  %-15s %s\n", $$1, $$2}'

## dev: Run the application in development mode
dev:
	wails dev

## build: Build the application for the current platform
build:
	wails build

## test: Run all backend tests
test:
	go test -v ./...

## clean: Remove build artifacts
clean:
	rm -rf build/bin/*

## deps: Install all dependencies (frontend and backend)
deps: backend-deps frontend-deps

## backend-deps: Download backend dependencies
backend-deps:
	go mod download

## frontend-deps: Install frontend dependencies
frontend-deps:
	cd frontend && npm install

## lint: Run golangci-lint if installed
lint:
	if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Install it from https://golangci-lint.run/"; \
	fi
