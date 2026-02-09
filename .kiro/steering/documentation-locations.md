# Documentation Location Guide

This file helps AI assistants (Kiro, planners) locate project documentation after reorganization.

## Documentation Structure

### Root Directory
- `README.md` - Main project documentation (user-facing)
- `LICENSE` - Project license
- `CHANGELOG.md` - Version history (if exists)

### User Documentation (`docs/`)
- `docs/README.md` - Documentation index
- `docs/configuration.md` - Configuration reference
- `docs/environment-variables.md` - Environment variables
- `docs/systemd-deployment.md` - Systemd deployment guide
- `docs/troubleshooting.md` - Troubleshooting guide
- `docs/fuzzing.md` - Security fuzzing guide
- `docs/secrets-sync.1` - Man page

### Development Documentation (`docs/development/`)
- `docs/development/PLAN.md` - Implementation plan
- `docs/development/ROADMAP.md` - Feature roadmap and releases
- `docs/development/COMPLETION.md` - Completion summary
- `docs/development/FINAL_REPORT.md` - Final project report
- `docs/development/SECURITY_AUDIT.md` - Security audit

### Examples (`examples/`)
- `examples/systemd/` - Systemd service files
- `examples/*.yaml` - Configuration examples

### Steering Guidelines (`.kiro/steering/`)
- `.kiro/steering/behavior/go_project_practices.md` - Go best practices
- `.kiro/steering/behavior/code_standards.md` - Code standards
- `.kiro/steering/behavior/task_rules.md` - Task management
- `.kiro/steering/behavior/question_rules.md` - Question handling

## Quick Reference for AI Assistants

**When user asks about:**
- Configuration → `docs/configuration.md`
- Systemd setup → `docs/systemd-deployment.md`
- Troubleshooting → `docs/troubleshooting.md`
- Project plan → `docs/development/PLAN.md`
- Roadmap/features → `docs/development/ROADMAP.md`
- Security → `docs/development/SECURITY_AUDIT.md` or `docs/fuzzing.md`
- Examples → `examples/` directory

## Rationale

Documentation was reorganized to:
1. Keep root directory clean (only essential files)
2. Separate user docs from development docs
3. Make navigation easier for new users
4. Follow standard open-source project layout
5. Scale better as documentation grows

## Migration Date

Reorganized: 2026-02-09
