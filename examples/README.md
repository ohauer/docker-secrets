# Examples

This directory contains example configurations and scripts for testing docker-secrets.

## Quick Start

### Option 1: Standalone Testing

Test the tool directly with Vault:

```bash
# 1. Build the binary
make build -C ..

# 2. Initialize Vault with test secrets
cd examples
./vault-init.sh

# 3. Run secrets-sync
mkdir -p .work/secrets
CONFIG_FILE=config.yaml ../bin/secrets-sync

# 4. Verify secrets (in another terminal)
ls -la .work/secrets/
cat .work/secrets/db-username
```

### Option 2: Docker Sidecar Pattern

Run the complete sidecar example with docker-compose:

```bash
# 1. Build docker image
make docker-build -C ..

# 2. Start the sidecar example
docker compose -f docker-compose.sidecar.yml up

# This starts:
# - Vault (on port 8202)
# - Secrets sidecar (syncs secrets)
# - Nginx app (uses synced secrets)
```

Note: The sidecar example includes its own Vault instance and needs to be initialized separately. See "Sidecar Example Setup" below.

## Working Directory

All generated files are stored in `examples/.work/`:
- `.work/secrets/` - Synced secret files
- `.work/vault-credentials.txt` - Vault credentials
- `.work/openbao-credentials.txt` - OpenBao credentials

This directory is gitignored to keep the repository clean.

## Files

### Scripts

- `vault-init.sh` - Initialize Vault (port 8200) with test secrets
- `openbao-init.sh` - Initialize OpenBao (port 8300) with test secrets
- `test-examples-vault.sh` - End-to-end test with Vault
- `test-examples-openbao.sh` - End-to-end test with OpenBao

### Configuration

- `config.yaml` - Example configuration for secrets-sync
- `docker-compose.sidecar.yml` - Sidecar pattern example

## Sidecar Example Setup

The docker-compose.sidecar.yml includes its own Vault instance on port 8202. To use it:

```bash
# Start services
docker compose -f docker-compose.sidecar.yml up -d

# Initialize Vault in the sidecar
docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN=dev-root-token vault-sidecar \
  vault kv put secret/common/tls/example-cert tlsCrt="cert-data" tlsKey="key-data"

docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN=dev-root-token vault-sidecar \
  vault kv put secret/database/prod/credentials username="dbuser" password="dbpass123"

docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN=dev-root-token vault-sidecar \
  vault kv put secret/app/config apiKey="api-key" apiSecret="api-secret"

# Restart sidecar to sync secrets
docker compose -f docker-compose.sidecar.yml restart secrets-sidecar

# Check logs
docker logs -f secrets-sidecar
```

## OpenBao Alternative

To use OpenBao instead of Vault:

```bash
./openbao-init.sh
```

OpenBao runs on port 8300 (to avoid conflicts with Vault on 8200). You'll need to create a separate config file or use the test script:

```bash
# Run the OpenBao test (creates config automatically)
./test-examples-openbao.sh
```

## Automated Testing

Two test scripts are provided to verify the complete workflow:

```bash
# Test with Vault (port 8200)
./test-examples-vault.sh

# Test with OpenBao (port 8300)
./test-examples-openbao.sh
```

These scripts:
- Start the secret backend
- Initialize with test secrets
- Build the binary
- Run secrets-sync
- Verify synced secrets
- Clean up automatically

## Test Secrets

Both Vault and OpenBao are initialized with:

### TLS Certificate
- **Path**: `secret/common/tls/example-cert`
- **Fields**: `tlsCrt`, `tlsKey`

### Database Credentials
- **Path**: `secret/database/prod/credentials`
- **Fields**: `username`, `password`

### Application Config
- **Path**: `secret/app/config`
- **Fields**: `apiKey`, `apiSecret`

## Authentication

### Token Authentication (Default)
```yaml
secretStore:
  authMethod: "token"
  token: "dev-root-token"
```

### AppRole Authentication
Get credentials from `.work/vault-credentials.txt` or `.work/openbao-credentials.txt`:

```yaml
secretStore:
  authMethod: "approle"
  roleId: "${VAULT_ROLE_ID}"
  secretId: "${VAULT_SECRET_ID}"
```

## Cleanup

```bash
# Stop sidecar example
docker compose -f ../docker-compose.sidecar.yml down

# Stop Vault
docker compose -f ../docker-compose.vault.yml down

# Stop OpenBao
docker compose -f ../docker-compose.openbao.yml down

# Remove working directory
rm -rf ./.work
```

## Notes

⚠️ **WARNING**: These are development examples only. Never use dev mode or these tokens in production!

- Dev mode stores data in memory (lost on restart)
- Root tokens have unlimited access
- TLS is disabled in these examples
- Secrets are stored in plaintext files
