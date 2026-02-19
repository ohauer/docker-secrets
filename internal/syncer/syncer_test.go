package syncer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ohauer/docker-secrets/internal/config"
	"github.com/ohauer/docker-secrets/internal/vault"
)

// createTestFactory creates a client factory for testing
func createTestFactory(client *vault.Client) ClientFactory {
	return func(creds config.CredentialSet) (*vault.Client, error) {
		return client, nil
	}
}

// createTestConfig creates a test config with default credentials
func createTestConfig() *config.Config {
	return &config.Config{
		SecretStore: config.SecretStore{
			AuthMethod: "token",
			Token:      "test-token",
		},
	}
}

func TestSyncSecret_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
            "data": {
                "data": {
                    "username": "testuser",
                    "password": "testpass"
                }
            }
        }`))
	}))
	defer server.Close()

	client, err := vault.NewClient(server.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	retryConfig := vault.RetryConfig{
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
		MaxRetries:     3,
	}

	syncer := NewSecretSyncer(createTestFactory(client), retryConfig)

	tmpDir := t.TempDir()
	cfg := createTestConfig()

	secret := config.Secret{
		Name:      "test-secret",
		Key:       "test/path",
		MountPath: "secret",
		KVVersion: "v2",
		Template: config.Template{
			Data: map[string]string{
				"username": "{{ .username }}",
				"password": "{{ .password }}",
			},
		},
		Files: []config.File{
			{Path: filepath.Join(tmpDir, "password"), Mode: "0600"}, // password comes first alphabetically
			{Path: filepath.Join(tmpDir, "username"), Mode: "0644"},
		},
	}

	ctx := context.Background()
	if err := syncer.SyncSecret(ctx, cfg, secret); err != nil {
		t.Fatalf("failed to sync secret: %v", err)
	}

	username, err := os.ReadFile(filepath.Join(tmpDir, "username"))
	if err != nil {
		t.Fatalf("failed to read username file: %v", err)
	}
	if string(username) != "testuser" {
		t.Errorf("expected 'testuser', got '%s'", string(username))
	}

	password, err := os.ReadFile(filepath.Join(tmpDir, "password"))
	if err != nil {
		t.Fatalf("failed to read password file: %v", err)
	}
	if string(password) != "testpass" {
		t.Errorf("expected 'testpass', got '%s'", string(password))
	}
}

func TestScheduler_AddSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
            "data": {
                "data": {
                    "key": "value"
                }
            }
        }`))
	}))
	defer server.Close()

	client, err := vault.NewClient(server.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	retryConfig := vault.RetryConfig{
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
		MaxRetries:     3,
	}

	syncer := NewSecretSyncer(createTestFactory(client), retryConfig)
	scheduler := NewScheduler(syncer)
	defer scheduler.Stop()

	tmpDir := t.TempDir()
	cfg := &config.Config{
		SecretStore: config.SecretStore{AuthMethod: "token", Token: "test-token"},
	}

	secret := config.Secret{
		Name:            "test-secret",
		Key:             "test/path",
		MountPath:       "secret",
		KVVersion:       "v2",
		RefreshInterval: 100 * time.Millisecond,
		Template: config.Template{
			Data: map[string]string{
				"key": "{{ .key }}",
			},
		},
		Files: []config.File{
			{Path: filepath.Join(tmpDir, "key"), Mode: "0600"},
		},
	}

	scheduler.AddSecret(cfg, secret)

	select {
	case result := <-scheduler.Results():
		if !result.Success {
			t.Errorf("expected success, got error: %v", result.Error)
		}
		if result.SecretName != "test-secret" {
			t.Errorf("expected secret name 'test-secret', got '%s'", result.SecretName)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for sync result")
	}
}

func TestScheduler_PeriodicSync(t *testing.T) {
	var syncCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&syncCount, 1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
            "data": {
                "data": {
                    "key": "value"
                }
            }
        }`))
	}))
	defer server.Close()

	client, err := vault.NewClient(server.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	retryConfig := vault.RetryConfig{
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
		MaxRetries:     3,
	}

	syncer := NewSecretSyncer(createTestFactory(client), retryConfig)
	scheduler := NewScheduler(syncer)
	defer scheduler.Stop()

	tmpDir := t.TempDir()
	cfg := &config.Config{
		SecretStore: config.SecretStore{AuthMethod: "token", Token: "test-token"},
	}

	secret := config.Secret{
		Name:            "test-secret",
		Key:             "test/path",
		MountPath:       "secret",
		KVVersion:       "v2",
		RefreshInterval: 100 * time.Millisecond,
		Template: config.Template{
			Data: map[string]string{
				"key": "{{ .key }}",
			},
		},
		Files: []config.File{
			{Path: filepath.Join(tmpDir, "key"), Mode: "0600"},
		},
	}

	scheduler.AddSecret(cfg, secret)

	time.Sleep(350 * time.Millisecond)

	count := atomic.LoadInt32(&syncCount)
	if count < 3 {
		t.Errorf("expected at least 3 syncs, got %d", count)
	}
}

func TestScheduler_RemoveSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
            "data": {
                "data": {
                    "key": "value"
                }
            }
        }`))
	}))
	defer server.Close()

	client, err := vault.NewClient(server.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	retryConfig := vault.RetryConfig{
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
		MaxRetries:     3,
	}

	syncer := NewSecretSyncer(createTestFactory(client), retryConfig)
	scheduler := NewScheduler(syncer)
	defer scheduler.Stop()

	tmpDir := t.TempDir()
	cfg := &config.Config{
		SecretStore: config.SecretStore{AuthMethod: "token", Token: "test-token"},
	}

	secret := config.Secret{
		Name:            "test-secret",
		Key:             "test/path",
		MountPath:       "secret",
		KVVersion:       "v2",
		RefreshInterval: 50 * time.Millisecond,
		Template: config.Template{
			Data: map[string]string{
				"key": "{{ .key }}",
			},
		},
		Files: []config.File{
			{Path: filepath.Join(tmpDir, "key"), Mode: "0600"},
		},
	}

	scheduler.AddSecret(cfg, secret)
	time.Sleep(100 * time.Millisecond)

	scheduler.RemoveSecret("test-secret")

	_, ok := scheduler.GetLastSyncTime("test-secret")
	if ok {
		t.Error("expected secret to be removed")
	}
}
