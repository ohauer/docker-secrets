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

	// Named credential sets for different namespaces/teams
	Credentials map[string]CredentialSet `yaml:"credentials,omitempty"`

	// TLS Configuration
	TLSSkipVerify bool   `yaml:"tlsSkipVerify,omitempty"` // Skip TLS verification (insecure)
	TLSCACert     string `yaml:"tlsCACert,omitempty"`     // Path to CA certificate file
	TLSCAPath     string `yaml:"tlsCAPath,omitempty"`     // Path to CA certificate directory
	TLSClientCert string `yaml:"tlsClientCert,omitempty"` // Path to client certificate
	TLSClientKey  string `yaml:"tlsClientKey,omitempty"`  // Path to client key
}

// CredentialSet defines authentication credentials
type CredentialSet struct {
	AuthMethod string `yaml:"authMethod"`
	Token      string `yaml:"token,omitempty"`
	RoleID     string `yaml:"roleId,omitempty"`
	SecretID   string `yaml:"secretId,omitempty"`
}

// Secret defines a single secret to sync
type Secret struct {
	Name            string        `yaml:"name"`
	Key             string        `yaml:"key"`
	MountPath       string        `yaml:"mountPath"`
	Namespace       string        `yaml:"namespace,omitempty"`   // OpenBao namespace override (optional)
	Credentials     string        `yaml:"credentials,omitempty"` // Named credential set (optional)
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

// ResolveCredentials returns the effective credentials for a secret
// Returns credential set name (empty string means use default credentials)
func (s *Secret) ResolveCredentials() string {
	return s.Credentials
}

// GetDefaultCredentials returns default credentials from SecretStore
func (ss *SecretStore) GetDefaultCredentials() CredentialSet {
	return CredentialSet{
		AuthMethod: ss.AuthMethod,
		Token:      ss.Token,
		RoleID:     ss.RoleID,
		SecretID:   ss.SecretID,
	}
}

// GetCredentials returns credentials by name, or default if name is empty
func (ss *SecretStore) GetCredentials(name string) (CredentialSet, bool) {
	if name == "" {
		return ss.GetDefaultCredentials(), true
	}
	creds, ok := ss.Credentials[name]
	return creds, ok
}
