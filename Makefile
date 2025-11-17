# Makefile for ShowTa云盘

# Variables
APP_NAME = showta
VERSION ?= 1.0.1
LDFLAGS = -X 'overlink.top/app/system/conf.AppVersion=$(VERSION)'
BIN_DIR = bin

# Default target
.PHONY: default
default: build

# Build for current platform
.PHONY: build
build: frontend
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME)$(shell [ "$(OS)" = "Windows_NT" ] && echo ".exe" || echo "") -ldflags="$(LDFLAGS)" .

# Build frontend
.PHONY: frontend
frontend:
	cd web && pnpm install && pnpm build
	rm -rf app/web/dist
	cp -r web/dist app/web/

# Development mode
.PHONY: dev
dev:
	@echo "Starting development environment..."
	@echo "Run 'cd web && pnpm serve' in another terminal for frontend development"
	go run main.go

# Build for all platforms
.PHONY: build-all
build-all: build-linux-amd64 build-linux-arm64 build-windows-amd64 build-windows-arm64 build-darwin-amd64 build-darwin-arm64

# Platform-specific builds
.PHONY: build-linux-amd64
build-linux-amd64: frontend
	mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 -ldflags="$(LDFLAGS)" .

.PHONY: build-linux-arm64
build-linux-arm64: frontend
	mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=arm64 go build -o $(BIN_DIR)/$(APP_NAME)-linux-arm64 -ldflags="$(LDFLAGS)" .

.PHONY: build-windows-amd64
build-windows-amd64: frontend
	mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-windows-amd64.exe -ldflags="$(LDFLAGS)" .

.PHONY: build-windows-arm64
build-windows-arm64: frontend
	mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=arm64 go build -o $(BIN_DIR)/$(APP_NAME)-windows-arm64.exe -ldflags="$(LDFLAGS)" .

.PHONY: build-darwin-amd64
build-darwin-amd64: frontend
	mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-darwin-amd64 -ldflags="$(LDFLAGS)" .

.PHONY: build-darwin-arm64
build-darwin-arm64: frontend
	mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(BIN_DIR)/$(APP_NAME)-darwin-arm64 -ldflags="$(LDFLAGS)" .

# Clean generated files
.PHONY: clean
clean:
	rm -f $(BIN_DIR)/$(APP_NAME)*
	rm -rf app/web/dist
	cd web && rm -rf dist node_modules

# Docker build
.PHONY: docker
docker: build
	docker build -t $(APP_NAME):$(VERSION) .

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build        - Build for current platform"
	@echo "  build-all    - Build for all platforms"
	@echo "  dev          - Start development environment"
	@echo "  clean        - Clean generated files"
	@echo "  docker       - Build Docker image"
	@echo "  help         - Show this help"
	@echo ""
	@echo "Platform-specific builds:"
	@echo "  build-linux-amd64   - Build for Linux AMD64"
	@echo "  build-linux-arm64   - Build for Linux ARM64"
	@echo "  build-windows-amd64 - Build for Windows AMD64"
	@echo "  build-windows-arm64 - Build for Windows ARM64"
	@echo "  build-darwin-amd64  - Build for macOS AMD64"
	@echo "  build-darwin-arm64  - Build for macOS ARM64"