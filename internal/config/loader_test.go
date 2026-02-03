package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_ValidConfig(t *testing.T) {
	_ = os.Setenv("VAULT_TOKEN", "test-token")
	defer func() { _ = os.Unsetenv("VAULT_TOKEN") }()

	cfg, err := Load("../../testdata/valid-config.yaml")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.SecretStore.Address != "https://vault.example.com" {
		t.Errorf("expected address 'https://vault.example.com', got: %s", cfg.SecretStore.Address)
	}

	if cfg.SecretStore.Token != "test-token" {
		t.Errorf("expected token 'test-token', got: %s", cfg.SecretStore.Token)
	}

	if len(cfg.Secrets) != 2 {
		t.Errorf("expected 2 secrets, got: %d", len(cfg.Secrets))
	}

	if cfg.Secrets[0].Name != "tls-cert" {
		t.Errorf("expected secret name 'tls-cert', got: %s", cfg.Secrets[0].Name)
	}

	if cfg.Secrets[0].RefreshInterval != 30*time.Minute {
		t.Errorf("expected refresh interval 30m, got: %v", cfg.Secrets[0].RefreshInterval)
	}
}

func TestLoad_InvalidConfig(t *testing.T) {
	_, err := Load("../../testdata/invalid-config.yaml")
	if err == nil {
		t.Fatal("expected error for invalid config, got nil")
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	_, err := Load("nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestValidate_MissingAddress(t *testing.T) {
	cfg := &Config{
		SecretStore: SecretStore{
			Address:    "",
			AuthMethod: "token",
			Token:      "test",
		},
		Secrets: []Secret{
			{
				Name:            "test",
				Key:             "test/path",
				MountPath:       "secret",
				KVVersion:       "v2",
				RefreshInterval: 5 * time.Minute,
				Template:        Template{Data: map[string]string{"key": "value"}},
				Files:           []File{{Path: "/test"}},
			},
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error for missing address, got nil")
	}
}

func TestValidate_InvalidAuthMethod(t *testing.T) {
	cfg := &Config{
		SecretStore: SecretStore{
			Address:    "https://vault.example.com",
			AuthMethod: "invalid",
		},
		Secrets: []Secret{
			{
				Name:            "test",
				Key:             "test/path",
				MountPath:       "secret",
				KVVersion:       "v2",
				RefreshInterval: 5 * time.Minute,
				Template:        Template{Data: map[string]string{"key": "value"}},
				Files:           []File{{Path: "/test"}},
			},
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error for invalid auth method, got nil")
	}
}

func TestValidate_NoSecrets(t *testing.T) {
	cfg := &Config{
		SecretStore: SecretStore{
			Address:    "https://vault.example.com",
			AuthMethod: "token",
			Token:      "test",
		},
		Secrets: []Secret{},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error for no secrets, got nil")
	}
}

func TestExpandEnvVars(t *testing.T) {
	_ = os.Setenv("TEST_TOKEN", "my-token")
	defer func() { _ = os.Unsetenv("TEST_TOKEN") }()

	cfg := &Config{
		SecretStore: SecretStore{
			Token: "${TEST_TOKEN}",
		},
	}

	ExpandEnvVars(cfg)

	if cfg.SecretStore.Token != "my-token" {
		t.Errorf("expected token 'my-token', got: %s", cfg.SecretStore.Token)
	}
}

func TestValidate_KVVersion(t *testing.T) {
	tests := []struct {
		name      string
		kvVersion string
		wantErr   bool
	}{
		{
			name:      "valid v1",
			kvVersion: "v1",
			wantErr:   false,
		},
		{
			name:      "valid v2",
			kvVersion: "v2",
			wantErr:   false,
		},
		{
			name:      "invalid version",
			kvVersion: "v3",
			wantErr:   true,
		},
		{
			name:      "empty version",
			kvVersion: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				SecretStore: SecretStore{
					Address:    "https://vault.example.com",
					AuthMethod: "token",
					Token:      "test",
				},
				Secrets: []Secret{
					{
						Name:            "test",
						Key:             "test/path",
						MountPath:       "secret",
						KVVersion:       tt.kvVersion,
						RefreshInterval: 5 * time.Minute,
						Template:        Template{Data: map[string]string{"key": "value"}},
						Files:           []File{{Path: "/test"}},
					},
				},
			}

			err := Validate(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidate_MissingMountPath(t *testing.T) {
	cfg := &Config{
		SecretStore: SecretStore{
			Address:    "https://vault.example.com",
			AuthMethod: "token",
			Token:      "test",
		},
		Secrets: []Secret{
			{
				Name:            "test",
				Key:             "test/path",
				MountPath:       "",
				KVVersion:       "v2",
				RefreshInterval: 5 * time.Minute,
				Template:        Template{Data: map[string]string{"key": "value"}},
				Files:           []File{{Path: "/test"}},
			},
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error for missing mountPath, got nil")
	}
}

func TestValidate_MissingKey(t *testing.T) {
	cfg := &Config{
		SecretStore: SecretStore{
			Address:    "https://vault.example.com",
			AuthMethod: "token",
			Token:      "test",
		},
		Secrets: []Secret{
			{
				Name:            "test",
				Key:             "",
				MountPath:       "secret",
				KVVersion:       "v2",
				RefreshInterval: 5 * time.Minute,
				Template:        Template{Data: map[string]string{"key": "value"}},
				Files:           []File{{Path: "/test"}},
			},
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
}
