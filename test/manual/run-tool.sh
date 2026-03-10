#!/bin/sh
# Run secrets-sync tool with test configuration

set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

log_message() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') run-tool - $1"
}

if [ ! -f "${PROJECT_ROOT}/bin/secrets-sync" ]; then
    log_message "Binary not found, building..."
    cd "${PROJECT_ROOT}"
    make build
fi

mkdir -p "${SCRIPT_DIR}/secrets"

export VAULT_ADDR="http://localhost:8200"
export VAULT_TOKEN="dev-root-token"
export CONFIG_FILE="${SCRIPT_DIR}/config.yaml"
export LOG_LEVEL="info"

log_message "Starting secrets-sync..."
log_message "Config: ${CONFIG_FILE}"
log_message "Output: ${SCRIPT_DIR}/secrets/"

"${PROJECT_ROOT}/bin/secrets-sync"
