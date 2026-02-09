# Project Roadmap - Docker Secrets Sidecar

## ✅ v0.1.0 RELEASED - 2026-02-09

**Current Version**: v0.1.0
**Status**: Production Ready with Systemd Support
**All Core Milestones**: COMPLETED

### Quick Stats
- **Total Tasks**: 17/17 core tasks completed (100%)
- **Security Hardening**: Complete (fuzzing, path validation, resource limits)
- **Systemd Support**: Complete (service, man page, install scripts)
- **Test Coverage**: 60+ tests + 5 fuzz tests, all passing
- **Container Image**: <20MB (FROM scratch)
- **Documentation**: Complete (README, guides, man page, examples)

### Recent Additions (v0.1.0)
- ✅ Systemd service support with security hardening
- ✅ SIGHUP signal handling for zero-downtime reload
- ✅ Comprehensive security hardening (TOCTOU prevention, symlink rejection)
- ✅ Fuzzing tests for input validation
- ✅ OS-aware path length limits
- ✅ Resource limits (1MB secrets, 10MB responses)
- ✅ Pre-commit hooks and GitHub Actions CI
- ✅ Man page documentation
- ✅ Duplicate path detection

---

## Vision

Create a production-ready, secure, and lightweight sidecar container for managing secrets from HashiCorp Vault or OpenBao in Docker/Podman environments. The tool should be simple to deploy, resilient to failures, and provide comprehensive observability.

## Milestones

### Milestone 1: Foundation (Week 1) ✅ COMPLETED
**Goal**: Establish project structure and core infrastructure

**Tasks**:
- [x] Task 1: Project scaffolding and configuration structures
- [x] Task 2: Configuration loader with validation
- [x] Task 2.5: Configuration file watcher with hot reload
- [x] Task 3: Structured logging with JSON output
- [x] Task 4: Vault/OpenBao client with authentication
- [x] Task 5: Circuit breaker integration

**Success Criteria**: ✅ ALL MET
- Project compiles and passes linting
- Configuration can be loaded from YAML and environment variables
- Vault authentication works with Token and AppRole
- Circuit breaker wraps Vault calls
- JSON logs are properly formatted

**Deliverables**: ✅ COMPLETED
- Working Go module with proper structure
- Configuration parsing and validation
- Vault client with authentication
- Structured logging infrastructure
- Circuit breaker integration

---

### Milestone 2: Core Functionality (Week 2) ✅ COMPLETED
**Goal**: Implement secret fetching, templating, and file writing

**Tasks**:
- [x] Task 6: Secret fetching with retry and backoff
- [x] Task 7: Template engine for field mapping
- [x] Task 8: File writer with atomic operations and permissions
- [x] Task 9: Secret synchronization orchestrator

**Success Criteria**: ✅ ALL MET
- Secrets can be fetched from Vault KV v2
- Templates correctly map secret fields to file content
- Files are written atomically with correct permissions
- Multiple secrets sync with different refresh intervals
- Exponential backoff works on failures

**Deliverables**: ✅ COMPLETED
- Secret fetcher with retry logic
- Go template engine for field mapping
- Atomic file writer with permission control
- Sync orchestrator with scheduling

---

### Milestone 3: Observability (Week 3) ✅ COMPLETED
**Goal**: Add health checks, metrics, and tracing

**Tasks**:
- [x] Task 10: Health and readiness endpoints with isready subcommand
- [x] Task 11: Prometheus metrics instrumentation
- [x] Task 12: Optional OpenTelemetry tracing
- [x] Task 13: Graceful shutdown handling

**Success Criteria**: ✅ ALL MET
- Health endpoints respond correctly
- `tool isready` command works for docker-compose
- Prometheus metrics are exposed and accurate
- Tracing can be enabled and exports spans
- Graceful shutdown completes in-flight operations

**Deliverables**: ✅ COMPLETED
- HTTP server with health/ready endpoints
- Status file for healthcheck subcommand
- Prometheus metrics instrumentation
- OpenTelemetry tracing integration
- Signal handling for graceful shutdown

---

### Milestone 4: Production Ready (Week 4) ✅ COMPLETED
**Goal**: Create container image and comprehensive documentation

**Tasks**:
- [x] Task 14: Dockerfile and container image
- [x] Task 15: Documentation and examples
- [x] Task 16: Dependency management configuration
- [x] Task 17: Development environment setup

**Success Criteria**: ✅ ALL MET
- Container image < 20MB
- Tool runs in scratch container
- Complete documentation with examples
- docker-compose example works
- Development environment ready

**Deliverables**: ✅ COMPLETED
- Minimal Dockerfile with scratch base
- Build scripts for static binary
- README with usage instructions
- Configuration documentation
- Example deployments (docker-compose, Vault/OpenBao dev environments)
- Troubleshooting guide
- Dependabot and Renovate configurations

---

## Release History

### v0.1.0 - Production Release (2026-02-09) ✅
**Focus**: Production-ready with systemd support and comprehensive security hardening

**Core Features**: ✅ ALL IMPLEMENTED
- Configuration from YAML with hot reload (SIGHUP)
- Token and AppRole authentication
- Secret fetching with exponential backoff retry
- Circuit breaker with configurable thresholds
- Template engine for field mapping (external-secrets-operator style)
- Atomic file writes with random temp files (TOCTOU prevention)
- Health checks and readiness probes (`isready` subcommand)
- Prometheus metrics (localhost-only by default)
- Optional OpenTelemetry tracing
- Graceful shutdown and reload
- Minimal container image (FROM scratch, <20MB)

**Security Hardening**: ✅ COMPLETE
- Random temp file names (TOCTOU attack prevention)
- Symlink and device file rejection
- Path validation (length limits, traversal prevention, Windows paths)
- File type validation (only regular files)
- Resource limits (1MB secrets, 10MB responses, 100 secrets max)
- Orphaned temp file cleanup
- Fuzzing tests for all input validation
- OS-aware path length limits

**Systemd Support**: ✅ COMPLETE
- Systemd unit file with extensive security hardening
- Installation and uninstallation scripts
- Man page (secrets-sync.1)
- SIGHUP reload support
- Environment file template
- Comprehensive deployment documentation

**Code Quality**: ✅ COMPLETE
- Pre-commit hooks (auto-fix whitespace, run tests)
- GitHub Actions CI (enforce quality, auto-fix issues)
- 60+ unit tests + 5 fuzz tests
- Duplicate path detection

**Documentation**: ✅ COMPLETE
- Complete README with systemd instructions
- Man page with full reference
- Systemd deployment guide
- Fuzzing guide
- Configuration documentation
- Troubleshooting guide
- Example deployments


**Target Audience**: Production use

**Status**: ✅ READY FOR RELEASE

---

## Long-term Roadmap (Post v1.0.0)

### v1.1.0 - Enhanced Secret Support ✅ COMPLETED (2026-02-04)
**Status**: Released

**Features**: ✅ ALL IMPLEMENTED
- ✅ Vault KV v1 engine support (per-secret kvVersion configuration)
- ✅ Per-secret mount path configuration
- ✅ Convert command for external-secrets-operator migration
- ✅ Vault query for automatic field detection
- ✅ AppRole authentication in convert command
- ✅ Support for special characters in field names (hyphens, dots)
- [ ] Dynamic secrets support (database credentials) - Deferred to v1.2.0
- [ ] Secret rotation hooks (execute command on change) - Deferred to v1.2.0
- [ ] Binary secret support (non-text files) - Deferred to v1.2.0

**Configuration Changes**:
- `kvVersion` and `mountPath` moved from `secretStore` to per-secret configuration
- Field `path` renamed to `key` in secret configuration

---

### v1.2.0 - Multi-Backend and Dynamic Secrets
**Estimated**: Q2 2026

**Features**:
- Dynamic secrets support (database credentials)
- Secret rotation hooks (execute command on change)
- Binary secret support (non-text files)
- AWS Secrets Manager backend
- Azure Key Vault backend
- Google Secret Manager backend
- Multiple secret stores per sidecar

---

### v1.3.0 - Advanced Features
**Estimated**: Q3 2026

**Features**:
- Secret caching with encryption at rest
- Webhook notifications on sync failures
- Secret versioning and rollback
- Audit logging for secret access

---

## Success Metrics

### Adoption Metrics
- **Target**: 100+ GitHub stars in first 3 months
- **Target**: 10+ production deployments in first 6 months
- **Target**: 5+ community contributions in first year

### Quality Metrics
- **Target**: 80%+ code coverage
- **Target**: < 5 open bugs at any time
- **Target**: < 48 hour response time on issues

### Performance Metrics
- **Target**: < 20MB container image size
- **Target**: < 50MB memory usage
- **Target**: < 5% CPU usage when idle
- **Target**: < 100ms secret sync latency

---

## Dependencies

### External Dependencies
- HashiCorp Vault or OpenBao instance
- Docker or Podman runtime
- (Optional) Kubernetes cluster for K8s examples
- (Optional) Prometheus for metrics collection
- (Optional) Jaeger/OTLP collector for tracing

### Go Dependencies
- `github.com/hashicorp/vault/api` - Vault client
- `github.com/sony/gobreaker` - Circuit breaker
- `github.com/prometheus/client_golang` - Metrics
- `github.com/fsnotify/fsnotify` - File watching
- `go.uber.org/zap` - Structured logging
- `go.opentelemetry.io/otel` - Tracing
- `gopkg.in/yaml.v3` - YAML parsing

---

## Communication Plan

### Development Updates
- Weekly progress updates in GitHub Discussions
- Milestone completion announcements
- Breaking changes documented in CHANGELOG.md

### Community Engagement
- GitHub Issues for bug reports and feature requests
- GitHub Discussions for questions and ideas
- Pull request reviews within 48 hours
- Monthly community calls (post v1.0.0)

### Documentation
- README.md - Quick start and overview
- docs/ - Detailed documentation
- examples/ - Working examples
- CHANGELOG.md - Version history
- CONTRIBUTING.md - Contribution guidelines

---

## Risk Mitigation

### Technical Risks
| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Circuit breaker tuning complexity | Medium | Medium | Extensive testing with various failure scenarios |
| Template engine edge cases | Medium | Low | Comprehensive test suite with real-world examples |
| File permission issues | Low | Medium | Test on multiple platforms (Linux, macOS) |
| Hot reload race conditions | High | Low | Careful synchronization, integration tests |

### Project Risks
| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Scope creep | High | Medium | Strict adherence to roadmap, defer features to future versions |
| Timeline delays | Medium | Medium | Buffer time in estimates, prioritize core features |
| Limited adoption | Medium | Low | Focus on documentation and examples, community engagement |
| Security vulnerabilities | High | Low | Security-first design, regular audits, responsible disclosure |

---

## Decision Log

### 2026-02-01: Core Design Decisions
- **Decision**: Use single YAML file with multiple secrets
- **Rationale**: Simpler deployment, easier to manage as ConfigMap
- **Alternatives considered**: One YAML per secret (rejected: too complex)

---

- **Decision**: Use FROM scratch for container image
- **Rationale**: Minimal attack surface, smallest image size
- **Alternatives considered**: Alpine (rejected: unnecessary overhead), distroless (rejected: still larger than scratch)

---

- **Decision**: Never exit on errors
- **Rationale**: Sidecar should be resilient, keep trying to sync secrets
- **Alternatives considered**: Exit and restart (rejected: causes pod restarts in K8s)

---

- **Decision**: File-based healthcheck with `isready` subcommand
- **Rationale**: Works in scratch container without external dependencies
- **Alternatives considered**: HTTP-only (rejected: doesn't work in docker-compose), file existence check (rejected: requires external tools)

---

- **Decision**: Support both Vault and OpenBao
- **Rationale**: API compatible, same client library works for both
- **Alternatives considered**: Vault only (rejected: OpenBao is growing in popularity)

---

## Review Schedule

- **Weekly**: Progress review against current milestone
- **Bi-weekly**: Roadmap review and adjustment
- **Monthly**: Community feedback review
- **Quarterly**: Long-term roadmap planning

---

## Approval

This roadmap is a living document and will be updated as the project evolves.

**Last Updated**: 2026-02-04
**Next Review**: 2026-02-11
