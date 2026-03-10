# Project Restructuring Proposal

## Current Issues

1. **Test values in unit tests** - `localhost:8200` hardcoded in `*_test.go` files
2. **Overlapping test directories** - Both `test-run/` and `examples/` serve similar purposes
3. **Documentation scattered** - Test instructions in multiple READMEs
4. **No clear separation** - Integration tests mixed with examples

## Proposed Structure

```
docker-secrets/
├── cmd/
│   └── secrets-sync/
├── internal/
│   └── */
│       └── *_test.go          # Unit tests only (mock/fake clients)
├── examples/
│   ├── README.md              # User-facing examples
│   ├── config.yaml            # Basic example
│   ├── config-approle.yaml    # AppRole example
│   ├── config-tls.yaml        # TLS example
│   ├── config-credential-sets.yaml
│   ├── config-openbao-namespaces.yaml
│   ├── docker-compose.sidecar.yml
│   └── systemd/
│       └── secrets-sync.service
├── test/
│   ├── integration/           # NEW: Integration tests
│   │   ├── README.md
│   │   ├── vault_test.sh      # Automated integration test
│   │   ├── openbao_test.sh
│   │   ├── docker_test.sh
│   │   └── fixtures/
│   │       ├── config.yaml
│   │       └── expected/
│   └── manual/                # NEW: Manual testing (was test-run/)
│       ├── README.md
│       ├── setup-vault.sh
│       ├── setup-openbao.sh
│       ├── run-tool.sh
│       ├── verify-secrets.sh
│       ├── generate-certs.sh
│       └── config.yaml
├── docs/
├── scripts/                   # Build/install scripts only
└── Makefile
```

## Changes

### 1. Unit Tests (`internal/*_test.go`)
**Current:** Hardcoded `localhost:8200`
**Proposed:** Use test constants or environment variables

```go
// internal/vault/testing.go (NEW)
package vault

const (
    TestVaultAddr = "http://localhost:8200"  // Only for tests
)

// Or use environment variable
func getTestVaultAddr() string {
    if addr := os.Getenv("TEST_VAULT_ADDR"); addr != "" {
        return addr
    }
    return "http://localhost:8200"
}
```

### 2. Integration Tests (`test/integration/`)
**Purpose:** Automated end-to-end testing
**Run by:** CI/CD, developers

```bash
# test/integration/vault_test.sh
#!/bin/bash
set -euo pipefail

# Start Vault
docker compose -f ../../docker-compose.vault.yml up -d
trap "docker compose -f ../../docker-compose.vault.yml down" EXIT

# Wait for ready
sleep 3

# Initialize with test data
docker exec vault-dev vault kv put secret/test/app password=test123

# Run secrets-sync
../../bin/secrets-sync --config fixtures/config.yaml &
PID=$!
trap "kill $PID 2>/dev/null || true" EXIT

# Wait and verify
sleep 3
test -f /tmp/test-secrets/password.txt
grep -q "test123" /tmp/test-secrets/password.txt

echo "✓ Integration test passed"
```

### 3. Manual Testing (`test/manual/`)
**Purpose:** Developer experimentation and debugging
**Run by:** Developers manually

- Rename `test-run/` → `test/manual/`
- Keep all helper scripts
- Clear README explaining it's for manual testing

### 4. Examples (`examples/`)
**Purpose:** User documentation and copy-paste configs
**Run by:** End users

- Remove test scripts (move to `test/`)
- Keep only config examples
- Keep docker-compose.sidecar.yml (it's a usage example)
- Simplified README focused on "how to use"

### 5. Makefile Targets

```makefile
# Unit tests (fast, no external dependencies)
test:
	go test -v -race ./...

# Integration tests (requires Docker)
test-integration:
	cd test/integration && ./vault_test.sh
	cd test/integration && ./openbao_test.sh
	cd test/integration && ./docker_test.sh

# All tests
test-all: test test-integration

# Manual test environment
test-manual:
	@echo "Starting manual test environment..."
	@cd test/manual && ./setup-vault.sh
	@echo "Run: cd test/manual && ./run-tool.sh"
```

## Migration Steps

1. **Create new structure**
   ```bash
   mkdir -p test/integration test/manual
   ```

2. **Move files**
   ```bash
   # Move test-run to manual
   mv test-run/* test/manual/
   rmdir test-run

   # Move test scripts from examples
   mv examples/test-examples-*.sh test/integration/
   mv examples/vault-init.sh test/manual/
   mv examples/openbao-init.sh test/manual/
   ```

3. **Create integration tests**
   - Extract automated tests from examples
   - Make them CI-friendly (exit codes, no interaction)

4. **Update unit tests**
   - Create `internal/vault/testing.go` with test constants
   - Replace hardcoded values with constants

5. **Update documentation**
   - `examples/README.md` - User-facing only
   - `test/manual/README.md` - Developer testing
   - `test/integration/README.md` - CI/CD integration

6. **Update .gitignore**
   ```
   /test/manual/.work/
   /test/manual/secrets/
   /test/integration/.work/
   ```

## Benefits

1. **Clear separation of concerns**
   - Unit tests: Fast, no dependencies
   - Integration tests: Automated, CI-friendly
   - Manual tests: Developer experimentation
   - Examples: User documentation

2. **No test values in production code**
   - Test constants isolated in `*_test.go` or `testing.go`
   - Environment variables for flexibility

3. **Better CI/CD**
   - `make test` - Fast unit tests
   - `make test-integration` - Full integration
   - `make test-all` - Everything

4. **Cleaner examples/**
   - Only user-facing configs
   - No test scripts
   - Clear documentation

5. **Easier maintenance**
   - Test infrastructure in `test/`
   - Examples don't change often
   - Clear what's for users vs developers

## Alternative: Minimal Change

If full restructure is too much, minimal changes:

1. **Move test-run/ → test/**
   ```bash
   mv test-run test
   ```

2. **Add test constants**
   ```go
   // internal/vault/testing.go
   package vault
   const TestVaultAddr = "http://localhost:8200"
   ```

3. **Update Makefile**
   ```makefile
   test-manual:
       @cd test && ./setup-vault.sh
   ```

4. **Update .gitignore**
   ```
   /test/.work/
   /test/secrets/
   ```

## Recommendation

**Go with full restructure** - It's a one-time effort that will:
- Make the project more professional
- Easier for contributors to understand
- Better CI/CD integration
- Cleaner separation of concerns

The current state is confusing with overlapping `test-run/` and `examples/` directories.
