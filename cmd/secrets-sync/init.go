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

  # KV engine version (v2 only supported)
  kvVersion: "v2"

  # KV mount path
  mountPath: "secret"

  # TLS Configuration (optional)
  # tlsCACert: "/certs/ca-bundle.pem"      # Custom CA certificate
  # tlsCAPath: "/etc/ssl/certs"            # CA certificate directory
  # tlsSkipVerify: false                   # Skip TLS verification (insecure)
  # tlsClientCert: "/certs/client.pem"     # Client certificate (mTLS)
  # tlsClientKey: "/certs/client-key.pem"  # Client key (mTLS)

# Secret Configuration
# Each secret maps template.data keys to files by position:
#   - First key in template.data -> First file in files list
#   - Second key in template.data -> Second file in files list
# The key names are just labels; actual file paths come from the files list.

secrets:
  # Example: TLS certificate
  - name: "tls-cert"
    path: "common/tls/example-cert"
    refreshInterval: "30m"
    template:
      data:
        tls.crt: '{{ .tlsCrt }}'   # -> /secrets/tls.crt
        tls.key: '{{ .tlsKey }}'   # -> /secrets/tls.key
    files:
      - path: "/secrets/tls.crt"
        mode: "0644"
      - path: "/secrets/tls.key"
        mode: "0600"

  # Example: Database credentials
  - name: "database-creds"
    path: "database/prod/credentials"
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

  # Example: API keys
  - name: "api-keys"
    path: "app/config"
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
