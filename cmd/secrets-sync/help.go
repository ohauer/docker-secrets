package main

import (
	"fmt"
	"os"
)

func printHelp() {
	fmt.Print(`secrets-sync - Sync secrets from Vault/OpenBao to filesystem

USAGE:
    secrets-sync [command] [flags]

COMMANDS:
    (none)      Run the secrets sync service (default)
    init        Generate example configuration file
    validate    Validate configuration file
    convert     Convert external-secrets YAML to docker-secrets format
    version     Show version information
    isready     Check if service is ready (for healthchecks)
    help        Show this help message

FLAGS:
    -c, --config <path>  Path to configuration file
    -h, --help           Show this help message

CONFIGURATION:
    Config file precedence (highest to lowest):
    1. --config/-c flag
    2. CONFIG_FILE environment variable
    3. ./config.yaml (current directory)
    4. /etc/secrets-sync/config.yaml (system-wide)

ENVIRONMENT VARIABLES:
    CONFIG_FILE              Path to configuration file
    VAULT_ADDR              Vault/OpenBao server address
    VAULT_TOKEN             Vault token for authentication
    VAULT_ROLE_ID           AppRole role ID
    VAULT_SECRET_ID         AppRole secret ID
    VAULT_CACERT            Path to CA certificate file
    VAULT_CAPATH            Path to CA certificate directory
    VAULT_SKIP_VERIFY       Skip TLS verification (insecure)
    VAULT_CLIENT_CERT       Path to client certificate (mTLS)
    VAULT_CLIENT_KEY        Path to client key (mTLS)
    LOG_LEVEL               Log level (debug, info, warn, error)
    WATCH_CONFIG            Enable config hot reload (default: false)

METRICS:
    METRICS_ADDR            Metrics server listen address (default: 127.0.0.1)
    METRICS_PORT            Metrics server port (default: 8080, range: 1025-65535)
    ENABLE_METRICS          Enable metrics/health endpoints (default: true)

EXAMPLES:
    # Run with config file (flag)
    secrets-sync --config config.yaml
    secrets-sync -c /etc/secrets-sync/config.yaml

    # Run with config file (env var)
    CONFIG_FILE=config.yaml secrets-sync

    # Run with default config location
    secrets-sync

    # Generate example config
    secrets-sync init > config.yaml

    # Validate config
    secrets-sync validate
    secrets-sync --config custom.yaml validate

    # Check version
    secrets-sync version

    # Healthcheck
    secrets-sync isready

    # Convert external-secrets to docker-secrets format
    secrets-sync convert external-secret.yaml --mount-path devops

    # Convert with vault query (auto-detect field names)
    secrets-sync convert external-secret.yaml --query-vault

For more information, see: https://github.com/ohauer/docker-secrets
`)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: secrets-sync [--config <path>] [command]\n")
	fmt.Fprintf(os.Stderr, "Run 'secrets-sync help' for more information.\n")
}
