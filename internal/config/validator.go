package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ohauer/docker-secrets/internal/filewriter"
)

// Validate checks if the configuration is valid
func Validate(cfg *Config) error {
	if err := validateSecretStore(&cfg.SecretStore); err != nil {
		return fmt.Errorf("secretStore: %w", err)
	}

	if len(cfg.Secrets) == 0 {
		return fmt.Errorf("at least one secret must be defined")
	}

	// Limit maximum number of secrets to prevent resource exhaustion
	if len(cfg.Secrets) > 100 {
		return fmt.Errorf("too many secrets defined (%d), maximum is 100", len(cfg.Secrets))
	}

	// Check for duplicate file paths
	if err := validateNoDuplicatePaths(cfg.Secrets); err != nil {
		return err
	}

	for i, secret := range cfg.Secrets {
		if err := validateSecret(&secret); err != nil {
			return fmt.Errorf("secrets[%d]: %w", i, err)
		}
	}

	return nil
}

func validateSecretStore(store *SecretStore) error {
	if store.Address == "" {
		return fmt.Errorf("address is required")
	}

	// Validate Vault address is a valid URL
	if err := validateVaultAddress(store.Address); err != nil {
		return err
	}

	if store.AuthMethod == "" {
		return fmt.Errorf("authMethod is required")
	}

	switch store.AuthMethod {
	case "token":
		if store.Token == "" {
			return fmt.Errorf("token is required for token auth")
		}
	case "approle":
		if store.RoleID == "" {
			return fmt.Errorf("roleId is required for approle auth")
		}
		if store.SecretID == "" {
			return fmt.Errorf("secretId is required for approle auth")
		}
	default:
		return fmt.Errorf("unsupported authMethod: %s (supported: token, approle)", store.AuthMethod)
	}

	// Validate TLS configuration
	if store.TLSCACert != "" {
		if _, err := os.Stat(store.TLSCACert); os.IsNotExist(err) {
			return fmt.Errorf("tlsCACert file does not exist: %s", store.TLSCACert)
		}
	}

	if store.TLSCAPath != "" {
		if info, err := os.Stat(store.TLSCAPath); os.IsNotExist(err) {
			return fmt.Errorf("tlsCAPath directory does not exist: %s", store.TLSCAPath)
		} else if !info.IsDir() {
			return fmt.Errorf("tlsCAPath is not a directory: %s", store.TLSCAPath)
		}
	}

	if store.TLSClientCert != "" && store.TLSClientKey == "" {
		return fmt.Errorf("tlsClientKey is required when tlsClientCert is set")
	}

	if store.TLSClientKey != "" && store.TLSClientCert == "" {
		return fmt.Errorf("tlsClientCert is required when tlsClientKey is set")
	}

	if store.TLSClientCert != "" {
		if _, err := os.Stat(store.TLSClientCert); os.IsNotExist(err) {
			return fmt.Errorf("tlsClientCert file does not exist: %s", store.TLSClientCert)
		}
	}

	if store.TLSClientKey != "" {
		if _, err := os.Stat(store.TLSClientKey); os.IsNotExist(err) {
			return fmt.Errorf("tlsClientKey file does not exist: %s", store.TLSClientKey)
		}
	}

	return nil
}

// validateVaultAddress validates the Vault address is a valid URL
func validateVaultAddress(address string) error {
	u, err := url.Parse(address)
	if err != nil {
		return fmt.Errorf("invalid address URL: %w", err)
	}

	if u.Scheme == "" {
		return fmt.Errorf("address must include scheme (http:// or https://)")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("address scheme must be http or https, got: %s", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("address must include host")
	}

	return nil
}

func validateSecret(secret *Secret) error {
	if secret.Name == "" {
		return fmt.Errorf("name is required")
	}

	if secret.Key == "" {
		return fmt.Errorf("key is required")
	}

	if secret.MountPath == "" {
		return fmt.Errorf("mountPath is required")
	}

	if secret.KVVersion == "" {
		return fmt.Errorf("kvVersion is required")
	}

	if secret.KVVersion != "v1" && secret.KVVersion != "v2" {
		return fmt.Errorf("kvVersion must be v1 or v2, got: %s", secret.KVVersion)
	}

	if secret.RefreshInterval <= 0 {
		return fmt.Errorf("refreshInterval must be positive")
	}

	// Enforce minimum refresh interval to prevent hammering Vault
	if secret.RefreshInterval < 30*time.Second {
		return fmt.Errorf("refreshInterval must be at least 30s, got: %s", secret.RefreshInterval)
	}

	if len(secret.Template.Data) == 0 {
		return fmt.Errorf("template.data must have at least one entry")
	}

	if len(secret.Files) == 0 {
		return fmt.Errorf("files must have at least one entry")
	}

	if len(secret.Template.Data) != len(secret.Files) {
		return fmt.Errorf("template.data and files must have the same number of entries")
	}

	for i := range secret.Files {
		if err := validateFile(&secret.Files[i]); err != nil {
			return fmt.Errorf("files[%d]: %w", i, err)
		}
	}

	return nil
}

func validateFile(file *File) error {
	if file.Path == "" {
		return fmt.Errorf("path is required")
	}

	// Resolve relative paths to absolute paths
	absPath, err := filepath.Abs(file.Path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}
	file.Path = filepath.Clean(absPath)

	// Validate path for security (same validation as at write time)
	if err := validateFilePath(file.Path); err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Set default mode if empty
	if file.Mode == "" {
		file.Mode = "0600"
	}

	// Validate mode is valid and secure
	if _, err := filewriter.ParseMode(file.Mode); err != nil {
		return fmt.Errorf("invalid mode '%s': %w", file.Mode, err)
	}

	// Validate owner if specified
	if file.Owner != "" {
		if _, err := filewriter.ParseOwner(file.Owner); err != nil {
			return fmt.Errorf("invalid owner '%s': %w", file.Owner, err)
		}
	}

	// Validate group if specified
	if file.Group != "" {
		if _, err := filewriter.ParseOwner(file.Group); err != nil {
			return fmt.Errorf("invalid group '%s': %w", file.Group, err)
		}
	}

	return nil
}

// validateFilePath validates file path for security
func validateFilePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Resolve to absolute path (handles relative paths securely)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Clean the path to remove any .. or . components
	cleanPath := filepath.Clean(absPath)

	// Ensure the cleaned path is still absolute
	if !filepath.IsAbs(cleanPath) {
		return fmt.Errorf("path must be absolute")
	}

	return nil
}

// ExpandEnvVars expands environment variables in configuration
func ExpandEnvVars(cfg *Config) {
	cfg.SecretStore.Address = expandEnv(cfg.SecretStore.Address)
	cfg.SecretStore.Namespace = expandEnv(cfg.SecretStore.Namespace)
	cfg.SecretStore.Token = expandEnv(cfg.SecretStore.Token)
	cfg.SecretStore.RoleID = expandEnv(cfg.SecretStore.RoleID)
	cfg.SecretStore.SecretID = expandEnv(cfg.SecretStore.SecretID)
	cfg.SecretStore.TLSCACert = expandEnv(cfg.SecretStore.TLSCACert)
	cfg.SecretStore.TLSCAPath = expandEnv(cfg.SecretStore.TLSCAPath)
	cfg.SecretStore.TLSClientCert = expandEnv(cfg.SecretStore.TLSClientCert)
	cfg.SecretStore.TLSClientKey = expandEnv(cfg.SecretStore.TLSClientKey)

	for i := range cfg.Secrets {
		cfg.Secrets[i].Namespace = expandEnv(cfg.Secrets[i].Namespace)
	}
}

func expandEnv(s string) string {
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		envVar := s[2 : len(s)-1]
		return os.Getenv(envVar)
	}
	return s
}

// validateNoDuplicatePaths checks that no two different secrets write to the same file path
func validateNoDuplicatePaths(secrets []Secret) error {
	pathToSecret := make(map[string]string) // path -> secret name

	for _, secret := range secrets {
		for _, file := range secret.Files {
			if existingSecret, found := pathToSecret[file.Path]; found {
				if existingSecret != secret.Name {
					// Different secrets writing to same path - race condition
					return fmt.Errorf("duplicate file path %q: used by both secret %q and secret %q (race condition)",
						file.Path, existingSecret, secret.Name)
				} else {
					// Same secret writing to same path multiple times - configuration error
					return fmt.Errorf("duplicate file path %q in secret %q (same path listed multiple times)",
						file.Path, secret.Name)
				}
			}
			pathToSecret[file.Path] = secret.Name
		}
	}

	return nil
}
