# Integration Tests

Automated integration tests for CI/CD pipelines.

## Purpose

These tests verify end-to-end functionality:
- Vault/OpenBao connectivity
- Secret syncing
- File creation and permissions
- Template rendering
- Error handling

## Running Tests

### All Integration Tests
```bash
make test-integration
```

### Individual Tests
```bash
# Vault integration test
./vault_test.sh

# OpenBao integration test (if available)
./openbao_test.sh
```

## Requirements

- Docker and Docker Compose
- Built binary (`make build`)
- Ports 8200, 8300 available

## Test Structure

Each test script:
1. Starts the secret backend (Vault/OpenBao)
2. Initializes with test data
3. Runs secrets-sync
4. Verifies synced secrets
5. Cleans up automatically (even on failure)

## Exit Codes

- `0` - All tests passed
- `1` - Test failed

## CI/CD Integration

These tests are designed to run in CI/CD pipelines:
- No user interaction required
- Automatic cleanup on exit
- Clear pass/fail status
- Detailed error messages

## Adding New Tests

1. Create `<name>_test.sh`
2. Follow the existing pattern:
   - Set `set -euo pipefail`
   - Use trap for cleanup
   - Verify all assertions
   - Exit with proper code
3. Make executable: `chmod +x <name>_test.sh`
4. Add to `make test-integration` in Makefile

## Debugging Failed Tests

If a test fails:
1. Check the error message
2. Run the test manually: `./<name>_test.sh`
3. Check Docker logs: `docker logs vault-dev`
4. Verify ports are available: `netstat -tuln | grep 8200`

## Manual Testing

For manual testing and experimentation, use `../manual/` instead.
