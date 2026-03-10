#!/bin/sh
# Test the examples with OpenBao end-to-end
# Run from examples directory: ./test-examples-openbao.sh

set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

log_message() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') test-examples-openbao - $1"
}

cleanup() {
    log_message "Cleaning up..."
    cd "${PROJECT_ROOT}"
    docker compose -f docker-compose.openbao.yml down 2>/dev/null || true
    rm -rf "${SCRIPT_DIR}/.work"
}

trap cleanup EXIT

log_message "Starting test..."

# 1. Initialize OpenBao
log_message "Step 1: Initialize OpenBao"
cd "${SCRIPT_DIR}"
./openbao-init.sh > /dev/null 2>&1

# 2. Verify credentials file
log_message "Step 2: Verify credentials file"
if [ ! -f ".work/openbao-credentials.txt" ]; then
    log_message "ERROR: Credentials file not created"
    exit 1
fi

# 3. Build binary
log_message "Step 3: Build binary"
cd "${PROJECT_ROOT}"
make build > /dev/null 2>&1

# 4. Create OpenBao config
log_message "Step 4: Create OpenBao config"
cd "${SCRIPT_DIR}"
cat > .work/config-openbao.yaml <<EOF
secretStore:
  address: "http://localhost:8300"
  authMethod: "token"
  token: "dev-root-token"
  kvVersion: "v2"
  mountPath: "secret"

secrets:
  - name: "tls-cert"
    path: "common/tls/example-cert"
    refreshInterval: "30m"
    template:
      data:
        tls.crt: '{{ .tlsCrt }}'
        tls.key: '{{ .tlsKey }}'
    files:
      - path: ".work/secrets/tls.crt"
        mode: "0644"
      - path: ".work/secrets/tls.key"
        mode: "0600"

  - name: "database-creds"
    path: "database/prod/credentials"
    refreshInterval: "5m"
    template:
      data:
        username: '{{ .username }}'
        password: '{{ .password }}'
    files:
      - path: ".work/secrets/db-username"
        mode: "0600"
      - path: ".work/secrets/db-password"
        mode: "0600"

  - name: "api-keys"
    path: "app/config"
    refreshInterval: "1h"
    template:
      data:
        api_key: '{{ .apiKey }}'
        api_secret: '{{ .apiSecret }}'
    files:
      - path: ".work/secrets/api-key"
        mode: "0600"
      - path: ".work/secrets/api-secret"
        mode: "0600"
EOF

# 5. Run secrets-sync
log_message "Step 5: Run secrets-sync"
mkdir -p .work/secrets
CONFIG_FILE=.work/config-openbao.yaml timeout 3 "${PROJECT_ROOT}/bin/secrets-sync" > /dev/null 2>&1 || true

# 6. Verify secrets
log_message "Step 6: Verify synced secrets"
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
log_message "  - .work/openbao-credentials.txt"
log_message "  - .work/secrets/tls.crt"
log_message "  - .work/secrets/tls.key"
log_message "  - .work/secrets/db-username"
log_message "  - .work/secrets/db-password"
log_message "  - .work/secrets/api-key"
log_message "  - .work/secrets/api-secret"
