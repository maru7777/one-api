# One API Database Migration Tool Makefile

# Build variables
BINARY_NAME=migrate
BUILD_DIR=build
GO_FILES=$(shell find . -name "*.go" -type f)

# Version information
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

.PHONY: all build clean test help install examples

# Default target
all: build

# Build the migration tool
build: $(BUILD_DIR)/$(BINARY_NAME)

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES)
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

# Install the tool to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) .
	@echo "Installed $(BINARY_NAME) to $(shell go env GOPATH)/bin"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Show help
help:
	@echo "One API Database Migration Tool"
	@echo ""
	@echo "Available targets:"
	@echo "  build     - Build the migration tool"
	@echo "  install   - Install the tool to GOPATH/bin"
	@echo "  clean     - Clean build artifacts"
	@echo "  test      - Run tests"
	@echo "  examples  - Show usage examples"
	@echo "  help      - Show this help message"
	@echo ""
	@echo "Usage after building:"
	@echo "  ./$(BUILD_DIR)/$(BINARY_NAME) --help"

# Show usage examples
examples:
	@echo "One API Database Migration Tool - Usage Examples"
	@echo ""
	@echo "1. Show migration plan:"
	@echo "   ./$(BUILD_DIR)/$(BINARY_NAME) -source-type=sqlite -source-dsn=\"./one-api.db\" \\"
	@echo "                    -target-type=mysql -target-dsn=\"user:pass@tcp(localhost:3306)/oneapi\" \\"
	@echo "                    -show-plan"
	@echo ""
	@echo "2. Validate migration:"
	@echo "   ./$(BUILD_DIR)/$(BINARY_NAME) -source-type=sqlite -source-dsn=\"./one-api.db\" \\"
	@echo "                    -target-type=mysql -target-dsn=\"user:pass@tcp(localhost:3306)/oneapi\" \\"
	@echo "                    -validate-only"
	@echo ""
	@echo "3. Dry run migration:"
	@echo "   ./$(BUILD_DIR)/$(BINARY_NAME) -source-type=sqlite -source-dsn=\"./one-api.db\" \\"
	@echo "                    -target-type=mysql -target-dsn=\"user:pass@tcp(localhost:3306)/oneapi\" \\"
	@echo "                    -dry-run -verbose"
	@echo ""
	@echo "4. Actual migration:"
	@echo "   ./$(BUILD_DIR)/$(BINARY_NAME) -source-type=sqlite -source-dsn=\"./one-api.db\" \\"
	@echo "                    -target-type=mysql -target-dsn=\"user:pass@tcp(localhost:3306)/oneapi\" \\"
	@echo "                    -verbose"
	@echo ""
	@echo "5. SQLite to PostgreSQL:"
	@echo "   ./$(BUILD_DIR)/$(BINARY_NAME) -source-type=sqlite -source-dsn=\"./one-api.db\" \\"
	@echo "                    -target-type=postgres -target-dsn=\"postgres://user:pass@localhost/oneapi?sslmode=disable\" \\"
	@echo "                    -verbose"
	@echo ""
	@echo "For more information, see README.md or run:"
	@echo "   ./$(BUILD_DIR)/$(BINARY_NAME) --help"

# Development targets
dev-build: build
	@echo "Development build complete"

# Quick test build
quick:
	go build -o $(BINARY_NAME) .
	@echo "Quick build complete: ./$(BINARY_NAME)"

# Cross-compilation targets
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .

build-all: build-linux build-windows build-darwin
	@echo "Cross-compilation complete"

# Check if required tools are available
check-deps:
	@echo "Checking dependencies..."
	@which go >/dev/null || (echo "Go is not installed" && exit 1)
	@echo "Dependencies OK"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@which golangci-lint >/dev/null || (echo "golangci-lint is not installed" && exit 1)
	golangci-lint run

# Run all checks
check: fmt lint test
	@echo "All checks passed"
