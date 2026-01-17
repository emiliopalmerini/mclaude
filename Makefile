.PHONY: all build test test-unit test-integration clean sqlc templ fmt lint run migrate reset install help

# Variables
BINARY_NAME := claude-watcher
BUILD_DIR := .
GO_FILES := $(shell find . -name '*.go' -not -path './sqlc/generated/*')

# Default target
all: tidy sqlc build

# Build the binary
build:
	go build -o $(BINARY_NAME) ./cmd/claude-watcher

# Build with version info
build-release:
	go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/claude-watcher

# Install to GOPATH/bin
install:
	go install ./cmd/claude-watcher

# Run all tests
test:
	go test -v ./...

# Run unit tests only (skip integration tests)
test-unit:
	go test -v -short ./...

# Run integration tests (requires CLAUDE_WATCHER_* env vars)
test-integration:
	go test -v -run Integration ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Generate sqlc code
sqlc:
	sqlc generate

# Generate templ templates (for web UI)
templ:
	templ generate

# Generate all code
generate: sqlc templ

# Format code
fmt:
	go fmt ./...
	gofmt -s -w .

# Lint code
lint:
	golangci-lint run ./...

# Tidy dependencies
tidy:
	go mod tidy

# Run the CLI
run: build
	./$(BINARY_NAME)

# Run migrations
migrate: build
	./$(BINARY_NAME) migrate

# Reset database (drop all tables)
reset: build
	./$(BINARY_NAME) reset

# Start web server (for development)
serve: build
	./$(BINARY_NAME) serve

# Development: watch and rebuild
dev:
	@echo "Watching for changes..."
	@while true; do \
		$(MAKE) build; \
		fswatch -1 $(GO_FILES) > /dev/null 2>&1 || inotifywait -q -e modify $(GO_FILES) 2>/dev/null || sleep 2; \
	done

# Help
help:
	@echo "claude-watcher - Personal analytics for Claude Code"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build targets:"
	@echo "  build           Build the binary"
	@echo "  build-release   Build optimized binary"
	@echo "  install         Install to GOPATH/bin"
	@echo "  clean           Remove build artifacts"
	@echo ""
	@echo "Test targets:"
	@echo "  test            Run all tests"
	@echo "  test-unit       Run unit tests only"
	@echo "  test-integration Run integration tests"
	@echo "  test-coverage   Run tests with coverage report"
	@echo ""
	@echo "Code generation:"
	@echo "  sqlc            Generate sqlc code"
	@echo "  templ           Generate templ templates"
	@echo "  generate        Generate all code (sqlc + templ)"
	@echo ""
	@echo "Code quality:"
	@echo "  fmt             Format code"
	@echo "  lint            Run linter"
	@echo "  tidy            Tidy go modules"
	@echo ""
	@echo "Database:"
	@echo "  migrate         Run database migrations"
	@echo "  reset           Reset database (drop all tables)"
	@echo ""
	@echo "Run:"
	@echo "  run             Build and run CLI"
	@echo "  serve           Start web server"
	@echo "  dev             Watch and rebuild on changes"
	@echo ""
	@echo "Other:"
	@echo "  all             tidy + sqlc + build"
	@echo "  help            Show this help"
