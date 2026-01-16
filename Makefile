.PHONY: build test clean run lint proto

# Variables
BINARY_DIR=bin
CMD_DIR=cmd
PROTO_DIR=proto
GO=go

# Build
build:
	@echo "Building binaries..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build -o $(BINARY_DIR)/predictor $(CMD_DIR)/predictor/main.go
	$(GO) build -o $(BINARY_DIR)/backtester $(CMD_DIR)/backtester/main.go

# Test
test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

test-integration:
	@echo "Running integration tests..."
	$(GO) test -v -tags=integration ./test/integration/...

test-coverage:
	@echo "Generating coverage report..."
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Lint
lint:
	@echo "Running linter..."
	golangci-lint run

# Clean
clean:
	@echo "Cleaning..."
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html
	rm -rf data/*.json data/*/*.json

# Run
run-predictor:
	@echo "Running predictor..."
	$(GO) run $(CMD_DIR)/predictor/main.go

run-backtester:
	@echo "Running backtester..."
	$(GO) run $(CMD_DIR)/backtester/main.go

# Dependencies
deps:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Proto
proto: ## Generate protobuf code
	@echo "Generating proto files..."
	@protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative $(PROTO_DIR)/*.proto

proto-lint: ## Lint proto files
	@echo "Linting proto files..."
	@if command -v buf >/dev/null 2>&1; then \
		buf lint $(PROTO_DIR); \
	else \
		echo "buf not installed. Install from https://buf.build"; \
	fi

proto-check: ## Check proto formatting
	@echo "Checking proto formatting..."
	@if command -v buf >/dev/null 2>&1; then \
		buf format --diff $(PROTO_DIR); \
	else \
		echo "buf not installed. Install from https://buf.build"; \
	fi

# Format
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Vet
vet:
	@echo "Vetting code..."
	$(GO) vet ./...

# All checks
check: fmt vet lint test

# Help
help:
	@echo "Available targets:"
	@echo "  build            - Build the binaries"
	@echo "  test             - Run tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-coverage    - Run tests with coverage"
	@echo "  lint             - Run linter"
	@echo "  clean            - Remove build artifacts"
	@echo "  run-predictor    - Run predictor application"
	@echo "  run-backtester   - Run backtester application"
	@echo "  deps             - Download dependencies"
	@echo "  proto            - Generate protobuf files"
	@echo "  proto-lint       - Lint proto files"
	@echo "  proto-check      - Check proto formatting"
	@echo "  fmt              - Format code"
	@echo "  vet              - Vet code"
	@echo "  check            - Run all checks (fmt, vet, lint, test)"
	@echo "  help             - Show this help message"
