.PHONY: all build clean test lint run deps help

# Build variables
BINARY_NAME=outb
BUILD_DIR=bin
CMD_PATH=./cmd/outb

# Go variables
GOFLAGS=-ldflags="-s -w"
GOARCH_ARM64=arm64
GOOS_LINUX=linux

all: deps build

## Build Commands

build: ## Build for current platform
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)

build-linux-arm64: ## Build for Linux ARM64 (Raspberry Pi, Jetson)
	@echo "Building $(BINARY_NAME) for Linux ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS_LINUX) GOARCH=$(GOARCH_ARM64) go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_PATH)

build-linux-amd64: ## Build for Linux AMD64
	@echo "Building $(BINARY_NAME) for Linux AMD64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS_LINUX) GOARCH=amd64 go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_PATH)

build-all: build build-linux-arm64 build-linux-amd64 ## Build for all platforms

## Development Commands

run: build ## Build and run
	@./$(BUILD_DIR)/$(BINARY_NAME)

deps: ## Download dependencies
	go mod download
	go mod tidy

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

fmt: ## Format code
	go fmt ./...
	goimports -w .

## Utility Commands

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

setup-config: ## Copy example config
	@if [ ! -f configs/config.yaml ]; then \
		cp configs/config.example.yaml configs/config.yaml; \
		echo "Created configs/config.yaml from example"; \
	else \
		echo "configs/config.yaml already exists"; \
	fi

## Help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
