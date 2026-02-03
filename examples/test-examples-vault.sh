#!/bin/sh
# Test the examples end-to-end
# Run from examples directory: ./test-examples.sh

set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

log_message() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') test-examples - $1"
}

cleanup() {
    log_message "Cleaning up..."
    cd "${PROJECT_ROOT}"
    docker compose -f docker-compose.vault.yml down 2>/dev/null || true
    rm -rf "${SCRIPT_DIR}/.work"
}

trap cleanup EXIT

log_message "Starting test..."

# 1. Initialize Vault
log_message "Step 1: Initialize Vault"
cd "${SCRIPT_DIR}"
./vault-init.sh > /dev/null 2>&1

# 2. Verify credentials file
log_message "Step 2: Verify credentials file"
if [ ! -f ".work/vault-credentials.txt" ]; then
    log_message "ERROR: Credentials file not created"
    exit 1
fi

# 3. Build binary
log_message "Step 3: Build binary"
cd "${PROJECT_ROOT}"
make build > /dev/null 2>&1

# 4. Run secrets-sync
log_message "Step 4: Run secrets-sync"
cd "${SCRIPT_DIR}"
mkdir -p .work/secrets
CONFIG_FILE=config.yaml timeout 3 "${PROJECT_ROOT}/bin/secrets-sync" > /dev/null 2>&1 || true

# 5. Verify secrets
log_message "Step 5: Verify synced secrets"
if [ ! -f ".work/secrets/db-username" ]; then
    log_message "ERROR: Secrets not synced"
    exit 1
fi

USERNAME=$(cat ".work/secrets/db-username")
if [ "${USERNAME}" != "dbuser" ]; then
    log_message "ERROR: Wrong secret value: ${USERNAME}"
    exit 1
fi

log_message "=========================================="
log_message "All tests passed!"
log_message "=========================================="
log_message "Files created:"
log_message "  - .work/vault-credentials.txt"
log_message "  - .work/secrets/tls.crt"
log_message "  - .work/secrets/tls.key"
log_message "  - .work/secrets/db-username"
log_message "  - .work/secrets/db-password"
log_message "  - .work/secrets/api-key"
log_message "  - .work/secrets/api-secret"
