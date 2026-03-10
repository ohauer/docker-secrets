# Docker Secrets Sidecar - Implementation TODO

## Milestone 1: Foundation (Week 1)

- [x] Task 1: Project scaffolding and configuration structures
  - [x] Initialize Go module with go 1.26.1
  - [x] Define YAML configuration structures
  - [x] Define environment variable configuration
  - [x] Create project structure (cmd/, internal/, pkg/)
  - [x] Add Makefile with lint, build, test targets

- [x] Task 2: Configuration loader with validation
  - [x] Implement YAML config file parser
  - [x] Implement environment variable parser
  - [x] Add validation logic for required fields
  - [x] Add unit tests for config parsing

- [x] Task 2.5: Configuration file watcher with hot reload
  - [x] Implement file system watcher using fsnotify
  - [x] Detect YAML config file changes
  - [x] Reload and validate new configuration
  - [x] Gracefully update secret sync jobs
  - [x] Make file watching optional via environment variable
  - [x] Add unit tests for config reload logic

- [x] Task 3: Structured logging with JSON output
  - [x] Integrate zap logger for JSON structured logging
  - [x] Create logging wrapper with standard fields
  - [x] Add log level configuration via environment variable
  - [x] Add unit tests for logging output format

- [x] Task 4: Vault/OpenBao client with authentication
  - [x] Implement Vault client wrapper using hashicorp/vault/api
  - [x] Add Token authentication support
  - [x] Add AppRole authentication support
  - [x] Add connection validation on startup
  - [x] Add unit tests with mock Vault server

- [x] Task 5: Circuit breaker integration
  - [x] Integrate sony/gobreaker library
  - [x] Wrap all Vault API calls with circuit breaker
  - [x] Configure circuit breaker thresholds via environment variables
  - [x] Add logging for circuit breaker state changes
  - [x] Add unit tests for circuit breaker behavior

## Milestone 2: Core Functionality (Week 2)

- [x] Task 6: Secret fetching with retry and backoff
  - [x] Implement secret fetch logic from Vault KV v2
  - [x] Add exponential backoff for failed fetches
  - [x] Implement best-effort startup
  - [x] Add unit tests for fetch and retry logic

- [x] Task 7: Template engine for field mapping
  - [x] Implement Go template engine for mapping secret fields
  - [x] Support external-secrets-operator template syntax
  - [x] Add template validation during config load
  - [x] Add unit tests for template rendering

- [x] Task 8: File writer with atomic operations and permissions
  - [x] Implement atomic file write (temp file + rename)
  - [x] Add configurable file permissions and ownership per secret
  - [x] Create parent directories if needed
  - [x] Add unit tests for file operations

- [x] Task 9: Secret synchronization orchestrator
  - [x] Implement main sync loop per secret with configurable refresh intervals
  - [x] Coordinate fetch, template, and write operations
  - [x] Track last successful sync time per secret
  - [x] Add unit tests for sync orchestration

## Milestone 3: Observability (Week 3)

- [x] Task 10: Health and readiness endpoints with isready subcommand
  - [x] Implement HTTP server with /health and /ready endpoints
  - [x] Add isready subcommand that reads status file
  - [x] Main process writes readiness state to status file
  - [x] Add configurable HTTP port and status file path
  - [x] Add unit tests for endpoints and isready subcommand

- [x] Task 11: Prometheus metrics instrumentation
  - [x] Add Prometheus metrics (fetch_total, errors_total, duration_seconds, breaker_state)
  - [x] Expose metrics on /metrics endpoint
  - [x] Add metric labels (secret_name, vault_path)
  - [x] Add unit tests for metric collection

- [x] Task 12: Optional OpenTelemetry tracing
  - [x] Add OpenTelemetry SDK integration
  - [x] Implement tracing for secret fetch and sync operations
  - [x] Make tracing optional via enable_tracing environment variable
  - [x] Add configuration for trace exporter endpoint
  - [x] Add unit tests for trace span creation

- [x] Task 13: Graceful shutdown handling
  - [x] Implement signal handling for SIGTERM/SIGINT
  - [x] Add graceful shutdown with timeout
  - [x] Ensure in-flight operations complete before exit
  - [x] Add integration test for shutdown behavior

## Milestone 4: Production Ready (Week 4)

- [x] Task 14: Dockerfile and container image
  - [x] Create minimal Dockerfile using FROM scratch
  - [x] Build statically compiled Go binary with CGO_ENABLED=0
  - [x] Add CA certificates for HTTPS connections
  - [x] Configure non-root user (nobody) in container
  - [x] Add .dockerignore file
  - [x] Build and test container image

- [x] Task 15: Documentation and examples
  - [x] Create README with usage instructions
  - [x] Add example YAML configurations
  - [x] Add example docker-compose.yml with sidecar pattern (primary)
  - [x] Add example Dockerfile for application container
  - [x] Document all environment variables
  - [x] Add troubleshooting guide
  - [x] Create Kubernetes deployment example (optional reference)

- [x] Task 16: Dependency management configuration
  - [x] Create Dependabot configuration for GitHub
  - [x] Create Renovate configuration for GitLab
  - [x] Configure update schedules and rules

- [x] Task 17: Development environment setup
  - [x] Create docker-compose for Vault development environment
  - [x] Create docker-compose for OpenBao development environment
  - [x] Create initialization scripts with test data
  - [x] Create development configuration examples
  - [x] Document usage in examples README


