#!/bin/sh
# Setup Vault with test secrets

set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

log_message() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') setup-vault - $1"
}

log_message "Starting Vault container..."
cd "${PROJECT_ROOT}"
docker compose -f docker-compose.vault.yml up -d vault 2>/dev/null || true

log_message "Waiting for Vault to be ready..."
sleep 5

VAULT_ADDR="http://localhost:8200"
VAULT_TOKEN="dev-root-token"

log_message "Creating test secrets..."

docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN="${VAULT_TOKEN}" vault-dev \
    vault kv put secret/common/tls/example-cert \
    tlsCrt="-----BEGIN CERTIFICATE-----
MIICxjCCAa4CCQD1234567890" \
    tlsKey="-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0B"

docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN="${VAULT_TOKEN}" vault-dev \
    vault kv put secret/database/prod/credentials \
    username="dbuser" \
    password="dbpass123"

docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN="${VAULT_TOKEN}" vault-dev \
    vault kv put secret/app/config \
    apiKey="api-key-12345" \
    apiSecret="api-secret-67890"

log_message "Test secrets created successfully"
log_message "Vault Address: ${VAULT_ADDR}"
log_message "Vault Token: ${VAULT_TOKEN}"
log_message ""
log_message "To stop Vault: docker compose -f docker-compose.vault.yml down"
