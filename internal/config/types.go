package config

import "time"

// Config represents the complete configuration
type Config struct {
	SecretStore SecretStore `yaml:"secretStore"`
	Secrets     []Secret    `yaml:"secrets"`
}

// SecretStore defines Vault/OpenBao connection settings
type SecretStore struct {
	Address    string `yaml:"address"`
	Namespace  string `yaml:"namespace,omitempty"` // OpenBao namespace (optional)
	AuthMethod string `yaml:"authMethod"`
	Token      string `yaml:"token"`
	RoleID     string `yaml:"roleId"`
	SecretID   string `yaml:"secretId"`

	// TLS Configuration
	TLSSkipVerify bool   `yaml:"tlsSkipVerify,omitempty"` // Skip TLS verification (insecure)
	TLSCACert     string `yaml:"tlsCACert,omitempty"`     // Path to CA certificate file
	TLSCAPath     string `yaml:"tlsCAPath,omitempty"`     // Path to CA certificate directory
	TLSClientCert string `yaml:"tlsClientCert,omitempty"` // Path to client certificate
	TLSClientKey  string `yaml:"tlsClientKey,omitempty"`  // Path to client key
}

// Secret defines a single secret to sync
type Secret struct {
	Name            string        `yaml:"name"`
	Key             string        `yaml:"key"`
	MountPath       string        `yaml:"mountPath"`
	Namespace       string        `yaml:"namespace,omitempty"` // OpenBao namespace override (optional)
	KVVersion       string        `yaml:"kvVersion"`
	RefreshInterval time.Duration `yaml:"refreshInterval"`
	Template        Template      `yaml:"template"`
	Files           []File        `yaml:"files"`
}

// Template defines how to map secret fields to file content
type Template struct {
	Data map[string]string `yaml:"data"`
}

// File defines output file configuration
type File struct {
	Path  string `yaml:"path"`
	Mode  string `yaml:"mode"`
	Owner string `yaml:"owner"`
	Group string `yaml:"group"`
}

// ResolveNamespace returns the effective namespace for a secret
// Per-secret namespace takes precedence over global namespace
func (s *Secret) ResolveNamespace(globalNamespace string) string {
	if s.Namespace != "" {
		return s.Namespace
	}
	return globalNamespace
}
