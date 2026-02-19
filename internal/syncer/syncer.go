package syncer

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/ohauer/docker-secrets/internal/config"
	"github.com/ohauer/docker-secrets/internal/filewriter"
	"github.com/ohauer/docker-secrets/internal/template"
	"github.com/ohauer/docker-secrets/internal/vault"
)

// ClientFactory creates Vault clients with specific credentials
type ClientFactory func(creds config.CredentialSet) (*vault.Client, error)

// SecretSyncer handles secret synchronization
type SecretSyncer struct {
	clientFactory ClientFactory
	clientPool    map[string]*vault.Client // Cache clients by credential set name
	writer        *filewriter.Writer
	retryConfig   vault.RetryConfig
}

// NewSecretSyncer creates a new secret syncer with a client factory
func NewSecretSyncer(factory ClientFactory, retryConfig vault.RetryConfig) *SecretSyncer {
	return &SecretSyncer{
		clientFactory: factory,
		clientPool:    make(map[string]*vault.Client),
		writer:        filewriter.NewWriter(),
		retryConfig:   retryConfig,
	}
}

// getOrCreateClient returns a cached client or creates a new one
func (s *SecretSyncer) getOrCreateClient(credName string, creds config.CredentialSet) (*vault.Client, error) {
	// Check cache
	if client, ok := s.clientPool[credName]; ok {
		return client, nil
	}

	// Create new client
	client, err := s.clientFactory(creds)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for credentials %q: %w", credName, err)
	}

	// Cache it
	s.clientPool[credName] = client
	return client, nil
}

// SyncSecret synchronizes a single secret
func (s *SecretSyncer) SyncSecret(ctx context.Context, cfg *config.Config, secret config.Secret) error {
	// Resolve credentials (per-secret overrides default)
	credName := secret.ResolveCredentials()
	creds, ok := cfg.SecretStore.GetCredentials(credName)
	if !ok {
		return fmt.Errorf("credentials %q not found", credName)
	}

	// Get or create client for these credentials
	client, err := s.getOrCreateClient(credName, creds)
	if err != nil {
		return err
	}

	// Resolve namespace (per-secret overrides global)
	namespace := secret.ResolveNamespace(cfg.SecretStore.Namespace)

	data, err := client.FetchSecretWithRetry(
		ctx,
		secret.MountPath,
		secret.Key,
		secret.KVVersion,
		namespace,
		s.retryConfig,
	)
	if err != nil {
		return fmt.Errorf("failed to fetch secret: %w", err)
	}

	engine := template.NewEngine()
	for name, tmpl := range secret.Template.Data {
		if err := engine.AddTemplate(name, tmpl); err != nil {
			return fmt.Errorf("failed to add template %s: %w", name, err)
		}
	}

	rendered, err := engine.RenderAll(map[string]interface{}(data))
	if err != nil {
		return fmt.Errorf("failed to render templates: %w", err)
	}

	if len(rendered) != len(secret.Files) {
		return fmt.Errorf("template count (%d) does not match file count (%d)", len(rendered), len(secret.Files))
	}

	// Sort template names for deterministic file mapping
	templateNames := make([]string, 0, len(secret.Template.Data))
	for name := range secret.Template.Data {
		templateNames = append(templateNames, name)
	}
	sort.Strings(templateNames)

	for i, file := range secret.Files {
		mode, err := filewriter.ParseMode(file.Mode)
		if err != nil {
			return fmt.Errorf("invalid mode for file %s: %w", file.Path, err)
		}

		owner, err := filewriter.ParseOwner(file.Owner)
		if err != nil {
			return fmt.Errorf("invalid owner for file %s: %w", file.Path, err)
		}

		group, err := filewriter.ParseOwner(file.Group)
		if err != nil {
			return fmt.Errorf("invalid group for file %s: %w", file.Path, err)
		}

		var content string
		if i < len(templateNames) {
			content = rendered[templateNames[i]]
		}

		fileConfig := filewriter.FileConfig{
			Path:  file.Path,
			Mode:  mode,
			Owner: owner,
			Group: group,
		}

		if err := s.writer.WriteFile(fileConfig, content); err != nil {
			return fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}
	}

	return nil
}

// SyncResult holds the result of a sync operation
type SyncResult struct {
	SecretName string
	Success    bool
	Error      error
	Timestamp  time.Time
}
