package config

import (
	"fmt"
	"os"
	"strings"
)

// Validate checks if the configuration is valid
func Validate(cfg *Config) error {
	if err := validateSecretStore(&cfg.SecretStore); err != nil {
		return fmt.Errorf("secretStore: %w", err)
	}

	if len(cfg.Secrets) == 0 {
		return fmt.Errorf("at least one secret must be defined")
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

	if len(secret.Template.Data) == 0 {
		return fmt.Errorf("template.data must have at least one entry")
	}

	if len(secret.Files) == 0 {
		return fmt.Errorf("files must have at least one entry")
	}

	if len(secret.Template.Data) != len(secret.Files) {
		return fmt.Errorf("template.data and files must have the same number of entries")
	}

	for i, file := range secret.Files {
		if err := validateFile(&file); err != nil {
			return fmt.Errorf("files[%d]: %w", i, err)
		}
	}

	return nil
}

func validateFile(file *File) error {
	if file.Path == "" {
		return fmt.Errorf("path is required")
	}

	if file.Mode == "" {
		file.Mode = "0600"
	}

	return nil
}

// ExpandEnvVars expands environment variables in configuration
func ExpandEnvVars(cfg *Config) {
	cfg.SecretStore.Address = expandEnv(cfg.SecretStore.Address)
	cfg.SecretStore.Token = expandEnv(cfg.SecretStore.Token)
	cfg.SecretStore.RoleID = expandEnv(cfg.SecretStore.RoleID)
	cfg.SecretStore.SecretID = expandEnv(cfg.SecretStore.SecretID)
	cfg.SecretStore.TLSCACert = expandEnv(cfg.SecretStore.TLSCACert)
	cfg.SecretStore.TLSCAPath = expandEnv(cfg.SecretStore.TLSCAPath)
	cfg.SecretStore.TLSClientCert = expandEnv(cfg.SecretStore.TLSClientCert)
	cfg.SecretStore.TLSClientKey = expandEnv(cfg.SecretStore.TLSClientKey)
}

func expandEnv(s string) string {
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		envVar := s[2 : len(s)-1]
		return os.Getenv(envVar)
	}
	return s
}
