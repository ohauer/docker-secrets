# Manual Testing Environment

This directory contains scripts for manual testing and experimentation with secrets-sync.

## Purpose

Use this environment for:
- Manual testing during development
- Debugging issues
- Experimenting with configurations
- Learning how the tool works

**Note:** This is NOT for automated testing. See `../integration/` for CI-friendly tests.

## Quick Start

```bash
# 1. Start Vault
docker compose -f ../../docker-compose.vault.yml up -d

# 2. Initialize with test secrets
./setup-vault.sh

# 3. Run the tool
./run-tool.sh

# 4. Verify secrets were synced
./verify-secrets.sh

# 5. Cleanup
docker compose -f ../../docker-compose.vault.yml down
rm -rf secrets/
```

## Scripts

### setup-vault.sh
Starts Vault and creates test secrets:
- `secret/common/tls/example-cert` - TLS certificate
- `secret/database/prod/credentials` - Database credentials
- `secret/app/config` - API keys

**Vault Details:**
- Address: http://localhost:8200
- Token: dev-root-token

### run-tool.sh
Runs secrets-sync with the test configuration. Synced secrets are written to `./secrets/`

### verify-secrets.sh
Displays all synced secret files and their contents for verification.

### generate-certs.sh
Generates test CA and certificates for TLS testing.

## Configuration

Edit `config.yaml` to customize which secrets to sync and where to write them.

## Cleanup

```bash
# Stop Vault
docker compose -f ../../docker-compose.vault.yml down

# Remove synced secrets
rm -rf secrets/ .work/
```

## OpenBao Alternative

To test with OpenBao instead of Vault:

```bash
# Start OpenBao (port 8300)
docker compose -f ../../docker-compose.openbao.yml up -d

# Initialize
./openbao-init.sh

# Edit config.yaml to use port 8300
# Then run normally
./run-tool.sh
```

## Notes

⚠️ **WARNING**: This is a development environment only!
- Never use dev mode in production
- Root tokens have unlimited access
- Secrets are stored in plaintext files
- Data is lost when containers restart
