# Environment Variables

## Configuration File

### CONFIG_FILE
- **Description**: Path to configuration file
- **Required**: No
- **Default**: Tries `./config.yaml`, then `/etc/secrets-sync/config.yaml`
- **Example**: `/etc/secrets-sync/config.yaml`
- **Note**: Can be overridden by `--config` / `-c` command-line flag

**Precedence** (highest to lowest):
1. `--config` / `-c` flag
2. `CONFIG_FILE` environment variable
3. `./config.yaml` (current directory)
4. `/etc/secrets-sync/config.yaml` (system-wide)

## Vault Connection

### VAULT_ADDR
- **Description**: Vault/OpenBao server address
- **Required**: Yes (if not in config file)
- **Example**: `https://vault.example.com`

### VAULT_TOKEN
- **Description**: Vault token for authentication
- **Required**: Yes (for token auth)
- **Example**: `s.xxxxxxxxxxxxx`

### VAULT_ROLE_ID
- **Description**: AppRole role ID
- **Required**: Yes (for approle auth)
- **Example**: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`

### VAULT_SECRET_ID
- **Description**: AppRole secret ID
- **Required**: Yes (for approle auth)
- **Example**: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`

## TLS Configuration

### VAULT_CACERT
- **Description**: Path to CA certificate file for custom/self-signed CAs
- **Required**: No
- **Example**: `/certs/ca-bundle.pem`

### VAULT_CAPATH
- **Description**: Path to directory containing CA certificates
- **Required**: No
- **Example**: `/etc/ssl/certs`

### VAULT_SKIP_VERIFY
- **Description**: Skip TLS certificate verification (insecure, dev only)
- **Default**: `false`
- **Example**: `true`

### VAULT_CLIENT_CERT
- **Description**: Path to client certificate for mTLS
- **Required**: No (required if VAULT_CLIENT_KEY is set)
- **Example**: `/certs/client-cert.pem`

### VAULT_CLIENT_KEY
- **Description**: Path to client key for mTLS
- **Required**: No (required if VAULT_CLIENT_CERT is set)
- **Example**: `/certs/client-key.pem`

**Note:** TLS environment variables override config file values.

## Configuration

### CONFIG_FILE
- **Description**: Path to configuration file
- **Default**: `/config.yaml`
- **Example**: `/etc/secrets-sync/config.yaml`

### WATCH_CONFIG
- **Description**: Enable configuration hot reload
- **Default**: `false`
- **Example**: `true`

## Circuit Breaker

### CIRCUIT_BREAKER_MAX_REQUESTS
- **Description**: Maximum requests allowed in half-open state
- **Default**: `3`
- **Example**: `5`

### CIRCUIT_BREAKER_INTERVAL
- **Description**: Interval for resetting failure counts
- **Default**: `60s`
- **Example**: `2m`

### CIRCUIT_BREAKER_TIMEOUT
- **Description**: Timeout before transitioning to half-open
- **Default**: `30s`
- **Example**: `1m`

## Retry Behavior

### INITIAL_BACKOFF
- **Description**: Initial backoff duration for retries
- **Default**: `1s`
- **Example**: `2s`

### MAX_BACKOFF
- **Description**: Maximum backoff duration
- **Default**: `5m`
- **Example**: `10m`

### BACKOFF_MULTIPLIER
- **Description**: Backoff multiplier for exponential backoff
- **Default**: `2.0`
- **Example**: `1.5`

## Observability

### LOG_LEVEL
- **Description**: Logging level
- **Default**: `info`
- **Options**: `debug`, `info`, `warn`, `error`
- **Example**: `debug`

### HTTP_ADDR
- **Description**: HTTP server listen address for health/metrics
- **Default**: `127.0.0.1` (localhost only)
- **Example**: `0.0.0.0` (all interfaces), `192.168.1.10`
- **Note**: For security, default binds to localhost only. Use `0.0.0.0` only if needed.

### HTTP_PORT
- **Description**: HTTP server port for health/metrics
- **Default**: `8080`
- **Example**: `8081`

### ENABLE_HTTP_SERVER
- **Description**: Enable HTTP server for health checks and metrics
- **Default**: `true`
- **Example**: `false`
- **Note**: When disabled, health checks via HTTP are not available. Use `isready` command instead.

### STATUS_FILE
- **Description**: Path to readiness status file
- **Default**: `/tmp/.ready-state`
- **Example**: `/var/run/secrets-sync/.ready`

### ENABLE_TRACING
- **Description**: Enable OpenTelemetry tracing
- **Default**: `false`
- **Example**: `true`

### OTEL_EXPORTER_ENDPOINT
- **Description**: OpenTelemetry exporter endpoint
- **Required**: Yes (if tracing enabled)
- **Example**: `http://jaeger:4318`

## Example Configuration

### Docker Compose

```yaml
environment:
  # Vault connection
  VAULT_ADDR: http://vault:8200
  VAULT_TOKEN: ${VAULT_TOKEN}

  # Configuration
  CONFIG_FILE: /config.yaml
  WATCH_CONFIG: "true"

  # Circuit breaker
  CIRCUIT_BREAKER_MAX_REQUESTS: "3"
  CIRCUIT_BREAKER_INTERVAL: "60s"
  CIRCUIT_BREAKER_TIMEOUT: "30s"

  # Retry
  INITIAL_BACKOFF: "1s"
  MAX_BACKOFF: "5m"
  BACKOFF_MULTIPLIER: "2.0"

  # Observability
  LOG_LEVEL: info
  HTTP_PORT: "8080"
  STATUS_FILE: /tmp/.ready-state
  ENABLE_TRACING: "false"
```

### Kubernetes

```yaml
env:
  - name: VAULT_ADDR
    value: "https://vault.example.com"
  - name: VAULT_TOKEN
    valueFrom:
      secretKeyRef:
        name: vault-token
        key: token
  - name: CONFIG_FILE
    value: "/config/config.yaml"
  - name: LOG_LEVEL
    value: "info"
  - name: HTTP_PORT
    value: "8080"
```

### Command Line

```bash
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=dev-root-token
export CONFIG_FILE=config.yaml
export LOG_LEVEL=debug
export WATCH_CONFIG=true

./secrets-sync
```
