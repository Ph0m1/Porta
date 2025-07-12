# Porta Gateway Makefile

# Variables
APP_NAME := porta
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S)
GO_VERSION := $(shell go version | awk '{print $$3}')
GIT_COMMIT := $(shell git rev-parse HEAD)

# Build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Directories
BUILD_DIR := build
DOCKER_DIR := docker
EXAMPLES_DIR := examples

# Default target
.PHONY: all
all: clean test build

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run with Docker Compose"
	@echo "  docker-stop    - Stop Docker Compose"
	@echo "  lint           - Run linters"
	@echo "  fmt            - Format code"
	@echo "  deps           - Download dependencies"
	@echo "  install        - Install the application"
	@echo "  dev            - Run in development mode"

# Build targets
.PHONY: build
build: deps
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@unset http_proxy https_proxy ftp_proxy && CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -ldflags '-extldflags "-static"' -o $(BUILD_DIR)/$(APP_NAME) ./$(EXAMPLES_DIR)/gin/
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

.PHONY: build-all
build-all: deps
	@echo "Building all examples..."
	@mkdir -p $(BUILD_DIR)
	@GO111MODULE=on go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-gin ./$(EXAMPLES_DIR)/gin/
	@GO111MODULE=on go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-mux ./$(EXAMPLES_DIR)/mux/
	@GO111MODULE=on go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-gorilla ./$(EXAMPLES_DIR)/gorilla/
	@GO111MODULE=on go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-negroni ./$(EXAMPLES_DIR)/negroni/
	@echo "All builds complete"

# Test targets
.PHONY: test
test:
	@echo "Running tests..."
	@GO111MODULE=on go test -v ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@GO111MODULE=on go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-race
test-race:
	@echo "Running tests with race detection..."
	@GO111MODULE=on go test -v -race ./...

.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	@GO111MODULE=on go test -v -bench=. -benchmem ./...

# Development targets
.PHONY: dev
dev: build
	@echo "Starting development server..."
	@./$(BUILD_DIR)/$(APP_NAME) -c ./$(EXAMPLES_DIR)/etc/config.yaml -d -l DEBUG

.PHONY: dev-gin
dev-gin: build
	@echo "Starting Gin development server..."
	@./$(BUILD_DIR)/$(APP_NAME)-gin -c ./$(EXAMPLES_DIR)/etc/config.yaml -d -l DEBUG

# Docker targets
.PHONY: docker-build
docker-build: build
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME):$(VERSION) .
	@docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "Docker image built: $(APP_NAME):$(VERSION)"

.PHONY: docker-run
docker-run:
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d
	@echo "Services started. Gateway available at http://localhost:8080"

.PHONY: docker-stop
docker-stop:
	@echo "Stopping Docker Compose services..."
	@docker-compose down

.PHONY: docker-logs
docker-logs:
	@docker-compose logs -f porta-gateway

.PHONY: docker-clean
docker-clean:
	@echo "Cleaning Docker resources..."
	@docker-compose down -v --remove-orphans
	@docker system prune -f

# Code quality targets
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

.PHONY: lint
lint:
	@echo "Running linters..."
	@golangci-lint run ./...

.PHONY: vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Dependency targets
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	@GO111MODULE=on go mod download
	@GO111MODULE=on go mod tidy

.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	@GO111MODULE=on go get -u ./...
	@GO111MODULE=on go mod tidy

# Installation targets
.PHONY: install
install: build
	@echo "Installing $(APP_NAME)..."
	@cp $(BUILD_DIR)/$(APP_NAME) $(GOPATH)/bin/
	@echo "$(APP_NAME) installed to $(GOPATH)/bin/"

# Clean targets
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean -cache -testcache -modcache

# Release targets
.PHONY: release
release: clean test build-all
	@echo "Creating release $(VERSION)..."
	@mkdir -p $(BUILD_DIR)/release
	@tar -czf $(BUILD_DIR)/release/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR) .
	@echo "Release created: $(BUILD_DIR)/release/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz"

# Setup development environment
.PHONY: setup-dev
setup-dev:
	@echo "Setting up development environment..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@mkdir -p $(DOCKER_DIR)
	@echo "Development environment setup complete"

# Generate mock data for testing
.PHONY: generate-mocks
generate-mocks:
	@echo "Generating mock data..."
	@mkdir -p $(DOCKER_DIR)/backend1-data $(DOCKER_DIR)/backend2-data
	@echo '{"message": "Hello from Backend 1", "service": "backend-1"}' > $(DOCKER_DIR)/backend1-data/index.json
	@echo '{"message": "Hello from Backend 2", "service": "backend-2"}' > $(DOCKER_DIR)/backend2-data/index.json

# Show project information
.PHONY: info
info:
	@echo "Project Information:"
	@echo "  Name: $(APP_NAME)"
	@echo "  Version: $(VERSION)"
	@echo "  Go Version: $(GO_VERSION)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Build Time: $(BUILD_TIME)" 