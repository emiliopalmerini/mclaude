.PHONY: all build test test-unit test-integration clean sqlc templ fmt lint run migrate reset install help

# Variables
BINARY_NAME := mclaude
BUILD_DIR := .
GO_FILES := $(shell find . -name '*.go' -not -path './sqlc/generated/*')
TEMPL_FILES := $(shell find . -name '*.templ')
SQL_FILES := $(shell find sqlc/queries -name '*.sql')

# Default target
all: build

# === Code Generation ===

# Generate sqlc code (depends on SQL files)
sqlc: $(SQL_FILES)
	sqlc generate

# Generate templ templates (depends on templ files)
templ: $(TEMPL_FILES)
	templ generate

# Generate all code
generate: sqlc templ

# === Build ===

# Build the binary (depends on generated code)
# CGO_ENABLED=1 required for go-libsql embedded database
build: generate
	CGO_ENABLED=1 go build -o $(BINARY_NAME) ./cmd/mclaude

# Build with version info
build-release: generate
	CGO_ENABLED=1 go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/mclaude

# Install to GOPATH/bin
install: generate
	CGO_ENABLED=1 go install ./cmd/mclaude

# === Testing ===

# Run all tests (depends on generated code)
test: generate
	CGO_ENABLED=1 go test -v ./...

# Run unit tests only (skip integration tests)
test-unit: generate
	CGO_ENABLED=1 go test -v -short ./...

# Run integration tests (requires MCLAUDE_* env vars)
test-integration: generate
	CGO_ENABLED=1 go test -v -run Integration ./...

# Run tests with coverage
test-coverage: generate
	CGO_ENABLED=1 go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# === Code Quality ===

# Format code
fmt:
	go fmt ./...
	gofmt -s -w .

# Lint code (depends on generated code)
lint: generate
	golangci-lint run ./...

# Tidy dependencies
tidy:
	go mod tidy

# Check everything (format, lint, test)
check: fmt lint test

# === Database ===

# Run migrations
migrate: build
	./$(BINARY_NAME) migrate

# Reset database (drop all tables)
reset: build
	./$(BINARY_NAME) migrate 0

# === Run ===

# Run the CLI
run: build
	./$(BINARY_NAME)

# Start web server (for development)
serve: build
	./$(BINARY_NAME) serve

# Development: watch and rebuild
dev:
	@echo "Watching for changes..."
	@while true; do \
		$(MAKE) build; \
		fswatch -1 $(GO_FILES) $(TEMPL_FILES) > /dev/null 2>&1 || inotifywait -q -e modify $(GO_FILES) $(TEMPL_FILES) 2>/dev/null || sleep 2; \
	done

# === Cleanup ===

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Clean generated code too
clean-all: clean
	rm -f sqlc/generated/*.go
	rm -f internal/web/templates/*_templ.go

# === Help ===

help:
	@echo "mclaude - Personal analytics for Claude Code"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build targets:"
	@echo "  all             Build everything (default)"
	@echo "  build           Generate code + build mclaude binary"
	@echo "  build-release   Generate code + build optimized binary"
	@echo "  install         Generate code + install to GOPATH/bin"
	@echo "  clean           Remove build artifacts"
	@echo "  clean-all       Remove build + generated code"
	@echo ""
	@echo "Test targets:"
	@echo "  test            Generate + run all tests"
	@echo "  test-unit       Generate + run unit tests only"
	@echo "  test-integration Generate + run integration tests"
	@echo "  test-coverage   Generate + run tests with coverage report"
	@echo ""
	@echo "Code generation:"
	@echo "  sqlc            Generate sqlc code from SQL"
	@echo "  templ           Generate Go code from templ templates"
	@echo "  generate        Generate all code (sqlc + templ)"
	@echo ""
	@echo "Code quality:"
	@echo "  fmt             Format code"
	@echo "  lint            Generate + run linter"
	@echo "  tidy            Tidy go modules"
	@echo "  check           Format + lint + test"
	@echo ""
	@echo "Database:"
	@echo "  migrate         Build + run database migrations"
	@echo "  reset           Build + reset database to version 0"
	@echo ""
	@echo "Run:"
	@echo "  run             Build + run CLI"
	@echo "  serve           Build + start web server"
	@echo "  dev             Watch files and rebuild on changes"
