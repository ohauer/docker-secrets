# Configuration

## Configuration File Location

The configuration file can be specified in multiple ways with the following precedence (highest to lowest):

1. **Command-line flag**: `--config` or `-c`
   ```bash
   ./secrets-sync --config /path/to/config.yaml
   ./secrets-sync -c config.yaml
   ```

2. **Environment variable**: `CONFIG_FILE`
   ```bash
   CONFIG_FILE=/path/to/config.yaml ./secrets-sync
   ```

3. **Default locations** (checked in order):
   - `./config.yaml` (current directory)
   - `/etc/secrets-sync/config.yaml` (system-wide)

## Configuration File Structure

The configuration file is a YAML file with two main sections: `secretStore` and `secrets`.

```yaml
secretStore:
  address: "https://vault.example.com"
  authMethod: "token"  # or "approle"
  token: "${VAULT_TOKEN}"
  kvVersion: "v2"
  mountPath: "secret"

secrets:
  - name: "secret-name"
    path: "path/to/secret"
    refreshInterval: "30m"
    template:
      data:
        key1: '{{ .field1 }}'
        key2: '{{ .field2 }}'
    files:
      - path: "/secrets/file1"
        mode: "0644"
      - path: "/secrets/file2"
        mode: "0600"
```

## Secret Store Configuration

### Required Fields

- `address` - Vault/OpenBao server address (e.g., `https://vault.example.com`)
- `authMethod` - Authentication method: `token` or `approle`

### Optional Fields

- `namespace` - OpenBao namespace (global default for all secrets)
- `credentials` - Named credential sets for different teams/namespaces
- `kvVersion` - KV engine version (default: `v2`)
- `mountPath` - KV mount path (default: `secret`)

### Token Authentication

```yaml
secretStore:
  address: "https://vault.example.com"
  authMethod: "token"
  token: "${VAULT_TOKEN}"
```

### AppRole Authentication

```yaml
secretStore:
  address: "https://vault.example.com"
  authMethod: "approle"
  roleId: "${VAULT_ROLE_ID}"
  secretId: "${VAULT_SECRET_ID}"
```

### Named Credential Sets

Use different credentials for different secrets/namespaces:

```yaml
secretStore:
  address: "https://vault.example.com"

  # Default credentials
  authMethod: "token"
  token: "${VAULT_TOKEN}"

  # Named credential sets
  credentials:
    team-a:
      authMethod: "token"
      token: "${TEAM_A_TOKEN}"
    team-b:
      authMethod: "approle"
      roleId: "${TEAM_B_ROLE_ID}"
      secretId: "${TEAM_B_SECRET_ID}"

secrets:
  - name: "default-secret"
    # Uses default credentials

  - name: "team-a-secret"
    credentials: "team-a"  # Uses team-a credentials

  - name: "team-b-secret"
    credentials: "team-b"  # Uses team-b credentials
```

**How it works:**
- Each credential set creates a separate authenticated Vault client
- Clients are cached and reused for efficiency
- If a secret doesn't specify `credentials`, it uses the default credentials
- Useful for multi-tenant environments or different namespace permissions

### OpenBao Namespace Support

OpenBao namespaces allow logical partitioning of secrets within a single OpenBao instance.

#### Global Namespace

Set a global namespace that applies to all secrets:

```yaml
secretStore:
  address: "https://openbao.example.com"
  authMethod: "token"
  token: "${VAULT_TOKEN}"
  namespace: "team-a"  # All secrets use this namespace by default
```

#### Per-Secret Namespace Override

Override the global namespace for specific secrets:

```yaml
secretStore:
  address: "https://openbao.example.com"
  authMethod: "token"
  token: "${VAULT_TOKEN}"
  namespace: "team-a"  # Global default

secrets:
  - name: "team-a-secret"
    key: "app/config"
    mountPath: "secret"
    kvVersion: "v2"
    # Uses global namespace (team-a)

  - name: "team-b-secret"
    key: "app/config"
    mountPath: "secret"
    namespace: "team-b"  # Overrides global namespace
    kvVersion: "v2"
```

#### Root Namespace

To access the root namespace when a global namespace is set, use an empty string:

```yaml
secretStore:
  namespace: "team-a"  # Global default

secrets:
  - name: "root-secret"
    namespace: ""  # Empty = root namespace
```

### TLS Configuration

#### Custom CA Certificate (Self-Signed)

```yaml
secretStore:
  address: "https://vault.example.com"
  authMethod: "token"
  token: "${VAULT_TOKEN}"
  tlsCACert: "/certs/ca-bundle.pem"
```

#### CA Certificate Directory

```yaml
secretStore:
  address: "https://vault.example.com"
  authMethod: "token"
  token: "${VAULT_TOKEN}"
  tlsCAPath: "/etc/ssl/certs"
```

#### Skip TLS Verification (Insecure - Dev Only)

```yaml
secretStore:
  address: "https://vault.example.com"
  authMethod: "token"
  token: "${VAULT_TOKEN}"
  tlsSkipVerify: true
```

#### Mutual TLS (mTLS)

```yaml
secretStore:
  address: "https://vault.example.com"
  authMethod: "token"
  token: "${VAULT_TOKEN}"
  tlsCACert: "/certs/ca-bundle.pem"
  tlsClientCert: "/certs/client-cert.pem"
  tlsClientKey: "/certs/client-key.pem"
```

### TLS Fields

- `tlsCACert` - Path to CA certificate file (for self-signed or internal CAs)
- `tlsCAPath` - Path to CA certificate directory
- `tlsSkipVerify` - Skip TLS verification (insecure, dev only)
- `tlsClientCert` - Path to client certificate (for mTLS)
- `tlsClientKey` - Path to client key (for mTLS)

**Note:** Environment variables take precedence over config file values.

## Secret Configuration

### Required Fields

- `name` - Unique name for the secret
- `path` - Path to secret in Vault (without mount path prefix)
- `mountPath` - KV secrets engine mount path
- `kvVersion` - KV engine version (`v1` or `v2`)
- `refreshInterval` - How often to refresh (e.g., `30m`, `1h`, `24h`)
- `template.data` - Map of template names to Go templates
- `files` - List of output files

### Optional Fields

- `namespace` - OpenBao namespace (overrides global namespace from secretStore)
- `credentials` - Named credential set to use (overrides default credentials)

### Template Syntax

Templates use Go template syntax with secret fields as variables:

```yaml
template:
  data:
    username: '{{ .username }}'
    password: '{{ .password }}'
    connection: 'postgresql://{{ .username }}:{{ .password }}@localhost/db'
```

**Important:** The keys in `template.data` are mapped to files **by position**:
- First key in `template.data` → First file in `files` list
- Second key in `template.data` → Second file in `files` list
- And so on...

Example:
```yaml
template:
  data:
    username: '{{ .username }}'  # Written to first file
    password: '{{ .password }}'  # Written to second file
files:
  - path: "/secrets/db-username"  # Receives 'username' value
    mode: "0600"
  - path: "/secrets/db-password"  # Receives 'password' value
    mode: "0600"
```

The key names in `template.data` are not used for file naming - they're just labels. The actual file paths come from the `files` list.

### File Configuration

Each file entry supports:

- `path` - Output file path (required, can be relative or absolute)
- `mode` - File permissions in octal (default: `0600`)
- `owner` - File owner UID (optional)
- `group` - File group GID (optional)

**Path Resolution:**
- Relative paths (e.g., `secrets/file.txt`) are resolved to absolute paths based on the current working directory
- Absolute paths (e.g., `/var/secrets/file.txt`) are used as-is
- All paths are validated for security (no path traversal allowed)

Example:

```yaml
files:
  - path: "/secrets/tls.crt"      # Absolute path
    mode: "0644"
  - path: "secrets/tls.key"       # Relative path (resolved to absolute)
    mode: "0600"
    owner: "1000"
    group: "1000"
```

## Environment Variable Expansion

Configuration values can reference environment variables using `${VAR_NAME}` syntax:

```yaml
secretStore:
  address: "${VAULT_ADDR}"
  token: "${VAULT_TOKEN}"
```

## Multiple Secrets

You can configure multiple secrets with different refresh intervals:

```yaml
secrets:
  - name: "tls-cert"
    path: "common/tls/cert"
    refreshInterval: "24h"
    # ...

  - name: "database-creds"
    path: "database/prod/creds"
    refreshInterval: "5m"
    # ...

  - name: "api-keys"
    path: "app/api-keys"
    refreshInterval: "1h"
    # ...
```

## Configuration Hot Reload

Enable configuration hot reload to update secrets without restart:

```bash
WATCH_CONFIG=true
```

When enabled, the tool will:
- Watch the configuration file for changes
- Reload and validate the new configuration
- Stop syncing removed secrets
- Start syncing new secrets
- Update refresh intervals for existing secrets

## Example Configurations

### TLS Certificate

```yaml
secrets:
  - name: "tls-cert"
    path: "common/tls/example-cert"
    refreshInterval: "24h"
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

### Database Credentials

```yaml
secrets:
  - name: "database-creds"
    path: "database/prod/credentials"
    refreshInterval: "5m"
    template:
      data:
        username: '{{ .username }}'
        password: '{{ .password }}'
    files:
      - path: "/secrets/db-username"
        mode: "0600"
      - path: "/secrets/db-password"
        mode: "0600"
```

### API Keys

```yaml
secrets:
  - name: "api-keys"
    path: "app/config"
    refreshInterval: "1h"
    template:
      data:
        api_key: '{{ .apiKey }}'
        api_secret: '{{ .apiSecret }}'
    files:
      - path: "/secrets/api-key"
        mode: "0600"
      - path: "/secrets/api-secret"
        mode: "0600"
```
