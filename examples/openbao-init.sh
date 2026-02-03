#!/bin/sh
# Initialize OpenBao with test secrets
# Run from examples directory: ./openbao-init.sh

set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
WORK_DIR="${SCRIPT_DIR}/.work"

log_message() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') openbao-init - $1"
}

# Create working directory
mkdir -p "${WORK_DIR}"

log_message "Starting OpenBao container..."
cd "${PROJECT_ROOT}"
docker compose -f docker-compose.openbao.yml up -d openbao 2>/dev/null || true

log_message "Waiting for OpenBao to be ready..."
sleep 5

BAO_ADDR="http://localhost:8300"
BAO_TOKEN="dev-root-token"

log_message "Enabling KV v2 secrets engine..."
docker exec -e BAO_ADDR=http://127.0.0.1:8200 -e BAO_TOKEN="${BAO_TOKEN}" openbao-dev \
    bao secrets enable -version=2 -path=secret kv 2>/dev/null || log_message "KV engine already enabled"

log_message "Creating test secrets..."

docker exec -e BAO_ADDR=http://127.0.0.1:8200 -e BAO_TOKEN="${BAO_TOKEN}" openbao-dev \
    bao kv put secret/common/tls/example-cert \
    tlsCrt="-----BEGIN CERTIFICATE-----
MIICxjCCAa4CCQD1234567890" \
    tlsKey="-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0B"

docker exec -e BAO_ADDR=http://127.0.0.1:8200 -e BAO_TOKEN="${BAO_TOKEN}" openbao-dev \
    bao kv put secret/database/prod/credentials \
    username="dbuser" \
    password="dbpass123"

docker exec -e BAO_ADDR=http://127.0.0.1:8200 -e BAO_TOKEN="${BAO_TOKEN}" openbao-dev \
    bao kv put secret/app/config \
    apiKey="api-key-12345" \
    apiSecret="api-secret-67890"

log_message "Creating AppRole auth..."
docker exec -e BAO_ADDR=http://127.0.0.1:8200 -e BAO_TOKEN="${BAO_TOKEN}" openbao-dev \
    bao auth enable approle 2>/dev/null || log_message "AppRole already enabled"

docker exec -e BAO_ADDR=http://127.0.0.1:8200 -e BAO_TOKEN="${BAO_TOKEN}" openbao-dev \
    bao write auth/approle/role/secrets-sync \
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

docker cp "${WORK_DIR}/policy.hcl" openbao-dev:/tmp/policy.hcl
docker exec -e BAO_ADDR=http://127.0.0.1:8200 -e BAO_TOKEN="${BAO_TOKEN}" openbao-dev \
    bao policy write secrets-sync-policy /tmp/policy.hcl

ROLE_ID=$(docker exec -e BAO_ADDR=http://127.0.0.1:8200 -e BAO_TOKEN="${BAO_TOKEN}" openbao-dev \
    bao read -field=role_id auth/approle/role/secrets-sync/role-id)
SECRET_ID=$(docker exec -e BAO_ADDR=http://127.0.0.1:8200 -e BAO_TOKEN="${BAO_TOKEN}" openbao-dev \
    bao write -field=secret_id -f auth/approle/role/secrets-sync/secret-id)

# Save credentials to working directory
cat > "${WORK_DIR}/openbao-credentials.txt" <<EOF
OpenBao Address: ${BAO_ADDR}
Root Token: ${BAO_TOKEN}

AppRole Credentials:
  Role ID:   ${ROLE_ID}
  Secret ID: ${SECRET_ID}

Test secrets created:
  - secret/common/tls/example-cert (tlsCrt, tlsKey)
  - secret/database/prod/credentials (username, password)
  - secret/app/config (apiKey, apiSecret)
EOF

log_message "=========================================="
log_message "OpenBao initialized successfully!"
log_message "=========================================="
log_message "OpenBao Address: ${BAO_ADDR}"
log_message "Root Token: ${BAO_TOKEN}"
log_message ""
log_message "AppRole Credentials:"
log_message "  Role ID:   ${ROLE_ID}"
log_message "  Secret ID: ${SECRET_ID}"
log_message ""
log_message "Credentials saved to: ${WORK_DIR}/openbao-credentials.txt"
log_message ""
log_message "To stop OpenBao:"
log_message "  docker compose -f ../docker-compose.openbao.yml down"
