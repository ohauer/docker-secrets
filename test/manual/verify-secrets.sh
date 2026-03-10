#!/bin/sh
# Verify synced secrets

set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "=== Synced Secret Files ==="
ls -lh "${SCRIPT_DIR}/secrets/"

echo ""
echo "=== Secret Contents ==="
echo "TLS Certificate:"
cat "${SCRIPT_DIR}/secrets/tls.crt"

echo ""
echo "Database Username:"
cat "${SCRIPT_DIR}/secrets/db-username"

echo ""
echo "Database Password:"
cat "${SCRIPT_DIR}/secrets/db-password"

echo ""
echo "API Key:"
cat "${SCRIPT_DIR}/secrets/api-key"

echo ""
echo "API Secret:"
cat "${SCRIPT_DIR}/secrets/api-secret"

echo ""
echo "✅ All secrets synced successfully"
