#!/bin/sh
# Initialize Vault with test secrets
# Run from examples directory: ./vault-init.sh

set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
WORK_DIR="${SCRIPT_DIR}/.work"

log_message() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') vault-init - $1"
}

# Create working directory
mkdir -p "${WORK_DIR}"

log_message "Starting Vault container..."
cd "${PROJECT_ROOT}"
docker compose -f docker-compose.vault.yml up -d vault 2>&1 | grep -v "level=warning" || true

log_message "Waiting for Vault to be ready..."
sleep 5

VAULT_ADDR="http://localhost:8200"
VAULT_TOKEN="dev-root-token"

log_message "Enabling KV v2 secrets engine..."
docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN="${VAULT_TOKEN}" vault-dev \
    vault secrets enable -version=2 -path=secret kv 2>/dev/null || log_message "KV engine already enabled"

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

log_message "Creating AppRole auth..."
docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN="${VAULT_TOKEN}" vault-dev \
    vault auth enable approle 2>/dev/null || log_message "AppRole already enabled"

docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN="${VAULT_TOKEN}" vault-dev \
    vault write auth/approle/role/secrets-sync \
    token_ttl=1h \
    token_max_ttl=4h \
    policies=secrets-sync-policy

log_message "Creating policy..."
cat > "${WORK_DIR}/policy.hcl" <<EOF
path "secret/data/*" {
  capabilities = ["read", "list"]
}
path "secret/metadata/*" {
  capabilities = ["read", "list"]
}
EOF

docker cp "${WORK_DIR}/policy.hcl" vault-dev:/tmp/policy.hcl
docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN="${VAULT_TOKEN}" vault-dev \
    vault policy write secrets-sync-policy /tmp/policy.hcl

ROLE_ID=$(docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN="${VAULT_TOKEN}" vault-dev \
    vault read -field=role_id auth/approle/role/secrets-sync/role-id)
SECRET_ID=$(docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN="${VAULT_TOKEN}" vault-dev \
    vault write -field=secret_id -f auth/approle/role/secrets-sync/secret-id)

# Save credentials to working directory
cat > "${WORK_DIR}/vault-credentials.txt" <<EOF
Vault Address: ${VAULT_ADDR}
Root Token: ${VAULT_TOKEN}

AppRole Credentials:
  Role ID:   ${ROLE_ID}
  Secret ID: ${SECRET_ID}

Test secrets created:
  - secret/common/tls/example-cert (tlsCrt, tlsKey)
  - secret/database/prod/credentials (username, password)
  - secret/app/config (apiKey, apiSecret)
EOF

log_message "=========================================="
log_message "Vault initialized successfully!"
log_message "=========================================="
log_message "Vault Address: ${VAULT_ADDR}"
log_message "Root Token: ${VAULT_TOKEN}"
log_message ""
log_message "AppRole Credentials:"
log_message "  Role ID:   ${ROLE_ID}"
log_message "  Secret ID: ${SECRET_ID}"
log_message ""
log_message "Credentials saved to: ${WORK_DIR}/vault-credentials.txt"
log_message ""
log_message "To stop Vault:"
log_message "  docker compose -f ../docker-compose.vault.yml down"
