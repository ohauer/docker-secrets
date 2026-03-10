# Examples

User-facing configuration examples for secrets-sync.

## Configuration Examples

### Basic Configuration
- `config.yaml` - Simple token authentication
- `config-credential-sets.yaml` - Multiple credential sets for multi-tenant
- `config-openbao-namespaces.yaml` - OpenBao with namespace support

### Docker Compose
- `docker-compose.sidecar.yml` - Sidecar pattern with Vault and application

### Systemd
- `systemd/secrets-sync.service` - Systemd service file

## Quick Start

### 1. Copy Example Config
```bash
cp examples/config.yaml my-config.yaml
```

### 2. Edit Configuration
```yaml
secretStore:
  address: "https://vault.example.com"
  authMethod: "token"
  token: "${VAULT_TOKEN}"

secrets:
  - name: "my-secret"
    key: "path/to/secret"
    mountPath: "secret"
    kvVersion: "v2"
    files:
      - path: "/secrets/my-secret.txt"
        mode: "0600"
```

### 3. Run
```bash
export VAULT_TOKEN=your-token
./bin/secrets-sync --config my-config.yaml
```

## Docker Sidecar Pattern

See `docker-compose.sidecar.yml` for a complete example of running secrets-sync as a sidecar container.

```bash
docker compose -f examples/docker-compose.sidecar.yml up
```

## Systemd Service

See [Systemd Deployment Guide](../docs/systemd-deployment.md) for installation instructions.

## Testing

For testing and development, see:
- `test/manual/` - Manual testing environment
- `test/integration/` - Automated integration tests

## Documentation

- [Configuration Reference](../docs/configuration.md)
- [Environment Variables](../docs/environment-variables.md)
- [Troubleshooting](../docs/troubleshooting.md)
