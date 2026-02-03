.PHONY: help all build test lint clean

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
	@echo "  all           - Run lint, test, and build"
	@echo "  build         - Build the binary"
	@echo "  build-static  - Build static binary for container"
	@echo "  test          - Run tests"
	@echo "  coverage      - Generate coverage report"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  clean         - Remove build artifacts"
	@echo "  run           - Run the application"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  docker-test   - Test Docker container"

all: lint test build

build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/secrets-sync

build-static:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -a $(LDFLAGS) -ldflags '-extldflags "-static"' -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/secrets-sync

test:
	$(GO) test -v -race -coverprofile=coverage.out ./...

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
