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

// SecretSyncer handles secret synchronization
type SecretSyncer struct {
	vaultClient *vault.Client
	writer      *filewriter.Writer
	retryConfig vault.RetryConfig
}

// NewSecretSyncer creates a new secret syncer
func NewSecretSyncer(vaultClient *vault.Client, retryConfig vault.RetryConfig) *SecretSyncer {
	return &SecretSyncer{
		vaultClient: vaultClient,
		writer:      filewriter.NewWriter(),
		retryConfig: retryConfig,
	}
}

// SyncSecret synchronizes a single secret
func (s *SecretSyncer) SyncSecret(ctx context.Context, cfg *config.Config, secret config.Secret) error {
	data, err := s.vaultClient.FetchSecretWithRetry(
		ctx,
		secret.MountPath,
		secret.Key,
		secret.KVVersion,
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
