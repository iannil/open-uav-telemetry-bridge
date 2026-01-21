.PHONY: all build clean test lint run deps help web-deps web-build web-clean build-with-web

# Build variables
BINARY_NAME=outb
BUILD_DIR=bin
CMD_PATH=./cmd/outb
WEB_DIR=web
EMBED_DIR=internal/web/dist

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

## Web UI Commands

web-deps: ## Install web dependencies
	cd $(WEB_DIR) && npm install

web-build: ## Build web UI and copy to embed directory
	@echo "Building Web UI..."
	cd $(WEB_DIR) && npm run build
	@rm -rf $(EMBED_DIR)
	@mkdir -p $(EMBED_DIR)
	cp -r $(WEB_DIR)/dist/* $(EMBED_DIR)/
	@echo "Web UI built and copied to $(EMBED_DIR)"

web-dev: ## Run web UI development server
	cd $(WEB_DIR) && npm run dev

web-clean: ## Clean web build artifacts
	@echo "Cleaning web artifacts..."
	@rm -rf $(WEB_DIR)/node_modules $(WEB_DIR)/dist $(EMBED_DIR)

build-with-web: web-build build ## Build with web UI embedded
