package main

import (
	"fmt"
)

func printInitConfig() {
	fmt.Print(`# Docker Secrets Sync Configuration
# See https://github.com/ohauer/docker-secrets for full documentation

secretStore:
  # Vault/OpenBao server address
  address: "https://vault.example.com"

  # Authentication method: token or approle
  authMethod: "token"

  # Token authentication (use environment variable: VAULT_TOKEN)
  token: "${VAULT_TOKEN}"

  # AppRole authentication (uncomment if using approle)
  # roleId: "${VAULT_ROLE_ID}"
  # secretId: "${VAULT_SECRET_ID}"

  # OpenBao namespace (optional, global default for all secrets)
  # namespace: "team-a"

  # TLS Configuration (optional)
  # tlsCACert: "/certs/ca-bundle.pem"      # Custom CA certificate
  # tlsCAPath: "/etc/ssl/certs"            # CA certificate directory
  # tlsSkipVerify: false                   # Skip TLS verification (insecure)
  # tlsClientCert: "/certs/client.pem"     # Client certificate (mTLS)
  # tlsClientKey: "/certs/client-key.pem"  # Client key (mTLS)

# Secret Configuration
# Each secret must specify:
#   - key: Path to the secret in Vault (e.g., "app/database/credentials")
#   - mountPath: KV secrets engine mount path (e.g., "secret")
#   - kvVersion: KV engine version - "v1" or "v2"
#   - namespace: (optional) OpenBao namespace override
#
# KV v1 vs v2:
#   - v1: Simple key-value store, no versioning, direct path access
#   - v2: Versioned secrets with metadata, path includes /data/ prefix (handled automatically)
#
# Namespace precedence:
#   - Per-secret namespace overrides global namespace
#   - Empty namespace ("") means root namespace
#
# Template mapping:
#   - First key in template.data -> First file in files list
#   - Second key in template.data -> Second file in files list
#   - The key names are just labels; actual file paths come from the files list

secrets:
  # Example: TLS certificate from KV v2
  - name: "tls-cert"
    key: "common/tls/example-cert"
    mountPath: "secret"
    kvVersion: "v2"
    refreshInterval: "30m"
    # namespace: ""  # Optional: override global namespace
    template:
      data:
        tls.crt: '{{ .tlsCrt }}'   # -> /secrets/tls.crt
        tls.key: '{{ .tlsKey }}'   # -> /secrets/tls.key
    files:
      - path: "/secrets/tls.crt"
        mode: "0644"
      - path: "/secrets/tls.key"
        mode: "0600"

  # Example: Database credentials from KV v2
  - name: "database-creds"
    key: "database/prod/credentials"
    mountPath: "secret"
    kvVersion: "v2"
    refreshInterval: "1h"
    template:
      data:
        username: '{{ .username }}'  # -> /secrets/db-username
        password: '{{ .password }}'  # -> /secrets/db-password
    files:
      - path: "/secrets/db-username"
        mode: "0600"
      - path: "/secrets/db-password"
        mode: "0600"

  # Example: API keys from KV v1 (legacy)
  - name: "api-keys"
    key: "app/config"
    mountPath: "kv"
    kvVersion: "v1"
    refreshInterval: "2h"
    template:
      data:
        apiKey: '{{ .apiKey }}'       # -> /secrets/api-key
        apiSecret: '{{ .apiSecret }}' # -> /secrets/api-secret
    files:
      - path: "/secrets/api-key"
        mode: "0600"
      - path: "/secrets/api-secret"
        mode: "0600"
`)
}
