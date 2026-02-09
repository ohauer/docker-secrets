.PHONY: help all build test fuzz lint clean install-systemd uninstall-systemd

BINARY_NAME=secrets-sync
BUILD_DIR=bin
GO=go
GOFLAGS=-v

VERSION=0.1.0
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildDate=$(BUILD_DATE)"

.DEFAULT_GOAL := help

help:
	@echo "Available targets:"
	@echo "  all              - Run lint, test, and build"
	@echo "  build            - Build the binary"
	@echo "  build-static     - Build static binary for container"
	@echo "  test             - Run tests"
	@echo "  fuzz             - Run fuzz tests (10s per test)"
	@echo "  coverage         - Generate coverage report"
	@echo "  lint             - Run linter"
	@echo "  fmt              - Format code"
	@echo "  vet              - Run go vet"
	@echo "  clean            - Remove build artifacts"
	@echo "  run              - Run the application"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Run Docker container"
	@echo "  docker-test      - Test Docker container"
	@echo "  install-systemd  - Install as systemd service (requires root)"
	@echo "  uninstall-systemd - Uninstall systemd service (requires root)"

all: lint test build

build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/secrets-sync

build-static:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -a $(LDFLAGS) -ldflags '-extldflags "-static"' -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/secrets-sync

test:
	$(GO) test -v -race -coverprofile=coverage.out ./...

fuzz:
	@echo "Running fuzz tests (10s each)..."
	$(GO) test -fuzz=FuzzValidatePath -fuzztime=10s ./internal/filewriter
	$(GO) test -fuzz=FuzzValidateMode -fuzztime=10s ./internal/filewriter
	$(GO) test -fuzz=FuzzRender -fuzztime=10s ./internal/template
	$(GO) test -fuzz=FuzzConfigLoad -fuzztime=10s ./internal/config
	$(GO) test -fuzz=FuzzValidateFilePath -fuzztime=10s ./internal/config
	@echo "All fuzz tests passed!"

coverage: test
	$(GO) tool cover -html=coverage.out -o coverage.html

lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

fmt:
	$(GO) fmt ./...
	gofmt -s -w .

vet:
	$(GO) vet ./...

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

run:
	$(GO) run ./cmd/secrets-sync

docker-build:
	docker build -t docker-secrets:latest .

docker-run:
	docker run --rm docker-secrets:latest

docker-test:
	docker build -t docker-secrets:test .
	docker run --rm docker-secrets:test isready; echo "Exit code: $$?"

install-systemd: build
	@echo "Installing secrets-sync as systemd service..."
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "Error: This target must be run as root (use sudo make install-systemd)"; \
		exit 1; \
	fi
	@./scripts/install-systemd.sh

uninstall-systemd:
	@echo "Uninstalling secrets-sync systemd service..."
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "Error: This target must be run as root (use sudo make uninstall-systemd)"; \
		exit 1; \
	fi
	@./scripts/uninstall-systemd.sh
