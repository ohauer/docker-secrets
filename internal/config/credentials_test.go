package config

import (
	"testing"
	"time"
)

func TestCredentialResolution(t *testing.T) {
	tests := []struct {
		name           string
		store          SecretStore
		secret         Secret
		expectedCreds  string
		expectedExists bool
	}{
		{
			name: "use default credentials",
			store: SecretStore{
				AuthMethod: "token",
				Token:      "default-token",
			},
			secret:         Secret{},
			expectedCreds:  "",
			expectedExists: true,
		},
		{
			name: "use named credentials",
			store: SecretStore{
				AuthMethod: "token",
				Token:      "default-token",
				Credentials: map[string]CredentialSet{
					"team-a": {
						AuthMethod: "token",
						Token:      "team-a-token",
					},
				},
			},
			secret: Secret{
				Credentials: "team-a",
			},
			expectedCreds:  "team-a",
			expectedExists: true,
		},
		{
			name: "non-existent credentials",
			store: SecretStore{
				AuthMethod: "token",
				Token:      "default-token",
			},
			secret: Secret{
				Credentials: "non-existent",
			},
			expectedCreds:  "non-existent",
			expectedExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			credName := tt.secret.ResolveCredentials()
			if credName != tt.expectedCreds {
				t.Errorf("ResolveCredentials() = %q, want %q", credName, tt.expectedCreds)
			}

			_, exists := tt.store.GetCredentials(credName)
			if exists != tt.expectedExists {
				t.Errorf("GetCredentials(%q) exists = %v, want %v", credName, exists, tt.expectedExists)
			}
		})
	}
}

func TestGetDefaultCredentials(t *testing.T) {
	store := SecretStore{
		AuthMethod: "approle",
		RoleID:     "test-role",
		SecretID:   "test-secret",
	}

	creds := store.GetDefaultCredentials()

	if creds.AuthMethod != "approle" {
		t.Errorf("AuthMethod = %q, want %q", creds.AuthMethod, "approle")
	}
	if creds.RoleID != "test-role" {
		t.Errorf("RoleID = %q, want %q", creds.RoleID, "test-role")
	}
	if creds.SecretID != "test-secret" {
		t.Errorf("SecretID = %q, want %q", creds.SecretID, "test-secret")
	}
}

func TestValidate_CredentialSets(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid credential sets",
			config: Config{
				SecretStore: SecretStore{
					Address:    "http://localhost:8200",
					AuthMethod: "token",
					Token:      "default-token",
					Credentials: map[string]CredentialSet{
						"team-a": {
							AuthMethod: "token",
							Token:      "team-a-token",
						},
						"team-b": {
							AuthMethod: "approle",
							RoleID:     "role-id",
							SecretID:   "secret-id",
						},
					},
				},
				Secrets: []Secret{
					{
						Name:            "test",
						Key:             "test/path",
						MountPath:       "secret",
						KVVersion:       "v2",
						RefreshInterval: 30 * time.Minute,
						Credentials:     "team-a",
						Template: Template{
							Data: map[string]string{"test": "{{ .value }}"},
						},
						Files: []File{
							{Path: "/tmp/test", Mode: "0600"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid credential set - missing token",
			config: Config{
				SecretStore: SecretStore{
					Address:    "http://localhost:8200",
					AuthMethod: "token",
					Token:      "default-token",
					Credentials: map[string]CredentialSet{
						"team-a": {
							AuthMethod: "token",
							// Missing Token
						},
					},
				},
				Secrets: []Secret{},
			},
			wantErr: true,
			errMsg:  "token is required",
		},
		{
			name: "secret references non-existent credentials",
			config: Config{
				SecretStore: SecretStore{
					Address:    "http://localhost:8200",
					AuthMethod: "token",
					Token:      "default-token",
				},
				Secrets: []Secret{
					{
						Name:            "test",
						Key:             "test/path",
						MountPath:       "secret",
						KVVersion:       "v2",
						RefreshInterval: 30 * time.Minute,
						Credentials:     "non-existent",
						Template: Template{
							Data: map[string]string{"test": "{{ .value }}"},
						},
						Files: []File{
							{Path: "/tmp/test", Mode: "0600"},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(&tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil {
					t.Errorf("Validate() error = nil, want error containing %q", tt.errMsg)
				} else {
					errStr := err.Error()
					found := false
					for i := 0; i <= len(errStr)-len(tt.errMsg); i++ {
						if errStr[i:i+len(tt.errMsg)] == tt.errMsg {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
					}
				}
			}
		})
	}
}
