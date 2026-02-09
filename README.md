# Docker Secrets Sidecar

A lightweight sidecar container for managing secrets from HashiCorp Vault or OpenBao in Docker/Podman environments. Continuously syncs secrets to the filesystem with configurable refresh intervals.

## Features

- 🔄 **Continuous Sync** - Automatically refreshes secrets at configurable intervals
- 🔐 **Multiple Auth Methods** - Supports Token and AppRole authentication
- 🔒 **TLS Support** - Custom CA certificates, mTLS, self-signed certificates
- 📝 **Template Engine** - Map secret fields to multiple files (external-secrets-operator style)
- 🛡️ **Circuit Breaker** - Prevents cascading failures with exponential backoff
- 📊 **Observability** - JSON logging, Prometheus metrics, optional OpenTelemetry tracing
- 🔧 **Hot Reload** - Configuration changes without restart
- 🐳 **Minimal Image** - FROM scratch, <20MB, runs as non-root
- ✅ **Health Checks** - Built-in healthcheck for docker-compose and Kubernetes

## Quick Start

### 1. Start Vault (for testing)

```bash
docker-compose -f docker-compose.vault.yml up -d
docker logs vault-init  # Get credentials
```

### 2. Create Configuration

```yaml
# config.yaml
secretStore:
  address: "http://localhost:8200"
  authMethod: "token"
  token: "dev-root-token"

secrets:
  - name: "tls-cert"
    key: "common/tls/example-cert"
    mountPath: "secret"
    kvVersion: "v2"
    refreshInterval: "30m"
    template:
      data:
        tls.crt: '{{ .tlsCrt }}'
        tls.key: '{{ .tlsKey }}'
    files:
      - path: "/secrets/tls.crt"
        mode: "0644"
      - path: "/secrets/tls.key"
        mode: "0600"
```

### 3. Run with Docker Compose

```bash
docker-compose -f examples/docker-compose.sidecar.yml up
```

## Installation

### Build from Source

```bash
make build
```

### Build Docker Image

```bash
make docker-build
# or
./build.sh
```

### Install as Systemd Service

For Linux systems, you can install secrets-sync as a systemd service:

```bash
# Automated installation
sudo make install-systemd

# Manual installation
make build
sudo cp bin/secrets-sync /usr/local/bin/
sudo mkdir -p /etc/secrets-sync
secrets-sync init | sudo tee /etc/secrets-sync/config.yaml
sudo cp examples/systemd/secrets-sync.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now secrets-sync

# Reload configuration without restart
sudo systemctl reload secrets-sync
```

See [Systemd Deployment Guide](docs/systemd-deployment.md) for detailed instructions.

## Usage

### Command Line Interface

The tool provides several commands for different use cases:

#### Run the Service

```bash
# Run with config file (using flag)
./secrets-sync --config config.yaml
./secrets-sync -c config.yaml

# Run with config file (using environment variable)
CONFIG_FILE=config.yaml ./secrets-sync

# Run with default config location (./config.yaml or /etc/secrets-sync/config.yaml)
./secrets-sync

# Run with environment variables
export VAULT_ADDR=https://vault.example.com:8443
export VAULT_ROLE_ID=your-role-id
export VAULT_SECRET_ID=your-secret-id
./secrets-sync --config config.yaml
```

**Config file precedence** (highest to lowest):
1. `--config` / `-c` flag
2. `CONFIG_FILE` environment variable
3. `./config.yaml` (current directory)
4. `/etc/secrets-sync/config.yaml` (system-wide)

#### Generate Sample Configuration

```bash
# Generate a sample config.yaml with all options
./secrets-sync init > config.yaml
```

#### Validate Configuration

```bash
# Validate config file without running
./secrets-sync validate
./secrets-sync --config custom-config.yaml validate
CONFIG_FILE=custom-config.yaml ./secrets-sync validate
```

#### Check Version

```bash
# Show version information
./secrets-sync version
```

#### Health Check

```bash
# Check if service is ready (for scripts/monitoring)
./secrets-sync isready
```

#### Convert from external-secrets-operator

Convert ExternalSecret resources to docker-secrets format (supports both YAML and JSON):

```bash
# Basic conversion (uses fallback for unknown fields)
./secrets-sync convert external-secret.yaml > config.yaml

# Works with JSON too
./secrets-sync convert external-secret.json > config.yaml

# Read from stdin (useful with kubectl)
kubectl get externalsecret my-secret -o json | ./secrets-sync convert - --query-vault > config.yaml
kubectl get externalsecrets -n namespace -o json | ./secrets-sync convert - --query-vault > config.yaml

# Query Vault for actual field names (recommended)
export VAULT_ADDR=https://vault.example.com:8443
export VAULT_TOKEN=your-token
./secrets-sync convert external-secret.yaml --query-vault > config.yaml

# Use AppRole authentication for querying
export VAULT_ADDR=https://vault.example.com:8443
export VAULT_ROLE_ID=your-role-id
export VAULT_SECRET_ID=your-secret-id
./secrets-sync convert external-secret.yaml --query-vault > config.yaml

# Convert multiple files (YAML or JSON)
./secrets-sync convert file1.yaml file2.json --query-vault > config.yaml

# Specify mount path manually
./secrets-sync convert external-secret.yaml --mount-path devops > config.yaml
```

The convert command:
- Supports YAML and JSON formats
- Supports single ExternalSecret, Kubernetes List, and multi-document YAML formats
- Auto-detects mount paths from secret keys
- Queries Vault for actual field names when `--query-vault` is used
- Generates complete config including secretStore section
- Comments out secrets that fail to query (permission denied)
- Handles special characters in field names (hyphens, dots)

#### Show Help

```bash
# Show all available commands
./secrets-sync help
./secrets-sync --help
./secrets-sync -h
```

### Docker Compose Sidecar Pattern

```yaml
services:
  app:
    image: your-app:latest
    volumes:
      - secrets:/secrets:ro
    depends_on:
      secrets-sidecar:
        condition: service_healthy

  secrets-sidecar:
    image: docker-secrets:v1.0.0
    volumes:
      - secrets:/secrets
      - ./config.yaml:/config.yaml:ro
    environment:
      VAULT_ADDR: http://vault:8200
      VAULT_TOKEN: ${VAULT_TOKEN}
      CONFIG_FILE: /config.yaml
    healthcheck:
      test: ["/app/secrets-sync", "isready"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  secrets:
```

## Configuration

See [docs/configuration.md](docs/configuration.md) for detailed configuration options.

### Environment Variables

See [docs/environment-variables.md](docs/environment-variables.md) for all available environment variables.

## Authentication

### Token Authentication

```yaml
secretStore:
  authMethod: "token"
  token: "${VAULT_TOKEN}"
```

### AppRole Authentication

```yaml
secretStore:
  authMethod: "approle"
  roleId: "${VAULT_ROLE_ID}"
  secretId: "${VAULT_SECRET_ID}"
```

### TLS with Custom CA (Self-Signed Certificates)

```yaml
secretStore:
  address: "https://vault.example.com"
  authMethod: "token"
  token: "${VAULT_TOKEN}"
  tlsCACert: "/certs/ca-bundle.pem"
```

Or using environment variables:

```bash
export VAULT_ADDR=https://vault.example.com
export VAULT_TOKEN=your-token
export VAULT_CACERT=/certs/ca-bundle.pem
./secrets-sync
```

See [docs/configuration.md](docs/configuration.md) for more TLS options (mTLS, CA path, skip verify).

## Observability

### Health Endpoints

- `GET /health` - Always returns 200 (liveness)
- `GET /ready` - Returns 200 when secrets synced (readiness)
- `GET /metrics` - Prometheus metrics

### Metrics

- `secret_fetch_total` - Total fetch attempts
- `secret_fetch_errors_total` - Total fetch errors
- `secret_sync_duration_seconds` - Sync duration histogram
- `circuit_breaker_state` - Circuit breaker state (0=closed, 1=half-open, 2=open)
- `secrets_configured` - Number of configured secrets
- `secrets_synced` - Number of successfully synced secrets

### Tracing

Enable OpenTelemetry tracing:

```bash
ENABLE_TRACING=true
OTEL_EXPORTER_ENDPOINT=http://jaeger:4318
```

## Examples

- [Docker Compose Sidecar](examples/docker-compose.sidecar.yml)
- [Vault Development Environment](docker-compose.vault.yml)
- [OpenBao Development Environment](docker-compose.openbao.yml)
- [Example Configurations](examples/)

## Troubleshooting

See [docs/troubleshooting.md](docs/troubleshooting.md) for common issues and solutions.

## Development

```bash
# Run tests
make test

# Run linter
make lint

# Build binary
make build

# Build Docker image
make docker-build
```

## Security

- Runs as non-root user (UID 65534)
- Minimal attack surface (FROM scratch)
- Atomic file writes
- Configurable file permissions
- Secrets never logged
- Circuit breaker prevents overwhelming Vault

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions welcome! Please read the contributing guidelines before submitting PRs.

## Support

- GitHub Issues: Report bugs and request features
- GitLab Issues: For internal deployments

## Acknowledgments

- Inspired by [external-secrets-operator](https://external-secrets.io/)
- Uses [HashiCorp Vault Go SDK](https://github.com/hashicorp/vault/tree/main/api)
- Circuit breaker by [sony/gobreaker](https://github.com/sony/gobreaker)
